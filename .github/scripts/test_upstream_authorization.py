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
import unittest
from unittest.mock import patch

import monthly_release as release
import notation_release_controller as controller
from test_monthly_release import FakeGitHub, body, plan as candidate, pull

CORE, GO, CLI = ("notaryproject/" + name for name in release.PROJECTS)
SOURCE, BRANCH, PR_HEAD, MERGED = (character * 40 for character in "abcd")
MONTH = "2026-10"
POLICY = {
    "schema": 1, "repositories": [CORE, GO, CLI],
    "dependency_versions": "same-major-minor-nondecreasing",
    "producer_versions": "exact-planned-stable-patch",
    "go_requirement": "unchanged-for-routine-updates",
    "reachable_vulnerabilities": "block", "optional_backports": "excluded",
    "workflow_updates": "explicit-approval", "maximum_assessment_age_hours": 24,
    "plan_approval": "active-cli-maintainer-comment",
    "publication": "maintainer-publishes-qualified-draft",
    "release_votes": {"notation-core-go": "branch-rules", "notation-go": "repo-maintainer-majority",
                      "notation": "repo-maintainer-majority"},
}
TREE = [{"path": controller.WORKER_PATH, "sha": "e" * 40, "type": "blob", "mode": "100644"}]
AUTOMATION = controller.digest({item["path"]: {key: item[key] for key in ("sha", "mode", "type")} for item in TREE})


