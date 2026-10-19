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
import re
import unittest

import release_bugfix_proposals as proposals
from monthly_release import APIError


REPOSITORY = "notaryproject/notation"
MAIN, RELEASE, BASE, FIRST, OTHER = (letter * 40 for letter in "abcde")


def changed_file():
    return {
        "filename": "verify.go", "status": "modified",
        "additions": 1, "deletions": 1, "changes": 2,
        "patch": "@@ -1 +1 @@\n-old\n+new",
    }


def history_record(sha, message="Source behavior changed"):
    return {"sha": sha, "commit": {"message": message}}


class FixtureAPI:
    """A strict GET-only fixture. Unknown requests are test failures."""

    def __init__(self, sources=None, status="diverged", release_head=RELEASE):
        self.sources = [FIRST, MAIN] if sources is None else sources
        self.release_head = release_head
        self.status = status
        self.behind = 1 if status in {"diverged", "behind"} else 0
        self.merge_base = (MAIN if status == "behind" else
                           BASE if self.behind else release_head)
        self.history = [history_record(release_head), history_record(BASE)]
        self.calls = []
        self.overrides = {}
        self.main_reads = 0
        self.drift_after = None
        self.commits = {}
        self.pulls = {}
        for number, sha in enumerate(self.sources, 1):
            self.commits[sha] = {
                "sha": sha, "parents": [{"sha": BASE}],
                "commit": {"message": "Return an error for invalid signature input"},
                "stats": {"additions": 1, "deletions": 1},
                "files": [changed_file()],
            }
            self.pulls[number] = {
                "number": number, "merged_at": "2026-10-18T10:00:00Z",
                "merge_commit_sha": sha,
                "base": {"repo": {"full_name": REPOSITORY}, "ref": "main", "sha": BASE},
                "head": {"sha": sha}, "changed_files": 1,
                "title": "feature: misleading title", "body": "Review the implementation.",
            }
        self.associated = {
            sha: [{"number": number}] for number, sha in enumerate(self.sources, 1)
        }

    def request(self, path, method="GET", payload=None):
        if method != "GET" or payload is not None:
            raise AssertionError("Evidence collection attempted a write")
        self.calls.append(path)
        prefix = f"repos/{REPOSITORY}/"
        if not path.startswith(prefix):
            raise AssertionError(f"Unexpected repository: {path}")
        path = path[len(prefix):]
        if path in self.overrides:
            result = self.overrides[path]
            if isinstance(result, Exception):
                raise result
            return copy.deepcopy(result)
        if path == "commits/main":
            self.main_reads += 1
            return {"sha": OTHER if self.drift_after == self.main_reads else MAIN}
        comparison = re.fullmatch(
            rf"compare/{self.release_head}\.\.\.{MAIN}\?per_page=100&page=(\d+)", path,
        )
        if comparison:
            page = int(comparison[1])
            return {
                "status": self.status, "ahead_by": len(self.sources),
                "behind_by": self.behind, "total_commits": len(self.sources),
                "base_commit": {"sha": self.release_head},
                "merge_base_commit": {"sha": self.merge_base},
                "commits": [history_record(sha) for sha in
                            self.sources[(page - 1) * 100:page * 100]],
            }
        history = re.fullmatch(rf"commits\?sha={self.release_head}&per_page=100&page=(\d+)", path)
        if history:
            page = int(history[1])
            return copy.deepcopy(self.history[(page - 1) * 100:page * 100])
        commit = re.fullmatch(r"commits/([a-f0-9]{40})\?per_page=100&page=1", path)
        if commit and commit[1] in self.commits:
            return copy.deepcopy(self.commits[commit[1]])
        associated = re.fullmatch(r"commits/([a-f0-9]{40})/pulls\?per_page=100&page=1", path)
        if associated and associated[1] in self.associated:
            return copy.deepcopy(self.associated[associated[1]])
        pull = re.fullmatch(r"pulls/(\d+)", path)
        if pull and int(pull[1]) in self.pulls:
            return copy.deepcopy(self.pulls[int(pull[1])])
        files = re.fullmatch(r"pulls/(\d+)/files\?per_page=100&page=1", path)
        if files and int(files[1]) in self.pulls:
            return [changed_file()]
        commits = re.fullmatch(r"pulls/(\d+)/commits\?per_page=100&page=1", path)
        if commits and int(commits[1]) in self.pulls:
            return [{"sha": self.pulls[int(commits[1])]["head"]["sha"]}]
        raise AssertionError(f"Unexpected request: {path}")


