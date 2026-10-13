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

"""Open checked backport PRs and preserve the documented separate release vote."""

import base64
import os
import pathlib
import re
import tempfile

import monthly_release as release


def content(api, repository, path, ref):
    item = api.request(f"repos/{repository}/contents/{path}?ref={ref}")
    if item.get("encoding") != "base64":
        raise ValueError(f"Cannot read release-vote evidence: {path}")
    return base64.b64decode(item["content"]).decode("utf-8")


def commit_identity(api, repository, sha, tree=None):
    commit = api.request(f"repos/{repository}/commits/{sha}")
    if commit["sha"] != sha or not commit["commit"]["verification"]["verified"]:
        raise ValueError("Release PR commit is not the exact verified signed commit")
    actual = commit["commit"]["tree"]["sha"]
    if tree is not None and actual != tree:
        raise ValueError("Release PR tree changed")
    return actual


def receipt(plan, phase, head, tree, base):
    return {
        "repository": plan["repository"], "actor": os.environ["MONTHLY_PATCH_ACTOR"],
        "base_ref": plan["branch"], "base_sha": base,
        "head_ref": f"monthly-patch/{plan['month']}/{plan['controller_plan_id']}/{phase}",
        "head_sha": head, "head_tree": tree, "tag": plan["tag"],
    }


def publish_pull(api, plan, expected, directory, environment, title):
    repository = plan["repository"]
    marker = f"<!-- monthly-release-pr:{plan['controller_plan_id']}:{expected['head_ref']} -->"
    remote = api.optional(f"repos/{repository}/git/ref/heads/{expected['head_ref']}")
    if remote is not None:
        if remote["object"]["sha"] != expected["head_sha"]:
            raise ValueError("Reserved release PR branch changed; never overwrite it")
    else:
        release.git(["-c", "push.followTags=false", "push",
                     f"--force-with-lease=refs/heads/{expected['head_ref']}:",
                     "origin", f"{expected['head_sha']}:refs/heads/{expected['head_ref']}"], directory, environment)
    commit_identity(api, repository, expected["head_sha"], expected["head_tree"])
    pulls = [pull for state in ("open", "closed") for pull in api.pulls(repository, expected["base_ref"], state)
             if pull["head"]["ref"] == expected["head_ref"]]
    if len(pulls) > 1:
        raise ValueError("Duplicate release PRs require reconciliation")
    if pulls:
        pull = pulls[0]
        if pull.get("user", {}).get("login") != expected["actor"] or marker not in (pull.get("body") or ""):
            raise ValueError("Reserved release PR has an untrusted identity")
    else:
        pull = api.request(f"repos/{repository}/pulls", "POST", {
            "title": title, "head": expected["head_ref"], "base": expected["base_ref"],
            "body": (
                f"{marker}\n\n"
                f"Monthly patch `{plan['tag']}` for approved controller plan `{plan['controller_plan_id']}`.\n\n"
                f"[Change logs](https://github.com/{repository}/compare/{plan['baseline']}...{expected['head_sha']})\n\n"
                f"[Repo-level maintainers at the PR base](https://github.com/{repository}/blob/{expected['base_sha']}/MAINTAINERS)\n\n"
                f"[Pinned assessed roster](https://github.com/{repository}/blob/{plan['maintainer_roster_sha']}/MAINTAINERS)\n\n"
                "This PR preserves ordinary CI and branch reviews. "
                + ("The separate release-voting PR requires a strict majority of repo-level "
                   "maintainer approvals and a normal merge commit. "
                   if repository != "notaryproject/notation-core-go" else
                   "Core uses its reviewed dependency-only policy and normal merge commit, without an invented metadata vote. ")
                +
                "Approval does not authorize optional source fixes or features."
            ),
        })
    expected["number"] = release_id(pull["number"])
    return pull


def release_id(value):
    if type(value) is not int or value <= 0:
        raise ValueError("Invalid release PR number")
    return value


def pull_identity(pull, expected):
    if (pull["number"] != expected["number"]
            or pull.get("user", {}).get("login") != expected["actor"]
            or pull["base"]["ref"] != expected["base_ref"]
            or pull["head"]["sha"] != expected["head_sha"]
            or pull["head"]["ref"] != expected["head_ref"]
            or (pull["head"].get("repo") or {}).get("full_name") != expected["repository"]):
        raise ValueError("Release PR identity or candidate changed")


