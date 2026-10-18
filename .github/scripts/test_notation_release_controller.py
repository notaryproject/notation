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
import copy
import base64
import datetime
import json
import os
import pathlib
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import monthly_release as release
import notation_release_controller as coordinator
from test_upstream_authorization import UpstreamAPI as ControllerAPI, inventory as canonical_inventory
from test_upstream_authorization import ApprovedController
from test_monthly_release import FakeGitHub, plan as worker_plan, body, pull
from test_monthly_release_checks import scanner_messages, stream

CORE, GO, CLI = coordinator.REPOSITORIES
MAIN, BRANCH, AUTOMATION, PR_HEAD = ("a" * 40, "b" * 40, "a" * 40, "c" * 40)
MONTH = "2026-10"

def inventory(changes=()):
    return canonical_inventory(changes)



def assessment(evidence, decisions=("skip", "skip", "skip")):
    return {
        "schema": 1, "month": evidence["month"], "inventory_sha256": coordinator.digest(evidence),
        "decisions": {repository: {"decision": decision, "reason": "Evidence-based decision",
                                  "evidence": [f"{repository}:dependencies"]}
                      for repository, decision in zip(coordinator.REPOSITORIES, decisions)},
    }




class ReleaseAssessmentTests(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ, {"MONTHLY_PATCH_ACTOR": "test-actor"})
        self.environment.start()
        self.addCleanup(self.environment.stop)

    def test_core_patch_includes_both_consumers(self):
        evidence = inventory((CORE,))
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        self.assertEqual(set(plan["decisions"]), set(coordinator.REPOSITORIES))
        self.assertNotIn("body", plan["snapshots"][CORE]["pulls"][0])

    def test_go_only_and_cli_only_do_not_invent_core_patch(self):
        for changes, decisions in (((GO,), ("skip", "release", "release")), ((CLI,), ("skip", "skip", "release"))):
            with self.subTest(changes=changes):
                evidence = inventory(changes)
                plan = coordinator.validate_assessment(evidence, assessment(evidence, decisions))
                self.assertEqual(plan["decisions"][CORE]["decision"], "skip")

    def test_no_changes_yields_three_skips(self):
        evidence = inventory()
        plan = coordinator.validate_assessment(evidence, assessment(evidence))
        self.assertEqual({item["decision"] for item in plan["decisions"].values()}, {"skip"})

    def test_downstream_cannot_skip_planned_or_deferred_producer(self):
        for decisions in (("release", "skip", "release"), ("release", "release", "skip"), ("defer", "skip", "skip")):
            evidence = inventory((CORE,))
            with self.subTest(decisions=decisions), self.assertRaisesRegex(ValueError, "downstream"):
                coordinator.validate_assessment(evidence, assessment(evidence, decisions))

    def test_missing_setup_is_deferred_not_skipped_or_released(self):
        evidence = inventory()
        evidence["repositories"][CORE]["blockers"] = ["Missing reviewed branch"]
        for decisions in (("skip", "skip", "skip"), ("release", "release", "release")):
            with self.subTest(decisions=decisions), self.assertRaises(ValueError):
                coordinator.validate_assessment(evidence, assessment(evidence, decisions))
        coordinator.validate_assessment(evidence, assessment(evidence, ("defer", "skip", "skip")))

    def test_known_cve_requires_fix_or_deferral(self):
        evidence = inventory()
        evidence["repositories"][CORE]["scans"]["."]["reachable"] = ["GO-2026-1234"]
        with self.assertRaises(ValueError):
            coordinator.validate_assessment(evidence, assessment(evidence))
        with self.assertRaisesRegex(ValueError, "empty patch"):
            coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        coordinator.validate_assessment(evidence, assessment(evidence, ("defer", "defer", "defer")))

    def test_informational_findings_do_not_force_empty_patch(self):
        evidence = inventory()
        evidence["repositories"][CORE]["scans"]["."]["informational"] = ["GO-2026-1234"]
        coordinator.validate_assessment(evidence, assessment(evidence))

    def test_extra_repo_fields_and_fabricated_evidence_are_rejected(self):
        evidence = inventory()
        valid = assessment(evidence)
        invalids = [
            {**valid, "command": "publish"},
            {**valid, "schema": True},
            {**valid, "inventory_sha256": "0" * 64},
            {**valid, "month": "2026-11"},
            {**valid, "decisions": {**valid["decisions"], "untrusted/notation": valid["decisions"][CLI]}},
        ]
        for invalid in invalids:
            with self.subTest(invalid=invalid), self.assertRaises(ValueError):
                coordinator.validate_assessment(evidence, invalid)
        invalid = copy.deepcopy(valid)
        invalid["decisions"][CLI]["evidence"] = ["notaryproject/notation:pr:999"]
        with self.assertRaisesRegex(ValueError, "fabricated"):
            coordinator.validate_assessment(evidence, invalid)
        invalid["decisions"][CLI]["evidence"] = [f"{CORE}:baseline"]
        with self.assertRaisesRegex(ValueError, "own repository"):
            coordinator.validate_assessment(evidence, invalid)

    def test_incomplete_scan_or_branch_namespace_is_rejected(self):
        for field, value in (("scans", {}), ("branch", "main"), ("main_head", "bad"), ("source_repository", "untrusted/notation")):
            evidence = inventory()
            evidence["repositories"][CORE][field] = value
            with self.subTest(field=field), self.assertRaises(ValueError):
                coordinator.validate_assessment(evidence, assessment(evidence))

    def test_duplicate_json_fields_are_rejected(self):
        with self.assertRaisesRegex(ValueError, "Duplicate JSON"):
            coordinator.load_json('{"schema":1,"schema":1}')

    def test_source_scan_preserves_reachable_findings_as_evidence(self):
        scanned = subprocess.CompletedProcess([], 0, stream(scanner_messages("source")), "")
        with tempfile.TemporaryDirectory() as temporary, patch.object(release, "git", return_value=BRANCH), patch.object(
            coordinator.subprocess, "run", return_value=scanned
        ):
            result = coordinator.scan_source(CORE, BRANCH, (".",), pathlib.Path(temporary))
            self.assertEqual(result["."]["reachable"], ["GO-2026-1234"])

    def test_source_scan_errors_do_not_become_clean_evidence(self):
        for scanned in (subprocess.CompletedProcess([], 1, "", "network failed"),
                        subprocess.CompletedProcess([], 0, '{"config":{}}', "")):
            with tempfile.TemporaryDirectory() as temporary, patch.object(release, "git", return_value=BRANCH), patch.object(
                coordinator.subprocess, "run", return_value=scanned
            ), self.assertRaises(ValueError):
                coordinator.scan_source(CORE, BRANCH, (".",), pathlib.Path(temporary))

    def test_comparison_must_be_complete(self):
        api = ControllerAPI()
        with patch.object(api, "request", return_value={"status": "ahead", "ahead_by": 2, "commits": [{"sha": MAIN}]}), self.assertRaisesRegex(ValueError, "Incomplete"):
            coordinator.compare_commits(api, CORE, BRANCH, MAIN)

    def test_workflow_registration_does_not_prove_installation_on_main(self):
        api = ControllerAPI()
        self.assertFalse(coordinator.installed_worker(api, CORE, AUTOMATION))
        required = {
            coordinator.WORKER_PATH: f"jobs:\n  release:\n    uses: {coordinator.HOST}/{coordinator.SHARED_PATH}@{AUTOMATION}\n    with:\n      worker_sha: {AUTOMATION}\n",
            coordinator.SHARED_PATH: "  workflow_call:\nMONTHLY_PATCH_COORDINATOR_REQUIRED: 'true'\n",
            ".github/scripts/notation_release_controller.py": "def worker_authorization(",
            ".github/scripts/monthly_release.py": "coordinator_authorization(",
            ".github/scripts/monthly_release_pr.py": "def advance(\ndef validate(",
            ".github/scripts/monthly_release_vote.py": "def validate_vote(",
            ".github/scripts/qualify_monthly_candidate.py": "def qualify(",
            ".github/scripts/build_monthly_assets.py": "def ",
            ".github/scripts/test_monthly_public_release.py": "def ",
        }
        def file(path):
            self.assertTrue(path.endswith("?ref=" + AUTOMATION))
            name = path.split("/contents/", 1)[1].split("?")[0]
            if name != coordinator.WORKER_PATH:
                self.assertTrue(path.startswith("repos/" + coordinator.HOST + "/contents/"))
            return {"encoding": "base64", "content": base64.b64encode(required[name].encode()).decode()}
        with patch.object(api, "optional", side_effect=file):
            self.assertTrue(coordinator.installed_worker(api, CORE, AUTOMATION))
            required[".github/scripts/monthly_release.py"] = "legacy independent publisher"
            self.assertFalse(coordinator.installed_worker(api, CORE, AUTOMATION))


