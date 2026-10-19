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
import unittest
from unittest.mock import patch

import monthly_release as release
import notation_release_controller as controller
import release_issue_topology as threads
from test_upstream_authorization import CORE, GO, CLI, UpstreamAPI, assessed, inventory
from test_monthly_release import plan as candidate_plan


class ReleaseThreadTests(unittest.TestCase):
    def setUp(self):
        env = patch.dict(os.environ, {"MONTHLY_PATCH_ACTOR": "test-actor"})
        env.start()
        self.addCleanup(env.stop)
        self.api = UpstreamAPI()
        self.plan = assessed(inventory((CORE, GO, CLI)))
        for repo in (CORE, GO, CLI):
            self.plan["snapshots"][repo]["pulls"][0]["title"] = f"Only {repo} [title](evil) @attacker"
        self.plan["plan_id"] = controller.digest({key: value for key, value in self.plan.items() if key != "plan_id"})
        self.controller = controller.Controller(self.api, "test-actor")

    def start(self):
        return self.controller.approve(self.plan, 22)

    def test_one_thread_per_repo_own_scope_dependency_links_only(self):
        issue, state = self.start()
        self.assertEqual({repo: len(items) for repo, items in self.api.issues_by_repo.items()},
                         {CORE: 1, GO: 1, CLI: 1})
        for repo in (CORE, GO):
            body = self.api.issues_by_repo[repo][0]["body"]
            self.assertIn(f"Only {repo}", body)
            for other in (CORE, GO, CLI):
                if other != repo:
                    self.assertNotIn(f"Only {other} ", body)
            self.assertNotIn("@attacker", body)
        visible = issue["body"].split("<summary>Technical plan", 1)[0]
        self.assertIn(f"Only {CLI}", visible)
        self.assertNotIn(f"Only {CORE}", visible)
        self.assertNotIn(f"Only {GO}", visible)
        self.assertIn(threads.issue_url(CORE, state["release_issues"][CORE]), visible)
        self.assertIn(threads.issue_url(GO, state["release_issues"][GO]), visible)
        self.assertNotIn("/approve-plan", visible)
        self.assertIn("/approve", visible)
        self.assertIn("Required CI and reviews", self.api.issues_by_repo[CORE][0]["body"])

    def test_retries_do_not_duplicate_threads_or_accept_ambiguous_matches(self):
        self.start()
        self.start()
        self.assertTrue(all(len(items) == 1 for items in self.api.issues_by_repo.values()))
        duplicate = copy.deepcopy(self.api.issues_by_repo[CORE][0])
        duplicate["number"] = 98
        self.api.issues_by_repo[CORE].append(duplicate)
        with self.assertRaisesRegex(ValueError, "Duplicate"):
            self.start()

    def test_creation_interruption_recovers_exact_placeholder_without_circular_hash(self):
        original = self.api.request
        def failing(path, method="GET", payload=None):
            if path == f"repos/{GO}/issues" and method == "POST":
                raise RuntimeError("interrupted Go issue creation")
            return original(path, method, payload)
        with patch.object(self.api, "request", side_effect=failing), self.assertRaises(RuntimeError):
            self.start()
        self.assertEqual(len(self.api.issues_by_repo[CORE]), 1)
        self.assertIn("wiring is incomplete", self.api.issues_by_repo[CORE][0]["body"])
        issue, state = self.start()
        self.assertEqual(state["release_issues"][CLI], issue["number"])
        self.assertTrue(all(len(items) == 1 for items in self.api.issues_by_repo.values()))

    def test_foreign_author_tampered_body_or_missing_route_is_rejected(self):
        issue, state = self.start()
        original = copy.deepcopy(self.api.issues_by_repo[CORE])
        for change in ({"body": "forged"}, {"user": {"login": "attacker"}}, {"state": "closed"}):
            self.api.issues_by_repo[CORE] = [{**original[0], **change}]
            with self.subTest(change=change), self.assertRaises(ValueError):
                self.controller.save(issue, state)
        self.api.issues_by_repo[CORE] = []
        with self.assertRaises(ValueError):
            self.controller.save(issue, state)

    def test_simple_approve_is_bound_to_revision_not_an_old_generic_comment(self):
        issue, state = self.start()
        self.api.approve_comment(issue, state)
        old_id = next(iter(self.api.comments))
        replacement = copy.deepcopy(self.plan)
        replacement["decisions"][CORE]["reason"] = "Changed immutable assessment"
        replacement["plan_id"] = controller.digest({k: v for k, v in replacement.items() if k != "plan_id"})
        issue, state = self.controller.replan(issue, state, replacement, 23)
        self.assertEqual(state["review_after_comment_id"], old_id)
        self.assertIsNone(controller.approval_comment(self.api, issue, state))
        self.api.approve_comment(issue, state)
        receipt = controller.approval_comment(self.api, issue, state)
        self.assertEqual(receipt["plan_id"], state["plan"]["plan_id"])
        self.assertGreater(receipt["comment_id"], old_id)

    def test_failure_targets_own_maintainers_and_only_links_dependency_blockers(self):
        issue, state = self.start()
        paths = []
        original = self.api.request
        def recording(path, method="GET", payload=None):
            paths.append(path)
            return original(path, method, payload)
        with patch.object(self.api, "request", side_effect=recording):
            self.controller.report_failure(issue, state, CORE, {"id": 42}, "source <script> failure", ["qualify"])
            self.controller.report_failure(issue, state, CORE, {"id": 42}, "source <script> failure", ["qualify"])
        self.assertTrue(any(path.startswith(f"repos/{CORE}/contents/MAINTAINERS?") for path in paths))
        self.assertFalse(any(path.startswith(f"repos/{CLI}/contents/MAINTAINERS?") for path in paths))
        self.assertEqual(len(self.api.comments), 3)
        for item in self.api.comments.values():
            if f"/repos/{CORE}/" in item["issue_url"]:
                self.assertIn("source &lt;script&gt; failure", item["body"])
                self.assertIn("@maintainer", item["body"])
            else:
                self.assertIn(f"Blocked by [{CORE}", item["body"])
                self.assertNotIn("source", item["body"])
                self.assertNotIn("@maintainer", item["body"])

    def test_qualified_draft_comments_in_existing_release_issue_not_another_thread(self):
        issue, state = self.start()
        candidate = candidate_plan(status="draft")
        candidate.update(controller_issue=issue["number"], controller_plan_id=state["plan"]["plan_id"],
                         release_issue=state["release_issues"][CORE], release_id=44, assets=[])
        release.Cycle(self.api, CORE, state["month"], "execute").save(candidate)
        counts = {repo: len(items) for repo, items in self.api.issues_by_repo.items()}
        before = self.api.issues_by_repo[CORE][0]["body"]
        with patch.dict(os.environ, {"GITHUB_RUN_ID": "123", "MONTHLY_PATCH_WORKER_SHA": "a" * 40}), \
                patch.object(release, "policy"), patch.object(release, "coordinator_authorization"), \
                patch.object(release, "assert_tag"), patch.object(release, "check_release_assets"):
            result = release.complete(self.api, candidate)
            self.assertEqual(result["draft_review_issue"], state["release_issues"][CORE])
            self.assertEqual(len(self.api.comments), 1)
            self.assertIn("Publish release", next(iter(self.api.comments.values()))["body"])
            self.assertEqual(release.check_draft_review(self.api, result)["body"], before)
            comment = self.api.comments[result["draft_review_comment"]]
            comment["body"] += "Unreviewed mutation"
            with self.assertRaisesRegex(ValueError, "comment changed"):
                release.check_draft_review(self.api, result)
        self.assertEqual({repo: len(items) for repo, items in self.api.issues_by_repo.items()}, counts)