def merge_pull(api, expected, pull, require_tip=True):
    pull_identity(pull, expected)
    repository = expected["repository"]
    if pull.get("merged_at"):
        sha = pull.get("merge_commit_sha", "")
        if not release.SHA.fullmatch(sha):
            raise ValueError("Merged release PR has no exact merge commit")
        commit = api.request(f"repos/{repository}/commits/{sha}")
        if ([item["sha"] for item in commit["parents"]] != [expected["base_sha"], expected["head_sha"]]
                or commit_identity(api, repository, sha, expected["head_tree"]) != expected["head_tree"]):
            raise ValueError("Release PR was not merged with the voted tree and exact normal merge parents")
        if require_tip and api.request(f"repos/{repository}/branches/{expected['base_ref']}")["commit"]["sha"] != sha:
            raise ValueError("Release branch changed after its approved PR merge")
        return sha
    if pull.get("state") != "open":
        raise ValueError("Release PR was closed without its required normal merge")
    if api.request(f"repos/{repository}/branches/{expected['base_ref']}")["commit"]["sha"] != expected["base_sha"]:
        raise ValueError("Release branch advanced outside the approved release PR")
    ready, reason = release.checks_ready(pull, api.checks(repository, expected["number"]))
    if not ready:
        return None
    result = api.request(f"repos/{repository}/pulls/{expected['number']}/merge", "PUT", {
        "sha": expected["head_sha"], "merge_method": "merge",
    })
    if result.get("merged") is not True:
        raise ValueError("GitHub refused the normal release PR merge")
    merged = api.request(f"repos/{repository}/pulls/{expected['number']}")
    return merge_pull(api, expected, merged)


def version_metadata(repository, original, tag):
    if repository.endswith("/notation-go"):
        pattern = r'(const signingAgent\s*=\s*")[^"]+(")'
        replacement = "notation-go/" + tag.removeprefix("v")
        if len(re.findall(pattern, original)) != 1:
            raise ValueError("Expected one documented signingAgent declaration")
        return re.sub(pattern, lambda match: match[1] + replacement + match[2], original)
    if repository.endswith("/notation"):
        result = original
        for field, value in (("Version", tag), ("BuildMetadata", "")):
            pattern = rf'(?m)^(\s*{field}\s*=\s*")[^"]*(")'
            if len(re.findall(pattern, result)) != 1:
                raise ValueError(f"Expected one documented release {field} declaration")
            result = re.sub(pattern, lambda match: match[1] + value + match[2], result)
        return result
    raise ValueError("This repository has no documented version-metadata voting step")


