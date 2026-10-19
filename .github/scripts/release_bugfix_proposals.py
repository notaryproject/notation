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

"""Read-only evidence for human-reviewed main-to-release bug-fix proposals.

``collect`` accepts immutable full SHAs, verifies the live main before and after
collection, and uses only GET requests. It never checks out or executes source.
Bounds fail closed, or produce explicit evidence blockers, never silent success.
``validate_proposals`` accepts AI assessments against a *trusted collector result*,
not an independently supplied evidence document. Its digest binds that snapshot;
it is not an authentication signature. Neither API grants worker authority.

Commit messages supply provenance, not proof of textual equivalence or bug-fix
intent. Squashes without provenance, nonstandard reverts, conflicts, and feature
prerequisites cannot be proved by this collector. Applicability always requires
separate review. Titles, labels, and PR mergeability are not classifiers.
"""

import hashlib
import json
import re
from urllib.parse import quote

from monthly_release import REPOSITORY, SHA


MAX_COMMITS = 200
MAX_HISTORY = 200
MAX_FILES = 100
MAX_PRS = 5
MAX_PR_COMMITS = 200
MAX_PATCH_BYTES = 100_000
MAX_EVIDENCE_BYTES = 10_000_000
MAX_REASON = 4000
CLASSIFICATIONS = {"candidate_bugfix", "feature", "not_bugfix", "needs_review"}
CHERRY_PICK = re.compile(r"(?m)^\(cherry picked from commit ([0-9a-f]{40})\)$")
REVERT = re.compile(r"(?m)^This reverts commit ([0-9a-f]{40})(?:\.|,)")


def digest(value):
    return hashlib.sha256(json.dumps(
        value, sort_keys=True, separators=(",", ":"), ensure_ascii=True,
    ).encode()).hexdigest()


def _sha(value):
    if not isinstance(value, str) or not SHA.fullmatch(value):
        raise ValueError("Expected an immutable full commit SHA")
    return value


def _get(api, repository, path):
    # Deliberately do not catch HTTP errors, including missing/inaccessible diffs.
    return api.request(f"repos/{repository}/{path}")


def _main(api, repository, expected):
    if _get(api, repository, "commits/main")["sha"] != expected:
        raise ValueError("Main drifted; recollect against the current main SHA")


