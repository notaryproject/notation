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
import datetime
import gzip
import hashlib
import io
import json
import os
import pathlib
import tempfile
import unittest
import urllib.error
from unittest.mock import patch

import monthly_release as release
import notation_release_controller as controller
import test_monthly_public_release as packages
from test_monthly_release import ENVIRONMENT, FakeGitHub, plan, pull
from test_upstream_authorization import CLI, CORE, MONTH, PR_HEAD, SOURCE, UpstreamAPI, assessed, inventory


class PlanReviewTests(unittest.TestCase):
    def setUp(self):
        self.env = patch.dict(os.environ, {
            **ENVIRONMENT, "GITHUB_REPOSITORY": CLI, "GITHUB_SHA": SOURCE,
            "MONTHLY_PATCH_COORDINATOR_REQUIRED": "true", "MONTHLY_PATCH_WORKER_SHA": SOURCE,
            "GITHUB_RUN_ID": "1", "NOTATION_RELEASE_AUTO_APPROVAL_ENABLED": "true",
        })
        self.env.start()
        self.addCleanup(self.env.stop)
        self.api = UpstreamAPI()
        self.controller = controller.Controller(self.api, "test-actor")
        self.plan = assessed(inventory())
        self.api.prs[(CORE, 1)] = pull(sha=PR_HEAD)
        self.issue, self.state = self.controller.approve(self.plan, 22)

    def approve(self):
        self.api.approve_comment(self.issue, self.state)
        return self.controller.advance(self.issue, self.state)

    def test_assessment_creates_only_review_issue_never_execution(self):
        self.assertEqual(self.state["status"], "review")
        self.controller.advance(self.issue, self.state)
        self.assertFalse(self.api.dispatched)
        self.assertTrue(all("/issues" in path for path, _, _ in self.api.writes))
        with self.assertRaisesRegex(ValueError, "disabled"):
            controller.authorize_routine(self.api, self.plan)

    def test_exact_unedited_active_maintainer_comment_dispatches(self):
        issue, state = self.approve()
        self.assertEqual(state["authorization"]["kind"], "maintainer")
        self.assertEqual(state["authorization"]["login"], "maintainer")
        self.assertEqual(len(self.api.dispatched), 1)
        self.assertIn(controller.plan_comment(self.plan), issue["body"])

    def test_unknown_bot_stale_wrong_issue_or_edited_comments_cannot_approve(self):
        self.api.approve_comment(self.issue, self.state)
        original = copy.deepcopy(self.api.comments[1])
        for changes in (
            {"user": {"login": "attacker", "type": "User"}},
            {"user": {"login": "maintainer", "type": "Bot"}},
            {"body": "/approve-plan " + "f" * 64},
            {"body": controller.plan_comment(self.plan) + "\nLooks good"},
            {"updated_at": "2099-01-02T00:00:00Z"},
            {"issue_url": "https://api.github.com/repos/untrusted/notation/issues/1"},
        ):
            self.api.comments[1] = {**original, **changes}
            self.controller.advance(self.issue, self.state)
            self.assertFalse(self.api.dispatched)

    def test_close_or_checklist_is_not_approval(self):
        for changes in ({"state": "closed"}, {"body": self.issue["body"] + "\n- [x] Approved"}):
            changed = {**self.issue, **changes}
            self.api.issues_by_repo[CLI] = [changed]
            with self.assertRaises(ValueError):
                self.controller.advance(changed, self.state)
            self.assertFalse(self.api.dispatched)

    def test_comment_edit_deletion_and_retired_maintainer_revoke_remaining_writes(self):
        issue, state = self.approve()
        receipt = copy.deepcopy(self.api.comments[1])
        for changes in ({"body": "revoked"}, {"updated_at": "2099-01-02T00:00:00Z"}):
            self.api.comments[1] = {**receipt, **changes}
            with self.assertRaisesRegex(ValueError, "revoked"):
                controller.require_plan_approval(self.api, issue, state)
        self.api.comments[1] = receipt
        self.api.roster = "# Repo-Level Maintainers\nReplacement <r@example.com> (@replacement)\n"
        with self.assertRaisesRegex(ValueError, "revoked"):
            controller.require_plan_approval(self.api, issue, state)
        self.api.comments.clear()
        with self.assertRaises(KeyError):
            controller.require_plan_approval(self.api, issue, state)

    def test_worker_cannot_use_unapproved_or_edited_plan(self):
        os.environ.update(MONTHLY_PATCH_CONTROLLER_ISSUE=str(self.issue["number"]),
                          MONTHLY_PATCH_CONTROLLER_PLAN=self.plan["plan_id"], MONTHLY_PATCH_CONTROLLER_ATTEMPT="1")
        with self.assertRaisesRegex(ValueError, "authorization"):
            controller.worker_authorization(self.api, CORE, MONTH)
        issue, state = self.approve()
        self.api.comments[1]["body"] = "revoked"
        with self.assertRaisesRegex(ValueError, "revoked"):
            controller.worker_authorization(self.api, CORE, MONTH)

    def test_current_refs_policy_and_age_are_rechecked_before_dispatch(self):
        self.api.approve_comment(self.issue, self.state)
        self.api.prs[(CORE, 1)]["head"]["sha"] = "f" * 40
        with self.assertRaisesRegex(ValueError, "PR head"):
            self.controller.advance(self.issue, self.state)
        self.assertFalse(self.api.dispatched)

    def test_stale_plan_approval_is_not_authority_for_new_assessment(self):
        self.approve()
        self.state["nodes"][CORE]["status"] = "pending"
        self.issue = self.controller.save(self.issue, self.state)
        evidence = inventory()
        evidence["collected_at"] = (datetime.datetime.now(datetime.timezone.utc) - datetime.timedelta(minutes=1)).isoformat()
        replacement = assessed(evidence)
        issue, state = self.controller.approve(replacement, 23, replan=True)
        self.assertEqual(state["status"], "review")
        dispatched = len(self.api.dispatched)
        self.controller.advance(issue, state)
        self.assertEqual(len(self.api.dispatched), dispatched)
        self.assertNotEqual(state["plan"]["plan_id"], self.plan["plan_id"])

    def test_stale_assessment_and_changed_policy_block_comment_approval(self):
        self.api.approve_comment(self.issue, self.state)
        self.state["plan"]["collected_at"] = (datetime.datetime.now(datetime.timezone.utc) - datetime.timedelta(days=2)).isoformat()
        self.state["plan"]["plan_id"] = controller.digest({key: value for key, value in self.state["plan"].items() if key != "plan_id"})
        self.issue = self.controller.save(self.issue, self.state)
        self.api.approve_comment(self.issue, self.state)
        with self.assertRaisesRegex(ValueError, "stale"):
            self.controller.advance(self.issue, self.state)
        self.assertFalse(self.api.dispatched)

    def test_explicit_scope_blocker_cannot_be_overridden_by_comment(self):
        self.state["review_blocker"] = "Workflow changes require a fresh configured-actor review request"
        self.issue = self.controller.save(self.issue, self.state)
        self.api.approve_comment(self.issue, self.state)
        with self.assertRaisesRegex(ValueError, "execution is blocked"):
            self.controller.advance(self.issue, self.state)
        self.assertFalse(self.api.dispatched)

    def test_stable_branch_lease_and_policy_changes_block_initial_dispatch(self):
        self.api.approve_comment(self.issue, self.state)
        self.api.changed[(CORE, "release-1.3")] = "f" * 40
        with self.assertRaisesRegex(ValueError, "Stable branch changed"):
            self.controller.advance(self.issue, self.state)
        self.api.changed.clear()
        self.api.policy["publication"] = "automatic"
        with self.assertRaisesRegex(ValueError, "Unsupported"):
            self.controller.advance(self.issue, self.state)
        self.assertFalse(self.api.dispatched)

    def test_approval_is_bound_to_exact_full_readable_scope(self):
        evidence = inventory()
        evidence["repositories"][CORE]["pulls"][0].update(
            title="Bump dependency | [injection](https://evil.invalid) @everyone",
            commits=[{"sha": "e" * 40, "subject": "additional"}, {"sha": PR_HEAD, "subject": "same"}])
        item = controller.validate_assessment(evidence, {
            "schema": 1, "month": MONTH, "inventory_sha256": controller.digest(evidence),
            "decisions": self.plan["decisions"],
        })
        rendered = controller.release_scope(item)
        self.assertEqual(rendered.count(f"/pull/1)"), 1)
        self.assertIn("Open PR proposed HEAD", rendered)
        self.assertIn("e" * 40, rendered)
        self.assertNotIn("[injection]", rendered)
        self.assertNotIn("@everyone", rendered)
        self.assertIn("not Dependabot source commits", rendered)
        self.assertNotIn("f" * 40, rendered)
        item["snapshots"][CORE]["merged"] = [{"number": 7, "commit": "d" * 40, "title": "Merged update"}]
        self.assertIn("Already merged source commit to backport", controller.release_scope(item))

    def test_missing_duplicate_or_changed_commit_inventory_cannot_be_approved(self):
        for commits in ([], [{"sha": "f" * 40}], [{"sha": PR_HEAD}, {"sha": PR_HEAD}]):
            evidence = inventory()
            evidence["repositories"][CORE]["pulls"][0]["commits"] = commits
            with self.assertRaisesRegex(ValueError, "commit list"):
                assessed(evidence)