class UpstreamAPI(FakeGitHub):
    def __init__(self):
        super().__init__()
        self.issues_by_repo = {repo: [] for repo in (CORE, GO, CLI)}
        self.changed = {}
        self.tree = copy.deepcopy(TREE)
        self.dispatched = []
        self.runs = {}
        self.policy = copy.deepcopy(POLICY)
        self.comparison = []
        self.prs = {}
        self.editor = "test-actor"
        self.comments = {}
        self.roster = "# Repo-Level Maintainers\nMaintainer <maintainer@example.com> (@maintainer)\n"

    def issue_editor(self, repository, number):
        return self.editor

    def optional(self, path):
        if "/contents/" in path and (controller.WORKER_PATH in path or "/MAINTAINERS?" in path):
            return self.request(path)
        return super().optional(path)

    def request(self, path, method="GET", payload=None):
        repository = "/".join(path.split("/")[1:3])
        suffix = "/".join(path.split("/")[3:])
        if path == "user":
            return {"login": "test-actor"}
        if method == "POST" and suffix.endswith("/dispatches"):
            self.dispatched.append((repository, payload))
            self.writes.append((path, method, payload))
            return None
        if method != "GET":
            return super().request(path, method, payload)

        if not suffix:
            return {"fork": False, "private": False, "default_branch": "main", "has_issues": True}
        if suffix.startswith("branches/"):
            name = suffix.removeprefix("branches/")
            return {"commit": {"sha": self.changed.get((repository, name), SOURCE if name == "main" else BRANCH)}}
        if suffix.startswith("git/trees/"):
            return {"truncated": False, "tree": self.tree}
        if suffix.startswith("git/ref/tags/"):
            return {"object": {"type": "tag", "sha": "e" * 40}}
        if suffix.startswith("git/tags/"):
            return {"object": {"sha": BRANCH}, "verification": {"verified": True}}
        if suffix == "releases/777":
            return {"draft": False}
        if suffix.startswith(f"contents/{controller.POLICY_PATH}"):
            return {"encoding": "base64", "content": base64.b64encode(json.dumps(self.policy).encode()).decode()}
        if suffix.startswith("issues/"):
            if suffix.startswith("issues/comments/"):
                return self.comments[int(suffix.rsplit("/", 1)[1])]
            return next(item for item in self.issues_by_repo[repository] if item["number"] == int(suffix.rsplit("/", 1)[1]))
        if suffix.startswith("contents/MAINTAINERS?"):
            return {"encoding": "base64", "content": base64.b64encode(self.roster.encode()).decode()}
        if suffix == "actions/workflows/monthly-patch-release.yml":
            return {"id": (CORE, GO, CLI).index(repository) + 10, "state": "active"}
        if "/runs?" in suffix:
            return {"workflow_runs": list(self.runs.get(repository, {}).values())}
        if suffix.startswith("actions/runs/"):
            registered = self.runs.get(repository, {}).get(int(suffix.rsplit("/", 1)[1]))
            if registered is not None:
                return registered
            state = next(controller.decode_issue(issue, "test-actor") for issue in self.issues_by_repo[CLI]
                         if controller.decode_issue(issue, "test-actor") is not None)
            node = state["nodes"][repository]
            return {
                "actor": {"login": "test-actor"}, "head_sha": node.get("dispatch_sha", SOURCE),
                "workflow_id": (CORE, GO, CLI).index(repository) + 10, "event": "workflow_dispatch",
                "display_title": controller.run_title(MONTH, state["plan"]["plan_id"], 1, node["attempt"]),
                "referenced_workflows": [{"path": controller.HOST + "/" + controller.SHARED_PATH + "@" + SOURCE,
                                          "sha": SOURCE}]}
        if suffix.startswith("compare/"):
            return {"status": "ahead", "ahead_by": len(self.comparison), "behind_by": 0,
                    "commits": [{"sha": sha} for sha in self.comparison]}
        if suffix.startswith("contents/") and controller.WORKER_PATH in suffix:
            reference = "./" + controller.SHARED_PATH if repository == CLI else controller.HOST + "/" + controller.SHARED_PATH + "@" + SOURCE
            pin = "${{ github.sha }}" if repository == CLI else SOURCE
            text = f"jobs:\n  release:\n    uses: {reference}\n    with:\n      worker_sha: {pin}\n"
            return {"encoding": "base64", "content": base64.b64encode(text.encode()).decode()}
        if suffix.startswith("pulls/"):
            return self.prs[(repository, int(suffix.rsplit("/", 1)[1]))]
        return super().request(path, method, payload)

    def worker(self, repository, issue, state, status="published", conclusion="success"):
        node = state["nodes"][repository]
        run_id = len(self.runs) + 100
        run = {
            "id": run_id, "actor": {"login": "test-actor"}, "head_sha": node.get("dispatch_sha", SOURCE),
            "workflow_id": (CORE, GO, CLI).index(repository) + 10,
            "event": "workflow_dispatch", "status": "completed", "conclusion": conclusion,
            "display_title": controller.run_title(MONTH, state["plan"]["plan_id"], issue["number"], node["attempt"]),
            "referenced_workflows": [{"path": controller.HOST + "/" + controller.SHARED_PATH + "@" + node.get("worker_sha", SOURCE),
                                      "sha": node.get("worker_sha", SOURCE)}],
        }
        self.runs.setdefault(repository, {})[run_id] = run
        producers = {
            name: {"repository": f"notaryproject/{name}", "version": state["nodes"][f"notaryproject/{name}"]["version"]}
            for name in release.PROJECTS[repository.split("/")[1]]
            if state["nodes"][f"notaryproject/{name}"]["status"] == "published"
        }
        worker = candidate(repository, "execute", MONTH, status)
        worker.update(controller_issue=issue["number"], controller_plan_id=state["plan"]["plan_id"], producers=producers,
                      assets=[], release_id=777)
        worker["maintainer_roster_sha"] = state["plan"]["snapshots"][repository]["automation_sha"]
        release.Cycle(self, repository, MONTH, "execute").save(worker)
        for directory in release.modules(repository):
            self.manifests[(repository, worker["commit"], directory)] = {
                "Require": [{"Path": f"github.com/notaryproject/{name}", "Version": item["version"]} for name, item in producers.items()],
            }
        return run

    def pages(self, path):
        if path.endswith("/comments"):
            return iter(self.comments.values())
        return super().pages(path)

    def approve_comment(self, issue, state):
        identifier = len(self.comments) + 1
        self.comments[identifier] = {
            "id": identifier, "body": controller.plan_comment(state["plan"]),
            "user": {"login": "maintainer", "type": "User"},
            "created_at": "2099-01-01T00:00:00Z", "updated_at": "2099-01-01T00:00:00Z",
            "issue_url": f"https://api.github.com/repos/{CLI}/issues/{issue['number']}",
        }


class ApprovedController(controller.Controller):
    """Old execution regressions use a real, explicit maintainer review receipt."""

    def approve(self, *args, **kwargs):
        issue, state = super().approve(*args, **kwargs)
        if state["status"] == "review":
            for repository, snapshot in state["plan"]["snapshots"].items():
                for item in snapshot["pulls"]:
                    update = pull(sha=item["head"])
                    update["head"]["repo"]["full_name"] = repository
                    self.api.prs[(repository, item["number"])] = update
            self.api.approve_comment(issue, state)
            return self.accept_review(issue, state)
        return issue, state