def _comparison(api, repository, base, head):
    commits = []
    snapshot = None
    for page in range(1, MAX_COMMITS // 100 + 2):
        data = _get(api, repository,
                    f"compare/{base}...{head}?per_page=100&page={page}")
        current = {
            "status": data["status"], "ahead_by": data["ahead_by"],
            "behind_by": data["behind_by"], "total_commits": data["total_commits"],
            "base": data["base_commit"]["sha"],
            "merge_base": data["merge_base_commit"]["sha"],
        }
        if snapshot is None:
            snapshot = current
            status, ahead, behind = (current[key] for key in
                                      ("status", "ahead_by", "behind_by"))
            if (current["base"] != base or status not in
                    {"ahead", "behind", "identical", "diverged"}):
                raise ValueError("Invalid comparison ancestry")
            _sha(current["merge_base"])
            if (type(ahead) is not int or type(behind) is not int or
                    min(ahead, behind) < 0 or current["total_commits"] != ahead):
                raise ValueError("Invalid comparison counts")
            expected_status = ("diverged" if ahead and behind else
                               "ahead" if ahead else "behind" if behind else "identical")
            if status != expected_status or (
                    not behind and current["merge_base"] != base):
                raise ValueError("Inconsistent comparison status or merge base")
            if ((status == "identical" and base != head) or
                    (status == "behind" and current["merge_base"] != head) or
                    (status == "diverged" and current["merge_base"] in {base, head})):
                raise ValueError("Comparison merge base contradicts ancestry")
            if ahead > MAX_COMMITS:
                raise ValueError("Comparison exceeds commit bound; manual review required")
        elif current != snapshot:
            raise ValueError("Comparison changed during pagination")
        batch = data["commits"]
        if not isinstance(batch, list) or len(batch) > 100:
            raise ValueError("Invalid comparison page")
        commits.extend(batch)
        shas = [_sha(item["sha"]) for item in commits]
        if len(shas) != len(set(shas)) or len(shas) > snapshot["ahead_by"]:
            raise ValueError("Duplicate or excess comparison commits")
        if len(shas) == snapshot["ahead_by"]:
            if shas and shas[-1] != head:
                raise ValueError("Comparison does not reach the pinned head")
            return {**snapshot, "source_shas": shas}
        if len(batch) < 100:
            raise ValueError("Incomplete comparison pagination")
    raise ValueError("Comparison pagination bound exceeded")


def _list(api, repository, path, limit):
    items = []
    for page in range(1, limit // 100 + 2):
        separator = "&" if "?" in path else "?"
        batch = _get(api, repository,
                     f"{path}{separator}per_page=100&page={page}")
        if not isinstance(batch, list) or len(batch) > 100:
            raise ValueError("Invalid list page")
        items.extend(batch)
        if len(items) > limit:
            return items[:limit], False
        if len(batch) < 100:
            return items, True
    raise ValueError("List pagination bound exceeded")


def _history(api, repository, head):
    records, complete = _list(api, repository, f"commits?sha={head}", MAX_HISTORY)
    shas = [_sha(item["sha"]) for item in records]
    if not shas or shas[0] != head or len(shas) != len(set(shas)):
        raise ValueError("Invalid pinned release history")
    return records, complete


def _patch_complete(patch, additions, deletions):
    """Reject missing or truncated unified hunks, including partial API patches."""
    if not isinstance(patch, str) or not patch or len(patch.encode()) > MAX_PATCH_BYTES:
        return False
    old = new = added = removed = 0
    saw_hunk = False
    for line in patch.splitlines():
        hunk = re.fullmatch(r"@@ -\d+(?:,(\d+))? \+\d+(?:,(\d+))? @@.*", line)
        if hunk:
            if old or new:
                return False
            old = int(hunk[1]) if hunk[1] is not None else 1
            new = int(hunk[2]) if hunk[2] is not None else 1
            saw_hunk = True
        elif line == r"\ No newline at end of file" and saw_hunk:
            continue
        elif not saw_hunk or not line or line[0] not in " +-":
            return False
        else:
            if line[0] in " -":
                old -= 1
            if line[0] in " +":
                new -= 1
            added += line[0] == "+"
            removed += line[0] == "-"
            if old < 0 or new < 0:
                return False
    return saw_hunk and old == new == 0 and added == additions and removed == deletions


def _files(repository, sha, files):
    result = []
    for item in files:
        path = item["filename"]
        if not isinstance(path, str) or not path or len(path) > 4096:
            raise ValueError("Invalid diff filename")
        patch = item.get("patch")
        available = _patch_complete(patch, item["additions"], item["deletions"])
        result.append({
            "filename": path, "previous_filename": item.get("previous_filename"),
            "status": item["status"], "additions": item["additions"],
            "deletions": item["deletions"], "changes": item["changes"],
            "patch": patch if available else None,
            "patch_available": available,
            "url": f"https://github.com/{repository}/commit/{sha}#diff-"
                   + hashlib.sha256(path.encode()).hexdigest(),
        })
    if len({item["filename"] for item in result}) != len(result):
        raise ValueError("Duplicate diff files")
    return result


def _commit(api, repository, sha):
    data = _get(api, repository, f"commits/{sha}?per_page=100&page=1")
    if data["sha"] != sha:
        raise ValueError("Commit endpoint returned another SHA")
    # The commit endpoint caps file output. Refuse to call a full page complete.
    files = _files(repository, sha, data["files"][:MAX_FILES])
    stats = data["stats"]
    complete = (len(data["files"]) < MAX_FILES and
                sum(item["additions"] for item in files) == stats["additions"] and
                sum(item["deletions"] for item in files) == stats["deletions"] and
                all(item["patch_available"] for item in files) and bool(files))
    return {
        "sha": sha, "parents": [_sha(item["sha"]) for item in data["parents"]],
        "message": data["commit"]["message"],
        "url": f"https://github.com/{repository}/commit/{sha}",
        "files": files, "diff_complete": complete,
    }


def _ancestor(api, repository, base, head):
    data = _get(api, repository, f"compare/{base}...{head}?per_page=1&page=1")
    return (data["base_commit"]["sha"] == base and
            data["merge_base_commit"]["sha"] == base and
            data["status"] in {"ahead", "identical"} and data["behind_by"] == 0)


def _pulls(api, repository, source, main, main_shas):
    associated, complete = _list(
        api, repository, f"commits/{source}/pulls", MAX_PRS,
    )
    pulls = []
    blockers = [] if complete else ["associated_pr_scan_incomplete"]
    seen = set()
    for association in associated:
        number = association["number"]
        if type(number) is not int or number <= 0 or number in seen:
            raise ValueError("Invalid associated PR number")
        seen.add(number)
        pull = _get(api, repository, f"pulls/{number}")
        if pull["number"] != number:
            raise ValueError("Associated PR identity mismatch")
        if not pull.get("merged_at"):
            continue
        if (pull["base"]["repo"]["full_name"] != repository or
                pull["base"]["ref"] != "main"):
            continue
        merge = _sha(pull["merge_commit_sha"])
        if merge not in main_shas and not _ancestor(api, repository, merge, main):
            blockers.append("pr_merge_not_on_pinned_main")
            continue
        commits, commits_complete = _list(
            api, repository, f"pulls/{number}/commits", MAX_PR_COMMITS,
        )
        commit_shas = [_sha(item["sha"]) for item in commits]
        if len(commit_shas) != len(set(commit_shas)):
            raise ValueError("Duplicate PR commits")
        if merge != source and (not commits_complete or source not in commit_shas):
            blockers.append("pr_source_membership_unverified")
            continue
        head = _sha(pull["head"]["sha"])
        base = _sha(pull["base"]["sha"])
        raw_files, files_complete = _list(
            api, repository, f"pulls/{number}/files", MAX_FILES,
        )
        files = _files(repository, merge, raw_files)
        diff_complete = (files_complete and bool(files) and
                         len(files) == pull["changed_files"] and
                         all(item["patch_available"] for item in files))
        pulls.append({
            "number": number, "repository": repository, "base_ref": "main",
            "base_sha": base, "head_sha": head, "merge_sha": merge,
            "merged_at": pull["merged_at"], "title": pull["title"],
            "body": pull.get("body"), "commit_shas": commit_shas,
            "commits_complete": commits_complete,
            "url": f"https://github.com/{repository}/pull/{number}",
            "immutable_diff_url": f"https://github.com/{repository}/compare/{base}...{head}",
            "files": files, "diff_complete": diff_complete,
        })
    if len(pulls) != 1:
        blockers.append("missing_or_ambiguous_merged_canonical_pr")
    if any(not pull["diff_complete"] for pull in pulls):
        blockers.append("pr_diff_unavailable_or_incomplete")
    return pulls, blockers


def _provenance(records):
    return [{
        "sha": _sha(item["sha"]),
        "cherry_picks": CHERRY_PICK.findall(item["commit"]["message"]),
        "reverts": REVERT.findall(item["commit"]["message"]),
    } for item in records]


def collect(api, repository, main_head, release_head):
    """Collect bounded evidence, not an execution plan or bug-fix classification.

    Missing patches (including binary files), uncertain provenance, multiple PRs,
    and merge commits block a bug-fix proposal. HTTP errors propagate. The caller
    owns selection and freshness of the release SHA; no release branch is assumed.
    A truncated release history explicitly blocks equivalence confirmation.
    """
    if not isinstance(repository, str) or not REPOSITORY.fullmatch(repository):
        raise ValueError("Expected a canonical Notary repository")
    _sha(main_head)
    _sha(release_head)
    _main(api, repository, main_head)
    comparison = _comparison(api, repository, release_head, main_head)
    release_records, history_complete = _history(api, repository, release_head)
    release_provenance = _provenance(release_records)
    release_shas = {item["sha"] for item in release_provenance}
    release_reverted = {sha for item in release_provenance for sha in item["reverts"]}
    commits = []
    for sha in comparison["source_shas"]:
        commits.append(_commit(api, repository, sha))
        if len(json.dumps(commits).encode()) > MAX_EVIDENCE_BYTES:
            raise ValueError("Evidence byte bound exceeded; manual review required")
    main_provenance = _provenance([
        {"sha": item["sha"], "commit": {"message": item["message"]}}
        for item in commits
    ])
    main_reverted = {sha for item in main_provenance for sha in item["reverts"]}
    candidates = []
    evidence_records = {}
    for commit in commits:
        source = commit["sha"]
        pulls, blockers = _pulls(
            api, repository, source, main_head, set(comparison["source_shas"]),
        )
        included = sorted(item["sha"] for item in release_provenance
                          if source in item["cherry_picks"])
        # Also handle a main-side cherry-pick whose original is on release.
        originals = CHERRY_PICK.findall(commit["message"])
        included.extend(sorted(set(originals) & release_shas))
        known_reverts = sorted({item["sha"] for item in main_provenance
                                if source in item["reverts"]})
        release_reverts = sorted({item["sha"] for item in release_provenance
                                   if source in item["reverts"] or
                                   set(item["reverts"]) & set(included)})
        inclusion = "not_observed"
        if included:
            inclusion = "cherry_pick_provenance"
            blockers.append("already_present_by_cherry_pick_provenance")
        if source in main_reverted or known_reverts or release_reverts:
            blockers.append("known_revert_requires_review")
        if REVERT.search(commit["message"]):
            blockers.append("revert_commit_requires_review")
        if any(sha in release_reverted for sha in included):
            blockers.append("release_equivalent_was_reverted")
        if not history_complete:
            blockers.append("release_history_incomplete")
        if not commit["diff_complete"]:
            blockers.append("source_diff_unavailable_or_incomplete")
        if len(commit["parents"]) != 1:
            blockers.append("non_single_parent_source")
        ids = []
        for kind, record in [("commit", commit)] + [("pr", pull) for pull in pulls]:
            evidence_id = f"{kind}:{digest(record)}"
            evidence_records[evidence_id] = record
            ids.append(evidence_id)
        candidates.append({
            "candidate_id": f"{repository}@{source}",
            "repository": repository, "source_sha": source,
            "main_member": True, "inclusion": inclusion,
            "equivalent_shas": sorted(set(included)),
            "known_revert_shas": known_reverts,
            "release_revert_shas": release_reverts,
            "pr_numbers": [pull["number"] for pull in pulls],
            "evidence_ids": ids,
            "blockers": sorted(set(blockers + ["applicability_unverified"])),
        })
        if len(json.dumps(evidence_records).encode()) > MAX_EVIDENCE_BYTES:
            raise ValueError("Evidence byte bound exceeded; manual review required")
    _main(api, repository, main_head)
    return {
        "schema": 1, "authority": "proposal_only", "automatic_inclusion": False,
        "repository": repository, "main_head": main_head, "release_head": release_head,
        "comparison": comparison, "release_history_complete": history_complete,
        "release_provenance": release_provenance,
        "candidates": candidates, "evidence": evidence_records,
        "limits": {
            "commits": MAX_COMMITS, "release_history": MAX_HISTORY,
            "files_per_diff": MAX_FILES, "associated_prs": MAX_PRS,
            "pr_commits": MAX_PR_COMMITS, "patch_bytes": MAX_PATCH_BYTES,
            "evidence_bytes": MAX_EVIDENCE_BYTES,
        },
    }


def _keys(value, expected):
    if not isinstance(value, dict) or set(value) != set(expected.split()):
        raise ValueError("Unexpected or missing proposal fields")


def _reason(value):
    if not isinstance(value, str) or not value.strip() or len(value) > MAX_REASON:
        raise ValueError("A bounded evidence-based rationale is required")


def validate_proposals(evidence, proposed):
    """Bind untrusted AI proposals to trusted, freshly collected evidence.

    The caller must retain the original collector result, not accept replacement
    evidence from the model. Rationale quality remains a human review obligation.
    Even a valid ``candidate_bugfix`` is blocked pending applicability review and
    cannot be included automatically. Unverified conflict/prerequisite claims
    add blockers, never execution permission.
    """
    _keys(proposed, "schema evidence_sha256 proposals")
    if (type(proposed["schema"]) is not int or proposed["schema"] != 1 or
            evidence["schema"] != 1 or
            evidence["authority"] != "proposal_only" or
            evidence["automatic_inclusion"] is not False or
            proposed["evidence_sha256"] != digest(evidence)):
        raise ValueError("Proposal evidence snapshot mismatch")
    if not isinstance(proposed["proposals"], list):
        raise ValueError("Expected a list of proposals")
    candidates = {item["candidate_id"]: item for item in evidence["candidates"]}
    source_shas = set(evidence["comparison"]["source_shas"])
    seen, validated = set(), []
    for proposal in proposed["proposals"]:
        _keys(proposal, "candidate_id repository source_sha pr_number classification "
              "evidence_ids rationale applicability")
        candidate_id = proposal["candidate_id"]
        if not isinstance(candidate_id, str) or candidate_id not in candidates or candidate_id in seen:
            raise ValueError("Unknown or duplicate candidate")
        seen.add(candidate_id)
        candidate = candidates[candidate_id]
        source = _sha(proposal["source_sha"])
        if (proposal["repository"] != evidence["repository"] or
                proposal["repository"] != candidate["repository"] or
                source != candidate["source_sha"] or source not in source_shas or
                not candidate["main_member"]):
            raise ValueError("Proposal repository/source/main membership mismatch")
        classification = proposal["classification"]
        if not isinstance(classification, str) or classification not in CLASSIFICATIONS:
            raise ValueError("Unsupported proposal classification")
        number = proposal["pr_number"]
        if number is not None and (
                type(number) is not int or number not in candidate["pr_numbers"]):
            raise ValueError("Fabricated or unrelated PR")
        citations = proposal["evidence_ids"]
        if (not isinstance(citations, list) or not citations or
                any(not isinstance(item, str) for item in citations) or
                len(citations) != len(set(citations)) or
                not set(citations) <= set(candidate["evidence_ids"])):
            raise ValueError("Fabricated or unrelated evidence citation")
        for evidence_id in citations:
            record = evidence["evidence"][evidence_id]
            kind, fingerprint = evidence_id.split(":", 1)
            if kind not in {"commit", "pr"} or digest(record) != fingerprint:
                raise ValueError("Evidence content changed")
        if not any(item.startswith("commit:") for item in citations):
            raise ValueError("Source diff evidence citation is required")
        if classification == "candidate_bugfix":
            if (set(candidate["blockers"]) - {"applicability_unverified"} or
                    candidate["pr_numbers"] != [number]):
                raise ValueError("Incomplete, ambiguous, included, or reverted source is blocked")
            if not any(item.startswith("pr:") and
                       evidence["evidence"][item]["number"] == number for item in citations):
                raise ValueError("Merged canonical PR diff citation is required")
        _reason(proposal["rationale"])
        applicability = proposal["applicability"]
        _keys(applicability, "status rationale prerequisite_shas")
        status = applicability["status"]
        if (not isinstance(status, str) or status not in
                {"unverified", "conflict", "missing_prerequisite", "not_applicable"}):
            raise ValueError("Applicability cannot be asserted as verified by this module")
        _reason(applicability["rationale"])
        prerequisites = applicability["prerequisite_shas"]
        if (not isinstance(prerequisites, list) or
                any(not isinstance(sha, str) or sha not in source_shas or sha == source
                    for sha in prerequisites) or len(prerequisites) != len(set(prerequisites))):
            raise ValueError("Fabricated or unrelated prerequisite")
        if (status == "missing_prerequisite") != bool(prerequisites):
            raise ValueError("Missing-prerequisite assessment must cite observed source SHAs")
        blockers = set(candidate["blockers"])
        if status != "unverified":
            blockers.add(f"applicability_{status}")
        validated.append({
            **proposal, "authority": "proposal_only", "automatic_inclusion": False,
            "blocked": True, "blockers": sorted(blockers),
        })
    return {
        "schema": 1, "authority": "proposal_only", "automatic_inclusion": False,
        "repository": evidence["repository"], "main_head": evidence["main_head"],
        "release_head": evidence["release_head"], "evidence_sha256": digest(evidence),
        "proposals": validated,
    }
