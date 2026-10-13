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
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import monthly_release as release


REPO = "notaryproject/notation-core-go"
COMMIT, PREVIOUS = "a" * 40, "b" * 40
ENVIRONMENT = {
    "GITHUB_REF": "refs/heads/main", "MONTHLY_PATCH_ENABLED": "true",
    "MONTHLY_PATCH_REHEARSAL_ENABLED": "true", "MONTHLY_PATCH_TOKEN_READY": "true",
    "MONTHLY_PATCH_SIGNING_KEY": "temporary-test-key", "MONTHLY_PATCH_SIGNER_EMAIL": "test@example.com",
    "MONTHLY_PATCH_SIGNER_LOGIN": "test-actor", "MONTHLY_PATCH_ACTOR": "test-actor",
}


def plan(repository=REPO, mode="execute", month="2026-10", status="ready"):
    return {
        "schema": 1, "repository": repository, "mode": mode, "month": month,
        "baseline": "v1.3.0", "baseline_published": "2026-01-01T00:00:00Z",
        "branch": "release-1.3" if mode == "execute" else "monthly-patch-test-release-1.3",
        "main": "main" if mode == "execute" else "monthly-patch-test-main",
        "tag": "v1.3.1" if mode == "execute" else f"v1.3.4-monthly-test.{month.replace('-', '')}",
        "commit": COMMIT, "previous_head": PREVIOUS, "status": status, "merged": [],
        "started_at": "2026-10-01T09:00:00+00:00",
    }


def body(state):
    return release.plan_marker(state) + "\n```json\n" + json.dumps(state) + "\n```"


def pull(number=1, sha=COMMIT, base="main"):
    return {
        "number": number, "state": "open", "draft": False,
        "user": {"login": "dependabot[bot]", "type": "Bot"},
        "base": {"ref": base}, "head": {"sha": sha, "repo": {"full_name": REPO}},
    }


def checks(sha=COMMIT):
    return {
        "headRefOid": sha, "mergeable": "MERGEABLE", "mergeStateStatus": "CLEAN",
        "reviewDecision": "APPROVED", "statusCheckRollup": [
            {"__typename": "CheckRun", "status": "COMPLETED", "conclusion": "SUCCESS"}],
    }


class FakeGitHub:
    def __init__(self):
        self.issues, self.writes, self.open_pulls, self.closed_pulls = [], [], [], []
        self.head = PREVIOUS
        self.releases = [{"tag_name": "v1.3.0", "published_at": "2026-01-01T00:00:00Z",
                          "draft": False, "prerelease": False}]
        self.files = [{"filename": "go.mod"}]
        self.check_result = checks()
        self.releases_by_repo, self.issues_by_repo, self.manifests = {}, {}, {}

    def pages(self, path):
        if "/issues?" in path:
            issues = self.issues_by_repo.get("/".join(path.split("/")[1:3]), self.issues)
            return iter(issues if "state=all" in path else [item for item in issues if item.get("state") != "closed"])
        if path.endswith("/releases"):
            return iter(self.releases_by_repo.get("/".join(path.split("/")[1:3]), self.releases))
        if path.endswith("/tags"):
            return iter([{"name": "v1.3.3-trial.4"}, {"name": "v9.0.0"}])
        if path.endswith("/files"):
            return iter(self.files)
        if path.endswith("/ssh_signing_keys"):
            return iter([{"key": self.public_key}])
        raise AssertionError(path)

    def request(self, path, method="GET", payload=None):
        if method != "GET":
            self.writes.append((path, method, payload))
            if "/issues" in path:
                issues = self.issues_by_repo.get("/".join(path.split("/")[1:3]), self.issues)
                number = int(path.rsplit("/", 1)[1]) if method == "PATCH" else len(issues) + 1
                issue = {
                    "number": number, "body": payload["body"], "state": payload["state"],
                    "user": {"login": "test-actor"},
                }
                issues[:] = [item for item in issues if item.get("number", 1) != number] + [issue]
                return issue
            if path.endswith("/dispatches"):
                return None
            if path.endswith("/merge"):
                return {"merged": True, "sha": COMMIT}
            raise AssertionError(path)
        if path.startswith("repos/") and len(path.split("/")) == 3:
            return {"default_branch": "main", "fork": path.startswith("repos/yizha1")}
        if "/branches/" in path:
            return {"commit": {"sha": self.head}}
        if "/pulls/" in path:
            return next(item for item in self.open_pulls if str(item["number"]) == path.split("/")[-1])
        if "/releases/tags/" in path:
            repository = "/".join(path.split("/")[1:3])
            return next(item for item in self.releases_by_repo.get(repository, self.releases) if item["tag_name"] == path.rsplit("/", 1)[1])
        raise AssertionError(path)

    def optional(self, path):
        if "/git/ref/tags/" in path or "/contents/" in path:
            return None
        return self.request(path)

    def pulls(self, repository, branch, state):
        return self.open_pulls if state == "open" else self.closed_pulls

    def checks(self, repository, number):
        return self.check_result

    def manifest(self, repository, ref, directory):
        return self.manifests.get((repository, ref, directory), {
            "Require": [{"Path": f"github.com/notaryproject/{name}", "Version": "v1.3.0"}
                        for name in release.PROJECTS[repository.split("/")[1]]],
        })


