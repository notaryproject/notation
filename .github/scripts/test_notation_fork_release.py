# Copyright The Notary Project Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import base64
import copy
import json
import os
import pathlib
import tempfile
import unittest
import zipfile
from unittest.mock import patch

import monthly_release as release
import check_notation_fork_setup as setup
import notation_fork_propagation as propagation
import notation_release_controller as controller
from test_monthly_release import plan as worker_plan, pull
from test_notation_release_controller import ControllerAPI, CORE, GO, CLI, MAIN, PR_HEAD, MONTH

PLAN_ID = "1" * 64
TARGET = "v1.3.4-monthly-test.202610"


def authorization():
    return {"issue": 12, "plan_id": PLAN_ID,
            "snapshot": {"main": "monthly-patch-test-main", "main_head": MAIN, "tag": TARGET},
            "producers": {"notation-core-go": {"repository": CORE, "version": TARGET}}}


def producer_pr():
    candidate = pull(number=99, sha=PR_HEAD, base="monthly-patch-test-main")
    candidate["user"] = {"login": "test-actor", "type": "User"}
    candidate["base"]["sha"] = MAIN
    candidate["head"].update(repo={"full_name": GO}, ref=controller.propagation_branch(MONTH, PLAN_ID))
    receipt = {"schema": 1, "repository": GO, "month": MONTH, "controller_issue": 12, "plan_id": PLAN_ID,
               "producers": authorization()["producers"], "base": MAIN, "head": PR_HEAD}
    candidate["body"] = f"<!-- notation-fork-propagation:{MONTH}:{PLAN_ID} -->\n```json\n{json.dumps(receipt)}\n```\n"
    return candidate


class PropagationAuthorizationTests(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ, {"MONTHLY_PATCH_ACTOR": "test-actor"})
        self.environment.start()
        self.addCleanup(self.environment.stop)
        self.api = ControllerAPI()
        self.commit = patch.object(self.api, "request", return_value={
            "commit": {"verification": {"verified": True}}, "author": {"login": "test-actor"}})
        self.commit.start()
        self.addCleanup(self.commit.stop)
        self.scope = patch.object(controller, "producer_pull", return_value=True)
        self.scope.start()
        self.addCleanup(self.scope.stop)

    def test_only_signed_actor_pr_with_current_exact_plan_is_accepted(self):
        self.assertTrue(controller.fork_propagation_pull(self.api, GO, producer_pr(), authorization()))
        for mutate in (
            lambda pr: pr["user"].update(login="different-actor"),
            lambda pr: pr["head"].update(sha=MAIN),
            lambda pr: pr["head"].update(ref="main"),
            lambda pr: pr["head"].update(repo={"full_name": "other/notation-go"}),
            lambda pr: pr["base"].update(ref="main"),
            lambda pr: pr.update(body="ordinary manually authored dependency update"),
        ):
            candidate = producer_pr()
            mutate(candidate)
            with self.subTest(candidate=candidate):
                self.assertFalse(controller.fork_propagation_pull(self.api, GO, candidate, authorization()))
        self.assertFalse(controller.fork_propagation_pull(self.api, "notaryproject/notation-go", producer_pr(), authorization()))
        self.assertFalse(controller.fork_propagation_pull(self.api, GO, producer_pr(), {**authorization(), "plan_id": "2" * 64}))

    def test_unsigned_or_wrong_signer_commit_blocks_merging(self):
        for verified, actor in ((False, "test-actor"), (True, "different-actor")):
            self.api.request.return_value = {"commit": {"verification": {"verified": verified}}, "author": {"login": actor}}
            with self.subTest(verified=verified, actor=actor), self.assertRaisesRegex(ValueError, "verified commit"):
                controller.fork_propagation_pull(self.api, GO, producer_pr(), authorization())

    def test_signed_pr_cannot_hide_unrelated_dependency_changes(self):
        controller.producer_pull.return_value = False
        self.assertFalse(controller.fork_propagation_pull(self.api, GO, producer_pr(), authorization()))

    def test_closed_pr_uses_immutable_receipt_base_for_scope_validation(self):
        candidate = producer_pr()
        candidate["base"]["sha"] = "e" * 40
        candidate.update(merged_at="2026-10-05T01:00:00Z", merge_commit_sha="f" * 40)
        self.assertTrue(controller.fork_propagation_pull(self.api, GO, candidate, authorization()))
        self.assertEqual(controller.producer_pull.call_args.args[2]["base"]["sha"], MAIN)

    def test_new_authorized_actor_pr_can_enter_worker_inventory(self):
        candidate = producer_pr()
        candidate.update(merged_at="2026-10-05T01:00:00Z", merge_commit_sha="f" * 40)
        self.api.closed_pulls = [candidate]
        self.assertEqual(release.inventory(self.api, GO, "monthly-patch-test-main", authorization())[0]["number"], 99)
        self.assertEqual(release.inventory(self.api, GO, "monthly-patch-test-main"), [])

    def test_plan_without_producers_never_authorizes_actor_updates(self):
        self.assertFalse(controller.fork_propagation_pull(self.api, GO, producer_pr(), {**authorization(), "producers": {}}))