class DraftAPI(FakeGitHub):
    def __init__(self):
        super().__init__()
        self.remote = None
        self.assets = []

    def pages(self, path):
        if path.endswith("/assets"):
            return iter(copy.deepcopy(self.assets))
        return super().pages(path)

    def request(self, path, method="GET", payload=None):
        if path == f"repos/{CORE}/releases" and method == "POST":
            self.writes.append((path, method, payload))
            self.remote = {**payload, "id": 44}
            self.releases.append(self.remote)
            return copy.deepcopy(self.remote)
        if path == f"repos/{CORE}/releases/44":
            if self.remote is None:
                raise release.APIError("HTTP 404 release removed")
            return copy.deepcopy(self.remote)
        if "/issues/" in path and method == "GET":
            return next(issue for issue in self.issues if issue["number"] == int(path.rsplit("/", 1)[1]))
        return super().request(path, method, payload)

    def upload(self, command):
        if not command[4].startswith("https://uploads.github.com/"):
            raise AssertionError("Expected the exact trusted draft upload endpoint")
        path = pathlib.Path(command[-1])
        content = path.read_bytes()
        item = {"id": len(self.assets) + 100, "name": path.name, "size": len(content),
                "digest": "sha256:" + hashlib.sha256(content).hexdigest()}
        self.assets.append(item)
        return json.dumps(item)


