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
import os
import pathlib
import tempfile
import unittest
from unittest.mock import patch

import monthly_release as release
import monthly_release_pr as prs
import monthly_release_vote as voting
from test_monthly_release import checks, plan
import test_monthly_release_vote as fixtures

BASE, HEAD, TREE, MERGE = (character * 40 for character in "abcd")
REPO = "notaryproject/notation-core-go"


def expected():
    return {"repository": REPO, "actor": "test-actor", "base_ref": "release-1.3", "base_sha": BASE,
            "head_ref": "monthly-patch/2026-10/plan/dependencies", "head_sha": HEAD, "head_tree": TREE,
            "tag": "v1.3.1", "number": 7}


class API:
    def __init__(self):
        self.pull = {
            "number": 7, "state": "open", "draft": False, "user": {"login": "test-actor"},
            "base": {"ref": "release-1.3", "sha": BASE, "repo": {"full_name": REPO}},
            "head": {"ref": expected()["head_ref"], "sha": HEAD, "repo": {"full_name": REPO}},
        }
        self.tip = BASE
        self.commit = {
            "sha": MERGE, "parents": [{"sha": BASE}, {"sha": HEAD}],
            "commit": {"tree": {"sha": TREE}, "verification": {"verified": True}},
        }
        self.checks_result = checks(HEAD)
        self.writes = []

    def request(self, path, method="GET", payload=None):
        if method != "GET":
            self.writes.append((path, method, payload))
            if path.endswith("/merge"):
                self.pull.update(state="closed", merged_at="2026-10-13T00:00:00Z", merge_commit_sha=MERGE)
                self.tip = MERGE
                return {"merged": True, "sha": MERGE}
            raise AssertionError(path)
        if path == f"repos/{REPO}":
            return {"default_branch": "main"}
        if "/branches/" in path:
            return {"commit": {"sha": self.tip}}
        if "/pulls/" in path:
            return self.pull
        if "/commits/" in path:
            return self.commit
        raise AssertionError(path)

    def checks(self, repository, number):
        return self.checks_result