class MonthlyReleaseTests(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ, ENVIRONMENT)
        self.environment.start()
        self.addCleanup(self.environment.stop)

    def test_stable_selection_uses_semver_and_excludes_trials(self):
        items = [{"tag_name": tag, "draft": draft, "prerelease": pre} for tag, draft, pre in [
            ("v1.9.9", False, False), ("v1.10.0", False, False),
            ("v2.0.0", True, False), ("v3.0.0-test", False, True), ("v01.11.0", False, False)]]
        self.assertEqual(release.latest_release(items)["tag_name"], "v1.10.0")
        with self.assertRaises(ValueError):
            release.latest_release([])

    def test_rehearsal_bumps_above_all_same_line_trial_tags(self):
        candidate = release.baseline_plan(FakeGitHub(), "yizha1/notation-core-go", "2026-10", "rehearse")
        self.assertEqual(candidate["tag"], "v1.3.4-monthly-test.202610")
        self.assertEqual(candidate["branch"], "monthly-patch-test-release-1.3")
        release.validate_plan(candidate)

    def test_policy_requires_identity_opt_in_trusted_ref_and_all_credentials(self):
        release.policy(REPO, "execute", ENVIRONMENT, {"default_branch": "main"})
        for name in ENVIRONMENT:
            if name == "MONTHLY_PATCH_REHEARSAL_ENABLED":
                continue
            environment = {**ENVIRONMENT, name: ""}
            with self.subTest(name=name), self.assertRaises(ValueError):
                release.policy(REPO, "execute", environment, {"default_branch": "main"})
        for repository, mode, metadata in [
            ("yizha1/notation", "execute", {"default_branch": "main"}),
            ("notaryproject/notation", "rehearse", {"fork": True}),
            ("yizha1/notation", "rehearse", {"fork": False}),
            ("other/product", "dry-run", {}),
        ]:
            with self.subTest(repository=repository, mode=mode), self.assertRaises(ValueError):
                release.policy(repository, mode, ENVIRONMENT, metadata)
        release.policy("yizha1/notation", "rehearse", ENVIRONMENT, {"fork": True})
        release.policy(REPO, "dry-run", {}, {})

    def test_plan_identity_namespace_and_signature_sha_guards(self):
        release.validate_plan(plan())
        for name, value in [
            ("repository", "other/product"), ("schema", 2), ("baseline", "v1.3.0-trial"),
            ("branch", "main"), ("main", "release-1.3"), ("tag", "v1.3.2"),
            ("commit", "a\ninjected"), ("previous_head", "bad"), ("status", "unknown"),
        ]:
            with self.subTest(name=name), self.assertRaises((ValueError, KeyError)):
                release.validate_plan({**plan(), name: value})
        malformed = body(plan()).replace('"schema": 1', '"schema": 1, "schema": 1')
        with self.assertRaisesRegex(ValueError, "Duplicate"):
            release.decode_state(malformed, REPO, "2026-10", "execute")
        with self.assertRaises(ValueError):
            release.decode_state(body(plan()) + "\n```json\n{}\n```", REPO, "2026-10", "execute")

    def test_untrusted_or_duplicate_cycle_issue_cannot_control_releases(self):
        api = FakeGitHub()
        api.issues = [{"body": body(plan()), "user": {"login": "attacker"}}]
        self.assertIsNone(release.Cycle(api, REPO, "2026-10", "execute").state)
        api.issues[0]["user"]["login"] = "test-actor"
        api.issues *= 2
        with self.assertRaisesRegex(ValueError, "Duplicate"):
            release.Cycle(api, REPO, "2026-10", "execute")
        self.assertFalse(api.writes)

    def test_unverified_older_month_is_resumed_not_accepted_as_new_baseline(self):
        api = FakeGitHub()
        api.issues = [{"body": body(plan(month="2026-09", status="verifying")), "user": {"login": "test-actor"}, "state": "open"}]
        self.assertEqual(release.resume_cycle(api, REPO, "2026-10", "execute"), ("2026-09", "monthly"))
        api.issues.append({"body": body(plan()), "user": {"login": "test-actor"}})
        with self.assertRaisesRegex(ValueError, "Multiple active"):
            release.resume_cycle(api, REPO, "2026-10", "execute")

    def test_only_actual_dependabot_targets_qualify(self):
        self.assertTrue(release.dependabot(pull(), "main"))
        for item in [
            {**pull(), "user": {"login": "dependabot[bot]", "type": "User"}},
            {**pull(), "user": {"login": "person", "type": "Bot"}},
            pull(base="feature"),
        ]:
            self.assertFalse(release.dependabot(item, "main"))

    def test_checks_reject_changed_head_pending_failed_missing_and_missing_review(self):
        self.assertEqual(release.checks_ready(pull(), checks()), (True, ""))
        for key, value in [
            ("headRefOid", PREVIOUS), ("mergeable", "UNKNOWN"),
            ("mergeStateStatus", "BLOCKED"), ("reviewDecision", "REVIEW_REQUIRED"),
            ("statusCheckRollup", []),
            ("statusCheckRollup", [{"status": "IN_PROGRESS", "conclusion": None}]),
            ("statusCheckRollup", [{"status": "COMPLETED", "conclusion": "FAILURE"}]),
            ("statusCheckRollup", [{"status": "COMPLETED", "conclusion": "SKIPPED"}]),
            ("statusCheckRollup", [{"__typename": "StatusContext", "state": "PENDING"}]),
        ]:
            with self.subTest(key=key, value=value):
                self.assertFalse(release.checks_ready(pull(), {**checks(), key: value})[0])
        self.assertFalse(release.checks_ready({**pull(), "draft": True}, checks())[0])

    def test_dependency_only_file_allowlist_includes_rename_source(self):
        for path in ("go.mod", "go.sum", "test/e2e/go.mod", "test/e2e/plugin/go.sum", ".github/workflows/build.yaml"):
            self.assertTrue(release.dependency_file(path))
        for path in ("cmd/notation/sign.go", "README.md", ".github/scripts/monthly_release.py", ".github/workflows/nested/build.yml"):
            self.assertFalse(release.dependency_file(path))
        api = FakeGitHub()
        api.files = [{"filename": "go.mod", "previous_filename": "sensitive.go"}]
        with self.assertRaises(ValueError):
            release.check_files(api, REPO, 1)

    def test_canonical_manifests_never_accept_trial_replacements(self):
        manifest = {"Require": [{"Path": "github.com/notaryproject/notation-core-go", "Version": "v1.3.1"}]}
        module = manifest["Require"][0]["Path"]
        self.assertEqual(release.requirement(manifest, module, REPO, "execute"), "v1.3.1")
        replacement = {"Old": {"Path": module}, "New": {"Path": "github.com/yizha1/notation-core-go", "Version": "v1.3.4-monthly-test.202610"}}
        manifest["Replace"] = [replacement]
        with self.assertRaises(ValueError):
            release.requirement(manifest, module, REPO, "execute")
        self.assertEqual(release.requirement(manifest, module, "yizha1/notation-core-go", "rehearse"), replacement["New"]["Version"])
        with self.assertRaises(ValueError):
            release.requirement(manifest, module, "wrong/notation-core-go", "rehearse")
        with self.assertRaisesRegex(ValueError, "explicit dependency"):
            release.requirement({"Require": []}, module, REPO, "execute")

    def test_inventory_orders_merges_chronologically_not_by_pr_number(self):
        api = FakeGitHub()
        api.closed_pulls = [
            {**pull(1), "merged_at": "2026-09-03T00:00:00Z", "merge_commit_sha": COMMIT},
            {**pull(2), "merged_at": "2026-09-02T00:00:00Z", "merge_commit_sha": PREVIOUS},
            {**pull(3), "merged_at": None},
        ]
        self.assertEqual([item["number"] for item in release.inventory(api, REPO, "main")], [2, 1])

    def test_reassessment_retains_recorded_producer_merge_without_reauthorizing_old_open_prs(self):
        api = FakeGitHub()
        producer = {**pull(14), "user": {"login": "test-actor"}, "merged_at": "2026-10-01T00:00:00Z",
                    "merge_commit_sha": COMMIT}
        api.closed_pulls = [producer]
        authorization = {"snapshot": {"pulls": [], "merged": []}}
        recorded = [{"number": 14, "commit": COMMIT}]
        with patch("notation_release_controller.fork_propagation_pull", return_value=False):
            self.assertEqual(release.inventory(api, REPO, "main", authorization, recorded),
                             [{"number": 14, "commit": COMMIT, "merged_at": producer["merged_at"]}])
            self.assertEqual(release.inventory(api, REPO, "main", authorization), [])
            api.closed_pulls = [{**producer, "merged_at": None}]
            self.assertEqual(release.inventory(api, REPO, "main", authorization, recorded), [])
            api.closed_pulls = [{**producer, "merge_commit_sha": PREVIOUS}]
            with self.assertRaisesRegex(ValueError, "Worker-approved dependency merge changed"):
                release.inventory(api, REPO, "main", authorization, recorded)

    def test_dry_run_existing_ready_tag_recovery_is_strictly_read_only(self):
        api = FakeGitHub()
        api.issues = [{"body": body(plan()), "user": {"login": "test-actor"}}]
        with tempfile.TemporaryDirectory() as temporary:
            result = release.prepare(api, REPO, "dry-run", "2026-10", pathlib.Path(temporary), temporary, True)
        self.assertFalse(result["writes"])
        self.assertEqual(result["action"], "preview")
        self.assertFalse(api.writes)

    def test_dry_run_pr_preview_never_merges(self):
        api = FakeGitHub()
        api.open_pulls = [pull()]
        with tempfile.TemporaryDirectory() as temporary:
            result = release.prepare(api, REPO, "dry-run", "2026-10", pathlib.Path(temporary), temporary, True)
        self.assertEqual(result["pulls"], [{"number": 1, "ready": True, "reason": ""}])
        self.assertFalse(api.writes)

    def test_blocked_earlier_pr_does_not_prevent_a_checked_cve_fix_merge(self):
        for failure in ("failed", "pending", "missing"):
            with self.subTest(failure=failure):
                api = FakeGitHub()
                head = "c" * 40
                api.open_pulls = [pull(3), pull(6, sha=head)]
                blocked = checks()
                if failure == "failed":
                    blocked["statusCheckRollup"][0]["conclusion"] = "FAILURE"
                elif failure == "pending":
                    blocked["statusCheckRollup"][0]["status"] = "IN_PROGRESS"
                else:
                    blocked["statusCheckRollup"] = []
                with patch.object(api, "checks", side_effect=lambda _, number: blocked if number == 3 else checks(head)), \
                     patch.object(release, "signing_environment", return_value=dict(os.environ)), \
                     patch.object(release, "git") as candidate_git:
                    result = release.prepare(api, REPO, "execute", "2026-10", ".", ".", True)
                self.assertEqual(result["action"], "waiting")
                self.assertIn("Dependabot #3:", result["reason"])
                self.assertEqual(result["merged"], [{"number": 6, "commit": COMMIT}])
                merges = [(path, payload) for path, method, payload in api.writes if method == "PUT"]
                self.assertEqual(merges, [(f"repos/{REPO}/pulls/6/merge", {"sha": head, "merge_method": "squash"})])
                candidate_git.assert_not_called()
                api.open_pulls = [pull(3)]
                api.check_result = blocked
                with patch.object(release, "signing_environment", return_value=dict(os.environ)):
                    resumed = release.prepare(api, REPO, "execute", "2026-10", ".", ".", False, True)
                self.assertEqual(resumed["action"], "waiting")
                self.assertEqual(resumed["merged"], result["merged"])
                self.assertEqual(sum(method == "PUT" for _, method, _ in api.writes), 1)
                self.assertEqual(len(api.issues), 1)

    def test_all_blocked_prs_are_reported_without_merging_or_backporting(self):
        api = FakeGitHub()
        api.open_pulls = [pull(3), pull(6)]
        api.check_result["statusCheckRollup"][0]["conclusion"] = "FAILURE"
        with patch.object(release, "signing_environment", return_value=dict(os.environ)), \
             patch.object(release, "git") as candidate_git:
            result = release.prepare(api, REPO, "execute", "2026-10", ".", ".", True)
        self.assertEqual(result["action"], "waiting")
        self.assertEqual(result["reason"], "Dependabot #3: A check failed; Dependabot #6: A check failed")
        self.assertFalse(result["merged"])
        self.assertFalse(any(method == "PUT" for _, method, _ in api.writes))
        candidate_git.assert_not_called()

    def test_changed_check_head_never_qualifies_a_later_merge(self):
        api = FakeGitHub()
        api.open_pulls = [pull(3), pull(6, sha="c" * 40)]
        api.check_result["statusCheckRollup"][0]["conclusion"] = "FAILURE"
        with patch.object(release, "signing_environment", return_value=dict(os.environ)):
            result = release.prepare(api, REPO, "execute", "2026-10", ".", ".", True)
        self.assertEqual(result["action"], "waiting")
        self.assertIn("Dependabot #6: Draft or changed pull-request head", result["reason"])
        self.assertFalse(any(method == "PUT" for _, method, _ in api.writes))

    def test_automatic_resumption_does_not_start_an_unscheduled_cycle(self):
        api = FakeGitHub()
        result = release.prepare(api, REPO, "execute", "2026-10", ".", ".", False)
        self.assertEqual(result["action"], "idle")
        self.assertFalse(api.writes)

    def test_empty_cycle_is_closed_as_skipped(self):
        api = FakeGitHub()
        with patch.object(release, "signing_environment", return_value=dict(os.environ)):
            result = release.prepare(api, REPO, "execute", "2026-10", ".", ".", True)
        self.assertEqual(result["status"], "skipped")
        self.assertEqual(api.issues[0]["state"], "closed")
        self.assertFalse(any(method == "PUT" for _, method, _ in api.writes))

    def test_unfinished_producer_without_new_public_release_does_not_block_consumer(self):
        for status in ("waiting", "failed", "verifying"):
            api = FakeGitHub()
            api.issues = [{"body": body(plan(status=status)), "user": {"login": "test-actor"}}]
            ready = release.producer_versions(api, "notaryproject/notation-go", "execute")
            self.assertEqual(ready["notation-core-go"]["version"], "v1.3.0")

    def test_consumer_does_not_require_any_same_month_producer_cycle(self):
        api = FakeGitHub()
        ready = release.producer_versions(api, "notaryproject/notation-go", "execute")
        self.assertEqual(ready["notation-core-go"]["version"], "v1.3.0")
        with patch.object(release, "signing_environment", return_value=dict(os.environ)):
            result = release.prepare(api, "notaryproject/notation-go", "execute", "2026-10", ".", ".", True)
        self.assertEqual(result["status"], "skipped")

    def test_unverified_automated_public_release_is_not_consumed_yet(self):
        api = FakeGitHub()
        state = plan(status="verifying")
        api.issues = [{"body": body(state), "user": {"login": "test-actor"}}]
        api.releases.append({"tag_name": "v1.3.1", "draft": False, "prerelease": False, "body": body(state)})
        self.assertEqual(release.producer_versions(api, "notaryproject/notation-go", "execute")["notation-core-go"]["version"], "v1.3.0")
        state["status"] = "published"
        api.issues[0]["body"] = body(state)
        with patch.object(release, "assert_tag", return_value=COMMIT):
            self.assertEqual(release.producer_versions(api, "notaryproject/notation-go", "execute")["notation-core-go"]["version"], "v1.3.1")

    def test_missing_new_core_dependabot_pr_waits_without_publishing_old_version(self):
        api = FakeGitHub()
        api.releases_by_repo[REPO] = [{"tag_name": "v1.3.1", "draft": False, "prerelease": False}]
        with patch.object(release, "signing_environment", return_value=dict(os.environ)):
            result = release.prepare(api, "notaryproject/notation-go", "execute", "2026-10", ".", ".", True)
        self.assertEqual(result["status"], "waiting")
        self.assertIn("Dependabot", result["reason"])
        self.assertIn("v1.3.1", result["reason"])
        self.assertFalse(any(path.endswith("/merge") for path, _, _ in api.writes))

    def test_late_core_release_opens_distinct_cycle_after_consumers_monthly_release(self):
        api = FakeGitHub()
        consumer = "notaryproject/notation-go"
        done = plan(repository=consumer, status="published")
        done["notified"] = ["notaryproject/notation"]
        api.issues = [{"number": 1, "body": body(done), "user": {"login": "test-actor"}, "state": "closed"}]
        api.releases_by_repo[consumer] = [{"tag_name": "v1.3.1", "published_at": "2026-10-01T09:00:00Z", "draft": False, "prerelease": False}]
        api.releases_by_repo[REPO] = [{"tag_name": "v1.3.1", "draft": False, "prerelease": False}]
        with patch.object(release, "signing_environment", return_value=dict(os.environ)):
            result = release.prepare(api, consumer, "execute", "2026-10", ".", ".", False, True)
            repeated = release.prepare(api, consumer, "execute", "2026-10", ".", ".", False, True)
        self.assertEqual(result["tag"], "v1.3.2")
        self.assertEqual(result["status"], "waiting")
        self.assertRegex(result["cycle_id"], r"^upstream-[0-9a-f]{64}$")
        self.assertEqual(repeated["cycle_id"], result["cycle_id"])
        self.assertEqual(len(api.issues), 2)
        self.assertEqual(release.Cycle(api, consumer, "2026-10", "execute").state["status"], "published")
        self.assertEqual(release.resume_cycle(api, consumer, "2026-11", "execute"), ("2026-10", result["cycle_id"]))
        self.assertIsNone(release.decode_state(body(result), consumer, "2026-10", "execute"))
        with self.assertRaises(ValueError):
            release.validate_plan({**result, "cycle_id": "untrusted"})

    def test_no_new_upstream_release_does_not_open_an_extra_consumer_cycle(self):
        api = FakeGitHub()
        consumer = "notaryproject/notation-go"
        done = plan(repository=consumer, status="published")
        done["notified"] = ["notaryproject/notation"]
        api.issues = [{"number": 1, "body": body(done), "user": {"login": "test-actor"}, "state": "closed"}]
        result = release.prepare(api, consumer, "execute", "2026-10", ".", ".", False, True)
        self.assertEqual(result["action"], "complete")
        self.assertFalse(api.writes)
        self.assertTrue(release.includes_version("v1.3.2", "v1.3.1"))
        self.assertFalse(release.includes_version("v1.3.0", "v1.3.1"))

    def test_cli_producer_snapshot_is_independent_of_both_libraries_monthly_status(self):
        api = FakeGitHub()
        versions = release.producer_versions(api, "notaryproject/notation", "execute")
        self.assertEqual(set(versions), {"notation-core-go", "notation-go"})
        self.assertEqual(release.producer_lag(api, "notaryproject/notation", "main", "execute", versions), [])
        api.releases_by_repo[REPO] = [{"tag_name": "v1.3.1", "draft": False, "prerelease": False}]
        versions = release.producer_versions(api, "notaryproject/notation", "execute")
        self.assertEqual(versions["notation-go"]["version"], "v1.3.0")
        self.assertIn("notation-core-go", release.producer_lag(api, "notaryproject/notation", "main", "execute", versions)[0])
        self.assertNotEqual(release.upstream_cycle_id(versions), release.upstream_cycle_id({**versions, "notation-go": {"repository": "notaryproject/notation-go", "version": "v1.3.1"}}))

    def test_dispatch_fanout_duplicates_and_fork_scope(self):
        api = FakeGitHub()
        progress = list(release.dispatch_consumers(api, REPO, "execute", "v1.3.1"))
        self.assertEqual(progress[-1], ["notaryproject/notation-go", "notaryproject/notation"])
        self.assertEqual([path for path, _, _ in api.writes], [
            "repos/notaryproject/notation-go/dispatches", "repos/notaryproject/notation/dispatches"])
        self.assertEqual(list(release.dispatch_consumers(api, REPO, "execute", "v1.3.1", progress[-1])), [])
        fork = FakeGitHub()
        list(release.dispatch_consumers(fork, "yizha1/notation-core-go", "rehearse", "v1.3.4-monthly-test.202610"))
        self.assertTrue(all(path.startswith("repos/yizha1/") for path, _, _ in fork.writes))
        with self.assertRaises(ValueError):
            list(release.dispatch_consumers(api, REPO, "execute", "v1.3.1", ["untrusted/notation"]))
        with patch.object(api, "request", side_effect=release.APIError("HTTP 403")), self.assertRaises(release.APIError):
            list(release.dispatch_consumers(api, REPO, "execute", "v1.3.1"))

    def test_notification_requires_completed_verified_cycle_and_records_progress(self):
        api = FakeGitHub()
        state = plan(status="published")
        api.issues = [{"number": 1, "body": body(state), "user": {"login": "test-actor"}, "state": "closed"}]
        with patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"):
            notified = release.notify(api, state)
            release.notify(api, state)
        self.assertEqual(notified["notified"], ["notaryproject/notation-go", "notaryproject/notation"])
        self.assertEqual(sum(path.endswith("/dispatches") for path, _, _ in api.writes), 2)
        api.issues[0]["body"] = body(plan(status="verifying"))
        with self.assertRaisesRegex(ValueError, "successfully verified"):
            release.notify(api, state)

    def test_partial_followup_notification_failure_is_resumed_without_duplicate_dispatches(self):
        api = FakeGitHub()
        state = plan(status="published", month="2026-09")
        state["cycle_id"] = release.upstream_cycle_id({"producer": {"repository": REPO, "version": "v1.3.1"}})
        api.issues = [{"number": 1, "body": body(state), "user": {"login": "test-actor"}, "state": "closed"}]
        request = api.request
        def fail_second_dispatch(path, method="GET", payload=None):
            if path == "repos/notaryproject/notation/dispatches":
                raise release.APIError("HTTP 503")
            return request(path, method, payload)
        with patch.object(api, "request", side_effect=fail_second_dispatch), patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"), self.assertRaises(release.APIError):
            release.notify(api, state)
        stored = release.Cycle(api, REPO, "2026-09", "execute", state["cycle_id"]).state
        self.assertEqual(stored["notified"], ["notaryproject/notation-go"])
        resumed = release.prepare(api, REPO, "execute", "2026-10", ".", ".", False, True)
        self.assertEqual(resumed["cycle_id"], state["cycle_id"])
        self.assertEqual(resumed["month"], "2026-09")
        with patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"):
            release.notify(api, resumed)
        paths = [path for path, _, _ in api.writes if path.endswith("/dispatches")]
        self.assertEqual(paths, ["repos/notaryproject/notation-go/dispatches", "repos/notaryproject/notation/dispatches"])

    def test_waiting_cycle_refreshes_targets_and_never_consumes_unverified_newer_version(self):
        api = FakeGitHub()
        consumer = "notaryproject/notation-go"
        state = plan(repository=consumer, status="waiting")
        state["producers"] = {"notation-core-go": {"repository": REPO, "version": "v1.3.1"}}
        api.issues = [{"number": 1, "body": body(state), "user": {"login": "test-actor"}, "state": "open"}]
        api.releases_by_repo[REPO] = [{"tag_name": "v1.3.2", "draft": False, "prerelease": False}]
        with patch.object(release, "signing_environment", return_value=dict(os.environ)):
            result = release.prepare(api, consumer, "execute", "2026-10", ".", ".", False, True)
        self.assertEqual(result["producers"]["notation-core-go"]["version"], "v1.3.2")
        self.assertIn("v1.3.2", result["reason"])
        api.manifests[(consumer, "main", ".")] = {"Require": [{"Path": "github.com/notaryproject/notation-core-go", "Version": "v1.3.3"}]}
        self.assertTrue(release.producer_lag(api, consumer, "main", "execute", result["producers"]))

    def test_ordinary_stable_release_announces_but_automated_release_waits_for_verification(self):
        api = FakeGitHub()
        api.releases[0].update(tag_name="v1.3.1", body="")
        with patch.dict(os.environ, {"GITHUB_REF": "refs/tags/v1.3.1"}):
            result = release.announce(api, REPO, "execute", "v1.3.1")
            self.assertEqual(len(result["notified"]), 2)
            api.writes.clear()
            api.releases[0]["body"] = body(plan(status="verifying"))
            self.assertEqual(release.announce(api, REPO, "execute", "v1.3.1")["action"], "idle")
            self.assertFalse(api.writes)
        with self.assertRaises(ValueError):
            release.announce(api, REPO, "execute", "v1.3.1")

    def test_baseline_drift_cannot_increment_a_stale_release(self):
        api = FakeGitHub()
        api.releases[0]["tag_name"] = "v1.4.0"
        with self.assertRaisesRegex(ValueError, "advanced"):
            release.baseline_plan(api, REPO, "2026-10", "execute", plan())

    def test_existing_cli_tag_publisher_must_exclude_monthly_actor(self):
        api = FakeGitHub()
        old = "    if: github.repository == 'notaryproject/notation'"
        for content, guarded in [(old, False), (old + " && github.actor != vars.MONTHLY_PATCH_ACTOR", True)]:
            workflow = {"encoding": "base64", "content": base64.b64encode(content.encode()).decode()}
            with patch.object(api, "optional", return_value=workflow):
                if guarded:
                    release.check_publisher_guard(api, "notaryproject/notation", "release-1.3")
                else:
                    with self.assertRaisesRegex(ValueError, "automation-actor guard"):
                        release.check_publisher_guard(api, "notaryproject/notation", "release-1.3")
        with patch.object(api, "optional", return_value=None):
            release.check_publisher_guard(api, "notaryproject/notation", "release-1.3")
        self.assertFalse(api.writes)

    def test_complete_is_idempotent_and_has_no_new_writes(self):
        api = FakeGitHub()
        api.issues = [{"body": body(plan(status="published")), "user": {"login": "test-actor"}, "state": "closed"}]
        self.assertEqual(release.prepare(api, REPO, "execute", "2026-10", ".", ".", True)["action"], "complete")
        self.assertFalse(api.writes)

    def test_tag_requires_annotated_signature_and_exact_commit(self):
        api = FakeGitHub()
        with patch.object(api, "request", side_effect=[
            {"object": {"type": "tag", "sha": PREVIOUS}},
            {"object": {"sha": COMMIT}, "verification": {"verified": True}},
        ]):
            self.assertEqual(release.verified_tag(api, REPO, "v1.3.1", COMMIT), PREVIOUS)
        for annotated in [
            {"object": {"sha": PREVIOUS}, "verification": {"verified": True}},
            {"object": {"sha": COMMIT}, "verification": {"verified": False}},
        ]:
            with patch.object(api, "request", side_effect=[{"object": {"type": "tag", "sha": PREVIOUS}}, annotated]), self.assertRaises(ValueError):
                release.verified_tag(api, REPO, "v1.3.1", COMMIT)

    def test_asset_recovery_rejects_changed_digest_size_visibility_and_extra_assets(self):
        api = FakeGitHub()
        candidate = {**plan(status="verifying"), "assets": [
            {"name": name, "size": 10, "sha256": "a" * 64}
            for name in ("notation-core-go_1.3.1_source.tar.gz", "notation-core-go_1.3.1_checksums.txt")
        ]}
        public = {"id": 1, "draft": False, "prerelease": False, "body": body(candidate)}
        assets = [{"name": item["name"], "size": 10, "digest": "sha256:" + "a" * 64} for item in candidate["assets"]]
        asset = assets[0]
        with patch.object(api, "request", return_value=public), patch.object(api, "pages", return_value=iter(assets)):
            release.check_public_assets(api, candidate)
        for changed in [[{**asset, "size": 11}, assets[1]], [{**asset, "digest": "sha256:changed"}, assets[1]], [asset, asset]]:
            with patch.object(api, "request", return_value=public), patch.object(api, "pages", return_value=iter(changed)), self.assertRaises(ValueError):
                release.check_public_assets(api, candidate)
        with patch.object(api, "request", return_value={**public, "draft": True}), self.assertRaises(ValueError):
            release.check_public_assets(api, candidate)

    def test_candidate_bundle_round_trip_without_remote_mutation(self):
        self.candidate_bundle_round_trip(REPO)

    def test_late_core_release_builds_a_second_consumer_patch_with_signed_backport(self):
        self.candidate_bundle_round_trip("notaryproject/notation-go", late_release=True)

    def candidate_bundle_round_trip(self, repository, late_release=False):
        with tempfile.TemporaryDirectory(prefix="monthly-engine-test-") as temporary:
            root = pathlib.Path(temporary)
            origin, checkout, output = root / "origin.git", root / "checkout", root / "output"
            subprocess.run(["git", "init", "--bare", "--quiet", str(origin)], check=True)
            subprocess.run(["git", "clone", "--quiet", str(origin), str(checkout)], check=True)
            for key, value in (("user.name", "Test"), ("user.email", "test@example.com"), ("commit.gpgsign", "false")):
                release.git(["config", key, value], checkout)
            before = "\nrequire github.com/notaryproject/notation-core-go v1.3.0\n" if late_release else ""
            after = "\nrequire github.com/notaryproject/notation-core-go v1.3.1\n" if late_release else ""
            (checkout / "go.mod").write_text("module example.com/test\n\ngo 1.26.0\n" + before)
            release.git(["add", "go.mod"], checkout)
            release.git(["commit", "-m", "Baseline"], checkout)
            previous = release.git(["rev-parse", "HEAD"], checkout)
            release.git(["branch", "-M", "main"], checkout)
            release.git(["branch", "release-1.3"], checkout)
            (checkout / "go.mod").write_text("module example.com/test\n\ngo 1.26.1\n" + after)
            release.git(["add", "go.mod"], checkout)
            release.git(["commit", "-m", "Dependency update"], checkout)
            source_commit = release.git(["rev-parse", "HEAD"], checkout)
            release.git(["push", "origin", "main", "release-1.3"], checkout)
            api = FakeGitHub()
            api.head = previous
            api.closed_pulls = [{**pull(), "merged_at": "2025-12-01T00:00:00Z", "merge_commit_sha": source_commit}]
            if late_release:
                done = plan(repository=repository, status="published")
                done["notified"] = ["notaryproject/notation"]
                api.issues = [{"number": 1, "body": body(done), "user": {"login": "test-actor"}, "state": "closed"}]
                api.releases_by_repo[repository] = [{"tag_name": "v1.3.1", "published_at": "2026-10-01T09:00:00Z", "draft": False, "prerelease": False}]
                api.releases_by_repo[REPO] = [{"tag_name": "v1.3.1", "draft": False, "prerelease": False}]
                api.manifests[(repository, "main", ".")] = {"Require": [{"Path": "github.com/notaryproject/notation-core-go", "Version": "v1.3.1"}]}
            output.mkdir()
            key = root / "test-key"
            subprocess.run(["ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(key)], check=True)
            api.public_key = key.with_suffix(".pub").read_text()
            with patch.dict(os.environ, {"MONTHLY_PATCH_SIGNING_KEY": key.read_text()}):
                candidate = release.prepare(api, repository, "execute", "2026-10", checkout, output, not late_release, late_release)
            self.assertEqual(candidate["action"], "qualify")
            self.assertEqual(candidate["minimum_go"], "1.26")
            self.assertNotEqual(candidate["commit"], source_commit)
            self.assertEqual(release.git([f"--git-dir={origin}", "rev-parse", "refs/heads/release-1.3"], checkout), previous)
            self.assertIn(source_commit, release.git(["log", "-1", "--format=%B", candidate["commit"]], checkout))
            if late_release:
                self.assertEqual(candidate["tag"], "v1.3.2")
                self.assertTrue(candidate["cycle_id"].startswith("upstream-"))
                self.assertIn("github.com/notaryproject/notation-core-go v1.3.1", release.git(["show", f"{candidate['commit']}:go.mod"], checkout))
                self.assertEqual(len(api.issues), 2)
            release.import_candidate(checkout, candidate, output / "candidate.bundle")
            self.assertEqual(release.git(["rev-parse", "HEAD"], checkout), candidate["commit"])
            with self.assertRaises(ValueError):
                release.import_candidate(checkout, {**candidate, "commit": COMMIT}, output / "candidate.bundle")
            if late_release:
                candidate.update(status="published", notified=["notaryproject/notation"])
                release.Cycle(api, repository, "2026-10", "execute", candidate["cycle_id"]).save(candidate)
                api.releases_by_repo[repository].append({"tag_name": "v1.3.2", "published_at": "2026-10-04T09:00:00Z", "draft": False, "prerelease": False})
                api.manifests[(repository, "v1.3.2", ".")] = api.manifests[(repository, "main", ".")]
                repeated = release.prepare(api, repository, "execute", "2026-10", checkout, output, False, True)
                self.assertEqual(repeated["action"], "complete")
                self.assertEqual(len(api.issues), 2)

    def test_signing_key_is_private_registered_and_removed_with_temporary_directory(self):
        api = FakeGitHub()
        with tempfile.TemporaryDirectory(prefix="monthly-signing-test-") as temporary:
            path = pathlib.Path(temporary)
            key = path / "test-key"
            subprocess.run(["ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", str(key)], check=True)
            api.public_key = key.with_suffix(".pub").read_text()
            secret = key.read_text()
            location = path / "signing"
            location.mkdir()
            with patch.dict(os.environ, {"MONTHLY_PATCH_SIGNING_KEY": secret, "MONTHLY_PATCH_SIGNER_NAME": ""}):
                environment = release.signing_environment(location, api)
            self.assertEqual((location / "signing-key").stat().st_mode & 0o777, 0o600)
            self.assertIn("ssh", environment.values())
            self.assertEqual(environment["GIT_CONFIG_VALUE_0"], "Monthly dependency release")
            identity = subprocess.run(["git", "var", "GIT_COMMITTER_IDENT"], cwd=path,
                                      env=environment, check=True, capture_output=True, text=True).stdout
            self.assertTrue(identity.startswith("Monthly dependency release <test@example.com>"))
            named = path / "named"
            named.mkdir()
            with patch.dict(os.environ, {"MONTHLY_PATCH_SIGNING_KEY": secret,
                                        "MONTHLY_PATCH_SIGNER_NAME": "Rehearsal signer"}):
                self.assertEqual(release.signing_environment(named, api)["GIT_CONFIG_VALUE_0"], "Rehearsal signer")
            mismatch = path / "unregistered"
            mismatch.mkdir()
            api.public_key = "ssh-ed25519 different"
            with patch.dict(os.environ, {"MONTHLY_PATCH_SIGNING_KEY": secret}), self.assertRaisesRegex(ValueError, "not registered"):
                release.signing_environment(mismatch, api)
        self.assertFalse(path.exists())

    def test_tag_push_uses_atomic_exact_leases_without_follow_tags(self):
        api = FakeGitHub()
        candidate = plan()
        api.issues = [{"number": 1, "body": body(candidate), "user": {"login": "test-actor"}}]
        with patch.object(release, "signing_environment", return_value=dict(os.environ)), patch.object(release, "verified_tag", return_value=PREVIOUS), patch.object(release, "git") as command:
            command.side_effect = [COMMIT, "", COMMIT, ""]
            result = release.tag_candidate(api, candidate, ".")
        args = command.call_args.args[0]
        self.assertIn("--atomic", args)
        self.assertIn("push.followTags=false", args)
        self.assertIn(f"--force-with-lease=refs/heads/release-1.3:{PREVIOUS}", args)
        self.assertIn("--force-with-lease=refs/tags/v1.3.1:", args)
        self.assertEqual(result["status"], "tagged")

    def test_tag_push_rechecks_stable_baseline_after_qualification(self):
        api = FakeGitHub()
        api.issues = [{"body": body(plan()), "user": {"login": "test-actor"}}]
        api.releases[0]["tag_name"] = "v1.4.0"
        with patch.object(release, "git") as command, self.assertRaisesRegex(ValueError, "advanced"):
            release.tag_candidate(api, plan(), ".")
        command.assert_not_called()


if __name__ == "__main__":
    unittest.main()
