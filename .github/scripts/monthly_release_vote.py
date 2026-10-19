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

"""Read-only release-vote qualification. Callers own PR creation and merging."""

import base64
import datetime
import re
from urllib.parse import quote


class VoteError(ValueError):
    """Release identity, scope, or provenance does not match the pinned plan."""


def _require(condition, message):
    if not condition:
        raise VoteError(message)


def _login(value):
    _require(isinstance(value, str) and re.fullmatch(r"[A-Za-z0-9-]+(?:\[bot\])?", value),
             "invalid GitHub login")
    return value.lower()


def parse_repo_maintainers(text):
    """Parse only the active repo-level section of the trusted MAINTAINERS file."""
    section = False
    found = False
    logins = []
    for line in text.splitlines():
        if line.startswith("# Repo-Level Maintainers"):
            _require(not found, "duplicate repo-level maintainer section")
            section = found = True
        elif section and line.startswith("# ") and not line.startswith("# Note:"):
            section = False
        elif section and line.strip() and not line.startswith("#"):
            match = re.fullmatch(r".+ <[^<>]+> \(@([A-Za-z0-9-]+)\)", line.strip())
            _require(match is not None, "malformed repo-level maintainer entry")
            logins.append(_login(match.group(1)))
    _require(found and logins, "no active repo-level maintainer roster")
    _require(len(set(logins)) == len(logins), "duplicate repo-level maintainer")
    return sorted(logins)


def _identity(expected, pr):
    for key in ("base_sha", "head_sha", "head_tree", "candidate_sha", "candidate_tree"):
        _require(re.fullmatch(r"[0-9a-f]{40}", expected.get(key, "")),
                 f"invalid expected {key}")
    _require(re.fullmatch(r"notaryproject/(notation|notation-go|notation-core-go)",
                         expected.get("repository", "")), "noncanonical repository")
    if expected["repository"] == "notaryproject/notation-core-go":
        _require(expected.get("vote_scope") == "dependency-candidate",
                 "Core release vote policy must declare dependency-candidate scope")
        _require(expected["head_sha"] == expected["candidate_sha"] and
                 expected["head_tree"] == expected["candidate_tree"],
                 "Core vote changed the qualified dependency candidate")
    else:
        _require(expected["head_sha"] != expected["candidate_sha"] and
                 expected["head_tree"] != expected["candidate_tree"],
                 "release vote has no distinct metadata commit")
    _require(re.fullmatch(r"release-\d+\.\d+", expected.get("base_ref", "")),
             "invalid stable release branch")
    _require(re.fullmatch(r"v\d+\.\d+\.\d+", expected.get("tag", "")),
             "invalid stable patch tag")
    _require(expected["base_ref"] == "release-" + expected["tag"][1:].rsplit(".", 1)[0],
             "release tag and branch differ")
    actor = _login(expected["actor"])
    _require(isinstance(expected.get("number"), int) and expected["number"] > 0,
             "missing pinned release PR number")
    _require(bool(expected.get("head_ref")), "missing pinned release PR branch")
    _require(pr.get("number") == expected["number"], "release PR number changed")
    _require(_login(pr.get("user", {}).get("login")) == actor, "release PR actor changed")
    for side in ("base", "head"):
        actual = pr.get(side, {})
        _require(actual.get("repo", {}).get("full_name") == expected["repository"],
                 f"release PR {side} repository changed")
        _require(actual.get("sha") == expected[f"{side}_sha"],
                 f"release PR {side} commit changed")
        _require(actual.get("ref") == expected[f"{side}_ref"],
                 f"release PR {side} branch changed")
    body = pr.get("body") or ""
    root = "https://github.com/" + expected["repository"]
    _require(root + "/blob/" + expected["base_sha"] + "/MAINTAINERS" in body,
             "release PR lacks pinned maintainer-list link")
    _require(root + "/compare/" in body, "release PR lacks changelog link")