def collect(api=None):
    return proposals.collect(api or FixtureAPI(), REPOSITORY, MAIN, RELEASE)


def assessment(evidence, classification="candidate_bugfix", index=0):
    candidate = evidence["candidates"][index]
    return {
        "schema": 1, "evidence_sha256": proposals.digest(evidence),
        "proposals": [{
            "candidate_id": candidate["candidate_id"], "repository": REPOSITORY,
            "source_sha": candidate["source_sha"],
            "pr_number": candidate["pr_numbers"][0] if candidate["pr_numbers"] else None,
            "classification": classification,
            "evidence_ids": candidate["evidence_ids"][:],
            "rationale": "The source diff adds the error return for invalid signature input.",
            "applicability": {
                "status": "unverified",
                "rationale": "Release branch context and prerequisites need separate review.",
                "prerequisite_shas": [],
            },
        }],
    }


class CollectionTests(unittest.TestCase):
    def test_read_only_pinned_diff_evidence_without_title_classifier(self):
        api = FixtureAPI()
        evidence = collect(api)
        self.assertEqual(evidence["authority"], "proposal_only")
        self.assertFalse(evidence["automatic_inclusion"])
        self.assertEqual(evidence["comparison"]["merge_base"], BASE)
        self.assertEqual(evidence["comparison"]["status"], "diverged")
        self.assertEqual(api.main_reads, 2)
        candidate = evidence["candidates"][0]
        self.assertTrue(candidate["main_member"])
        self.assertNotIn("classification", candidate)
        self.assertEqual(candidate["blockers"], ["applicability_unverified"])
        commit, pull = (evidence["evidence"][item] for item in candidate["evidence_ids"])
        self.assertIn(FIRST, commit["files"][0]["url"])
        self.assertIn(FIRST, pull["immutable_diff_url"])
        self.assertIn("patch", commit["files"][0])
        self.assertEqual(pull["title"], "feature: misleading title")
        self.assertNotIn("/labels", "\n".join(api.calls))

    def test_ancestry_already_present_is_not_a_candidate(self):
        for status in ("identical", "behind"):
            with self.subTest(status=status):
                release = MAIN if status == "identical" else RELEASE
                api = FixtureAPI([], status, release)
                evidence = proposals.collect(api, REPOSITORY, MAIN, release)
                self.assertEqual(evidence["candidates"], [])
                self.assertEqual(evidence["comparison"]["source_shas"], [])
                self.assertFalse(any("/pulls" in path for path in api.calls))

    def test_ahead_comparison_uses_release_as_merge_base(self):
        evidence = collect(FixtureAPI(status="ahead"))
        self.assertEqual(evidence["comparison"]["merge_base"], RELEASE)

    def test_cherry_picked_equivalence_blocks_inclusion(self):
        api = FixtureAPI()
        api.history[0]["commit"]["message"] += f"\n\n(cherry picked from commit {FIRST})"
        evidence = collect(api)
        candidate = evidence["candidates"][0]
        self.assertEqual(candidate["inclusion"], "cherry_pick_provenance")
        self.assertEqual(candidate["equivalent_shas"], [RELEASE])
        with self.assertRaisesRegex(ValueError, "included"):
            proposals.validate_proposals(evidence, assessment(evidence))
        proposals.validate_proposals(evidence, assessment(evidence, "needs_review"))

    def test_main_cherry_pick_of_release_original(self):
        api = FixtureAPI()
        api.commits[FIRST]["commit"]["message"] += (
            f"\n\n(cherry picked from commit {RELEASE})"
        )
        self.assertIn("already_present_by_cherry_pick_provenance",
                      collect(api)["candidates"][0]["blockers"])

    def test_main_revert_and_revert_commit_require_review(self):
        api = FixtureAPI()
        api.commits[MAIN]["commit"]["message"] = f"Undo behavior\n\nThis reverts commit {FIRST}."
        evidence = collect(api)
        self.assertEqual(evidence["candidates"][0]["known_revert_shas"], [MAIN])
        self.assertIn("known_revert_requires_review", evidence["candidates"][0]["blockers"])
        self.assertIn("revert_commit_requires_review", evidence["candidates"][1]["blockers"])
        with self.assertRaises(ValueError):
            proposals.validate_proposals(evidence, assessment(evidence))

    def test_reverted_release_equivalent_is_not_treated_as_safe(self):
        api = FixtureAPI()
        api.history = [
            history_record(RELEASE, f"Undo pick\n\nThis reverts commit {OTHER}."),
            history_record(OTHER, f"Pick\n\n(cherry picked from commit {FIRST})"),
            history_record(BASE),
        ]
        candidate = collect(api)["candidates"][0]
        self.assertEqual(candidate["release_revert_shas"], [RELEASE])
        self.assertIn("release_equivalent_was_reverted", candidate["blockers"])

    def test_history_truncation_is_an_explicit_blocker(self):
        api = FixtureAPI()
        api.history = [history_record(RELEASE)] + [
            history_record(f"{index:040x}") for index in range(201)
        ]
        evidence = collect(api)
        self.assertFalse(evidence["release_history_complete"])
        self.assertEqual(len(evidence["release_provenance"]), proposals.MAX_HISTORY)
        self.assertIn("release_history_incomplete", evidence["candidates"][0]["blockers"])

    def test_comparison_paginates_and_checks_full_count(self):
        sources = [f"{index:040x}" for index in range(100)] + [MAIN]
        api = FixtureAPI(sources)
        evidence = collect(api)
        self.assertEqual(len(evidence["candidates"]), 101)
        self.assertTrue(any("compare/" in path and "page=2" in path for path in api.calls))
        api = FixtureAPI(sources)
        first_path = f"compare/{RELEASE}...{MAIN}?per_page=100&page=1"
        bad = api.request(f"repos/{REPOSITORY}/{first_path}")
        bad["commits"] = bad["commits"][:1]
        api.overrides[first_path] = bad
        with self.assertRaisesRegex(ValueError, "Incomplete comparison"):
            collect(api)

    def test_duplicate_or_drifting_comparison_fails_closed(self):
        for modification in ("duplicate", "count", "merge_base", "status"):
            api = FixtureAPI()
            path = f"compare/{RELEASE}...{MAIN}?per_page=100&page=1"
            bad = api.request(f"repos/{REPOSITORY}/{path}")
            if modification == "duplicate":
                bad["commits"][1] = bad["commits"][0]
            elif modification == "count":
                bad["total_commits"] += 1
            elif modification == "merge_base":
                bad["merge_base_commit"]["sha"] = "not-a-sha"
            else:
                bad["status"] = "ahead"
            api.overrides[path] = bad
            with self.subTest(modification=modification), self.assertRaises(ValueError):
                collect(api)

    def test_comparison_bound_does_not_silently_truncate(self):
        api = FixtureAPI([f"{index:040x}" for index in range(201)] + [MAIN])
        with self.assertRaisesRegex(ValueError, "commit bound"):
            collect(api)

    def test_main_drift_at_either_boundary_is_rejected(self):
        for boundary in (1, 2):
            api = FixtureAPI()
            api.drift_after = boundary
            with self.subTest(boundary=boundary), self.assertRaisesRegex(ValueError, "Main drifted"):
                collect(api)

    def test_fabricated_commit_response_is_rejected(self):
        api = FixtureAPI()
        api.commits[FIRST]["sha"] = OTHER
        with self.assertRaisesRegex(ValueError, "another SHA"):
            collect(api)

    def test_missing_binary_or_truncated_patch_is_blocked(self):
        for mode in ("missing", "oversized", "stats", "truncated_hunk", "truncated_changes"):
            api = FixtureAPI()
            if mode == "missing":
                del api.commits[FIRST]["files"][0]["patch"]
            elif mode == "oversized":
                api.commits[FIRST]["files"][0]["patch"] = "x" * (proposals.MAX_PATCH_BYTES + 1)
            elif mode == "stats":
                api.commits[FIRST]["stats"]["additions"] += 1
            elif mode == "truncated_hunk":
                api.commits[FIRST]["files"][0]["patch"] = "@@ -1,2 +1,2 @@\n-old\n+new"
            else:
                api.commits[FIRST]["files"][0]["additions"] += 1
            evidence = collect(api)
            with self.subTest(mode=mode):
                self.assertIn("source_diff_unavailable_or_incomplete",
                              evidence["candidates"][0]["blockers"])
                with self.assertRaises(ValueError):
                    proposals.validate_proposals(evidence, assessment(evidence))
                proposals.validate_proposals(evidence, assessment(evidence, "needs_review"))

    def test_pr_file_count_and_missing_patch_block_full_diff_claim(self):
        for mode in ("count", "patch"):
            api = FixtureAPI()
            if mode == "count":
                api.pulls[1]["changed_files"] = 2
            else:
                file = changed_file()
                del file["patch"]
                api.overrides["pulls/1/files?per_page=100&page=1"] = [file]
            with self.subTest(mode=mode):
                self.assertIn("pr_diff_unavailable_or_incomplete",
                              collect(api)["candidates"][0]["blockers"])

    def test_http_errors_are_not_swallowed(self):
        for path in (f"commits/{FIRST}?per_page=100&page=1",
                     "pulls/1/files?per_page=100&page=1",
                     f"commits/{FIRST}/pulls?per_page=100&page=1"):
            api = FixtureAPI()
            api.overrides[path] = APIError("HTTP 404 missing evidence")
            with self.subTest(path=path), self.assertRaises(APIError):
                collect(api)

    def test_only_associated_merged_canonical_main_pr_is_evidence(self):
        for field in ("unmerged", "repository", "branch", "membership"):
            api = FixtureAPI()
            if field == "unmerged":
                api.pulls[1]["merged_at"] = None
            elif field == "repository":
                api.pulls[1]["base"]["repo"]["full_name"] = "fork/notation"
            elif field == "branch":
                api.pulls[1]["base"]["ref"] = "release-1.3"
            else:
                api.pulls[1]["merge_commit_sha"] = MAIN
                api.pulls[1]["head"]["sha"] = OTHER
            with self.subTest(field=field):
                evidence = collect(api)
                self.assertEqual(evidence["candidates"][0]["pr_numbers"], [])
                self.assertIn("missing_or_ambiguous_merged_canonical_pr",
                              evidence["candidates"][0]["blockers"])
                with self.assertRaises(ValueError):
                    proposals.validate_proposals(evidence, assessment(evidence))

    def test_multiple_associated_merged_prs_are_ambiguous(self):
        api = FixtureAPI()
        api.pulls[2]["merge_commit_sha"] = FIRST
        api.associated[FIRST] += [{"number": 2}]
        candidate = collect(api)["candidates"][0]
        self.assertEqual(candidate["pr_numbers"], [1, 2])
        self.assertIn("missing_or_ambiguous_merged_canonical_pr", candidate["blockers"])

    def test_merge_commit_source_requires_manual_review(self):
        api = FixtureAPI()
        api.commits[FIRST]["parents"].append({"sha": OTHER})
        self.assertIn("non_single_parent_source", collect(api)["candidates"][0]["blockers"])

    def test_patch_validation_handles_context_multiple_hunks_and_no_newline(self):
        self.assertTrue(proposals._patch_complete(
            "@@ -1,2 +1,2 @@ function\n context\n-old\n+new\n"
            "@@ -9 +9 @@\n-before\n+after\n\\ No newline at end of file", 2, 2,
        ))
        for patch in ("@@ -1 +1 @@\n-old", "non-diff text", "@@ -1 +1 @@\n-old\n+new\n+extra"):
            with self.subTest(patch=patch):
                self.assertFalse(proposals._patch_complete(patch, 1, 1))

    def test_reject_noncanonical_repository_and_mutable_refs_without_requests(self):
        api = FixtureAPI()
        for repository, main, release in (
                ("fork/notation", MAIN, RELEASE),
                (REPOSITORY, "main", RELEASE),
                (REPOSITORY, MAIN, "release-1.3")):
            with self.subTest(repository=repository, main=main), self.assertRaises(ValueError):
                proposals.collect(api, repository, main, release)
        self.assertEqual(api.calls, [])


