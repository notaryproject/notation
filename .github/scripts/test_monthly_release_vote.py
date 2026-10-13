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
import unittest

import monthly_release_vote as voting


BASE, HEAD, TREE, CANDIDATE_TREE, MERGE, ROSTER = [char * 40 for char in "abcdef"]
MAINTAINERS = """# Org-Level Maintainers (in alphabetical order)
Org Person <org@example.com> (@org)
# Repo-Level Maintainers (in alphabetical order)
# Note: This is for the repo
A Person <a@example.com> (@alice)
B Person <b@example.com> (@bob)
C Person <c@example.com> (@carol)
# Emeritus Repo-Level Maintainers (in alphabetical order)
Old Person <old@example.com> (@old)
"""


class VoteTests(unittest.TestCase):
    def setUp(self):
        self.expected = {
            "repository": "notaryproject/notation-go", "number": 42, "actor": "robot",
            "base_ref": "release-1.3", "base_sha": BASE,
            "head_ref": "monthly/2026-10/vote", "head_sha": HEAD, "head_tree": TREE,
            "candidate_sha": BASE, "candidate_tree": CANDIDATE_TREE, "tag": "v1.3.1",
            "signer_login": "robot", "merge_committer_login": "web-flow",
            "maintainer_roster_sha": ROSTER,
            "metadata_before": 'package signer\nconst signingAgent = "notation-go/1.3.0"\n',
            "metadata_after": 'package signer\nconst signingAgent = "notation-go/1.3.1"\n',
        }
        self.pr = {
            "number": 42, "state": "open", "draft": False, "user": {"login": "robot"},
            "base": {"ref": self.expected["base_ref"], "sha": BASE,
                     "repo": {"full_name": self.expected["repository"]}},
            "head": {"ref": self.expected["head_ref"], "sha": HEAD,
                     "repo": {"full_name": self.expected["repository"]}},
            "body": f"https://github.com/notaryproject/notation-go/blob/{BASE}/MAINTAINERS\n"
                    f"https://github.com/notaryproject/notation-go/compare/v1.3.0...v1.3.1",
            "merged_at": None, "merge_commit_sha": None,
        }
        self.reviews = [self.review("alice", 1), self.review("bob", 2)]
        self.diff = {
            "files": [{"filename": "signer/signer.go", "status": "modified"}],
            "before": self.expected["metadata_before"], "after": self.expected["metadata_after"],
        }

    def review(self, login, ident, state="APPROVED", head=HEAD):
        return {"id": ident, "user": {"login": login}, "state": state, "commit_id": head,
                "submitted_at": f"2026-10-13T05:00:{ident:02d}Z"}

    def assess(self):
        return voting.assess_vote(self.expected, self.pr, self.reviews, MAINTAINERS, self.diff)

    def merge(self):
        self.pr.update(state="closed", merged_at="2026-10-13T06:00:00Z", merge_commit_sha=MERGE)
        return {"sha": MERGE, "parents": [{"sha": BASE}, {"sha": HEAD}],
                "commit": {"tree": {"sha": TREE},
                           "verification": {"verified": True, "reason": "valid"}},
                "committer": {"login": "web-flow"}}

    def test_strict_majority_and_roster(self):
        result = self.assess()
        self.assertEqual(result["status"], "ready")
        self.assertEqual(result["required"], 2)
        self.assertEqual(result["maintainers"], ["alice", "bob", "carol"])

    def test_insufficient_votes_wait_with_evidence(self):
        self.reviews = self.reviews[:1]
        result = self.assess()
        self.assertEqual(result["status"], "waiting")
        self.assertEqual(result["approvals"], ["alice"])
        self.assertIn("majority", result["reason"])

    def test_org_and_emeritus_votes_do_not_count(self):
        self.reviews = [self.review("org", 1), self.review("old", 2)]
        self.assertEqual(self.assess()["approvals"], [])

    def test_case_insensitive_unique_votes(self):
        self.reviews.append(self.review("ALICE", 3))
        self.assertEqual(self.assess()["approvals"], ["alice", "bob"])

    def test_latest_review_supersedes_even_if_results_unordered(self):
        self.reviews.insert(0, self.review("alice", 3, "CHANGES_REQUESTED"))
        result = self.assess()
        self.assertEqual(result["status"], "waiting")
        self.assertEqual(result["blocking"], ["alice"])

    def test_dismissal_revokes_vote(self):
        self.reviews.append(self.review("alice", 3, "DISMISSED"))
        self.assertEqual(self.assess()["approvals"], ["bob"])

    def test_comment_does_not_replace_vote(self):
        self.reviews.append(self.review("alice", 3, "COMMENTED"))
        self.assertEqual(self.assess()["status"], "ready")

    def test_stale_approval_never_counts(self):
        self.reviews[0]["commit_id"] = BASE
        self.assertEqual(self.assess()["status"], "waiting")

    def test_missing_and_malformed_roster_fail(self):
        for text in ("", "# Org-Level Maintainers\nA <a@x> (@alice)",
                     "# Repo-Level Maintainers\nbad", MAINTAINERS.replace("(@bob)", "(@alice)")):
            with self.subTest(text=text), self.assertRaises(voting.VoteError):
                voting.parse_repo_maintainers(text)

    def test_roster_section_not_confused_with_emeritus(self):
        text = MAINTAINERS.replace("# Repo-Level Maintainers", "# Emeritus Repo-Level Maintainers")
        with self.assertRaises(voting.VoteError):
            voting.parse_repo_maintainers(text)

    def test_review_corruption_rejected(self):
        for field, value in (("id", 1), ("submitted_at", "bad"),
                             ("submitted_at", "2026-01-01T00:00:00"), ("state", "PENDING")):
            with self.subTest(field=field, value=value):
                original = copy.deepcopy(self.reviews)
                self.reviews[1][field] = value
                with self.assertRaises(voting.VoteError):
                    self.assess()
                self.reviews = original

    def test_changed_pr_identity_rejected(self):
        for side, field, value in (
            ("head", "sha", BASE), ("base", "sha", HEAD),
            ("head", "ref", "attacker"), ("base", "ref", "main"),
            ("head", "repo", {"full_name": "fork/notation-go"}),
            ("base", "repo", {"full_name": "fork/notation-go"}),
        ):
            with self.subTest(side=side, field=field):
                pr = copy.deepcopy(self.pr)
                self.pr[side][field] = value
                with self.assertRaises(voting.VoteError):
                    self.assess()
                self.pr = pr
        for field, value in (("number", 43), ("user", {"login": "attacker"}), ("body", "")):
            with self.subTest(field=field):
                pr = copy.deepcopy(self.pr)
                self.pr[field] = value
                with self.assertRaises(voting.VoteError):
                    self.assess()
                self.pr = pr

    def test_draft_waits_and_closed_unmerged_rejects(self):
        self.pr["draft"] = True
        self.assertEqual(self.assess()["status"], "waiting")
        self.pr.update(draft=False, state="closed")
        with self.assertRaises(voting.VoteError):
            self.assess()

    def test_source_change_rejected_even_when_pinned(self):
        self.expected["metadata_after"] += "func arbitrary() {}\n"
        self.diff["after"] = self.expected["metadata_after"]
        with self.assertRaises(voting.VoteError):
            self.assess()

    def test_other_file_or_renamed_metadata_rejected(self):
        for files in ([{"filename": "go.mod", "status": "modified"}],
                      [{"filename": "signer/signer.go", "status": "renamed"}],
                      self.diff["files"] + [{"filename": "go.sum", "status": "modified"}]):
            with self.subTest(files=files), self.assertRaises(voting.VoteError):
                voting.assess_vote(self.expected, self.pr, self.reviews, MAINTAINERS,
                                   {**self.diff, "files": files})

    def test_wrong_metadata_version_rejected(self):
        self.expected["metadata_after"] = self.diff["after"] = self.diff["before"]
        with self.assertRaises(voting.VoteError):
            self.assess()

    def test_cli_version_and_build_metadata_only(self):
        self.expected["repository"] = "notaryproject/notation"
        self.expected["metadata_before"] = 'package version\nvar (\n\tVersion = "v1.3.0"\n\tBuildMetadata = "unreleased"\n\tGitCommit = ""\n)\n'
        self.expected["metadata_after"] = 'package version\nvar (\n\tVersion = "v1.3.1"\n\tBuildMetadata = ""\n\tGitCommit = ""\n)\n'
        for side in ("head", "base"):
            self.pr[side]["repo"]["full_name"] = self.expected["repository"]
        self.pr["body"] = self.pr["body"].replace("notation-go/", "notation/")
        self.diff = {"files": [{"filename": "internal/version/version.go", "status": "modified"}],
                     "before": self.expected["metadata_before"],
                     "after": self.expected["metadata_after"]}
        self.assertEqual(self.assess()["status"], "ready")
        self.diff["after"] = self.expected["metadata_after"] = self.diff["after"].replace(
            'GitCommit = ""', 'GitCommit = "unexpected"')
        with self.assertRaises(voting.VoteError):
            self.assess()

    def test_core_requires_declared_candidate_scope_without_metadata(self):
        self.expected.update(repository="notaryproject/notation-core-go",
                             vote_scope="dependency-candidate", candidate_sha=HEAD,
                             candidate_tree=TREE)
        for side in ("head", "base"):
            self.pr[side]["repo"]["full_name"] = self.expected["repository"]
        self.pr["body"] = self.pr["body"].replace("notation-go/", "notation-core-go/")
        self.diff = {"files": [], "before": None, "after": None}
        self.assertEqual(self.assess()["status"], "ready")
        del self.expected["vote_scope"]
        with self.assertRaises(voting.VoteError):
            self.assess()

    def test_normal_verified_merge_receipt(self):
        commit = self.merge()
        receipt = voting.qualify_release_commit(self.expected, self.pr, commit)
        self.assertEqual(receipt["sha"], HEAD)
        self.assertEqual(receipt["merge_commit_sha"], MERGE)
        self.assertEqual(receipt["tree"], TREE)

    def test_merge_provenance_violations(self):
        commit = self.merge()
        mutations = [
            {"parents": [{"sha": BASE}]}, {"parents": [{"sha": HEAD}, {"sha": BASE}]},
            {"parents": [{"sha": BASE}, {"sha": HEAD}, {"sha": ROSTER}]},
            {"sha": BASE}, {"committer": {"login": "attacker"}},
            {"commit": {"tree": {"sha": CANDIDATE_TREE},
                        "verification": {"verified": True, "reason": "valid"}}},
            {"commit": {"tree": {"sha": TREE},
                        "verification": {"verified": False, "reason": "unsigned"}}},
        ]
        for mutation in mutations:
            with self.subTest(mutation=mutation), self.assertRaises(voting.VoteError):
                voting.qualify_release_commit(self.expected, self.pr, {**commit, **mutation})

    def test_unauthorized_expected_merge_committer_rejected(self):
        commit = self.merge()
        self.expected["merge_committer_login"] = "attacker"
        commit["committer"]["login"] = "attacker"
        with self.assertRaisesRegex(voting.VoteError, "configured actor"):
            voting.qualify_release_commit(self.expected, self.pr, commit)

    def test_invalid_pinned_plan_identity(self):
        for key, value in (("head_sha", "unknown"), ("tag", "v1.4.0"),
                           ("base_ref", "main"), ("candidate_sha", HEAD),
                           ("repository", "fork/notation-go"), ("number", 0),
                           ("head_ref", "")):
            with self.subTest(key=key):
                expected = copy.deepcopy(self.expected)
                self.expected[key] = value
                with self.assertRaises(voting.VoteError):
                    self.assess()
                self.expected = expected

    def api(self):
        merge = self.merge()
        self.expected["receipt"] = voting.qualify_release_commit(self.expected, self.pr, merge)
        repo = self.expected["repository"]
        def content(text):
            return {"type": "file", "encoding": "base64",
                    "content": base64.b64encode(text.encode()).decode()}
        signed = {"verified": True, "reason": "valid"}
        responses = {
            f"repos/{repo}/pulls/42": self.pr,
            f"repos/{repo}/contents/MAINTAINERS?ref={ROSTER}": content(MAINTAINERS),
            f"repos/{repo}/contents/signer/signer.go?ref={BASE}":
                content(self.expected["metadata_before"]),
            f"repos/{repo}/contents/signer/signer.go?ref={HEAD}":
                content(self.expected["metadata_after"]),
            f"repos/{repo}/compare/{BASE}...{HEAD}":
                {"status": "ahead", "ahead_by": 1, "behind_by": 0, "files": self.diff["files"]},
            f"repos/{repo}/commits/{BASE}":
                {"sha": BASE, "commit": {"tree": {"sha": CANDIDATE_TREE}}},
            f"repos/{repo}/commits/{HEAD}":
                {"sha": HEAD, "parents": [{"sha": BASE}], "committer": {"login": "robot"},
                 "commit": {"tree": {"sha": TREE}, "verification": signed,
                            "message": "bump: release v1.3.1"}},
            f"repos/{repo}/commits/{MERGE}": merge,
            f"repos/{repo}/git/ref/heads/release-1.3": {"object": {"sha": MERGE}},
        }
        class API:
            def request(api, path):
                return responses[path]
            def pages(api, path):
                self.assertEqual(path, f"repos/{repo}/pulls/42/reviews")
                return iter(self.reviews)
        return API(), responses

    def plan(self):
        return {"repository": self.expected["repository"], "tag": self.expected["tag"],
                "commit": HEAD, "previous_head": MERGE, "baseline": "v1.3.0",
                "maintainer_roster_sha": ROSTER,
                "release_vote": self.expected}

    def test_revalidate_all_evidence_before_publish(self):
        api, _ = self.api()
        self.assertEqual(voting.validate_vote(api, self.plan()), self.expected["receipt"])

    def test_revalidate_changed_tip_or_votes_or_receipt(self):
        api, responses = self.api()
        path = "repos/notaryproject/notation-go/git/ref/heads/release-1.3"
        responses[path]["object"]["sha"] = HEAD
        with self.assertRaisesRegex(voting.VoteError, "branch moved"):
            voting.validate_vote(api, self.plan())
        responses[path]["object"]["sha"] = MERGE
        self.reviews.append(self.review("alice", 3, "DISMISSED"))
        with self.assertRaisesRegex(voting.VoteError, "majority"):
            voting.validate_vote(api, self.plan())
        self.reviews.pop()
        self.expected["receipt"]["tree"] = CANDIDATE_TREE
        with self.assertRaisesRegex(voting.VoteError, "receipt changed"):
            voting.validate_vote(api, self.plan())

    def test_revalidate_unsigned_head_and_changed_candidate(self):
        api, responses = self.api()
        head = responses[f"repos/notaryproject/notation-go/commits/{HEAD}"]
        head["commit"]["verification"]["verified"] = False
        with self.assertRaisesRegex(voting.VoteError, "signature"):
            voting.validate_vote(api, self.plan())
        head["commit"]["verification"]["verified"] = True
        responses[f"repos/notaryproject/notation-go/commits/{BASE}"]["commit"]["tree"]["sha"] = TREE
        with self.assertRaisesRegex(voting.VoteError, "candidate tree"):
            voting.validate_vote(api, self.plan())

    def test_revalidate_uses_only_assessed_current_roster(self):
        api, _ = self.api()
        plan = self.plan()
        plan["maintainer_roster_sha"] = HEAD
        with self.assertRaisesRegex(voting.VoteError, "trusted maintainer roster"):
            voting.validate_vote(api, plan)

    def test_revalidate_rejects_missing_assessed_roster(self):
        api, _ = self.api()
        plan = self.plan()
        del plan["maintainer_roster_sha"]
        with self.assertRaisesRegex(voting.VoteError, "trusted maintainer roster"):
            voting.validate_vote(api, plan)

    def test_revalidate_rejects_extra_metadata_commits(self):
        api, responses = self.api()
        responses[f"repos/notaryproject/notation-go/compare/{BASE}...{HEAD}"]["ahead_by"] = 2
        with self.assertRaisesRegex(voting.VoteError, "candidate scope"):
            voting.validate_vote(api, self.plan())

    def test_revalidate_rejects_changed_metadata_parent(self):
        api, responses = self.api()
        responses[f"repos/notaryproject/notation-go/commits/{HEAD}"]["parents"] = [{"sha": ROSTER}]
        with self.assertRaisesRegex(voting.VoteError, "parent changed"):
            voting.validate_vote(api, self.plan())

    def test_revalidate_rejects_nonconventional_voting_commit(self):
        api, responses = self.api()
        responses[f"repos/notaryproject/notation-go/commits/{HEAD}"]["commit"]["message"] = "release"
        with self.assertRaisesRegex(voting.VoteError, "conventional"):
            voting.validate_vote(api, self.plan())

    def test_revalidate_rejects_tagging_merge_instead_of_voted_head(self):
        api, _ = self.api()
        plan = self.plan()
        plan["commit"] = MERGE
        with self.assertRaisesRegex(voting.VoteError, "voted head"):
            voting.validate_vote(api, plan)

    def test_app_actor_and_nonroster_bot_review(self):
        self.expected["actor"] = "release-app[bot]"
        self.pr["user"]["login"] = "release-app[bot]"
        self.reviews.append(self.review("dependabot[bot]", 3))
        self.assertEqual(self.assess()["approvals"], ["alice", "bob"])

    def test_vote_can_include_unmerged_qualified_dependency_candidate(self):
        candidate_sha = "1" * 40
        self.expected["candidate_sha"] = candidate_sha
        api, responses = self.api()
        root = "repos/notaryproject/notation-go"
        candidate = copy.deepcopy(responses[f"{root}/commits/{BASE}"])
        candidate.update(sha=candidate_sha, parents=[{"sha": BASE}],
                         committer={"login": "robot"})
        candidate["commit"]["verification"] = {"verified": True, "reason": "valid"}
        responses[f"{root}/commits/{candidate_sha}"] = candidate
        responses[f"{root}/compare/{candidate_sha}...{HEAD}"] = responses.pop(
            f"{root}/compare/{BASE}...{HEAD}")
        responses[f"{root}/contents/signer/signer.go?ref={candidate_sha}"] = responses.pop(
            f"{root}/contents/signer/signer.go?ref={BASE}")
        responses[f"{root}/commits/{HEAD}"]["parents"] = [{"sha": candidate_sha}]
        self.assertEqual(voting.validate_vote(api, self.plan())["sha"], HEAD)
        candidate["parents"] = [{"sha": ROSTER}]
        with self.assertRaisesRegex(voting.VoteError, "candidate parent"):
            voting.validate_vote(api, self.plan())


if __name__ == "__main__":
    unittest.main()