class BugfixPlanIntegrationTests(unittest.TestCase):
    def setUp(self):
        import test_release_bugfix_proposals as fixtures
        self.evidence = fixtures.collect()
        self.inventory = inventory((CLI,))
        self.inventory["repositories"][CLI]["bugfix_evidence"] = self.evidence
        self.assessment = {
            "schema": 1, "month": self.inventory["month"],
            "inventory_sha256": controller.digest(self.inventory),
            "decisions": {repo: {"decision": "release" if repo == CLI else "skip",
                                 "reason": "Dependabot scope only", "evidence": [f"{repo}:dependencies"]}
                          for repo in (CORE, GO, CLI)},
            "bugfix_proposals": {CORE: None, GO: None,
                                CLI: fixtures.assessment(self.evidence)},
        }
        self.assessment["bugfix_proposals"][CLI]["proposals"] += fixtures.assessment(self.evidence, index=1)["proposals"]

    def test_proposals_bound_to_evidence_not_added_to_worker_scope(self):
        result = controller.validate_assessment(self.inventory, self.assessment)
        self.assertEqual(result["snapshots"][CLI]["pulls"][0]["number"], 1)
        self.assertEqual(len(result["snapshots"][CLI]["pulls"]), 1)
        self.assertTrue(all(item["blocked"] and item["automatic_inclusion"] is False
                            for item in result["bugfix_proposals"][CLI]["proposals"]))
        routes = {CORE: 20, GO: 16, CLI: 26}
        self.assertIn("Proposed bug-fix backports", threads.scope_body(result, CLI, routes))
        self.assertNotIn("Proposed bug-fix backports", threads.scope_body(result, CORE, routes))
        self.assertIn("not included or executable", threads.scope_body(result, CLI, routes))

    def test_missing_assessment_fabricated_source_or_changed_main_rejected(self):
        del self.assessment["bugfix_proposals"]
        with self.assertRaisesRegex(ValueError, "all three"):
            controller.validate_assessment(self.inventory, self.assessment)
        self.setUp()
        self.assessment["bugfix_proposals"][CLI]["proposals"][0]["source_sha"] = "f" * 40
        with self.assertRaises(ValueError):
            controller.validate_assessment(self.inventory, self.assessment)
        self.setUp()
        self.inventory["repositories"][CLI]["bugfix_evidence"]["main_head"] = "f" * 40
        with self.assertRaisesRegex(ValueError, "source/release"):
            controller.validate_assessment(self.inventory, self.assessment)

    def test_all_new_main_commits_must_be_classified_not_silently_ignored(self):
        self.assessment["bugfix_proposals"][CLI]["proposals"] = []
        with self.assertRaisesRegex(ValueError, "every observed"):
            controller.validate_assessment(self.inventory, self.assessment)

    def test_features_and_already_included_reverted_candidates_are_not_proposed(self):
        proposals = self.assessment["bugfix_proposals"][CLI]["proposals"]
        proposals[0]["classification"] = "feature"
        proposals[1]["classification"] = "not_bugfix"
        result = controller.validate_assessment(self.inventory, self.assessment)
        self.assertNotIn("Proposed bug-fix backports", threads.scope_body(result, CLI, {CORE: 20, GO: 16, CLI: 26}))


if __name__ == "__main__":
    unittest.main()