class ValidationTests(unittest.TestCase):
    def setUp(self):
        self.evidence = collect()
        self.proposed = assessment(self.evidence)

    def test_each_classification_is_only_a_blocked_proposal(self):
        for classification in proposals.CLASSIFICATIONS:
            with self.subTest(classification=classification):
                validated = proposals.validate_proposals(
                    self.evidence, assessment(self.evidence, classification),
                )
                result = validated["proposals"][0]
                self.assertEqual(result["classification"], classification)
                self.assertTrue(result["blocked"])
                self.assertFalse(result["automatic_inclusion"])
                self.assertEqual(result["authority"], "proposal_only")
                self.assertIn("applicability_unverified", result["blockers"])

    def test_fabricated_repository_source_pr_and_citations_are_rejected(self):
        for field, value in (
                ("repository", "fork/notation"), ("source_sha", OTHER),
                ("candidate_id", f"{REPOSITORY}@{OTHER}"), ("pr_number", 999),
                ("pr_number", True), ("evidence_ids", ["commit:" + "f" * 64]),
                ("classification", "include"), ("evidence_ids", [])):
            proposed = copy.deepcopy(self.proposed)
            proposed["proposals"][0][field] = value
            with self.subTest(field=field, value=value), self.assertRaises(ValueError):
                proposals.validate_proposals(self.evidence, proposed)

    def test_other_candidates_evidence_is_not_accepted(self):
        self.proposed["proposals"][0]["evidence_ids"] = (
            self.evidence["candidates"][1]["evidence_ids"]
        )
        with self.assertRaisesRegex(ValueError, "unrelated evidence"):
            proposals.validate_proposals(self.evidence, self.proposed)

    def test_bugfix_requires_both_commit_and_merged_pr_diff_citations(self):
        for prefix in ("commit:", "pr:"):
            proposed = copy.deepcopy(self.proposed)
            proposed["proposals"][0]["evidence_ids"] = [
                item for item in proposed["proposals"][0]["evidence_ids"]
                if item.startswith(prefix)
            ]
            with self.subTest(prefix=prefix), self.assertRaises(ValueError):
                proposals.validate_proposals(self.evidence, proposed)

    def test_snapshot_digest_binds_heads_and_evidence(self):
        for field in ("main_head", "release_head", "repository"):
            evidence = copy.deepcopy(self.evidence)
            evidence[field] = OTHER
            with self.subTest(field=field), self.assertRaisesRegex(ValueError, "snapshot"):
                proposals.validate_proposals(evidence, self.proposed)
        evidence = copy.deepcopy(self.evidence)
        record_id = evidence["candidates"][0]["evidence_ids"][0]
        evidence["evidence"][record_id]["files"][0]["patch"] = "invented evidence"
        proposed = assessment(evidence)
        with self.assertRaisesRegex(ValueError, "content changed"):
            proposals.validate_proposals(evidence, proposed)

    def test_conflicts_and_feature_prerequisites_add_blockers_not_authority(self):
        for status, prerequisites in (
                ("conflict", []), ("not_applicable", []),
                ("missing_prerequisite", [MAIN])):
            proposed = copy.deepcopy(self.proposed)
            proposed["proposals"][0]["applicability"] = {
                "status": status, "rationale": "The release lacks the source feature context.",
                "prerequisite_shas": prerequisites,
            }
            with self.subTest(status=status):
                result = proposals.validate_proposals(self.evidence, proposed)["proposals"][0]
                self.assertIn(f"applicability_{status}", result["blockers"])
                self.assertTrue(result["blocked"])
                self.assertFalse(result["automatic_inclusion"])

    def test_missing_applicability_or_fabricated_prerequisite_is_rejected(self):
        cases = [
            {"status": "verified", "rationale": "Safe", "prerequisite_shas": []},
            {"status": "missing_prerequisite", "rationale": "Feature", "prerequisite_shas": [OTHER]},
            {"status": "missing_prerequisite", "rationale": "Feature", "prerequisite_shas": []},
            {"status": "missing_prerequisite", "rationale": "Feature", "prerequisite_shas": [FIRST]},
            {"status": "unverified", "rationale": "", "prerequisite_shas": []},
            {},
        ]
        for applicability in cases:
            proposed = copy.deepcopy(self.proposed)
            proposed["proposals"][0]["applicability"] = applicability
            with self.subTest(applicability=applicability), self.assertRaises(ValueError):
                proposals.validate_proposals(self.evidence, proposed)

    def test_exact_fields_and_duplicate_proposals_fail_closed(self):
        for location in ("envelope", "proposal", "applicability", "duplicate", "reason"):
            proposed = copy.deepcopy(self.proposed)
            if location == "duplicate":
                proposed["proposals"] *= 2
            elif location == "reason":
                proposed["proposals"][0]["rationale"] = " " * 100
            else:
                target = (proposed if location == "envelope" else
                          proposed["proposals"][0] if location == "proposal" else
                          proposed["proposals"][0]["applicability"])
                target["execute"] = True
            with self.subTest(location=location), self.assertRaises(ValueError):
                proposals.validate_proposals(self.evidence, proposed)

    def test_no_proposals_is_valid_and_does_not_mutate_input(self):
        before = copy.deepcopy(self.evidence)
        self.proposed["proposals"] = []
        validated = proposals.validate_proposals(self.evidence, self.proposed)
        self.assertEqual(validated["proposals"], [])
        self.assertEqual(before, self.evidence)


if __name__ == "__main__":
    unittest.main()