class ReleaseControllerTests(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ, {"MONTHLY_PATCH_ACTOR": "test-actor"})
        self.environment.start()
        self.addCleanup(self.environment.stop)
        self.api = ControllerAPI()
        self.controller = ApprovedController(self.api, "test-actor")

    def approved(self, changes=(CORE,), decisions=("release", "release", "release")):
        evidence = inventory(changes)
        plan = coordinator.validate_assessment(evidence, assessment(evidence, decisions))
        return self.controller.approve(plan, 22)

    def test_ordered_releases_are_dispatched_only_after_public_verification(self):
        issue, state = self.approved()
        issue, state = self.controller.advance(issue, state)
        self.assertEqual([repository for repository, _ in self.api.dispatched], [CORE])
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(len(self.api.dispatched), 1)
        self.api.worker(CORE, issue, state)
        with patch.object(release, "assert_tag") as signature, patch.object(release, "check_public_assets") as assets:
            issue, state = self.controller.advance(issue, state)
            signature.assert_called_once()
            assets.assert_called_once()
            self.assertEqual([repository for repository, _ in self.api.dispatched], [CORE, GO])
            self.api.worker(GO, issue, state)
            issue, state = self.controller.advance(issue, state)
            self.assertEqual([repository for repository, _ in self.api.dispatched], [CORE, GO, CLI])
            self.api.worker(CLI, issue, state)
            issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["status"], "completed")
        self.assertEqual(issue["state"], "closed")
        self.controller.advance(issue, state)
        self.assertEqual(len(self.api.dispatched), 3)

    def test_go_and_cli_only_skip_unnecessary_work(self):
        for changes, decisions, expected in (
            ((GO,), ("skip", "release", "release"), GO),
            ((CLI,), ("skip", "skip", "release"), CLI),
        ):
            self.api = ControllerAPI()
            self.controller = ApprovedController(self.api, "test-actor")
            issue, state = self.approved(changes, decisions)
            self.controller.advance(issue, state)
            self.assertEqual(self.api.dispatched[0][0], expected)

    def test_no_change_cycle_finishes_without_dispatch(self):
        issue, state = self.approved((), ("skip", "skip", "skip"))
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["status"], "completed")
        self.controller.advance(issue, state)
        self.assertEqual(self.api.dispatched, [])

    def test_deferred_producer_prevents_consumer_release(self):
        with self.assertRaisesRegex(ValueError, "Resolve deferrals"):
            self.approved((CORE,), ("defer", "release", "release"))
        self.assertEqual(self.api.dispatched, [])
        self.assertTrue(all("/issues" in path for path, _, _ in self.api.writes))

    def test_failed_worker_requires_explicit_retry_and_keeps_patch_version(self):
        issue, state = self.approved()
        issue, state = self.controller.advance(issue, state)
        self.api.worker(CORE, issue, state, "ready", "failure")
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["nodes"][CORE]["status"], "failed")
        self.controller.advance(issue, state)
        self.assertEqual(len(self.api.dispatched), 1)
        issue, state = self.controller.advance(issue, state, retry_failed=True)
        self.assertEqual(state["nodes"][CORE]["attempt"], 2)
        self.assertEqual(state["plan"]["snapshots"][CORE]["tag"], "v1.3.1")
        self.assertEqual([repo for repo, _ in self.api.dispatched], [CORE, CORE])

    def test_waiting_dependency_worker_is_resumed_not_replanned(self):
        issue, state = self.approved()
        issue, state = self.controller.advance(issue, state)
        self.api.worker(CORE, issue, state, "waiting")
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["nodes"][CORE]["status"], "pending")
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["nodes"][CORE]["attempt"], 2)
        self.assertEqual(len(self.api.dispatched), 2)

    def test_in_progress_worker_does_not_allow_consumers(self):
        issue, state = self.approved()
        issue, state = self.controller.advance(issue, state)
        self.api.worker(CORE, issue, state, "verifying")["status"] = "in_progress"
        self.controller.advance(issue, state)
        self.assertEqual(len(self.api.dispatched), 1)


    def test_explicit_reassessment_refreshes_refs_without_another_patch(self):
        issue, state = self.approved()
        evidence = inventory((CORE,))
        evidence["repositories"][CORE]["pulls"][0]["head"] = "e" * 40
        evidence["repositories"][CORE]["pulls"][0]["commits"] = [{"sha": "e" * 40, "subject": "Updated dependency"}]
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        issue, state = self.controller.approve(plan, 23, replan=True)
        self.assertEqual(state["plan"]["snapshots"][CORE]["pulls"][0]["head"], "e" * 40)
        self.assertEqual(state["plan"]["snapshots"][CORE]["tag"], "v1.3.1")
        self.assertEqual(state["revisions"][0]["source_run"], 23)

    def test_reassessment_cannot_modify_inflight_or_public_version(self):
        issue, state = self.approved()
        evidence = inventory((CORE,))
        evidence["repositories"][CORE]["pulls"][0]["head"] = "e" * 40
        evidence["repositories"][CORE]["pulls"][0]["commits"] = [{"sha": "e" * 40, "subject": "Updated dependency"}]
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        self.controller.advance(issue, state)
        with self.assertRaisesRegex(ValueError, "in-flight"):
            self.controller.approve(plan, 23, replan=True)
        state["nodes"][CORE]["status"] = "pending"
        self.controller.save(issue, state)
        with patch.object(self.api, "optional", return_value={"object": {"sha": MAIN}}), self.assertRaisesRegex(ValueError, "public candidate"):
            self.controller.approve(plan, 23, replan=True)

    def verification_reassessment(self):
        issue, state = self.approved()
        self.api.worker(CORE, issue, state, "verifying", "failure")
        state["nodes"][CORE]["status"] = "failed"
        self.controller.save(issue, state)
        evidence = inventory((GO,))
        source = evidence["repositories"][CORE]
        source.update(release_head=MAIN, source_commit=MAIN, automation_sha="e" * 40, main_head="e" * 40)
        self.api.changed[(CORE, source["branch"])] = MAIN
        self.api.changed[(CORE, "main")] = "e" * 40
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("skip", "release", "release")))
        return issue, state, plan

    def recover_verification(self, plan):
        original = self.api.optional
        def tag(path):
            return {"object": {"sha": MAIN}} if path.startswith(f"repos/{CORE}/git/ref/tags/") else original(path)
        with patch.object(self.api, "optional", side_effect=tag), \
             patch.object(release, "assert_tag") as signature, \
             patch.object(release, "check_public_assets") as assets:
            result = self.controller.approve(plan, 23, replan=True, recover_public_verification=True)
            signature.assert_called_once()
            assets.assert_called_once()
        return result

    def test_explicit_public_recovery_changes_only_worker_code_and_binds_asset_identity(self):
        issue, previous, plan = self.verification_reassessment()
        old_snapshot = copy.deepcopy(previous["plan"]["snapshots"][CORE])
        issue, state = self.recover_verification(plan)
        snapshot = state["plan"]["snapshots"][CORE]
        self.assertEqual({key: snapshot[key] for key in old_snapshot if key not in {"automation_sha", "main_head", "release_head"}},
                         {key: old_snapshot[key] for key in old_snapshot if key not in {"automation_sha", "main_head", "release_head"}})
        self.assertEqual(snapshot["release_head"], MAIN)
        self.assertEqual(snapshot["automation_sha"], "e" * 40)
        self.assertEqual(snapshot["verification_recovery"]["from_plan_id"], previous["plan"]["plan_id"])
        self.assertEqual(state["plan"]["decisions"][CORE]["decision"], "release")
        self.assertEqual(issue["number"], 1)
        self.assertEqual(state["nodes"][CORE]["status"], "pending")
        self.assertFalse(self.api.dispatched)
        coordinator.validate_state(state)

    def test_draft_recovery_preserves_release_id_assets_and_requires_new_plan_review(self):
        issue, previous, plan = self.verification_reassessment()
        cycle = release.Cycle(self.api, CORE, MONTH, "execute")
        cycle.state["status"] = "draft"
        cycle.save(cycle.state)
        original = self.api.optional
        with patch.object(self.api, "optional", side_effect=lambda path: {"object": {"sha": MAIN}}
                          if path.startswith(f"repos/{CORE}/git/ref/tags/") else original(path)), \
             patch.object(release, "assert_tag"), patch.object(release, "check_release_assets") as assets:
            issue, state = self.controller.approve(plan, 23, replan=True, recover_public_verification=True)
        assets.assert_called_once()
        self.assertTrue(assets.call_args.args[2])
        self.assertEqual(state["authorization"]["kind"], "maintainer")
        self.assertEqual(state["source_run"], 23)
        receipt = state["plan"]["snapshots"][CORE]["verification_recovery"]
        saved = release.Cycle(self.api, CORE, MONTH, "execute").state
        self.assertEqual(receipt["candidate_sha256"],
                         coordinator.digest({key: saved.get(key) for key in ("tag", "commit", "assets", "release_id")}))
        self.assertEqual(saved["release_id"], 777)
        self.assertFalse(self.api.dispatched)

    def test_verification_recovery_cannot_start_a_new_cycle_or_use_the_old_assessment(self):
        evidence = inventory((CORE,))
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        with self.assertRaisesRegex(ValueError, "existing approved cycle"):
            self.controller.approve(plan, 22, replan=True, recover_public_verification=True)
        self.controller.approve(plan, 22)
        with self.assertRaisesRegex(ValueError, "fresh changed assessment"):
            self.controller.approve(plan, 23, replan=True, recover_public_verification=True)
        with self.assertRaisesRegex(ValueError, "explicit replan"):
            self.controller.approve(plan, 23, recover_public_verification=True)

    def test_public_recovery_cannot_replace_tagged_but_unpublished_candidate(self):
        issue, state, plan = self.verification_reassessment()
        cycle = release.Cycle(self.api, CORE, MONTH, "execute")
        cycle.state["status"] = "tagged"
        cycle.save(cycle.state)
        with self.assertRaisesRegex(ValueError, "exact public verification candidate"):
            self.recover_verification(plan)

    def test_public_recovery_cannot_change_candidate_branch_or_baseline(self):
        for field, value in (("release_head", BRANCH), ("baseline", "v1.4.0"), ("workflow_id", 90)):
            with self.subTest(field=field):
                self.api = ControllerAPI()
                self.controller = ApprovedController(self.api, "test-actor")
                _, _, plan = self.verification_reassessment()
                plan["snapshots"][CORE][field] = value
                plan["plan_id"] = coordinator.digest({key: value for key, value in plan.items() if key != "plan_id"})
                with self.assertRaisesRegex(ValueError, "cannot change the public candidate"):
                    self.recover_verification(plan)

    def test_public_recovery_refuses_bad_signature_assets_and_stale_automation(self):
        for failure in ("signature", "assets", "stale"):
            with self.subTest(failure=failure):
                self.api = ControllerAPI()
                self.controller = ApprovedController(self.api, "test-actor")
                _, _, plan = self.verification_reassessment()
                if failure == "stale":
                    self.api.changed[(CORE, "main")] = "f" * 40
                target = "assert_tag" if failure == "signature" else "check_public_assets"
                original = self.api.optional
                with patch.object(self.api, "optional", side_effect=lambda path: {"object": {"sha": MAIN}} if path.startswith(f"repos/{CORE}/git/ref/tags/") else original(path)), \
                     patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"), \
                     patch.object(release, target, side_effect=ValueError("Gate rejected")):
                    writes = len(self.api.writes)
                    with self.assertRaisesRegex(ValueError, "stale|Gate rejected"):
                        self.controller.approve(plan, 23, replan=True, recover_public_verification=True)
                    self.assertEqual(len(self.api.writes), writes)

    def test_public_recovery_rebinds_verification_without_git_or_new_publication(self):
        _, _, plan = self.verification_reassessment()
        issue, state = self.recover_verification(plan)
        candidate = copy.deepcopy(release.Cycle(self.api, CORE, MONTH, "execute").state)
        issue, state = self.controller.advance(issue, state)
        environment = {
            "MONTHLY_PATCH_COORDINATOR_REQUIRED": "true", "GITHUB_REF": "refs/heads/main",
            "GITHUB_SHA": "e" * 40, "MONTHLY_PATCH_CONTROLLER_ISSUE": str(issue["number"]),
            "MONTHLY_PATCH_CONTROLLER_PLAN": state["plan"]["plan_id"], "MONTHLY_PATCH_CONTROLLER_ATTEMPT": "1",
            "MONTHLY_PATCH_ENABLED": "true", "MONTHLY_PATCH_TOKEN_READY": "true",
            "MONTHLY_PATCH_WORKER_SHA": AUTOMATION, "GITHUB_RUN_ID": "1",
            "MONTHLY_PATCH_SIGNING_KEY": "test-key", "MONTHLY_PATCH_SIGNER_EMAIL": "test@example.com",
            "MONTHLY_PATCH_SIGNER_LOGIN": "test-actor",
        }
        with patch.dict(os.environ, environment), patch.object(release, "assert_tag"), \
             patch.object(release, "check_public_assets"), patch.object(release, "check_draft_review"), patch.object(release, "git") as git:
            result = release.prepare(self.api, CORE, "execute", MONTH, pathlib.Path("."), pathlib.Path("."), True)
            self.assertEqual(result["action"], "verify")
            self.assertEqual(result["controller_plan_id"], state["plan"]["plan_id"])
            for field in ("tag", "commit", "assets"):
                self.assertEqual(result[field], candidate[field])
            git.assert_not_called()
            for field, value in (("status", "ready"), ("commit", BRANCH), ("assets", [{"name": "unexpected"}])):
                with self.subTest(field=field), self.assertRaises(ValueError):
                    coordinator.worker_authorization(self.api, CORE, MONTH, {**result, field: value})
            coordinator.worker_authorization(self.api, CORE, MONTH, result)
            with self.assertRaisesRegex(ValueError, "cannot build or republish"):
                release.publish(self.api, result, pathlib.Path("."))
            cycle = release.Cycle(self.api, CORE, MONTH, "execute")
            cycle.state["status"] = "published"
            cycle.save(cycle.state)
            coordinator.worker_authorization(self.api, CORE, MONTH, notification=True)
            with self.assertRaises(ValueError):
                coordinator.worker_authorization(self.api, CORE, MONTH)
            with self.assertRaisesRegex(ValueError, "cannot authorize writing"):
                coordinator.worker_authorization(self.api, CORE, MONTH, result, notification=True)
            with tempfile.TemporaryDirectory() as output, patch.object(release, "GitHub", return_value=self.api), \
                 patch("sys.argv", ["monthly_release.py", "notify-controller", "--repository", CORE,
                                   "--mode", "execute", "--month", MONTH, "--output", output]):
                self.assertEqual(release.main(), 0)
            hints = [payload for path, method, payload in self.api.writes
                     if path == f"repos/{CLI}/dispatches" and method == "POST"]
            self.assertEqual(len(hints), 1)
            self.assertEqual(hints[0]["event_type"], "notation-release-progress")
            cycle.state["status"] = "verifying"
            cycle.save(cycle.state)
            with self.assertRaises(ValueError):
                coordinator.worker_authorization(self.api, CORE, MONTH, candidate)
            cycle.state["status"] = "merging"
            cycle.save(cycle.state)
            with self.assertRaisesRegex(ValueError, "cannot change or rebuild"):
                coordinator.worker_authorization(self.api, CORE, MONTH)
        self.assertFalse(any(method == "PUT" for _, method, _ in self.api.writes))

    def test_unacknowledged_dispatch_is_not_blindly_reissued(self):
        issue, state = self.approved()
        original = self.api.request
        with patch.object(self.api, "request", wraps=self.api.request) as request:
            def interrupted(path, method="GET", payload=None):
                if path.endswith("/dispatches"):
                    raise release.APIError("HTTP 503 ambiguous request")
                return original(path, method, payload)
            request.side_effect = interrupted
            with self.assertRaises(release.APIError):
                self.controller.advance(issue, state)
        issue, restored = self.controller.records()[0]
        self.assertEqual(restored["nodes"][CORE]["status"], "dispatching")
        self.controller.advance(issue, restored)
        self.assertEqual(self.api.dispatched, [])

    def test_one_immutable_plan_per_month(self):
        issue, state = self.approved()
        same_issue, same_state = self.controller.approve(state["plan"], 22)
        self.assertEqual(same_issue["number"], issue["number"])
        self.assertEqual(same_state, state)
        evidence = inventory((CLI,))
        changed = coordinator.validate_assessment(evidence, assessment(evidence, ("skip", "skip", "release")))
        with self.assertRaisesRegex(ValueError, "immutable"):
            self.controller.approve(changed, 23)

    def test_overlapping_months_cannot_be_approved(self):
        self.approved()
        evidence = inventory((CORE,))
        evidence["month"] = "2026-11"
        for snapshot in evidence["repositories"].values():
            snapshot["tag"] = "v1.3.1"
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        with self.assertRaisesRegex(ValueError, "active cycle"):
            self.controller.approve(plan, 23)

    def test_stale_assessment_and_changed_refs_require_reassessment(self):
        evidence = inventory((CORE,))
        evidence["collected_at"] = (datetime.datetime.now(datetime.timezone.utc) - datetime.timedelta(days=8)).isoformat()
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        with self.assertRaisesRegex(ValueError, "stale"):
            self.controller.approve(plan, 22)
        self.api.changed[(CORE, "main")] = PR_HEAD
        with self.assertRaisesRegex(ValueError, "refs changed"):
            self.approved()

    def test_changed_worker_code_is_not_dispatched(self):
        issue, state = self.approved()
        self.api.changed[(CORE, "main")] = PR_HEAD
        with self.assertRaisesRegex(ValueError, "Source changed"):
            self.controller.advance(issue, state)
        self.assertEqual(self.api.dispatched, [])

    def test_untrusted_issues_cannot_authorize_or_block_controller(self):
        issue, state = self.approved()
        impostor = copy.deepcopy(issue)
        impostor.update(number=99, user={"login": "attacker"}, body=issue["body"].replace(MONTH, "2026-11"))
        self.api.issues_by_repo[CLI].append(impostor)
        self.assertEqual(len(self.controller.records()), 1)
        self.assertIsNone(coordinator.decode_issue(impostor, "test-actor"))

    def test_duplicate_or_mutated_trusted_records_block(self):
        issue, state = self.approved()
        duplicated = {**issue, "number": 99}
        self.api.issues_by_repo[CLI].append(duplicated)
        with self.assertRaisesRegex(ValueError, "Duplicate"):
            self.controller.records()
        state["plan"]["snapshots"][CORE]["tag"] = "v1.3.99"
        with self.assertRaisesRegex(ValueError, "plan changed"):
            coordinator.validate_state(state)

    def test_wrong_worker_actor_code_or_title_does_not_complete(self):
        for field, value in (("actor", {"login": "attacker"}), ("head_sha", PR_HEAD), ("display_title", "unrelated")):
            self.api = ControllerAPI()
            self.controller = ApprovedController(self.api, "test-actor")
            issue, state = self.approved()
            issue, state = self.controller.advance(issue, state)
            run = self.api.worker(CORE, issue, state)
            run[field] = value
            if field == "display_title":
                self.controller.advance(issue, state)
                self.assertEqual(state["nodes"][CORE]["status"], "dispatching")
            else:
                with self.subTest(field=field), self.assertRaises(ValueError):
                    self.controller.advance(issue, state)
            self.assertEqual(len(self.api.dispatched), 1)

    def test_missing_cycle_or_unverified_assets_cannot_release_consumer(self):
        issue, state = self.approved()
        issue, state = self.controller.advance(issue, state)
        self.api.worker(CORE, issue, state)
        self.api.issues_by_repo[CORE] = []
        issue, state = self.controller.advance(issue, state)
        self.assertEqual(state["nodes"][CORE]["status"], "failed")
        self.assertEqual(len(self.api.dispatched), 1)

    def test_asset_validation_failure_blocks_consumer(self):
        issue, state = self.approved()
        issue, state = self.controller.advance(issue, state)
        self.api.worker(CORE, issue, state)
        with patch.object(release, "assert_tag"), patch.object(release, "check_public_assets", side_effect=ValueError("bad asset digest")), self.assertRaisesRegex(ValueError, "digest"):
            self.controller.advance(issue, state)
        self.assertEqual(len(self.api.dispatched), 1)

    def test_published_consumer_must_include_the_exact_planned_library(self):
        issue, state = self.approved()
        issue, state = self.controller.advance(issue, state)
        self.api.worker(CORE, issue, state)
        with patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"):
            issue, state = self.controller.advance(issue, state)
            self.api.worker(GO, issue, state)
            self.api.manifests[(GO, MAIN, ".")]["Require"][0]["Version"] = "v1.3.0"
            with self.assertRaisesRegex(ValueError, "planned producer"):
                self.controller.advance(issue, state)
        self.assertEqual([repo for repo, _ in self.api.dispatched], [CORE, GO])