def advance(api, plan, directory, environment, cycle):
    repository = plan["repository"]
    if plan["mode"] != "execute":
        raise ValueError("Release voting is restricted to canonical execution")
    if "release_backport" not in plan:
        tree = release.git(["rev-parse", f"{plan['commit']}^{{tree}}"], directory)
        expected = receipt(plan, "dependencies", plan["commit"], tree, plan["previous_head"])
        publish_pull(api, plan, expected, directory, environment, f"build(deps): backport dependencies for {plan['tag']}")
        plan.update(status="voting", release_backport=expected)
        cycle.save(plan)
    expected = plan["release_backport"]
    pull = api.request(f"repos/{repository}/pulls/{expected['number']}")
    merged = merge_pull(api, expected, pull, require_tip="release_vote" not in plan)
    if merged is None:
        plan.update(status="voting", reason=f"Waiting for backport PR #{expected['number']} checks and required reviews")
        cycle.save(plan)
        return {**plan, "action": "waiting"}
    expected["merge_commit_sha"] = merged
    if repository.endswith("/notation-core-go"):
        plan.update(status="ready", commit=expected["head_sha"], previous_head=merged, reason="")
        cycle.save(plan)
        return {**plan, "action": "qualify"}
    from monthly_release_vote import assess_vote
    path = "signer/signer.go" if repository.endswith("/notation-go") else "internal/version/version.go"
    if "release_vote" not in plan:
        release.git(["fetch", "--no-tags", "origin", merged], directory)
        cycle.save(plan)
        with tempfile.TemporaryDirectory(prefix="monthly-release-vote-") as temporary:
            worktree = pathlib.Path(temporary) / "vote"
            release.git(["worktree", "add", "--detach", str(worktree), merged], directory)
            try:
                original = (worktree / path).read_text(encoding="utf-8")
                updated = version_metadata(repository, original, plan["tag"])
                if original == updated:
                    raise ValueError("Release version metadata is already present without this cycle's vote")
                (worktree / path).write_text(updated, encoding="utf-8")
                release.git(["add", "--", path], worktree)
                release.git(["commit", "-S", "-m", f"bump: tag and release {plan['tag']}"], worktree, environment)
                head = release.git(["rev-parse", "HEAD"], worktree)
                tree = release.git(["rev-parse", "HEAD^{tree}"], worktree)
                vote = {
                    **receipt(plan, "vote", head, tree, merged),
                    "candidate_sha": merged, "candidate_tree": expected["head_tree"],
                    "maintainer_roster_sha": plan["maintainer_roster_sha"],
                    "metadata_before": original, "metadata_after": updated,
                    "signer_login": os.environ["MONTHLY_PATCH_SIGNER_LOGIN"],
                    "merge_committer_login": "web-flow",
                }
                publish_pull(api, plan, vote, worktree, environment, f"bump: tag and release {plan['tag']}")
                plan["release_vote"] = vote
                cycle.save(plan)
            finally:
                release.git(["worktree", "remove", str(worktree)], directory)
    vote = plan["release_vote"]
    pull = api.request(f"repos/{repository}/pulls/{vote['number']}")
    reviews = list(api.pages(f"repos/{repository}/pulls/{vote['number']}/reviews"))
    roster = content(api, repository, "MAINTAINERS", vote["maintainer_roster_sha"])
    compare = api.request(f"repos/{repository}/compare/{vote['candidate_sha']}...{vote['head_sha']}")
    if compare.get("status") != "ahead" or compare.get("ahead_by") != 1 or compare.get("behind_by") != 0:
        raise ValueError("Release vote must contain exactly one metadata commit")
    diff = {
        "files": [{"filename": item["filename"], "status": item["status"]} for item in compare.get("files", [])],
        "before": content(api, repository, path, vote["candidate_sha"]),
        "after": content(api, repository, path, vote["head_sha"]),
    }
    result = assess_vote(vote, pull, reviews, roster, diff)
    if result["status"] != "ready":
        plan.update(status="voting", reason=result.get("reason", "Waiting for repo-maintainer release votes"))
        cycle.save(plan)
        return {**plan, "action": "waiting"}
    final = merge_pull(api, vote, pull)
    if final is None:
        plan.update(status="voting", reason=f"Release vote PR #{vote['number']} still needs ordinary CI or branch reviews")
        cycle.save(plan)
        return {**plan, "action": "waiting"}
    vote["merge_commit_sha"] = final
    from monthly_release_vote import qualify_release_commit
    vote["receipt"] = qualify_release_commit(vote, api.request(f"repos/{repository}/pulls/{vote['number']}"),
                                             api.request(f"repos/{repository}/commits/{final}"))
    plan.update(status="ready", commit=vote["head_sha"], previous_head=final, reason="")
    validate(api, plan)
    cycle.save(plan)
    return {**plan, "action": "qualify"}


def validate(api, plan):
    expected = plan.get("release_backport")
    if not expected:
        raise ValueError("Canonical publication requires a checked release backport PR")
    pull = api.request(f"repos/{plan['repository']}/pulls/{expected['number']}")
    pull_identity(pull, expected)
    if not pull.get("merged_at") or pull.get("merge_commit_sha") != expected.get("merge_commit_sha"):
        raise ValueError("Release backport merge provenance changed")
    if plan["repository"].endswith("/notation-core-go"):
        final = merge_pull(api, expected, pull)
        if plan["commit"] != expected["head_sha"] or plan["previous_head"] != final:
            raise ValueError("Core publication does not match its qualified dependency PR")
        return expected
    from monthly_release_vote import validate_vote
    return validate_vote(api, plan)