def inventory(changes=(CORE,)):
    return {
        "schema": 1, "host": CLI, "month": MONTH, "collected_at": controller.utc_now(),
        "repositories": {
            repo: {
                "baseline": "v1.3.0", "main": "main", "branch": "release-1.3", "tag": "v1.3.1",
                "main_head": SOURCE, "release_head": BRANCH, "automation_sha": SOURCE,
                "automation_digest": AUTOMATION, "workflow_id": index + 10,
                "worker_repository": controller.HOST, "worker_path": controller.SHARED_PATH,
                "worker_sha": SOURCE, "worker_digest": AUTOMATION,
                "baseline_sha": BRANCH, "baseline_tag_object": "e" * 40,
                "source_repository": repo, "source_commit": BRANCH, "blockers": [],
                "pulls": [{"number": 1, "head": PR_HEAD, "commits": [{"sha": PR_HEAD, "subject": "Dependency patch"}]}]
                         if repo in changes else [],
                "merged": [], "manifests": {directory: {} for directory in release.modules(repo)},
                "scans": {directory: {"reachable": [], "informational": [], "advisories": {}} for directory in release.modules(repo)},
            }
            for index, repo in enumerate((CORE, GO, CLI))
        },
    }


def assessed(evidence, decisions=("release", "release", "release")):
    assessment = {
        "schema": 1, "month": MONTH, "inventory_sha256": controller.digest(evidence),
        "decisions": {
            repo: {"decision": decision, "reason": "Assessed update",
                   "evidence": [f"{repo}:dependencies"]}
            for repo, decision in zip((CORE, GO, CLI), decisions)
        },
    }
    return controller.validate_assessment(evidence, assessment)