class PropagationPipelineTests(unittest.TestCase):
    def test_real_go_update_creates_one_signed_module_only_pr_and_resumes(self):
        self.check_real_go_update()

    def test_real_go_update_includes_a_producer_required_direct_dependency_bump(self):
        self.check_real_go_update(direct_dependency=True)

    def check_real_go_update(self, direct_dependency=False):
        with tempfile.TemporaryDirectory(prefix="notation-fork-integration-") as temporary:
            root = pathlib.Path(temporary)
            proxy, tree, remote = root / "proxy", root / "tree", root / "remote.git"
            tree.mkdir()
            old = "v1.3.3-monthly-test.202609"
            module = "github.com/yizha1/notation-core-go"
            versions = proxy / module / "@v"
            versions.mkdir(parents=True)
            for version in (old, TARGET):
                mod = "module github.com/notaryproject/notation-core-go\n\ngo 1.26.0\n"
                if direct_dependency:
                    mod += "\nrequire example.com/shared " + ("v1.0.0" if version == old else "v1.2.0") + "\n"
                (versions / (version + ".mod")).write_text(mod)
                (versions / (version + ".info")).write_text(json.dumps({"Version": version, "Time": "2026-10-01T00:00:00Z"}))
                with zipfile.ZipFile(versions / (version + ".zip"), "w") as archive:
                    archive.writestr(module + "@" + version + "/go.mod", mod)
                    body = ('package core\n\nimport "example.com/shared"\n\nfunc Value() int { return shared.Value() }\n'
                            if direct_dependency else "package core\n\nfunc Value() int { return 1 }\n")
                    archive.writestr(module + "@" + version + "/core.go", body)
            if direct_dependency:
                shared = proxy / "example.com/shared" / "@v"
                shared.mkdir(parents=True)
                for version in ("v1.0.0", "v1.2.0"):
                    mod = "module example.com/shared\n\ngo 1.26.0\n"
                    (shared / (version + ".mod")).write_text(mod)
                    (shared / (version + ".info")).write_text(json.dumps({"Version": version, "Time": "2026-10-01T00:00:00Z"}))
                    with zipfile.ZipFile(shared / (version + ".zip"), "w") as archive:
                        archive.writestr("example.com/shared@" + version + "/go.mod", mod)
                        archive.writestr("example.com/shared@" + version + "/shared.go", "package shared\n\nfunc Value() int { return 1 }\n")
            (tree / "go.mod").write_text(
                "module github.com/notaryproject/notation-go\n\ngo 1.26.0\n\n"
                "require github.com/notaryproject/notation-core-go v1.3.0\n\n"
                f"replace github.com/notaryproject/notation-core-go => {module} {old}\n"
            )
            (tree / "main.go").write_text(
                'package notation\n\nimport core "github.com/notaryproject/notation-core-go"\n\nfunc Value() int { return core.Value() }\n'
            )
            if direct_dependency:
                (tree / "go.mod").write_text((tree / "go.mod").read_text() + "\nrequire example.com/shared v1.0.0\n")
                (tree / "main.go").write_text(
                    'package notation\n\nimport (core "github.com/notaryproject/notation-core-go"; "example.com/shared")\n\n'
                    "func Value() int { return core.Value() + shared.Value() }\n"
                )
            environment = {"GOWORK": "off", "GOTOOLCHAIN": "local", "GOPATH": str(root / "go"),
                           "GOMODCACHE": str(root / "modules"), "GOPROXY": proxy.as_uri(), "GOSUMDB": "off",
                           "MONTHLY_PATCH_ACTOR": "test-actor", "MONTHLY_PATCH_SIGNER_LOGIN": "test-actor",
                           "MONTHLY_PATCH_SIGNER_EMAIL": "test@example.com", "MONTHLY_PATCH_TOKEN_READY": "true",
                           "MONTHLY_PATCH_REHEARSAL_ENABLED": "true"}
            with patch.dict(os.environ, environment):
                release.run(["go", "mod", "tidy"], tree)
                release.run(["git", "init", "--quiet"], tree)
                release.run(["git", "add", "."], tree)
                release.run(["git", "-c", "user.name=Test", "-c", "user.email=test@example.com",
                             "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "Baseline"], tree)
                base = release.run(["git", "rev-parse", "HEAD"], tree).strip()
                release.run(["git", "init", "--bare", "--quiet", str(remote)])
                release.run(["git", "push", "--quiet", str(remote), "HEAD:refs/heads/monthly-patch-test-main"], tree)
                key = root / "key"
                release.run(["ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(key)])
                public = key.with_suffix(".pub").read_text()
                plan = worker_plan(GO, "rehearse", MONTH, "waiting")
                auth = authorization()
                auth["snapshot"]["main_head"] = base
                plan.update(controller_issue=12, controller_plan_id=PLAN_ID, producers=auth["producers"])
                before = propagation.manifests(tree, GO)["."]

                class LocalAPI(ControllerAPI):
                    def optional(self, path):
                        if "/git/ref/heads/" in path:
                            return None
                        return super().optional(path)

                    def request(self, path, method="GET", payload=None):
                        if "/branches/" in path:
                            return {"commit": {"sha": base}}
                        if method == "POST" and path.endswith("/pulls"):
                            self.writes.append((path, method, payload))
                            head = release.run(["git", f"--git-dir={remote}", "rev-parse", payload["head"]]).strip()
                            item = {"number": 99, "html_url": "https://github.com/yizha1/notation-go/pull/99",
                                    "user": {"login": "test-actor"}, "body": payload["body"],
                                    "base": {"ref": plan["main"], "sha": base},
                                    "head": {"ref": payload["head"], "sha": head, "repo": {"full_name": GO}}}
                            self.open_pulls.append(item)
                            return item
                        if "/commits/" in path:
                            return {"commit": {"verification": {"verified": True}}, "author": {"login": "test-actor"}}
                        return super().request(path, method, payload)

                    def manifest(self, repository, ref, directory):
                        if repository == CORE:
                            return {"Go": "1.26.0", "Require": [{"Path": "example.com/shared", "Version": "v1.2.0"}]
                                    if direct_dependency else []}
                        if ref == base:
                            return before
                        modfile = root / "remote.mod"
                        modfile.write_text(release.run(["git", f"--git-dir={remote}", "show", ref + ":go.mod"]))
                        return json.loads(release.run(["go", "mod", "edit", "-json", "-modfile=" + str(modfile)]))

                api = LocalAPI()
                api.public_key = public
                actual_git = release.git
                def local_git(command, directory, environment=None):
                    if command[:3] == ["remote", "add", "origin"]:
                        command = [*command[:3], str(remote)]
                    return actual_git(command, directory, environment)
                with patch.dict(os.environ, {"MONTHLY_PATCH_SIGNING_KEY": key.read_text()}), patch.object(
                    controller, "worker_authorization", return_value=auth
                ), patch.object(release, "Cycle") as cycle, patch.object(release, "git", side_effect=local_git):
                    cycle.return_value.state = plan
                    prepared = propagation.prepare(api, plan, root / "prepared")
                    self.assertEqual(set(prepared["files"]), {"go.mod", "go.sum"})
                    result = propagation.publish(api, prepared, root / "published")
                    self.assertEqual(result["pull_request"], 99)
                    head = api.open_pulls[0]["head"]["sha"]
                    if direct_dependency:
                        requirements = {entry["Path"]: entry["Version"] for entry in api.manifest(GO, head, ".")["Require"]}
                        self.assertEqual(requirements["example.com/shared"], "v1.2.0")
                    signers = root / "allowed-signers"
                    signers.write_text("test@example.com " + public)
                    release.run(["git", "-c", "gpg.ssh.allowedSignersFile=" + str(signers),
                                 f"--git-dir={remote}", "verify-commit", head])
                    names = release.run(["git", f"--git-dir={remote}", "diff", "--name-only", base, head]).splitlines()
                    self.assertEqual(set(names), {"go.mod", "go.sum"})
                    self.assertEqual(propagation.publish(api, prepared, root / "retry")["pull_request"], 99)
                    self.assertEqual(len(api.writes), 1)

    def test_direct_dependency_changes_are_limited_to_the_producer_required_floor(self):
        before = {"Go": "1.26.0",
                  "Require": [{"Path": "github.com/notaryproject/notation-core-go", "Version": "v1.3.0"},
                              {"Path": "example.com/shared", "Version": "v1.0.0"}],
                  "Replace": [{"Old": {"Path": "github.com/notaryproject/notation-core-go"},
                               "New": {"Path": "github.com/yizha1/notation-core-go", "Version": "v1.3.1-trial.4"}}]}
        after = copy.deepcopy(before)
        after["Replace"][0]["New"]["Version"] = TARGET
        metadata = {"notation-core-go": {"Go": "1.26.0",
                                         "Require": [{"Path": "example.com/shared", "Version": "v1.2.0"}]}}
        for version, expected in (("v1.0.0", False), ("v1.1.0", False), ("v1.2.0", True), ("v1.3.0", False)):
            after["Require"][1]["Version"] = version
            with self.subTest(version=version):
                self.assertEqual(controller.producer_manifests({".": before}, {".": after}, authorization()["producers"],
                                                              metadata, "go1.27.1"), expected)
        before["Require"][1]["Version"] = "v1.3.0"
        after["Require"][1]["Version"] = "v1.2.0"
        self.assertFalse(controller.producer_manifests({".": before}, {".": after}, authorization()["producers"],
                                                      metadata, "go1.27.1"))
        after["Require"][1]["Version"] = "v1.3.0"
        self.assertTrue(controller.producer_manifests({".": before}, {".": after}, authorization()["producers"],
                                                     metadata, "go1.27.1"))
        before["Require"][1]["Version"] = "v1.0.0"
        after["Require"][1]["Version"] = "v1.2.0"
        after["Require"].append({"Path": "example.com/unrelated", "Version": "v1.0.0"})
        self.assertFalse(controller.producer_manifests({".": before}, {".": after}, authorization()["producers"],
                                                      metadata, "go1.27.1"))
        after["Require"].pop()
        after["Replace"][0]["New"]["Version"] = before["Replace"][0]["New"]["Version"]
        self.assertFalse(controller.producer_manifests({".": before}, {".": after}, authorization()["producers"],
                                                      metadata, "go1.27.1"))

    def test_module_version_ordering_matches_go_semver_precedence(self):
        ordered = ["v1.0.0-alpha", "v1.0.0-alpha.1", "v1.0.0-alpha.beta", "v1.0.0-beta",
                   "v1.0.0-beta.2", "v1.0.0-beta.11", "v1.0.0-rc.1", "v1.0.0", "v1.2.0", "v1.10.0"]
        self.assertEqual(sorted(reversed(ordered), key=controller.module_version), ordered)
        self.assertEqual(controller.module_version("v2.0.0+incompatible"), controller.module_version("v2.0.0"))
        for invalid in ("v01.0.0", "v1.0.0-01", "v1.0", "v1.0.0+", None):
            with self.subTest(version=invalid), self.assertRaises(ValueError):
                controller.module_version(invalid)

    def test_shared_dependency_uses_highest_changed_producer_requirement(self):
        producers = {name: {"repository": f"yizha1/{name}", "version": TARGET}
                     for name in ("notation-core-go", "notation-go")}
        before = {"Go": "1.26.0",
                  "Require": [{"Path": f"github.com/notaryproject/{name}", "Version": "v1.3.0"} for name in producers]
                             + [{"Path": "example.com/shared", "Version": "v1.0.0"}],
                  "Replace": [{"Old": {"Path": f"github.com/notaryproject/{name}"},
                               "New": {"Path": f"github.com/yizha1/{name}", "Version": "v1.3.1-trial.4"}}
                              for name in producers]}
        after = copy.deepcopy(before)
        for replacement in after["Replace"]:
            replacement["New"]["Version"] = TARGET
        metadata = {name: {"Go": "1.26.0", "Require": [{"Path": "example.com/shared", "Version": version}]}
                    for name, version in zip(producers, ("v1.10.0", "v1.2.0"))}
        for version, expected in (("v1.2.0", False), ("v1.10.0", True)):
            after["Require"][-1]["Version"] = version
            with self.subTest(version=version):
                self.assertEqual(controller.producer_manifests({".": before}, {".": after}, producers, metadata, "go1.27.1"),
                                 expected)

    def test_source_head_must_contain_only_recorded_approved_merges(self):
        api = ControllerAPI()
        plan = worker_plan(GO, "rehearse", MONTH, "waiting")
        plan["merged"] = [{"number": 1, "commit": "c" * 40}]
        with patch.object(release, "Cycle") as cycle, patch.object(controller, "worker_authorization"), patch.object(
            controller, "compare_commits", return_value={"d" * 40}
        ), self.assertRaisesRegex(ValueError, "outside the approved"):
            cycle.return_value.state = plan
            propagation.source_head(api, plan, {**authorization(), "snapshot": {**authorization()["snapshot"], "main_head": PR_HEAD}})

    def test_no_plan_and_no_producer_cannot_generate_updates(self):
        plan = worker_plan(GO, "rehearse", MONTH, "waiting")
        with tempfile.TemporaryDirectory() as temporary, patch.object(controller, "worker_authorization", return_value=None):
            with self.assertRaisesRegex(ValueError, "authorized fork"):
                propagation.prepare(ControllerAPI(), plan, pathlib.Path(temporary))
        with tempfile.TemporaryDirectory() as temporary, patch.object(
            controller, "worker_authorization", return_value={**authorization(), "producers": {}}
        ):
            self.assertEqual(propagation.prepare(ControllerAPI(), plan, pathlib.Path(temporary))["action"], "none")

    def test_existing_approved_dependency_pr_waits_without_duplicate_creation(self):
        api = ControllerAPI()
        api.open_pulls = [pull(base="monthly-patch-test-main")]
        plan = worker_plan(GO, "rehearse", MONTH, "waiting")
        with tempfile.TemporaryDirectory() as temporary, patch.object(
            controller, "worker_authorization", return_value=authorization()
        ), patch.object(controller, "scoped_pull", return_value=True), patch.object(propagation, "source_head") as head:
            result = propagation.prepare(api, plan, pathlib.Path(temporary))
        self.assertEqual(result["action"], "none")
        head.assert_not_called()
        self.assertEqual(api.writes, [])

    def test_malformed_and_out_of_scope_artifacts_fail_before_checkout(self):
        plan = worker_plan(GO, "rehearse", MONTH, "waiting")
        valid = {"schema": 1, "action": "pull-request", "plan": plan, "base": MAIN,
                 "original": {name: None for name in propagation.module_paths(GO)},
                 "files": {"go.mod": "module example\n"}}
        for artifact in (
            {**valid, "schema": True},
            {**valid, "files": {"../../outside": "content"}},
            {**valid, "files": {"go.mod": 42}},
            {**valid, "files": {"go.mod": "x" * 2_000_001}},
            {**valid, "original": {}},
        ):
            with tempfile.TemporaryDirectory() as temporary, patch.object(
                controller, "worker_authorization", return_value=authorization()
            ), patch.object(propagation, "source_head", return_value=MAIN), patch.object(release, "policy"), patch.object(
                release, "git"
            ) as git, self.subTest(artifact=list(artifact)), self.assertRaises(ValueError):
                propagation.publish(ControllerAPI(), artifact, pathlib.Path(temporary))
            git.assert_not_called()

    def test_source_job_has_no_signing_or_write_credential(self):
        workflow = pathlib.Path(__file__).parent.parent / "workflows/monthly-patch-release.yml"
        source = workflow.read_text().split("  package:\n", 1)[1].split("  publish:\n", 1)[0]
        self.assertNotIn("secrets.", source)
        self.assertNotIn("MONTHLY_PATCH_SIGNING_KEY", source)
        publisher = workflow.read_text().split("  publish:\n", 1)[1].split("  verify:\n", 1)[0]
        self.assertNotIn("go mod tidy", publisher)
        self.assertIn("monthly_release.py publish", publisher)
        self.assertNotIn("build_monthly_assets.py", publisher)

    def test_fork_ci_covers_same_repo_prs_and_uses_no_release_secrets(self):
        workflow = (pathlib.Path(__file__).parent.parent / "workflows/notation-release-validation.yml").read_text()
        self.assertIn("  pull_request:", workflow)
        self.assertIn("test_upstream_authorization.py", workflow)
        self.assertNotIn("secrets.", workflow)
        self.assertNotIn("head.repo.full_name !=", workflow)


class SetupReadinessTests(unittest.TestCase):
    def setUp(self):
        class SetupAPI(ControllerAPI):
            def __init__(self):
                super().__init__()
                self.missing = set()
                self.secrets = {repo: ["MONTHLY_PATCH_TOKEN", "MONTHLY_PATCH_SIGNING_KEY"]
                                for repo in controller.REPOSITORIES}
                self.secrets[CLI].append("COPILOT_GITHUB_TOKEN")
                self.variables = {repo: {
                    "MONTHLY_PATCH_ACTOR": "test-actor", "MONTHLY_PATCH_SIGNER_LOGIN": "test-actor",
                    "MONTHLY_PATCH_SIGNER_EMAIL": "test@example.com", "MONTHLY_PATCH_REHEARSAL_ENABLED": "true",
                    "NOTATION_RELEASE_COORDINATOR_ENABLED": "true",
                } for repo in controller.REPOSITORIES}

            def request(self, path, method="GET", payload=None):
                repository = "/".join(path.split("/")[1:3])
                if "/actions/secrets?" in path:
                    entries = [{"name": name} for name in self.secrets[repository]]
                    return {"total_count": len(entries), "secrets": entries}
                if "/actions/variables?" in path:
                    entries = [{"name": name, "value": value} for name, value in self.variables[repository].items()]
                    return {"total_count": len(entries), "variables": entries}
                return super().request(path, method, payload)

            def optional(self, path):
                if any(value in path for value in self.missing):
                    return None
                if "contents/" in path:
                    value = ("target-branch: monthly-patch-test-main\n" if "dependabot.yml" in path
                             else "  pull_request:\nqualify_monthly_candidate.py\n")
                    return {"encoding": "base64", "content": base64.b64encode(value.encode()).decode()}
                return self.request(path)

            def manifest(self, repository, commit, directory="."):
                producers = release.PROJECTS[repository.split("/")[1]]
                return {
                    "Require": [{"Path": f"github.com/notaryproject/{name}", "Version": "v1.3.0"}
                                for name in producers],
                    "Replace": [{"Old": {"Path": f"github.com/notaryproject/{name}"},
                                 "New": {"Path": f"github.com/yizha1/{name}", "Version": TARGET}}
                                for name in producers],
                }

        self.api = SetupAPI()
        for name, value in (
            ("fork_metadata", {"has_issues": True}),
            ("installed_worker", True),
        ):
            mocked = patch.object(controller, name, return_value=value)
            mocked.start()
            self.addCleanup(mocked.stop)
        for name, value in (
            ("baseline_plan", {"main": "monthly-patch-test-main", "branch": "monthly-patch-test-release-1.3",
                               "baseline": "v1.3.0"}),
            ("check_publisher_guard", None),
        ):
            mocked = patch.object(release, name, return_value=value)
            mocked.start()
            self.addCleanup(mocked.stop)

    def test_complete_fork_setup_is_ready_without_writes(self):
        result = setup.check(self.api, MONTH, dependabot_confirmed=True)
        self.assertTrue(result["ready"])
        self.assertEqual(set(result["repositories"]), set(controller.REPOSITORIES))
        self.assertEqual(self.api.writes, [])

    def test_configuration_does_not_prove_dependabot_enabled_on_a_fork(self):
        result = setup.check(self.api, MONTH)
        self.assertFalse(result["ready"])
        for state in result["repositories"].values():
            self.assertEqual(len(state["blockers"]), 1)
            self.assertIn("Enable Dependabot version updates", state["blockers"][0])

    def test_missing_branch_ci_and_token_are_reported_together(self):
        self.api.missing.update((f"repos/{CORE}/contents/.github/workflows/notation-fork-ci.yml",
                                 f"repos/{CLI}/branches/monthly-patch-test-release-1.3"))
        self.api.secrets[GO].remove("MONTHLY_PATCH_TOKEN")
        result = setup.check(self.api, MONTH, dependabot_confirmed=True)
        self.assertFalse(result["ready"])
        self.assertEqual(len(result["repositories"][CORE]["blockers"]), 2)
        self.assertIn("Install Actions secret MONTHLY_PATCH_TOKEN", result["repositories"][GO]["blockers"])
        self.assertIn("Create reviewed branch monthly-patch-test-release-1.3", result["repositories"][CLI]["blockers"])

    def test_actor_and_signer_must_match_the_coordinator(self):
        self.api.variables[CORE]["MONTHLY_PATCH_ACTOR"] = "other-actor"
        self.api.variables[CORE]["MONTHLY_PATCH_SIGNER_LOGIN"] = "other-actor"
        self.api.variables[GO]["MONTHLY_PATCH_SIGNER_LOGIN"] = "different-signer"
        result = setup.check(self.api, MONTH, dependabot_confirmed=True)
        self.assertFalse(result["ready"])
        self.assertIn("Use the same MONTHLY_PATCH_ACTOR as the CLI coordinator", result["repositories"][CORE]["blockers"])
        self.assertIn("Use the actor's signing identity for fork producer PRs", result["repositories"][GO]["blockers"])

    def test_metadata_collection_paginates_without_reading_secret_values(self):
        class MetadataAPI:
            def __init__(self):
                self.paths = []
            def request(self, path):
                self.paths.append(path)
                page = int(path.rsplit("=", 1)[1])
                names = range(100) if page == 1 else range(100, 101)
                return {"total_count": 101, "secrets": [{"name": f"SECRET_{number}"} for number in names]}
        api = MetadataAPI()
        self.assertEqual(len(setup.named_metadata(api, "repos/example/repo/actions/secrets", "secrets")), 101)
        self.assertEqual(len(api.paths), 2)
        self.assertTrue(all("/actions/secrets?" in path for path in api.paths))

    def test_incomplete_setup_metadata_fails_explicitly(self):
        api = ControllerAPI()
        with patch.object(api, "request", return_value={"total_count": 2, "secrets": [{"name": "ONE"}]}), self.assertRaisesRegex(ValueError, "Incomplete"):
            setup.named_metadata(api, "repos/example/repo/actions/secrets", "secrets")


if __name__ == "__main__":
    unittest.main()