def qualify_release_vote(expected, pr, reviews, maintainer_logins):
    """Return ready/waiting evidence for the immutable, current-head release vote.

    Expected pins repository, number, actor, base_ref/base_sha, head_ref/head_sha,
    head_tree, candidate_sha/candidate_tree, and tag. Callers must derive the
    roster from the assessed default-branch commit, not actor-provided login lists.
    This result is not CI, source-scope, or vulnerability certification.
    """
    _identity(expected, pr)
    roster = {_login(login) for login in maintainer_logins}
    _require(roster and len(roster) == len(maintainer_logins), "invalid maintainer roster")
    latest = {}
    seen = set()
    for review in reviews:
        login = _login(review.get("user", {}).get("login"))
        if login not in roster:
            continue
        ident = review.get("id")
        _require(isinstance(ident, int) and ident > 0 and ident not in seen,
                 "invalid or duplicate review ID")
        seen.add(ident)
        submitted = review.get("submitted_at")
        _require(isinstance(submitted, str), "review has no submission time")
        try:
            timestamp = datetime.datetime.fromisoformat(submitted.replace("Z", "+00:00"))
        except ValueError as error:
            raise VoteError("invalid review submission time") from error
        _require(timestamp.tzinfo is not None, "review timestamp has no timezone")
        state = review.get("state")
        _require(state in {"APPROVED", "CHANGES_REQUESTED", "DISMISSED", "COMMENTED"},
                 "invalid submitted review state")
        # Comments do not revoke a prior approval or a changes-requested review.
        if state == "COMMENTED":
            continue
        if login not in latest or (timestamp, ident) > latest[login][0]:
            latest[login] = ((timestamp, ident), review)
    approvals = sorted(login for login, (_, review) in latest.items()
                       if review["state"] == "APPROVED"
                       and review.get("commit_id") == expected["head_sha"])
    blocking = sorted(login for login, (_, review) in latest.items()
                      if review["state"] == "CHANGES_REQUESTED")
    required = len(roster) // 2 + 1
    evidence = {"approvals": approvals, "required": required, "maintainers": sorted(roster),
                "blocking": blocking, "head_sha": expected["head_sha"],
                "number": expected["number"]}
    _require(pr.get("state") in {"open", "closed"}, "invalid release PR state")
    _require(pr["state"] != "closed" or bool(pr.get("merged_at")),
             "release PR closed without merge")
    reason = ("release PR remains draft" if pr.get("draft") else
              "maintainer changes requested" if blocking else
              "current-head maintainer majority absent" if len(approvals) < required else "")
    return {**evidence, "status": "waiting" if reason else "ready",
            "reason": reason or "current-head maintainer majority satisfied"}


def _signed(commit, signer):
    verification = commit.get("commit", {}).get("verification", {})
    _require(verification.get("verified") is True and verification.get("reason") == "valid",
             "release commit signature is not verified")
    _require(_login(commit.get("committer", {}).get("login")) == _login(signer),
             "release commit committer provenance changed")


def qualify_release_commit(expected, pr, commit):
    """Qualify the exact normal merge commit, never a squash/rebase or admin bypass."""
    _identity(expected, pr)
    _require(pr.get("state") == "closed" and bool(pr.get("merged_at")),
             "release voting PR is not merged")
    sha = commit.get("sha")
    _require(re.fullmatch(r"[0-9a-f]{40}", sha or ""), "invalid final commit")
    _require(sha == pr.get("merge_commit_sha"), "final commit is not release PR merge")
    parents = [parent.get("sha") for parent in commit.get("parents", [])]
    _require(parents == [expected["base_sha"], expected["head_sha"]],
             "release PR did not use the exact normal merge parents")
    tree = commit.get("commit", {}).get("tree", {}).get("sha")
    _require(tree == expected["head_tree"], "merge tree differs from voted tree")
    _require(_login(expected["merge_committer_login"]) in
             {_login(expected["actor"]), "web-flow"},
             "merge committer is neither configured actor nor GitHub normal merger")
    _signed(commit, expected["merge_committer_login"])
    return {"sha": expected["head_sha"], "merge_commit_sha": sha,
            "tree": tree, "number": expected["number"],
            "head_sha": expected["head_sha"], "base_sha": expected["base_sha"],
            "repository": expected["repository"], "tag": expected["tag"],
            "committer_login": expected["merge_committer_login"]}


def _metadata(expected, release_diff):
    """Allow only the verified upstream version fields, not arbitrary source edits."""
    repo = expected["repository"]
    if repo == "notaryproject/notation-core-go":
        _require(expected.get("vote_scope") == "dependency-candidate"
                 and expected["head_sha"] == expected["candidate_sha"]
                 and expected["head_tree"] == expected["candidate_tree"],
                 "Core vote changed the qualified dependency candidate")
        _require(release_diff.get("files") == [] and
                 release_diff.get("before") is None and release_diff.get("after") is None,
                 "Core vote must not add metadata or unexplained source changes")
        return
    path = {"notaryproject/notation-go": "signer/signer.go",
            "notaryproject/notation": "internal/version/version.go"}.get(repo)
    _require(path is not None, "repository release metadata policy is not defined")
    _require(release_diff.get("files") == [{"filename": path, "status": "modified"}],
             "release vote changes files outside version metadata")
    before, after = release_diff.get("before"), release_diff.get("after")
    _require(isinstance(before, str) and isinstance(after, str),
             "missing release metadata contents")
    _require(before == expected["metadata_before"] and after == expected["metadata_after"],
             "release metadata differs from pinned plan")
    if repo.endswith("/notation-go"):
        pattern = r'(?m)^const signingAgent = "notation-go/[^"\n]+"$'
        _require(len(re.findall(pattern, before)) == 1, "ambiguous signingAgent metadata")
        wanted = re.sub(pattern, 'const signingAgent = "notation-go/' +
                        expected["tag"][1:] + '"', before)
    else:
        wanted = before
        for field, value in (("Version", expected["tag"]), ("BuildMetadata", "")):
            pattern = rf'(?m)^([ \t]*){field} = "[^"\n]*"$'
            _require(len(re.findall(pattern, wanted)) == 1, f"ambiguous {field} metadata")
            wanted = re.sub(pattern, lambda match: match.group(1) + field + ' = "' +
                            value + '"', wanted)
    _require(before != after and after == wanted, "release vote is not metadata-only")