class UpstreamAuthorizationTests(unittest.TestCase):
    def setUp(self):
        self.api = UpstreamAPI()
        self.env = patch.dict(os.environ, {
            "GITHUB_REPOSITORY": CLI, "GITHUB_REF": "refs/heads/main", "GITHUB_SHA": SOURCE,
            "MONTHLY_PATCH_ACTOR": "test-actor", "MONTHLY_PATCH_COORDINATOR_REQUIRED": "true",
            "NOTATION_RELEASE_AUTO_APPROVAL_ENABLED": "true",
            "MONTHLY_PATCH_WORKER_SHA": SOURCE, "GITHUB_RUN_ID": "1",
        })
        self.env.start()
        self.addCleanup(self.env.stop)
        self.controller = ApprovedController(self.api, "test-actor")

    def dispatch(self):
        issue, state = self.controller.approve(assessed(inventory()), 22)
        issue, state = self.controller.advance(issue, state)
        os.environ.update({
            "MONTHLY_PATCH_CONTROLLER_ISSUE": str(issue["number"]),
            "MONTHLY_PATCH_CONTROLLER_PLAN": state["plan"]["plan_id"],
            "MONTHLY_PATCH_CONTROLLER_ATTEMPT": "1",
        })
        return issue, state

    def test_canonical_plan_and_dispatch_use_stable_versions_and_real_main(self):
        issue, state = self.dispatch()
        self.assertEqual(controller.HOST, CLI)
        self.assertEqual(state["mode"], "execute")
        self.assertEqual(self.api.dispatched[0][0], CORE)
        self.assertEqual(self.api.dispatched[0][1]["inputs"]["mode"], "execute")
        self.assertEqual(controller.worker_authorization(self.api, CORE, MONTH)["snapshot"]["tag"], "v1.3.1")
        with self.assertRaises(ValueError):
            controller.worker_authorization(self.api, "yizha1/notation-core-go", MONTH)

    def test_changed_automation_blocks_dispatch_and_worker(self):
        issue, state = self.dispatch()
        self.api.tree[0]["sha"] = "f" * 40
        with self.assertRaisesRegex(ValueError, "code changed"):
            controller.worker_authorization(self.api, CORE, MONTH)
        state["nodes"][CORE]["status"] = "pending"
        issue = self.controller.save(issue, state)
        with self.assertRaisesRegex(ValueError, "code changed"):
            self.controller.advance(issue, state)

    def test_recorded_dependency_merge_does_not_invalidate_automation(self):
        issue, state = self.dispatch()
        worker = candidate(CORE, "execute", MONTH, "waiting")
        worker.update(controller_issue=issue["number"], controller_plan_id=state["plan"]["plan_id"],
                      merged=[{"number": 1, "commit": MERGED}], producers={})
        release.Cycle(self.api, CORE, MONTH, "execute").save(worker)
        self.api.changed[(CORE, "main")] = MERGED
        self.api.comparison = [MERGED]
        merged = pull(sha=PR_HEAD)
        merged.update(merged_at="2026-10-01T00:00:00Z", merge_commit_sha=MERGED, state="closed")
        self.api.prs[(CORE, 1)] = merged
        self.assertEqual(controller.source_checkpoint(self.api, CORE, state["plan"]["snapshots"][CORE], MONTH), MERGED)
        controller.worker_authorization(self.api, CORE, MONTH)
        self.api.comparison.append("f" * 40)
        with self.assertRaisesRegex(ValueError, "outside recorded"):
            controller.worker_authorization(self.api, CORE, MONTH)

    def test_cli_relative_worker_revision_advances_only_with_recorded_source_merges(self):
        issue, state = self.controller.approve(assessed(inventory((CLI,)), ("skip", "skip", "release")), 22)
        issue, state = self.controller.advance(issue, state)
        worker = candidate(CLI, "execute", MONTH, "waiting")
        worker.update(controller_issue=issue["number"], controller_plan_id=state["plan"]["plan_id"],
                      merged=[{"number": 1, "commit": MERGED}], producers={}, maintainer_roster_sha=SOURCE)
        self.api.worker(CLI, issue, state, status="waiting")
        release.Cycle(self.api, CLI, MONTH, "execute").save(worker)
        merged = pull(sha=PR_HEAD)
        merged["head"]["repo"]["full_name"] = CLI
        merged.update(merged_at="2026-10-01T00:00:00Z", merge_commit_sha=MERGED, state="closed")
        self.api.prs[(CLI, 1)] = merged
        self.api.changed[(CLI, "main")] = MERGED
        self.api.comparison = [MERGED]
        issue, state = self.controller.advance(issue, state)
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["nodes"][CLI]["dispatch_sha"], MERGED)
        self.assertEqual(state["nodes"][CLI]["worker_sha"], MERGED)
        run = self.api.worker(CLI, issue, state, status="waiting")
        release.Cycle(self.api, CLI, MONTH, "execute").save(worker)
        os.environ.update(GITHUB_SHA=MERGED, MONTHLY_PATCH_WORKER_SHA=MERGED, GITHUB_RUN_ID=str(run["id"]),
                          MONTHLY_PATCH_CONTROLLER_ISSUE=str(issue["number"]),
                          MONTHLY_PATCH_CONTROLLER_PLAN=state["plan"]["plan_id"],
                          MONTHLY_PATCH_CONTROLLER_ATTEMPT=str(state["nodes"][CLI]["attempt"]))
        controller.worker_authorization(self.api, CLI, MONTH, worker)
        self.api.tree[0]["sha"] = "f" * 40
        with self.assertRaisesRegex(ValueError, "code changed"):
            controller.worker_authorization(self.api, CLI, MONTH, worker)

    def test_unrecorded_main_changes_require_reassessment(self):
        issue, state = self.dispatch()
        self.api.changed[(CORE, "main")] = MERGED
        with self.assertRaisesRegex(ValueError, "trusted worker"):
            controller.worker_authorization(self.api, CORE, MONTH)

    def test_policy_is_disabled_by_default_and_exactly_scoped(self):
        with patch.dict(os.environ, {"NOTATION_RELEASE_AUTO_APPROVAL_ENABLED": ""}), self.assertRaisesRegex(ValueError, "disabled"):
            controller.authorize_routine(self.api, assessed(inventory()))
        self.api.policy["optional_backports"] = "automatic"
        with self.assertRaisesRegex(ValueError, "Unsupported"):
            controller.review_scope(self.api, assessed(inventory()))

    def test_manual_review_scope_checks_semantics_not_agent_rationale(self):
        update = pull(sha=PR_HEAD)
        update["base"]["sha"] = SOURCE
        self.api.prs[(CORE, 1)] = update
        before = {"Go": "1.24", "Require": [{"Path": "example/dependency", "Version": "v1.2.0"}]}
        after = copy.deepcopy(before)
        after["Require"][0]["Version"] = "v1.2.1"
        self.api.manifests[(CORE, SOURCE, ".")] = before
        self.api.manifests[(CORE, PR_HEAD, ".")] = after
        receipt = controller.review_scope(self.api, assessed(inventory()))
        self.assertEqual(receipt["kind"], "pending")
        for version in ("v2.0.0", "v1.3.0", "v1.1.9", "v1.2.1-rc.1"):
            after["Require"][0]["Version"] = version
            with self.subTest(version=version), self.assertRaises(ValueError):
                controller.review_scope(self.api, assessed(inventory()))

    def test_reachable_findings_and_workflow_changes_need_explicit_resolution(self):
        evidence = inventory()
        evidence["repositories"][CORE]["scans"]["."]["reachable"] = ["GO-2026-1234"]
        with self.assertRaisesRegex(ValueError, "Reachable"):
            controller.review_scope(self.api, assessed(evidence))
        update = pull(sha=PR_HEAD)
        update["base"]["sha"] = SOURCE
        self.api.prs[(CORE, 1)] = update
        self.api.files = [{"filename": ".github/workflows/build.yml"}]
        with self.assertRaisesRegex(ValueError, "explicit actor"):
            controller.review_scope(self.api, assessed(inventory()))

    def test_exceptional_authorization_requires_manual_configured_actor(self):
        for event, actor in (("schedule", "test-actor"), ("workflow_dispatch", "attacker")):
            with patch.dict(os.environ, {"GITHUB_EVENT_NAME": event, "GITHUB_ACTOR": actor}), self.assertRaises(ValueError):
                controller.explicit_actor()
        with patch.dict(os.environ, {"GITHUB_EVENT_NAME": "workflow_dispatch", "GITHUB_ACTOR": "test-actor"}):
            controller.explicit_actor()

    def test_late_producer_updates_require_exact_canonical_versions(self):
        update = pull(number=99, sha=PR_HEAD)
        update["head"]["repo"]["full_name"] = GO
        update["base"]["sha"] = SOURCE
        before = {"Go": "1.24", "Require": [{"Path": "github.com/notaryproject/notation-core-go", "Version": "v1.3.0"}]}
        after = copy.deepcopy(before)
        after["Require"][0]["Version"] = "v1.3.1"
        self.api.manifests = {
            (GO, SOURCE, "."): before, (GO, PR_HEAD, "."): after,
            (CORE, "v1.3.1", "."): {"Go": "1.24"},
        }
        producers = {"notation-core-go": {"repository": CORE, "version": "v1.3.1"}}
        with patch.object(release, "run", return_value="go1.27.1"):
            self.assertTrue(controller.producer_pull(self.api, GO, update, producers))
            after["Require"][0]["Version"] = "v1.3.2"
            self.assertFalse(controller.producer_pull(self.api, GO, update, producers))
            after["Require"][0]["Version"] = "v1.3.1"
            after["Replace"] = [{"Old": {"Path": "github.com/notaryproject/notation-core-go"},
                                 "New": {"Path": "github.com/yizha1/notation-core-go", "Version": "v1.3.1"}}]
            with self.assertRaises(ValueError):
                controller.producer_pull(self.api, GO, update, producers)

    def test_source_and_packaging_jobs_receive_no_release_secrets(self):
        workflow = (pathlib.Path(__file__).parent.parent / "workflows/monthly-patch-worker.yml").read_text()
        for name, next_name in (("qualify", "package"), ("package", "publish"), ("verify", "complete")):
            section = workflow.split(f"  {name}:\n", 1)[1].split(f"  {next_name}:\n", 1)[0]
            self.assertNotIn("secrets.", section)
        publisher = workflow.split("  publish:\n", 1)[1].split("  verify:\n", 1)[0]
        self.assertNotIn("build_monthly_assets.py", publisher)
        self.assertNotIn("go test", publisher)
        callback = workflow.split("  notify-controller:\n", 1)[1]
        self.assertIn("inputs.mode == 'execute'", callback)
        self.assertIn("github.actor == vars.MONTHLY_PATCH_ACTOR", callback)
        self.assertIn("github.ref == 'refs/heads/main'", callback)

    def test_worker_roster_pin_must_match_assessed_main(self):
        issue, state = self.dispatch()
        worker = candidate(CORE, "execute", MONTH, "waiting")
        worker.update(controller_issue=issue["number"], controller_plan_id=state["plan"]["plan_id"], producers={})
        with self.assertRaisesRegex(ValueError, "roster"):
            controller.worker_authorization(self.api, CORE, MONTH, worker)
        worker["maintainer_roster_sha"] = SOURCE
        controller.worker_authorization(self.api, CORE, MONTH, worker)
        worker["maintainer_roster_sha"] = BRANCH
        with self.assertRaisesRegex(ValueError, "roster"):
            controller.worker_authorization(self.api, CORE, MONTH, worker)

    def test_public_recovery_preserves_distinct_voted_commit_and_merge_tip(self):
        issue, previous = self.controller.approve(assessed(inventory()), 22)
        worker = candidate(CORE, "execute", MONTH, "verifying")
        worker.update(commit=PR_HEAD, previous_head=MERGED, release_backport={"number": 7},
                      controller_issue=issue["number"], controller_plan_id=previous["plan"]["plan_id"], assets=[])
        release.Cycle(self.api, CORE, MONTH, "execute").save(worker)
        evidence = inventory()
        evidence["repositories"][CORE].update(main_head="f" * 40, automation_sha="f" * 40, release_head=MERGED)
        self.api.changed[(CORE, "main")] = "f" * 40
        self.api.changed[(CORE, "release-1.3")] = MERGED
        replacement = assessed(evidence)
        optional = self.api.optional
        self.api.changed[(CORE, "release-1.3")] = SOURCE
        with patch.object(self.api, "optional", side_effect=lambda path: {"object": {"sha": PR_HEAD}} if path == f"repos/{CORE}/git/ref/tags/v1.3.1" else optional(path)), \
             patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"), \
             self.assertRaisesRegex(ValueError, "stale"):
            self.controller.approve(replacement, 23, replan=True, recover_public_verification=True)
        self.api.changed[(CORE, "release-1.3")] = MERGED
        with patch.object(self.api, "optional", side_effect=lambda path: {"object": {"sha": PR_HEAD}} if path == f"repos/{CORE}/git/ref/tags/v1.3.1" else optional(path)), \
             patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"):
            issue, state = self.controller.approve(replacement, 23, replan=True, recover_public_verification=True)
        self.assertEqual(state["plan"]["snapshots"][CORE]["main_head"], "f" * 40)
        self.assertEqual(release.Cycle(self.api, CORE, MONTH, "execute").state["commit"], PR_HEAD)
        self.assertEqual(release.Cycle(self.api, CORE, MONTH, "execute").state["previous_head"], MERGED)
    def test_canonical_core_go_cli_order_and_verified_producer_authority(self):
        issue, state = self.dispatch()
        core_run = self.api.worker(CORE, issue, state)
        with patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"):
            issue, state = self.controller.advance(issue, state)
            self.assertEqual([repo for repo, _ in self.api.dispatched], [CORE, GO])
            authorization = controller.worker_authorization(self.api, GO, MONTH)
            self.assertEqual(authorization["producers"], {"notation-core-go": {"repository": CORE, "version": "v1.3.1"}})
            core_run["actor"]["login"] = "forged"
            with self.assertRaisesRegex(ValueError, "trusted successful"):
                controller.worker_authorization(self.api, GO, MONTH)
            core_run["actor"]["login"] = "test-actor"
            self.api.worker(GO, issue, state)
            issue, state = self.controller.advance(issue, state)
            self.assertEqual([repo for repo, _ in self.api.dispatched], [CORE, GO, CLI])
            self.api.worker(CLI, issue, state)
            issue, state = self.controller.advance(issue, state)
            self.assertEqual(state["status"], "completed")

    def test_verification_record_alone_cannot_advance_before_worker_run_success(self):
        issue, state = self.dispatch()
        run = self.api.worker(CORE, issue, state)
        run["status"] = "in_progress"
        issue, state = self.controller.advance(issue, state)
        self.assertEqual([repo for repo, _ in self.api.dispatched], [CORE])
        run.update(status="completed", conclusion="failure")
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["nodes"][CORE]["status"], "failed")
        self.assertEqual([repo for repo, _ in self.api.dispatched], [CORE])

    def test_signed_baseline_and_coordinator_receipt_cannot_be_bypassed(self):
        self.dispatch()
        with patch.dict(os.environ, {"MONTHLY_PATCH_COORDINATOR_REQUIRED": ""}), self.assertRaisesRegex(ValueError, "always require"):
            controller.worker_authorization(self.api, CORE, MONTH)
        with patch.object(release, "verified_tag", return_value="f" * 40), self.assertRaisesRegex(ValueError, "baseline"):
            controller.worker_authorization(self.api, CORE, MONTH)

    def test_issue_author_alone_cannot_authorize_a_forged_state_edit(self):
        self.dispatch()
        self.api.editor = "attacker"
        with self.assertRaisesRegex(ValueError, "edited outside"):
            controller.worker_authorization(self.api, CORE, MONTH)
        with self.assertRaisesRegex(ValueError, "edited outside"):
            self.controller.records()


if __name__ == "__main__":
    unittest.main()