class DraftReleaseTests(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ, {
            **ENVIRONMENT, "GITHUB_RUN_ID": "123", "GITHUB_REPOSITORY": CORE,
            "MONTHLY_PATCH_WORKER_SHA": SOURCE,
        })
        self.environment.start()
        self.addCleanup(self.environment.stop)
        for name in ("coordinator_authorization", "assert_tag"):
            guard = patch.object(release, name, return_value=None)
            guard.start()
            self.addCleanup(guard.stop)
        self.api = DraftAPI()
        self.state = plan(status="tagged")
        self.state.update(controller_issue=10, controller_plan_id="a" * 64,
                          merged=[{"number": 1, "commit": PR_HEAD, "title": "Dependency update"}])
        validate = patch("monthly_release_pr.validate")
        validate.start()
        self.addCleanup(validate.stop)
        release.Cycle(self.api, CORE, MONTH, "execute").save(self.state)
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.directory = pathlib.Path(directory.name)
        for name in release.asset_names("notation-core-go", self.state["tag"]):
            (self.directory / name).write_bytes(("qualified bytes for " + name).encode())

    def stage(self):
        with patch.object(release, "run", side_effect=self.api.upload):
            return release.publish(self.api, self.state, self.directory)

    def prepare(self):
        return release.prepare(self.api, CORE, "execute", MONTH, self.directory, self.directory, True)

    def qualify(self):
        return release.complete(self.api, self.stage())

    def test_true_draft_stable_tag_and_exact_recorded_assets(self):
        draft = self.stage()
        self.assertEqual(draft["status"], "draft")
        self.assertEqual(draft["release_id"], 44)
        self.assertTrue(self.api.remote["draft"])
        self.assertFalse(self.api.remote["prerelease"])
        self.assertEqual(len(draft["assets"]), 2)
        self.assertTrue(all(item["id"] >= 100 for item in draft["assets"]))
        self.assertFalse(any("Qualified draft release review" in issue["body"] for issue in self.api.issues))
        self.assertEqual(self.prepare()["action"], "draft-check")

    def test_only_successful_draft_checks_open_review_issue_then_wait_for_human(self):
        result = self.qualify()
        self.assertEqual(result["status"], "awaiting-publication")
        review = release.check_draft_review(self.api, result)
        self.assertIn("Publish release", review["body"])
        self.assertIn("already public", review["body"])
        self.assertIn(PR_HEAD, review["body"])
        self.assertIn("Dependabot source", review["body"])
        self.assertEqual(self.prepare()["action"], "waiting")
        self.assertTrue(self.api.remote["draft"])
        self.assertFalse(any(path.endswith("/dispatches") for path, _, _ in self.api.writes))

    def test_human_publication_requires_same_identity_then_anonymous_gate_before_complete(self):
        review = self.qualify()
        self.api.remote["draft"] = False
        result = self.prepare()
        self.assertEqual(result["action"], "verify")
        self.assertEqual(result["release_id"], review["release_id"])
        self.assertEqual(result["status"], "verifying")
        done = release.complete(self.api, result)
        self.assertEqual(done["status"], "published")
        self.assertFalse(any(method == "PATCH" and "/releases/" in path for path, method, _ in self.api.writes))

    def test_published_before_draft_quality_checks_cannot_become_qualified(self):
        draft = self.stage()
        self.api.remote["draft"] = False
        with self.assertRaisesRegex(ValueError, "visibility"):
            release.complete(self.api, draft)
        with self.assertRaisesRegex(ValueError, "visibility"):
            self.prepare()
        self.assertEqual(len(self.api.issues), 1)

    def test_changed_asset_id_bytes_size_and_extra_assets_block_publication_observation(self):
        self.qualify()
        self.api.remote["draft"] = False
        original = copy.deepcopy(self.api.assets)
        for changes in ({"id": 999}, {"digest": "sha256:" + "f" * 64}, {"size": 1}):
            self.api.assets[0] = {**original[0], **changes}
            with self.assertRaisesRegex(ValueError, "integrity"):
                self.prepare()
        self.api.assets = original + [original[0]]
        with self.assertRaisesRegex(ValueError, "assets differ"):
            self.prepare()

    def test_replaced_release_tag_or_prerelease_is_rejected(self):
        self.qualify()
        self.api.remote["draft"] = False
        original = copy.deepcopy(self.api.remote)
        for changes in ({"id": 45}, {"tag_name": "v1.3.2"}, {"prerelease": True}):
            self.api.remote = {**original, **changes}
            with self.assertRaises(ValueError):
                self.prepare()
        self.api.remote = None
        with self.assertRaises(release.APIError):
            self.prepare()

    def test_review_issue_mutation_cannot_authorize_progress(self):
        review = self.qualify()
        issue = next(item for item in self.api.issues if item["number"] == review["draft_review_issue"])
        issue["body"] += "\nForged approval"
        with self.assertRaisesRegex(ValueError, "review issue"):
            self.prepare()

    def test_unrecorded_existing_release_is_not_adopted_even_if_draft(self):
        self.api.releases.append({"tag_name": self.state["tag"], "draft": True})
        with self.assertRaisesRegex(ValueError, "Unrecorded existing"):
            self.stage()
        self.assertIsNone(self.api.remote)

    def test_retry_uploads_missing_assets_but_never_replaces_or_rebuilds_bytes(self):
        original = self.api.upload
        calls = []
        def fail_second(command):
            calls.append(command)
            if len(calls) == 2:
                raise release.APIError("HTTP 503 upload failed")
            return original(command)
        with patch.object(release, "run", side_effect=fail_second), self.assertRaises(release.APIError):
            release.publish(self.api, self.state, self.directory)
        saved = release.Cycle(self.api, CORE, MONTH, "execute").state
        self.assertEqual(saved["status"], "staging")
        self.assertEqual(saved["release_id"], 44)
        path = next(self.directory.iterdir())
        content = path.read_bytes()
        path.write_bytes(b"different build")
        with self.assertRaisesRegex(ValueError, "originally staged"):
            release.publish(self.api, saved, self.directory)
        path.write_bytes(content)
        with patch.object(release, "run", side_effect=original):
            resumed = release.publish(self.api, saved, self.directory)
        self.assertEqual(resumed["release_id"], 44)
        self.assertEqual(len(self.api.assets), 2)
        self.assertEqual(sum(method == "POST" and path.endswith("/releases") for path, method, _ in self.api.writes), 1)

    def test_review_issue_retry_reuses_original_receipt_and_issue(self):
        draft = self.stage()
        save = release.Cycle.save
        def fail_final(cycle, state):
            if state["status"] == "awaiting-publication":
                raise release.APIError("HTTP 503 state save failed")
            return save(cycle, state)
        with patch.object(release.Cycle, "save", fail_final), self.assertRaises(release.APIError):
            release.complete(self.api, draft)
        saved = release.Cycle(self.api, CORE, MONTH, "execute").state
        self.assertEqual(saved["status"], "draft")
        self.assertEqual(len(self.api.issues), 2)
        with patch.dict(os.environ, {"GITHUB_RUN_ID": "124"}):
            resumed = release.complete(self.api, saved)
        self.assertEqual(len(self.api.issues), 2)
        self.assertEqual(resumed["draft_verification"]["run_id"], 123)

    def test_native_jobs_cannot_access_release_secrets_and_never_publish(self):
        root = pathlib.Path(__file__).resolve().parents[1]
        workflow = (root / "workflows/monthly-patch-worker.yml").read_text()
        native = workflow.split("  verify:\n", 1)[1].split("  complete:\n", 1)[0]
        self.assertNotIn("secrets.", native)
        self.assertIn("GH_TOKEN: ${{ github.token }}", native)
        self.assertIn("needs.verify.result == 'success'", workflow)
        self.assertNotIn("draft: false", workflow)
        self.assertNotIn("--draft=false", workflow)
        self.assertIn("inputs.mode == 'execute'", workflow)

    def test_draft_manifest_cannot_replace_candidate_scope_during_review(self):
        draft = self.stage()
        forged = copy.deepcopy(draft)
        forged["merged"].append({"number": 999, "commit": "f" * 40})
        result = release.complete(self.api, forged)
        body = release.check_draft_review(self.api, result)["body"]
        self.assertNotIn("/pull/999", body)
        self.assertNotIn("f" * 40, body)

    def test_unqualified_public_manifest_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "qualified reviewed"):
            release.check_public_assets(self.api, self.state)