class ReleasePRTests(unittest.TestCase):
    def setUp(self):
        self.api = API()

    def test_backport_waits_for_normal_reviews_and_ci(self):
        self.api.checks_result["reviewDecision"] = "REVIEW_REQUIRED"
        self.assertIsNone(prs.merge_pull(self.api, expected(), self.api.pull))
        self.assertEqual(self.api.writes, [])

    def test_normal_merge_pins_head_and_preserves_merge_tip(self):
        result = prs.merge_pull(self.api, expected(), self.api.pull)
        self.assertEqual(result, MERGE)
        self.assertEqual(self.api.writes[0][2], {"sha": HEAD, "merge_method": "merge"})
        self.assertEqual(self.api.tip, MERGE)

    def test_concurrent_base_change_blocks_merge(self):
        self.api.tip = HEAD
        with self.assertRaisesRegex(ValueError, "advanced"):
            prs.merge_pull(self.api, expected(), self.api.pull)
        self.assertEqual(self.api.writes, [])

    def test_changed_head_or_actor_cannot_merge(self):
        for field, value in (("user", {"login": "attacker"}), ("head", {**self.api.pull["head"], "sha": BASE})):
            with self.subTest(field=field):
                changed = {**self.api.pull, field: value}
                with self.assertRaises(ValueError):
                    prs.merge_pull(self.api, expected(), changed)
        self.assertEqual(self.api.writes, [])

    def test_squash_wrong_tree_or_unsigned_merge_cannot_qualify(self):
        prs.merge_pull(self.api, expected(), self.api.pull)
        for commit in (
            {**self.api.commit, "parents": [{"sha": BASE}]},
            {**self.api.commit, "commit": {"tree": {"sha": BASE}, "verification": {"verified": True}}},
            {**self.api.commit, "commit": {"tree": {"sha": TREE}, "verification": {"verified": False}}},
        ):
            with patch.object(self.api, "request", side_effect=lambda path, method="GET", payload=None: commit if "/commits/" in path else {"commit": {"sha": MERGE}}), self.assertRaises(ValueError):
                prs.merge_pull(self.api, expected(), self.api.pull)

    def test_closed_unmerged_pr_is_not_skipped_or_published(self):
        self.api.pull["state"] = "closed"
        with self.assertRaisesRegex(ValueError, "closed without"):
            prs.merge_pull(self.api, expected(), self.api.pull)

    def test_required_version_update_has_no_other_source_changes(self):
        before = 'package signer\nconst signingAgent = "notation-go/1.3.0"\nfunc Sign() {}\n'
        self.assertEqual(prs.version_metadata("notaryproject/notation-go", before, "v1.3.1"),
                         before.replace("notation-go/1.3.0", "notation-go/1.3.1"))
        cli = 'var (\n\tVersion = "v1.3.0"\n\tBuildMetadata = "unreleased"\n\tGitCommit = ""\n)\n'
        self.assertEqual(prs.version_metadata("notaryproject/notation", cli, "v1.3.1"),
                         cli.replace("v1.3.0", "v1.3.1").replace('BuildMetadata = "unreleased"', 'BuildMetadata = ""'))
        with self.assertRaises(ValueError):
            prs.version_metadata(REPO, before, "v1.3.1")

    def test_ambiguous_version_declaration_is_rejected(self):
        for text in ("", 'Version = "v1"\nVersion = "v2"\nBuildMetadata = ""\n'):
            with self.assertRaises(ValueError):
                prs.version_metadata("notaryproject/notation", text, "v1.3.1")

    def test_core_release_uses_exact_dependency_head_not_merge_commit(self):
        candidate = plan(REPO)
        candidate.update(commit=HEAD, previous_head=BASE, controller_plan_id="e" * 64,
                         release_backport=expected())
        cycle = unittest.mock.Mock()
        with patch.dict(os.environ, {"MONTHLY_PATCH_ACTOR": "test-actor"}):
            result = prs.advance(self.api, candidate, pathlib.Path("."), {}, cycle)
        self.assertEqual(result["action"], "qualify")
        self.assertEqual(result["commit"], HEAD)
        self.assertEqual(result["previous_head"], MERGE)
        prs.validate(self.api, result)

    def test_canonical_tag_push_keeps_already_merged_release_tip(self):
        candidate = plan(REPO)
        candidate.update(commit=HEAD, previous_head=MERGE, release_backport=expected(), controller_issue=1)
        commands = []
        def git(command, directory, environment=None):
            commands.append(command)
            if command[:2] == ["rev-parse", "HEAD"] or command[:1] == ["rev-parse"]:
                return HEAD
            return ""
        cycle = unittest.mock.Mock()
        cycle.return_value.state = copy.deepcopy(candidate)
        with patch.object(release, "policy"), patch.object(release, "coordinator_authorization"), \
             patch.object(release, "Cycle", cycle), patch.object(release, "baseline_plan"), \
             patch.object(release, "signing_environment", return_value={}), \
             patch.object(prs, "validate"), patch.object(release, "assert_tag", return_value=TREE), \
             patch.object(release, "git", side_effect=git), patch.object(self.api, "optional", return_value=None, create=True):
            release.tag_candidate(self.api, candidate, pathlib.Path("."))
        push = next(command for command in commands if "push" in command)
        self.assertIn(f"{MERGE}:refs/heads/release-1.3", push)
        self.assertNotIn(f"{HEAD}:refs/heads/release-1.3", push)
        self.assertIn(f"--force-with-lease=refs/heads/release-1.3:{MERGE}", push)
        self.assertIn("--force-with-lease=refs/tags/v1.3.1:", push)

    def vote_fixture(self):
        fixture = fixtures.VoteTests()
        fixture.setUp()
        reader, responses = fixture.api()
        repo = fixture.expected["repository"]
        before, dependency_head = "1" * 40, "2" * 40
        backport = {
            "repository": repo, "number": 7, "actor": "robot", "base_ref": "release-1.3",
            "base_sha": before, "head_ref": "monthly/dependencies", "head_sha": dependency_head,
            "head_tree": fixtures.CANDIDATE_TREE, "merge_commit_sha": fixtures.BASE, "tag": "v1.3.1",
        }
        responses[f"repos/{repo}/pulls/7"] = {
            "number": 7, "user": {"login": "robot"}, "state": "closed", "merged_at": "2026-10-13T00:00:00Z",
            "base": {"ref": "release-1.3", "sha": before, "repo": {"full_name": repo}},
            "head": {"ref": "monthly/dependencies", "sha": dependency_head, "repo": {"full_name": repo}},
            "merge_commit_sha": fixtures.BASE,
        }
        responses[f"repos/{repo}/commits/{fixtures.BASE}"].update(
            parents=[{"sha": before}, {"sha": dependency_head}])
        responses[f"repos/{repo}/commits/{fixtures.BASE}"]["commit"]["verification"] = {"verified": True}
        responses[f"repos/{repo}/branches/release-1.3"] = {"commit": {"sha": fixtures.MERGE}}
        candidate = {**fixture.plan(), "mode": "execute", "status": "voting", "release_backport": backport}
        return fixture, reader, responses, candidate

    def test_existing_separate_vote_completes_without_recreating_or_rewinding_prs(self):
        fixture, reader, responses, candidate = self.vote_fixture()
        cycle = unittest.mock.Mock()
        result = prs.advance(reader, candidate, pathlib.Path("."), {}, cycle)
        self.assertEqual(result["action"], "qualify")
        self.assertEqual(result["commit"], fixtures.HEAD)
        self.assertEqual(result["previous_head"], fixtures.MERGE)
        self.assertEqual(result["release_vote"]["receipt"], fixture.expected["receipt"])
        prs.validate(reader, result)

    def test_separate_vote_waits_for_current_head_majority(self):
        fixture, reader, responses, candidate = self.vote_fixture()
        fixture.reviews.pop()
        result = prs.advance(reader, candidate, pathlib.Path("."), {}, unittest.mock.Mock())
        self.assertEqual(result["action"], "waiting")
        self.assertEqual(result["status"], "voting")
        self.assertIn("majority", result["reason"])

    def test_vote_stage_rejects_source_changes_even_after_majority_approval(self):
        fixture, reader, responses, candidate = self.vote_fixture()
        compare = responses[f"repos/{fixture.expected['repository']}/compare/{fixtures.BASE}...{fixtures.HEAD}"]
        compare["files"].append({"filename": "feature.go", "status": "added"})
        with self.assertRaises(voting.VoteError):
            prs.advance(reader, candidate, pathlib.Path("."), {}, unittest.mock.Mock())

    def test_reassessment_rebinds_only_an_unchanged_roster(self):
        for changed in (False, True):
            with self.subTest(changed=changed):
                candidate = plan(REPO, "execute", status="voting")
                candidate.update(controller_plan_id="1" * 64, maintainer_roster_sha=BASE,
                                 release_vote={"maintainer_roster_sha": BASE})
                authorization = {"plan_id": "2" * 64, "snapshot": {"automation_sha": HEAD}}
                cycle = unittest.mock.Mock(state=candidate)
                def roster(api, repository, path, sha):
                    return "new maintainers" if changed and sha == HEAD else "same maintainers"
                with patch.object(release, "policy"), patch.object(release, "coordinator_authorization", return_value=authorization), \
                     patch.object(release, "resume_cycle", return_value=("2026-10", "monthly")), \
                     patch.object(release, "Cycle", return_value=cycle), patch.object(prs, "content", side_effect=roster), \
                     patch.object(release, "signing_environment", return_value={}), \
                     patch.object(prs, "advance", return_value={**candidate, "action": "waiting"}) as advance:
                    if changed:
                        with self.assertRaisesRegex(ValueError, "new release vote"):
                            release.prepare(self.api, REPO, "execute", "2026-10", pathlib.Path("."), ".")
                        cycle.save.assert_not_called()
                        advance.assert_not_called()
                    else:
                        release.prepare(self.api, REPO, "execute", "2026-10", pathlib.Path("."), ".")
                        self.assertEqual(candidate["maintainer_roster_sha"], HEAD)
                        self.assertEqual(candidate["release_vote"]["maintainer_roster_sha"], HEAD)
                        cycle.save.assert_called_once()


if __name__ == "__main__":
    unittest.main()