class WorkerAuthorizationTests(unittest.TestCase):
    def setUp(self):
        self.api = ControllerAPI()
        self.environment = patch.dict(os.environ, {
            "MONTHLY_PATCH_ACTOR": "test-actor", "MONTHLY_PATCH_COORDINATOR_REQUIRED": "true",
            "GITHUB_REF": "refs/heads/main", "GITHUB_SHA": AUTOMATION,
            "MONTHLY_PATCH_WORKER_SHA": AUTOMATION, "GITHUB_RUN_ID": "1",
        })
        self.environment.start()
        self.addCleanup(self.environment.stop)
        evidence = inventory((CORE,))
        plan = coordinator.validate_assessment(evidence, assessment(evidence, ("release", "release", "release")))
        self.controller = ApprovedController(self.api, "test-actor")
        self.issue, self.state = self.controller.approve(plan, 22)
        self.issue, self.state = self.controller.advance(self.issue, self.state)
        os.environ.update({
            "MONTHLY_PATCH_CONTROLLER_ISSUE": str(self.issue["number"]),
            "MONTHLY_PATCH_CONTROLLER_PLAN": plan["plan_id"], "MONTHLY_PATCH_CONTROLLER_ATTEMPT": "1",
        })

    def test_only_current_dispatched_node_is_authorized(self):
        authorization = coordinator.worker_authorization(self.api, CORE, MONTH)
        self.assertEqual(authorization["issue"], self.issue["number"])
        with self.assertRaises(ValueError):
            coordinator.worker_authorization(self.api, GO, MONTH)
        with self.assertRaises(ValueError):
            coordinator.worker_authorization(self.api, "untrusted/notation-core-go", MONTH)



    def test_missing_stale_or_changed_receipt_is_rejected(self):
        for variable, value in (("MONTHLY_PATCH_CONTROLLER_ISSUE", ""), ("MONTHLY_PATCH_CONTROLLER_PLAN", "0" * 64),
                                ("MONTHLY_PATCH_CONTROLLER_ATTEMPT", "2"), ("GITHUB_SHA", PR_HEAD),
                                ("GITHUB_REF", "refs/heads/other")):
            with patch.dict(os.environ, {variable: value}), self.subTest(variable=variable), self.assertRaises(ValueError):
                coordinator.worker_authorization(self.api, CORE, MONTH)

    def test_candidate_must_keep_fixed_tag_mode_and_producers(self):
        candidate = worker_plan(CORE, "execute", MONTH)
        candidate.update(controller_issue=self.issue["number"], controller_plan_id=self.state["plan"]["plan_id"],
                         producers={}, maintainer_roster_sha=AUTOMATION)
        coordinator.worker_authorization(self.api, CORE, MONTH, candidate)
        for field, value in (("tag", "v1.3.99"), ("mode", "rehearse"),
                             ("cycle_id", "upstream-" + "1" * 64), ("producers", {"notation-core-go": {}})):
            with self.subTest(field=field), self.assertRaises(ValueError):
                coordinator.worker_authorization(self.api, CORE, MONTH, {**candidate, field: value})

    def test_producer_versions_are_taken_from_verified_plan_not_latest_unplanned_release(self):
        self.api.worker(CORE, self.issue, self.state)
        with patch.object(release, "assert_tag"), patch.object(release, "check_public_assets"):
            self.issue, self.state = self.controller.advance(self.issue, self.state)
            authorization = coordinator.worker_authorization(self.api, GO, MONTH)
        self.assertEqual(authorization["producers"], {"notation-core-go": {"repository": CORE, "version": "v1.3.1"}})

    def test_assessed_pr_heads_cannot_change(self):
        authorization = coordinator.worker_authorization(self.api, CORE, MONTH)
        candidate = pull(sha=PR_HEAD)
        self.assertTrue(coordinator.scoped_pull(self.api, CORE, candidate, authorization))
        candidate["head"]["sha"] = MAIN
        with self.assertRaisesRegex(ValueError, "head changed"):
            coordinator.scoped_pull(self.api, CORE, candidate, authorization)

    def test_new_unplanned_third_party_update_is_not_added(self):
        authorization = coordinator.worker_authorization(self.api, CORE, MONTH)
        self.assertFalse(coordinator.scoped_pull(self.api, CORE, pull(number=99), authorization))

    def test_new_producer_pr_must_only_update_planned_producers(self):
        candidate = pull(number=99, sha=PR_HEAD)
        candidate["head"]["repo"]["full_name"] = GO
        candidate["base"]["sha"] = MAIN
        base = {"Go": "1.26", "Require": [{"Path": "github.com/notaryproject/notation-core-go", "Version": "v1.3.0"}],
                }
        after = copy.deepcopy(base)
        after["Require"][0]["Version"] = "v1.3.1"
        self.api.manifests = {(GO, MAIN, "."): base, (GO, PR_HEAD, "."): after}
        self.api.manifests[(CORE, "v1.3.1", ".")] = {"Go": "1.26"}
        producers = {"notation-core-go": {"repository": CORE, "version": "v1.3.1"}}
        self.assertTrue(coordinator.producer_pull(self.api, GO, candidate, producers))
        after["Require"].append({"Path": "unrelated/module", "Version": "v2.0.0"})
        self.assertFalse(coordinator.producer_pull(self.api, GO, candidate, producers))
        after["Require"].pop()
        after["Go"] = "1.27"
        self.assertFalse(coordinator.producer_pull(self.api, GO, candidate, producers))

    def test_propagation_allows_transitive_requirements_and_required_go_floor(self):
        candidate = pull(number=99, sha=PR_HEAD)
        candidate["head"]["repo"]["full_name"] = GO
        candidate["base"]["sha"] = MAIN
        module = "github.com/notaryproject/notation-core-go"
        base = {"Go": "1.26", "Require": [{"Path": module, "Version": "v1.3.0"}],
                }
        after = copy.deepcopy(base)
        after["Go"] = "1.26.2"
        after["Require"][0]["Version"] = "v1.3.1"
        after["Require"].append({"Path": "transitive/module", "Version": "v1.2.3", "Indirect": True})
        self.api.manifests = {(GO, MAIN, "."): base, (GO, PR_HEAD, "."): after,
                              (CORE, "v1.3.1", "."): {"Go": "1.26.2"}}
        producers = {"notation-core-go": {"repository": CORE, "version": "v1.3.1"}}
        self.assertTrue(coordinator.producer_pull(self.api, GO, candidate, producers))
        after["Toolchain"] = "go1.27.1"
        with patch.object(release, "run", return_value="go1.27.1\n"):
            self.assertTrue(coordinator.producer_pull(self.api, GO, candidate, producers))
        after["Go"] = "1.27"
        self.assertFalse(coordinator.producer_pull(self.api, GO, candidate, producers))

    def test_merged_propagation_keeps_its_observed_provenance_after_base_moves(self):
        authorization = coordinator.worker_authorization(self.api, CORE, MONTH)
        candidate = pull(number=99)
        candidate.update(state="closed", merged_at="2026-10-04T00:00:00Z", merge_commit_sha=BRANCH)
        self.assertTrue(coordinator.scoped_pull(self.api, CORE, candidate, authorization,
                                               [{"number": 99, "commit": BRANCH}]))
        candidate["merge_commit_sha"] = MAIN
        with self.assertRaisesRegex(ValueError, "merge changed"):
            coordinator.scoped_pull(self.api, CORE, candidate, authorization, [{"number": 99, "commit": BRANCH}])

    def test_cli_can_merge_separate_core_and_go_propagation_prs(self):
        candidate = pull(number=99, sha=PR_HEAD)
        candidate["head"]["repo"]["full_name"] = CLI
        candidate["base"]["sha"] = MAIN
        base = {
            "Go": "1.26", "Require": [{"Path": f"github.com/notaryproject/{name}", "Version": "v1.3.0"}
                                     for name in ("notation-core-go", "notation-go")],
        }
        after = copy.deepcopy(base)
        after["Require"][0]["Version"] = "v1.3.1"
        producers = {name: {"repository": f"notaryproject/{name}", "version": "v1.3.1"}
                     for name in ("notation-core-go", "notation-go")}
        self.api.manifests = {(item["repository"], item["version"], "."): {"Go": "1.26"} for item in producers.values()}
        for directory in release.modules(CLI):
            self.api.manifests[(CLI, MAIN, directory)] = base
            self.api.manifests[(CLI, PR_HEAD, directory)] = after
        self.assertTrue(coordinator.producer_pull(self.api, CLI, candidate, producers))

    def test_worker_policy_includes_authorization_before_merging(self):
        with patch.object(release, "policy"), patch.object(release, "resume_cycle") as resume, patch.dict(
            os.environ, {"MONTHLY_PATCH_CONTROLLER_PLAN": "0" * 64}
        ), self.assertRaises(ValueError):
            release.prepare(self.api, CORE, "execute", MONTH, pathlib.Path("."), pathlib.Path("."))
        resume.assert_not_called()


    def test_manual_automation_files_keep_repository_license_headers(self):
        root = pathlib.Path(__file__).parent.parent
        header = "\n".join(pathlib.Path(release.__file__).read_text().splitlines()[:13])
        files = [root / "scripts/notation_release_controller.py", pathlib.Path(__file__),
                 root / "workflows/monthly-patch-worker.yml",
                 root / "workflows/monthly-patch-release.yml", root / "workflows/notation-release-validation.yml"]
        controller = root / "workflows/notation-release-controller.yml"
        if controller.exists():
            files.append(controller)
        for file in files:
            with self.subTest(file=file):
                self.assertTrue(file.read_text().startswith(header))


if __name__ == "__main__":
    unittest.main()