def assess_vote(expected, pr, reviews, roster_text, release_diff):
    """Qualify votes after verifying roster and exact metadata-only release scope."""
    _metadata(expected, release_diff)
    return qualify_release_vote(expected, pr, reviews, parse_repo_maintainers(roster_text))


def _contents(api, repository, path, ref):
    result = api.request(f"repos/{repository}/contents/{path}?ref={quote(ref, safe='')}")
    _require(result.get("encoding") == "base64" and result.get("type") == "file",
             "unsupported repository content encoding")
    try:
        return base64.b64decode("".join(result["content"].split()), validate=True).decode("utf-8")
    except (ValueError, UnicodeError, KeyError) as error:
        raise VoteError("invalid repository file contents") from error


def validate_vote(api, plan):
    """Re-read all release evidence before qualification, tagging, or publishing.

    api.request(path) returns REST JSON; api.pages(path) returns all list entries.
    plan.release_vote contains the expected pins plus an immutable receipt.
    Maintainer roster is read at plan.maintainer_roster_sha, the assessed default
    branch commit. Callers bind that commit to the trusted automation digest and
    require reassessment when the roster changes.
    The caller must enforce required checks/reviews through normal branch rules.
    """
    expected = plan["release_vote"]
    repo = expected["repository"]
    _require(plan["repository"] == repo and plan["tag"] == expected["tag"],
             "plan and release vote identity differ")
    _require(re.fullmatch(r"[0-9a-f]{40}", expected.get("maintainer_roster_sha", "")) and
             expected["maintainer_roster_sha"] == plan.get("maintainer_roster_sha"),
             "missing trusted maintainer roster commit")
    pr = api.request(f"repos/{repo}/pulls/{expected['number']}")
    reviews = list(api.pages(f"repos/{repo}/pulls/{expected['number']}/reviews"))
    roster = _contents(api, repo, "MAINTAINERS", expected["maintainer_roster_sha"])
    compare = api.request(f"repos/{repo}/compare/{expected['candidate_sha']}..."
                          f"{expected['head_sha']}")
    core = repo == "notaryproject/notation-core-go"
    _require(compare.get("status") == ("identical" if core else "ahead") and
             compare.get("ahead_by") == (0 if core else 1) and compare.get("behind_by") == 0,
             "release vote candidate scope changed")
    candidate = api.request(f"repos/{repo}/commits/{expected['candidate_sha']}")
    head = api.request(f"repos/{repo}/commits/{expected['head_sha']}")
    _require(candidate.get("sha") == expected["candidate_sha"]
             and candidate.get("commit", {}).get("tree", {}).get("sha") ==
             expected["candidate_tree"], "dependency candidate tree changed")
    if expected["candidate_sha"] != expected["base_sha"]:
        _require([item.get("sha") for item in candidate.get("parents", [])] ==
                 [expected["base_sha"]], "qualified candidate parent changed")
        _signed(candidate, expected["signer_login"])
    _require(head.get("sha") == expected["head_sha"]
             and head.get("commit", {}).get("tree", {}).get("sha") == expected["head_tree"],
             "voted head tree changed")
    _require([item.get("sha") for item in head.get("parents", [])] ==
             [expected["base_sha"] if core else expected["candidate_sha"]],
             "voted commit parent changed")
    _signed(head, expected["signer_login"])
    message = head.get("commit", {}).get("message", "").splitlines()
    _require(message and re.fullmatch(r"[a-z][a-z0-9-]*(?:\([^)]+\))?!?: .+", message[0]),
             "release voting commit lacks a conventional commit message")
    path = ("signer/signer.go" if repo.endswith("/notation-go")
            else "internal/version/version.go")
    diff = {"files": [{"filename": item["filename"], "status": item["status"]}
                      for item in compare.get("files", [])],
            "before": None if core else _contents(api, repo, path, expected["candidate_sha"]),
            "after": None if core else _contents(api, repo, path, expected["head_sha"])}
    vote = assess_vote(expected, pr, reviews, roster, diff)
    _require(vote["status"] == "ready", vote["reason"])
    commit = api.request(f"repos/{repo}/commits/{pr['merge_commit_sha']}")
    receipt = qualify_release_commit(expected, pr, commit)
    _require(receipt == expected["receipt"], "immutable release vote receipt changed")
    _require(plan["commit"] == receipt["sha"] and
             plan["previous_head"] == receipt["merge_commit_sha"],
             "plan does not reference voted head and final merge tip")
    branch = api.request(f"repos/{repo}/git/ref/heads/{quote(expected['base_ref'], safe='/')}")
    _require(branch.get("object", {}).get("sha") == receipt["merge_commit_sha"],
             "release branch moved after voting")
    return receipt