class DraftDownloadTests(unittest.TestCase):
    def test_direct_download_uses_recorded_id_and_read_only_token(self):
        response = io.BytesIO(b"qualified")
        with patch.dict(os.environ, {"GH_TOKEN": "readonly"}), patch.object(packages.urllib.request, "build_opener") as opener:
            opener.return_value.open.return_value = response
            self.assertEqual(packages.download_draft(CORE, 100), b"qualified")
            request = opener.return_value.open.call_args.args[0]
            self.assertEqual(request.full_url, f"https://api.github.com/repos/{CORE}/releases/assets/100")
            self.assertEqual(request.headers["Authorization"], "Bearer readonly")

    def test_redirect_download_never_forwards_authorization(self):
        redirect = urllib.error.HTTPError("https://api.github.com", 302, "redirect",
                                         {"Location": "https://release-assets.githubusercontent.com/signed"}, None)
        with patch.dict(os.environ, {"GH_TOKEN": "readonly"}), \
             patch.object(packages.urllib.request, "build_opener") as opener, \
             patch.object(packages.urllib.request, "urlopen", return_value=io.BytesIO(b"bytes")) as download:
            opener.return_value.open.side_effect = redirect
            self.assertEqual(packages.download_draft(CORE, 100), b"bytes")
            self.assertEqual(download.call_args.args, ("https://release-assets.githubusercontent.com/signed",))
            self.assertNotIn("readonly", str(download.call_args))

    def test_denied_missing_network_and_untrusted_redirects_fail_closed(self):
        for code in (401, 403, 404, 500):
            error = urllib.error.HTTPError("https://api.github.com", code, "denied", {}, None)
            with patch.dict(os.environ, {"GH_TOKEN": "readonly"}), \
                 patch.object(packages.urllib.request, "build_opener") as opener, \
                 patch.object(packages.urllib.request, "urlopen") as fallback:
                opener.return_value.open.side_effect = error
                with self.assertRaisesRegex(ValueError, "no write-token fallback"):
                    packages.download_draft(CORE, 100)
                fallback.assert_not_called()
        for location in ("http://release-assets.githubusercontent.com/x", "https://evil.invalid/x"):
            with patch.dict(os.environ, {"GH_TOKEN": "readonly"}), \
                 patch.object(packages.urllib.request, "build_opener") as opener:
                opener.return_value.open.side_effect = urllib.error.HTTPError("", 302, "", {"Location": location}, None)
                with self.assertRaisesRegex(ValueError, "Untrusted"):
                    packages.download_draft(CORE, 100)

    def test_draft_native_gate_hashes_downloaded_bytes_and_writes_evidence(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = pathlib.Path(temporary)
            source = root / "source"
            source.mkdir()
            candidate = plan(status="draft")
            candidate["release_id"] = 44
            names = release.asset_names("notation-core-go", candidate["tag"])
            content = {name: gzip.compress(b"source") if name.endswith(".tar.gz") else b"checksums" for name in names}
            candidate["assets"] = [{"id": index + 100, "name": name, "size": len(data),
                                    "sha256": hashlib.sha256(data).hexdigest()} for index, (name, data) in enumerate(content.items())]
            by_id = {item["id"]: content[item["name"]] for item in candidate["assets"]}
            archive = next(name for name in names if name.endswith(".tar.gz"))
            with patch.object(packages, "assert_tag"), patch.object(packages, "check_release_assets"), \
                 patch.object(packages, "git", return_value=candidate["commit"]), \
                 patch.object(packages, "download_draft", side_effect=lambda repository, identifier: by_id[identifier]), \
                 patch.object(packages, "verify_checksums", return_value=[archive]), \
                 patch.object(packages, "library_source_archive", return_value=b"source"):
                packages.verify(candidate, source, "linux_amd64", root / "passed")
                evidence = json.loads((root / "passed/verified.json").read_text())
                self.assertTrue(evidence["draft_downloads"])
                self.assertFalse(evidence["public_downloads"])
                by_id[100] = b"tampered"
                with self.assertRaisesRegex(ValueError, "Downloaded bytes differ"):
                    packages.verify(candidate, source, "linux_amd64", root / "failed")
                self.assertFalse((root / "failed/verified.json").exists())
            by_id[100] = content[candidate["assets"][0]["name"]]
            candidate["status"] = "verifying"
            def anonymous(url, timeout):
                self.assertIsInstance(url, str)
                self.assertTrue(url.startswith(f"https://github.com/{CORE}/releases/download/"))
                return io.BytesIO(content[url.rsplit("/", 1)[1]])
            with patch.object(packages, "assert_tag"), patch.object(packages, "check_public_assets"), \
                 patch.object(packages, "git", return_value=candidate["commit"]), \
                 patch.object(packages, "download_draft") as authenticated, \
                 patch.object(packages.urllib.request, "urlopen", side_effect=anonymous), \
                 patch.object(packages, "verify_checksums", return_value=[archive]), \
                 patch.object(packages, "library_source_archive", return_value=b"source"):
                packages.verify(candidate, source, "linux_amd64", root / "public")
                authenticated.assert_not_called()
                self.assertTrue(json.loads((root / "public/verified.json").read_text())["public_downloads"])


if __name__ == "__main__":
    unittest.main()
