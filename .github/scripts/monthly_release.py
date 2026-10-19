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

"""Coordinate one dependency-only patch cycle using gh and git."""

import argparse
import base64
import copy
import datetime
import hashlib
import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile

from monthly_release_checks import asset_names

PROJECTS = {
    "notation-core-go": (),
    "notation-go": ("notation-core-go",),
    "notation": ("notation-core-go", "notation-go"),
}
STABLE = re.compile(r"v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$")
SHA = re.compile(r"[0-9a-f]{40}$")
MONTH = re.compile(r"\d{4}-(0[1-9]|1[0-2])$")
CYCLE = re.compile(r"monthly|upstream-[0-9a-f]{64}$")
REPOSITORY = re.compile(r"notaryproject/(notation-core-go|notation-go|notation)$")
FINAL = {"published", "skipped"}


def run(command, cwd=None, env=None):
    result = subprocess.run(
        command, cwd=cwd, env=env, capture_output=True, text=True, encoding="utf-8"
    )
    if result.returncode:
        raise RuntimeError(
            f"Command failed ({result.returncode}): {' '.join(command)}\n"
            f"{result.stdout}{result.stderr}"
        )
    return result.stdout


class APIError(RuntimeError):
    def __init__(self, message):
        super().__init__(message)
        match = re.search(r"HTTP (\d{3})", message)
        self.status = int(match[1]) if match else None


class GitHub:
    def request(self, path, method="GET", payload=None):
        command = ["gh", "api", path, "--method", method]
        if payload is not None:
            command.extend(["--input", "-"])
        result = subprocess.run(
            command, input=json.dumps(payload) if payload is not None else None,
            capture_output=True, text=True, encoding="utf-8",
        )
        if result.returncode:
            raise APIError(f"GitHub request failed: {method} {path}\n{result.stderr}")
        return json.loads(result.stdout) if result.stdout.strip() else None

    def optional(self, path):
        try:
            return self.request(path)
        except APIError as error:
            if error.status != 404:
                raise
            return None

    def pages(self, path):
        for page in range(1, 1001):
            separator = "&" if "?" in path else "?"
            items = self.request(f"{path}{separator}per_page=100&page={page}")
            if not isinstance(items, list):
                raise ValueError(f"Invalid paginated response: {path}")
            yield from items
            if len(items) < 100:
                return
        raise ValueError("Pagination limit reached; discovery is incomplete")

    def pulls(self, repository, branch, state):
        return list(self.pages(
            f"repos/{repository}/pulls?state={state}&base={branch}&sort=updated"
        ))

    def manifest(self, repository, branch, directory):
        path = "go.mod" if directory == "." else f"{directory}/go.mod"
        result = self.request(f"repos/{repository}/contents/{path}?ref={branch}")
        if result.get("encoding") != "base64":
            raise ValueError(f"Missing module manifest: {path}")
        with tempfile.TemporaryDirectory(prefix="monthly-manifest-") as temporary:
            modfile = pathlib.Path(temporary) / "go.mod"
            modfile.write_bytes(base64.b64decode(result["content"], validate=False))
            return json.loads(run(
                ["go", "mod", "edit", "-json", f"-modfile={modfile}"], cwd=temporary
            ))

    def checks(self, repository, number):
        return json.loads(run([
            "gh", "pr", "view", str(number), "--repo", repository,
            "--json", "headRefOid,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup",
        ]))

    def issue_editor(self, repository, number):
        owner, name = repository.split("/")
        query = (
            "query($owner:String!,$name:String!,$number:Int!){"
            "repository(owner:$owner,name:$name){issue(number:$number){"
            "lastEditedAt editor{login}}}}"
        )
        result = self.request("graphql", "POST", {
            "query": query, "variables": {"owner": owner, "name": name, "number": number},
        })
        if result.get("errors"):
            raise ValueError("Cannot authenticate release record editor")
        issue = result["data"]["repository"]["issue"]
        if issue is None:
            raise ValueError("Release record no longer exists")
        if issue["lastEditedAt"] is not None:
            editor = (issue.get("editor") or {}).get("login")
            if not editor:
                raise ValueError("Edited release record has no authenticated editor")
            return editor
        return None


def trusted_editor(api, repository, issue):
    if repository.startswith("notaryproject/") and os.environ.get("GITHUB_REPOSITORY", "").startswith("notaryproject/"):
        editor = api.issue_editor(repository, issue["number"])
        if editor is not None and editor != os.environ.get("MONTHLY_PATCH_ACTOR"):
            raise ValueError("Release state was edited outside the configured actor")


def semver(tag):
    match = STABLE.fullmatch(tag)
    return tuple(map(int, match.groups())) if match else None


def latest_release(releases):
    eligible = [
        item for item in releases
        if not item["draft"] and not item["prerelease"] and semver(item["tag_name"])
    ]
    if not eligible:
        raise ValueError("No published stable release; a reviewed baseline is required")
    return max(eligible, key=lambda item: semver(item["tag_name"]))


def modules(repository):
    return (".", "test/e2e", "test/e2e/plugin") if repository.split("/")[1] == "notation" else (".",)


def cycle_marker(month, mode, cycle_id="monthly"):
    if not MONTH.fullmatch(month) or mode != "execute" or not CYCLE.fullmatch(cycle_id):
        raise ValueError("Invalid cycle identity")
    suffix = "" if cycle_id == "monthly" else f":{cycle_id}"
    return f"<!-- dependency-patch:{month}:{mode}{suffix} -->"


def plan_marker(plan):
    return cycle_marker(plan["month"], plan["mode"], plan.get("cycle_id", "monthly"))


def policy(repository, mode, environment, metadata):
    if not REPOSITORY.fullmatch(repository) or mode not in {"dry-run", "execute"}:
        raise ValueError("Unsupported repository or mode")
    if mode == "dry-run":
        return
    if mode == "execute":
        if repository.split("/")[0] != "notaryproject":
            raise ValueError("Stable execution is restricted to the canonical repositories")
        if environment.get("MONTHLY_PATCH_ENABLED") != "true":
            raise ValueError("Monthly production execution is not enabled")
        if environment.get("GITHUB_REF") != "refs/heads/main" or metadata["default_branch"] != "main":
            raise ValueError("Production orchestration must run from the trusted default main")
    for key in ("MONTHLY_PATCH_TOKEN_READY", "MONTHLY_PATCH_SIGNING_KEY",
                "MONTHLY_PATCH_SIGNER_EMAIL", "MONTHLY_PATCH_SIGNER_LOGIN", "MONTHLY_PATCH_ACTOR"):
        if not environment.get(key) or (key.endswith("READY") and environment[key] != "true"):
            raise ValueError(f"Missing execution prerequisite: {key}")


def dependency_file(path):
    return (
        path in {"go.mod", "go.sum", "test/e2e/go.mod", "test/e2e/go.sum",
                 "test/e2e/plugin/go.mod", "test/e2e/plugin/go.sum"}
        or bool(re.fullmatch(r"\.github/workflows/[^/]+\.ya?ml", path))
    )


def dependabot(pull, branch):
    return (
        pull["user"]["login"] == "dependabot[bot]"
        and pull["user"].get("type") == "Bot"
        and pull["base"]["ref"] == branch
    )


def checks_ready(pull, checks):
    if pull.get("draft") or checks["headRefOid"] != pull["head"]["sha"]:
        return False, "Draft or changed pull-request head"
    if checks["mergeable"] != "MERGEABLE" or checks["mergeStateStatus"] != "CLEAN":
        return False, "Pull request is not cleanly mergeable under branch protection"
    if checks["reviewDecision"] not in ("", None, "APPROVED"):
        return False, "Required review is missing"
    results = checks["statusCheckRollup"]
    if not results:
        return False, "No CI evidence; missing checks never qualify a merge"
    successes = 0
    for result in results:
        if result.get("__typename") == "StatusContext":
            if result.get("state") != "SUCCESS":
                return False, "A commit status is not successful"
            successes += 1
        else:
            if result.get("status") != "COMPLETED":
                return False, "A check is pending"
            if result.get("conclusion") not in {"SUCCESS", "SKIPPED", "NEUTRAL"}:
                return False, "A check failed"
            successes += result.get("conclusion") == "SUCCESS"
    return (True, "") if successes else (False, "No successful checks")


def requirement(manifest, module, producer_repository, mode):
    if mode not in {"execute", "dry-run"}:
        raise ValueError("Unsupported dependency requirement mode")
    entries = [item for item in manifest.get("Require") or [] if item["Path"] == module]
    if len(entries) != 1:
        raise ValueError(f"Missing or duplicate explicit dependency requirement: {module}")
    version = entries[0]["Version"]
    replacements = [
        item for item in manifest.get("Replace") or []
        if item["Old"]["Path"] == module and item["Old"].get("Version", "") in ("", version)
    ]
    if mode == "execute":
        if manifest.get("Replace"):
            raise ValueError("Production release manifests must not contain trial replacements")
        return version
    return replacements[0]["New"].get("Version", "") if replacements else version


def decode_state(body, repository, month, mode, cycle_id="monthly"):
    marker = cycle_marker(month, mode, cycle_id)
    if marker not in (body or ""):
        return None
    blocks = re.findall(r"```json\n(.*?)\n```", body, flags=re.DOTALL)
    if len(blocks) != 1:
        raise ValueError("Malformed or duplicate cycle state")
    def unique(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError("Duplicate cycle state field")
            result[key] = value
        return result
    state = json.loads(blocks[0], object_pairs_hook=unique)
    if state.get("schema") != 1 or state.get("repository") != repository:
        raise ValueError("Cycle state belongs to another repository or schema")
    if state.get("month") != month or state.get("mode") != mode or state.get("cycle_id", "monthly") != cycle_id:
        raise ValueError("Cycle state identity changed")
    if state.get("status") not in {"merging", "waiting", "voting", "ready", "tagged", "staging", "draft",
                                  "awaiting-publication", "verifying", "failed", *FINAL}:
        raise ValueError("Unknown cycle outcome")
    baseline = semver(state.get("baseline", ""))
    if baseline is None:
        raise ValueError("Cycle has no stable baseline")
    major, minor, patch = baseline
    expected_branch = f"release-{major}.{minor}"
    expected_main = "main"
    if state.get("tag") != f"v{major}.{minor}.{patch + 1}":
        raise ValueError("Patch tag does not increment the stable baseline")
    if state.get("branch") != expected_branch or state.get("main") != expected_main:
        raise ValueError("Cycle branches escape the approved execution namespace")
    if state["status"] in {"ready", "tagged", "staging", "draft", "awaiting-publication", "verifying", "published"} and not SHA.fullmatch(state.get("commit", "")):
        raise ValueError("Cycle has no valid candidate commit")
    if state["status"] in {"ready", "tagged", "staging", "draft", "awaiting-publication", "verifying", "published"} and not SHA.fullmatch(state.get("previous_head", "")):
        raise ValueError("Cycle has no valid original release-branch head")
    return state


def validate_plan(plan):
    repository, month, mode = (plan[key] for key in ("repository", "month", "mode"))
    if not REPOSITORY.fullmatch(repository):
        raise ValueError("Invalid plan repository")
    body = plan_marker(plan) + "\n```json\n" + json.dumps(plan) + "\n```"
    return decode_state(body, repository, month, mode, plan.get("cycle_id", "monthly"))


def resume_cycle(api, repository, month, mode):
    active = []
    for issue in api.pages(f"repos/{repository}/issues?state=open"):
        if "pull_request" in issue:
            continue
        if issue.get("user", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR"):
            continue
        for previous, suffix in re.findall(rf"<!-- dependency-patch:(\d{{4}}-(?:0[1-9]|1[0-2])):{mode}(?::(upstream-[0-9a-f]{{64}}))? -->", issue.get("body") or ""):
            if previous > month:
                raise ValueError("A future monthly cycle requires reconciliation")
            cycle_id = suffix or "monthly"
            state = decode_state(issue["body"], repository, previous, mode, cycle_id)
            if state["status"] not in FINAL:
                if issue.get("user", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR"):
                    raise ValueError("Active cycle has an unexpected author")
                active.append((previous, cycle_id))
    if len(active) > 1:
        raise ValueError("Multiple active cycles require reconciliation before release")
    return active[0] if active else None


class Cycle:
    def __init__(self, api, repository, month, mode, cycle_id="monthly"):
        self.api, self.repository, self.month, self.mode = api, repository, month, mode
        self.cycle_id = cycle_id
        self.issue = None
        self.state = None
        matches = []
        for issue in api.pages(f"repos/{repository}/issues?state=all"):
            if "pull_request" in issue:
                continue
            if issue.get("user", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR"):
                continue
            state = decode_state(issue.get("body"), repository, month, mode, cycle_id)
            if state is not None:
                trusted_editor(api, repository, issue)
                if not os.environ.get("MONTHLY_PATCH_ACTOR"):
                    raise ValueError("Cycle record was not created by the configured automation identity")
                matches.append((issue, state))
        if len(matches) > 1:
            raise ValueError("Duplicate cycle records require manual reconciliation")
        if matches:
            self.issue, self.state = matches[0]

    def save(self, state):
        validate_plan(state)
        if (state["repository"], state["month"], state["mode"], state.get("cycle_id", "monthly")) != (self.repository, self.month, self.mode, self.cycle_id):
            raise ValueError("Cycle write changed its identity")
        title = f"Dependency patch: {self.month} ({self.mode}, {self.cycle_id})"
        body = (
            f"{plan_marker(state)}\n"
            f"# {title}\n\n"
            f"Outcome: **{state['status']}**. {state.get('reason', '')}\n\n"
            "```json\n" + json.dumps(state, indent=2) + "\n```\n"
        )
        payload = {"body": body, "state": "closed" if state["status"] in FINAL else "open"}
        if self.issue:
            self.issue = self.api.request(
                f"repos/{self.repository}/issues/{self.issue['number']}", "PATCH", payload
            )
        else:
            payload["title"] = title
            self.issue = self.api.request(f"repos/{self.repository}/issues", "POST", payload)
        if self.issue.get("user", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR"):
            raise ValueError("Cycle write used an unexpected automation identity")
        trusted_editor(self.api, self.repository, self.issue)
        self.state = state


def producer_versions(api, repository, mode):
    owner, project = repository.split("/")
    ready = {}
    for producer in PROJECTS[project]:
        target = f"notaryproject/{producer}"
        releases = list(api.pages(f"repos/{target}/releases"))
        candidates = []
        for item in releases:
            if item["draft"] or item["prerelease"]:
                continue
            version = semver(item["tag_name"])
            if version is None:
                continue
            candidates.append((version, item))
        selected = None
        for _, item in sorted(candidates, key=lambda pair: pair[0], reverse=True):
            marker = re.search(r"<!-- dependency-patch:(\d{4}-\d{2}):(execute)(?::(upstream-[0-9a-f]{64}))? -->", item.get("body") or "")
            if marker:
                cycle = Cycle(api, target, marker[1], marker[2], marker[3] or "monthly")
                state = cycle.state
                if not state or state["status"] != "published" or state["tag"] != item["tag_name"]:
                    continue
                assert_tag(api, state)
            selected = item
            break
        if selected is None:
            raise ValueError(f"No usable public stable producer release: {target}")
        ready[producer] = {"repository": target, "version": selected["tag_name"]}
    return ready


def includes_version(actual, expected):
    if actual == expected:
        return True
    current, required = semver(actual), semver(expected)
    return current is not None and required is not None and current >= required


def producer_lag(api, repository, ref, mode, producers, at_least=False):
    manifests = {directory: api.manifest(repository, ref, directory) for directory in modules(repository)} if producers else {}
    return producer_lag_from_manifests(repository, manifests, mode, producers, at_least)


def producer_lag_from_manifests(repository, manifests, mode, producers, at_least=False):
    lag = []
    for directory in modules(repository):
        manifest = manifests[directory] if producers else None
        for name, producer in producers.items():
            module = f"github.com/notaryproject/{name}"
            if directory != "." and not any(item["Path"] == module for item in manifest.get("Require") or []):
                continue
            actual = requirement(manifest, module, producer["repository"], mode)
            matches = includes_version(actual, producer["version"]) if at_least else actual == producer["version"]
            if not matches:
                lag.append(f"{directory}: wait for Dependabot to update {module} to {producer['version']}")
    return lag


def upstream_cycle_id(producers):
    versions = {name: {key: item[key] for key in ("repository", "version")} for name, item in producers.items()}
    digest = hashlib.sha256(json.dumps(versions, sort_keys=True).encode()).hexdigest()
    return f"upstream-{digest}"


def pending_notification(api, repository, mode):
    owner, producer = repository.split("/")
    targets = {f"{owner}/{name}" for name, dependencies in PROJECTS.items() if producer in dependencies}
    if not targets:
        return None
    pending = []
    for issue in api.pages(f"repos/{repository}/issues?state=all"):
        if "pull_request" in issue or issue.get("user", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR"):
            continue
        for month, suffix in re.findall(rf"<!-- dependency-patch:(\d{{4}}-\d{{2}}):{mode}(?::(upstream-[0-9a-f]{{64}}))? -->", issue.get("body") or ""):
            state = decode_state(issue["body"], repository, month, mode, suffix or "monthly")
            if state["status"] == "published" and targets - set(state.get("notified", [])):
                pending.append(state)
    return min(pending, key=lambda state: (state["month"], state["tag"])) if pending else None


def verified_tag(api, repository, tag, commit):
    ref = api.request(f"repos/{repository}/git/ref/tags/{tag}")
    if ref["object"]["type"] != "tag":
        raise ValueError("Expected an immutable, signed annotated release tag")
    annotated = api.request(f"repos/{repository}/git/tags/{ref['object']['sha']}")
    if annotated["object"]["sha"] != commit or not annotated["verification"]["verified"]:
        raise ValueError("Release tag signature or qualified commit does not match")
    return ref["object"]["sha"]


def check_release_assets(api, plan, draft):
    identifier = plan.get("release_id")
    if type(identifier) is not int or identifier <= 0:
        raise ValueError("Release manifest has no recorded release ID")
    release = api.request(f"repos/{plan['repository']}/releases/{identifier}")
    if (release.get("id") != identifier or release.get("tag_name") != plan["tag"]
            or release.get("draft") is not draft or release.get("prerelease") is not False):
        raise ValueError("Release visibility or stable/prerelease identity changed")
    if plan_marker(plan) not in (release.get("body") or "") or plan["commit"] not in release["body"]:
        raise ValueError("Release does not identify the qualified cycle")
    assets = list(api.pages(f"repos/{plan['repository']}/releases/{release['id']}/assets"))
    expected = {item["name"]: item for item in plan["assets"]}
    if set(expected) != set(asset_names(plan["repository"].split("/")[1], plan["tag"])) or len(expected) != len(plan["assets"]):
        raise ValueError("Publication manifest has an invalid asset set")
    if not all(re.fullmatch(r"[0-9a-f]{64}", item.get("sha256", "")) and type(item.get("size")) is int and item["size"] > 0 for item in expected.values()):
        raise ValueError("Publication manifest has invalid asset hashes or sizes")
    if len(assets) != len(expected) or {item["name"] for item in assets} != set(expected):
        raise ValueError("Public release assets differ from the qualified manifest")
    for item in assets:
        match = expected[item["name"]]
        if (type(match.get("id")) is not int or match["id"] <= 0 or item.get("id") != match["id"]
                or item["size"] != match["size"] or item.get("digest") != f"sha256:{match['sha256']}"):
            raise ValueError(f"Public release asset integrity changed: {item['name']}")
    return release


def check_public_assets(api, plan):
    if not plan.get("draft_verification") or not plan.get("draft_review_issue"):
        raise ValueError("Public release was not a qualified reviewed draft")
    check_draft_review(api, plan)
    return check_release_assets(api, plan, False)


def draft_review_body(plan):
    from notation_release_controller import safe_title
    record = {key: plan[key] for key in ("repository", "month", "tag", "commit", "release_id",
                                        "assets", "draft_verification", "controller_issue")}
    record["controller_plan_id"] = plan.get("draft_plan_id", plan["controller_plan_id"])
    marker = f"<!-- monthly-release-review:{plan['repository']}:{plan['month']}:{plan['tag']} -->"
    scope = ["## Actual release commits", "",
             "| Change kind | PR | Exact commit |", "|---|---|---|"]
    for item in plan.get("backports", plan["merged"]):
        scope.append(f"| Dependabot source | [#{item['number']}](https://github.com/{plan['repository']}/pull/{item['number']}) "
                     f"{safe_title(item.get('title', 'Dependency update'))} | `{item['commit']}` |")
        if item.get("backport_commit"):
            scope.append(f"| Signed dependency backport of PR #{item['number']} | Stable candidate source change | `{item['backport_commit']}` |")
    backport = plan.get("release_backport", {})
    scope.append(f"| Qualified stable dependency candidate/backport | Ordinary reviewed release PR | `{backport.get('head_sha', plan['commit'])}` |")
    label = "Signed release HEAD (Core dependency candidate)" if plan["repository"].endswith("/notation-core-go") else "Signed voted release HEAD (metadata, not Dependabot)"
    scope.append(f"| {label} | Repository release governance | `{plan['commit']}` |")
    scope.append(f"| Normal stable-branch merge tip | Preserved branch lease | `{plan['previous_head']}` |")
    return (
        f"{marker}\n# Qualified draft release review\n\n"
        f"Review [{plan['tag']} draft](https://github.com/{plan['repository']}/releases/tag/{plan['tag']}) "
        f"(release ID {plan['release_id']}). Source, scan, packaged-asset, and downloaded-draft native "
        "checks passed for the exact candidate and recorded bytes below.\n\n"
        f"Plan: https://github.com/notaryproject/notation/issues/{plan['controller_issue']}\n\n"
        "**Maintainer action:** inspect the draft release, all assets, hashes, and retained evidence. "
        "In GitHub Releases, open this exact draft, choose Edit, and choose Publish release "
        "without changing its tag, commit, assets, or stable release identity. "
        "Do not publish if any check or evidence is missing. Closing this issue or commenting "
        "does not publish or authorize the bot to publish. Only a maintainer publishes in this phase. "
        "Automation will verify the same release ID and bytes anonymously after publication "
        "before completing the cycle or releasing consumers.\n\n"
        "The signed annotated tag is already public. A draft does not hide or revoke the tag. "
        "Do not move or recreate it. Publication automation is not implemented.\n\n"
        + "\n".join(scope) + "\n\n"
        + f"```json\n{json.dumps(record, indent=2)}\n```\n"
    )


def check_draft_review(api, plan):
    number = plan.get("draft_review_issue")
    if type(number) is not int or number <= 0:
        raise ValueError("Missing qualified draft review issue")
    issue = api.request(f"repos/{plan['repository']}/issues/{number}")
    if plan.get("draft_review_comment"):
        import release_issue_topology as threads
        from notation_release_controller import HOST, controller_body, decode_issue
        actor = os.environ.get("MONTHLY_PATCH_ACTOR")
        root = api.request(f"repos/{HOST}/issues/{plan['controller_issue']}")
        state = decode_issue(root, actor)
        if (state is None or root.get("body") != controller_body(state)
                or state["plan"]["plan_id"] != plan["controller_plan_id"]
                or state.get("release_issues", {}).get(plan["repository"]) != number):
            raise ValueError("Qualified draft release issue route changed")
        trusted_editor(api, HOST, root)
        if plan["repository"] != HOST:
            threads.checked_issue(api, actor, plan["repository"], number,
                                  threads.scope_body(state["plan"], plan["repository"], state["release_issues"]))
        comment = api.request(f"repos/{plan['repository']}/issues/comments/{plan['draft_review_comment']}")
        if (comment.get("user", {}).get("login") != actor
                or comment.get("body") != plan["draft_review_body"]
                or comment.get("issue_url") != f"https://api.github.com/repos/{plan['repository']}/issues/{number}"):
            raise ValueError("Qualified draft review comment changed")
        return issue
    if ("pull_request" in issue or issue.get("user", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR")
            or issue.get("body") != draft_review_body(plan)):
        raise ValueError("Qualified draft review issue changed or has an untrusted author")
    trusted_editor(api, plan["repository"], issue)
    return issue


def assert_tag(api, plan):
    current = verified_tag(api, plan["repository"], plan["tag"], plan["commit"])
    if plan.get("tag_object") and plan["tag_object"] != current:
        raise ValueError("Immutable annotated tag object changed")
    return current


def signing_environment(path, api):
    key = os.environ.get("MONTHLY_PATCH_SIGNING_KEY", "")
    email = os.environ.get("MONTHLY_PATCH_SIGNER_EMAIL", "")
    if not key or not email or "\n" in email:
        raise ValueError("A dedicated registered SSH signing key and signer email are required")
    key_path = pathlib.Path(path) / "signing-key"
    descriptor = os.open(key_path, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    with os.fdopen(descriptor, "w") as output:
        output.write(key.rstrip() + "\n")
    public = run(["ssh-keygen", "-y", "-P", "", "-f", str(key_path)]).split()[:2]
    login = os.environ.get("MONTHLY_PATCH_SIGNER_LOGIN", "")
    if not re.fullmatch(r"[A-Za-z0-9-]+", login):
        raise ValueError("A registered GitHub signing-key owner is required")
    keys = list(api.pages(f"users/{login}/ssh_signing_keys"))
    if not any(item["key"].split()[:2] == public for item in keys):
        raise ValueError("Release key is not registered as the declared GitHub user's SSH signing key")
    settings = {
        "user.name": os.environ.get("MONTHLY_PATCH_SIGNER_NAME") or "Monthly dependency release",
        "user.email": email, "gpg.format": "ssh", "user.signingkey": str(key_path),
        "commit.gpgsign": "true",
    }
    environment = os.environ.copy()
    environment["GIT_CONFIG_COUNT"] = str(len(settings))
    for index, (name, value) in enumerate(settings.items()):
        environment[f"GIT_CONFIG_KEY_{index}"] = name
        environment[f"GIT_CONFIG_VALUE_{index}"] = value
    return environment


def git(command, directory, environment=None):
    return run([
        "git", "-c", "credential.helper=", "-c",
        "credential.helper=!gh auth git-credential", *command,
    ], cwd=directory, env=environment).strip()


def baseline_plan(api, repository, month, mode, state=None):
    source = repository
    baseline = latest_release(list(api.pages(f"repos/{source}/releases")))
    major, minor, patch = semver(baseline["tag_name"])
    branch = f"release-{major}.{minor}"
    tag = f"v{major}.{minor}.{patch + 1}"
    if state and state["baseline"] != baseline["tag_name"]:
        raise ValueError("Stable baseline advanced during the cycle; replan before any release")
    return {
        "schema": 1, "repository": repository, "month": month, "mode": mode,
        "baseline": baseline["tag_name"], "baseline_published": baseline["published_at"],
        "branch": branch, "main": "main",
        "tag": tag, "status": "merging", "merged": [],
        "started_at": datetime.datetime.now(datetime.timezone.utc).isoformat(),
    }


def inventory(api, repository, branch, authorization=None, merged=()):
    result = []
    for pull in api.pulls(repository, branch, "closed"):
        eligible = dependabot(pull, branch)
        if not eligible and authorization:
            from notation_release_controller import scoped_pull
            recorded = any(item["number"] == pull["number"] for item in merged)
            eligible = (scoped_pull(api, repository, pull, authorization, merged)
                        if recorded and pull.get("merged_at") else
                        False)
        if eligible and pull.get("merged_at"):
            if not SHA.fullmatch(pull.get("merge_commit_sha") or ""):
                raise ValueError("Merged Dependabot PR has no valid source commit")
            result.append({
                "number": pull["number"], "commit": pull["merge_commit_sha"],
                "merged_at": pull["merged_at"],
            })
    return sorted(result, key=lambda item: (item["merged_at"], item["number"]))


def check_files(api, repository, number):
    files = list(api.pages(f"repos/{repository}/pulls/{number}/files"))
    if not files or not all(
        dependency_file(item["filename"])
        and (not item.get("previous_filename") or dependency_file(item["previous_filename"]))
        for item in files
    ):
        raise ValueError(f"Dependabot PR #{number} changes files outside the dependency allowlist")


def check_publisher_guard(api, repository, branch):
    if repository.split("/")[1] != "notation":
        return
    workflow = api.optional(f"repos/{repository}/contents/.github/workflows/release-github.yml?ref={branch}")
    if workflow is None:
        return
    if workflow.get("encoding") != "base64":
        raise ValueError("Cannot inspect the release branch's existing tag publisher")
    content = base64.b64decode(workflow["content"]).decode("utf-8")
    guard = "    if: github.repository == 'notaryproject/notation' && github.actor != vars.MONTHLY_PATCH_ACTOR"
    if guard not in content.splitlines():
        raise ValueError("Install the monthly automation-actor guard in release-github.yml on the release branch before enabling monthly publication")


def coordinator_authorization(api, repository, mode, month, plan=None, allow_previous=False, notification=False):
    if mode == "dry-run":
        return None
    from notation_release_controller import worker_authorization
    return worker_authorization(api, repository, month, plan, allow_previous, notification)


def candidate_bundle(directory, output, state):
    git(["fetch", "--no-tags", "origin", state["commit"]], directory)
    git(["update-ref", "refs/heads/monthly-patch-candidate", state["commit"]], directory)
    git(["bundle", "create", str(pathlib.Path(output) / "candidate.bundle"),
         "refs/heads/monthly-patch-candidate"], directory)
    return {**state, "action": "qualify"}


def prepare(api, repository, mode, month, directory, output, start=False, reconcile=False):
    metadata = api.request(f"repos/{repository}")
    policy(repository, mode, os.environ, metadata)
    authorization = coordinator_authorization(api, repository, mode, month)
    effective_mode = "execute"
    active = resume_cycle(api, repository, month, effective_mode)
    if authorization and active and active != (month, "monthly"):
        raise ValueError("Another release cycle is active; coordinator authorization cannot replace it")
    cycle_id = "monthly"
    if active:
        month, cycle_id = active
    cycle = Cycle(api, repository, month, effective_mode, cycle_id)
    if authorization and cycle.state:
        coordinator_authorization(api, repository, mode, month, cycle.state, allow_previous=True)
        if cycle.state.get("controller_plan_id") != authorization["plan_id"] and cycle.state["status"] in {
            "merging", "waiting", "voting", "ready", "draft", "awaiting-publication", "verifying",
        }:
            cycle.state["controller_plan_id"] = authorization["plan_id"]
            if mode == "execute" and cycle.state["status"] not in {"draft", "awaiting-publication", "verifying"}:
                roster = authorization["snapshot"]["automation_sha"]
                if cycle.state.get("release_vote"):
                    from monthly_release_pr import content
                    vote = cycle.state["release_vote"]
                    if content(api, repository, "MAINTAINERS", roster) != content(api, repository, "MAINTAINERS", vote["maintainer_roster_sha"]):
                        raise ValueError("Maintainer roster changed; a new release vote is required")
                    vote["maintainer_roster_sha"] = roster
                cycle.state["maintainer_roster_sha"] = roster
            cycle.save(cycle.state)
    discovered = None
    if not active and reconcile and mode != "dry-run":
        pending = pending_notification(api, repository, effective_mode)
        if pending:
            return {**pending, "action": "complete"}
    if not active and reconcile and (mode != "dry-run" or cycle.state) and (not start or (cycle.state and cycle.state["status"] in FINAL)):
        discovered = producer_versions(api, repository, effective_mode)
        baseline = baseline_plan(api, repository, month, mode)
        ref = baseline["baseline"]
        if producer_lag(api, repository, ref, effective_mode, discovered, at_least=True):
            cycle_id = upstream_cycle_id(discovered)
            cycle = Cycle(api, repository, month, effective_mode, cycle_id)
            if cycle.state and cycle.state["status"] in FINAL:
                raise ValueError("Completed upstream cycle is missing from the current release; reconcile before publishing")
        elif not cycle.state:
            return {"action": "idle", "reason": "No active cycle or unconsumed public producer release"}
    if mode == "dry-run" and cycle.state:
        return {**cycle.state, "action": "preview", "writes": False}
    if cycle.state and cycle.state["status"] in FINAL:
        return {**cycle.state, "action": "complete"}
    if mode != "dry-run" and not cycle.state and not start and cycle_id == "monthly":
        return {"action": "idle", "reason": "No active monthly cycle to resume"}
    if mode == "execute" and authorization and cycle.state and cycle.state["status"] == "voting":
        from monthly_release_pr import advance
        with tempfile.TemporaryDirectory(prefix="monthly-vote-signing-") as signing:
            result = advance(api, cycle.state, directory, signing_environment(signing, api), cycle)
        return candidate_bundle(directory, output, result) if result["action"] == "qualify" else result
    if cycle.state and cycle.state["status"] in {"tagged", "staging", "draft", "awaiting-publication", "verifying"}:
        assert_tag(api, cycle.state)
        if cycle.state["status"] in {"awaiting-publication", "verifying"}:
            check_draft_review(api, cycle.state)
            remote = api.request(f"repos/{repository}/releases/{cycle.state['release_id']}")
            if remote.get("draft") is True:
                check_release_assets(api, cycle.state, True)
                return {**cycle.state, "action": "waiting", "reason": "Qualified draft awaits maintainer publication"}
            check_public_assets(api, cycle.state)
            cycle.state["status"] = "verifying"
            cycle.save(cycle.state)
            return {**cycle.state, "action": "verify"}
        if cycle.state["status"] == "draft":
            check_release_assets(api, cycle.state, True)
            return {**cycle.state, "action": "draft-check"}
        return {**cycle.state, "action": "package"}
    if cycle.state and cycle.state["status"] == "ready":
        existing_tag = api.optional(f"repos/{repository}/git/ref/tags/{cycle.state['tag']}")
        if existing_tag:
            cycle.state["tag_object"] = assert_tag(api, cycle.state)
            cycle.state["status"] = "tagged"
            cycle.save(cycle.state)
            return {**cycle.state, "action": "package"}
        if mode == "execute" and authorization and cycle.state.get("release_backport"):
            from monthly_release_pr import validate
            validate(api, cycle.state)
            return candidate_bundle(directory, output, cycle.state)
    state = cycle.state or baseline_plan(api, repository, month, mode)
    state["cycle_id"] = cycle_id
    if authorization:
        snapshot = authorization["snapshot"]
        if any(state[key] != snapshot[key] for key in ("baseline", "main", "branch", "tag")):
            raise ValueError("Assessed release baseline or proposed patch version changed")
        state.update(controller_issue=authorization["issue"], controller_plan_id=authorization["plan_id"])
        if authorization.get("release_issue"):
            state["release_issue"] = authorization["release_issue"]
        if mode == "execute":
            state["maintainer_roster_sha"] = snapshot["automation_sha"]
    baseline_plan(api, repository, month, mode, state)
    branch = api.optional(f"repos/{repository}/branches/{state['branch']}")
    main = api.optional(f"repos/{repository}/branches/{state['main']}")
    if branch is None or main is None:
        return {**state, "action": "blocked", "reason": "Reviewed main/release branches must exist before execution"}
    if authorization and not cycle.state and (
        branch["commit"]["sha"] != snapshot["release_head"]
        or main["commit"]["sha"] != snapshot["main_head"]
    ):
        raise ValueError("Assessed source refs changed before worker execution")
    if mode == "execute":
        check_publisher_guard(api, repository, state["branch"])
    producers = (authorization["producers"] if authorization else
                 discovered if discovered is not None else producer_versions(api, repository, effective_mode))
    state["producers"] = producers
    def eligible(pull):
        if dependabot(pull, state["main"]):
            return True
        if authorization:
            return False
        return False
    open_pulls = sorted([item for item in api.pulls(repository, state["main"], "open") if eligible(item)],
                        key=lambda item: item["number"])
    if authorization:
        from notation_release_controller import scoped_pull
        open_pulls = [pull for pull in open_pulls if scoped_pull(api, repository, pull, authorization)]
    if mode == "dry-run":
        previews = []
        for pull in open_pulls:
            check_files(api, repository, pull["number"])
            ready, reason = checks_ready(pull, api.checks(repository, pull["number"]))
            previews.append({"number": pull["number"], "ready": ready, "reason": reason})
        return {**state, "action": "preview", "pulls": previews,
                "producer_waits": producer_lag(api, repository, state["main"], effective_mode, producers),
                "writes": False}
    with tempfile.TemporaryDirectory(prefix="monthly-signing-") as signing:
        environment = signing_environment(signing, api)
        cycle.save(state)
        blocked = []
        for initial in open_pulls:
            pull = api.request(f"repos/{repository}/pulls/{initial['number']}")
            if not eligible(pull) or pull.get("state") != "open":
                raise ValueError("PR identity changed after discovery; retry from a fresh inventory")
            if pull["head"]["repo"]["full_name"] != repository:
                raise ValueError("Dependabot head is outside the target repository")
            if authorization and not scoped_pull(api, repository, pull, authorization):
                raise ValueError("Dependency PR changed after coordinator scope validation")
            check_files(api, repository, pull["number"])
            ready, reason = checks_ready(pull, api.checks(repository, pull["number"]))
            if not ready:
                blocked.append(f"Dependabot #{pull['number']}: {reason}")
                continue
            merged = api.request(
                f"repos/{repository}/pulls/{pull['number']}/merge", "PUT",
                {"sha": pull["head"]["sha"], "merge_method": "squash"},
            )
            if merged.get("merged") is not True or not SHA.fullmatch(merged.get("sha") or ""):
                raise ValueError("GitHub did not confirm a normal, non-bypass squash merge")
            state["merged"].append({"number": pull["number"], "commit": merged["sha"]})
            cycle.save(state)
        if blocked:
            state.update(status="waiting", reason="; ".join(blocked))
            cycle.save(state)
            return {**state, "action": "waiting"}
        waiting = producer_lag(api, repository, state["main"], mode, producers)
        if waiting:
            state.update(status="waiting", reason="; ".join(waiting))
            cycle.save(state)
            return {**state, "action": "waiting"}
        merged_pulls = inventory(api, repository, state["main"], authorization, state["merged"])
        if authorization:
            merged_pulls = [
                item for item in merged_pulls
                if scoped_pull(api, repository, api.request(f"repos/{repository}/pulls/{item['number']}"), authorization, state["merged"])
            ]
        if not merged_pulls:
            if cycle_id != "monthly":
                raise ValueError("Required producer update has no merged Dependabot backport provenance")
            state.update(status="skipped", reason="No merged dependency PRs")
            cycle.save(state)
            return {**state, "action": "complete"}
        git(["fetch", "--no-tags", "origin",
             f"+refs/heads/{state['main']}:refs/remotes/origin/{state['main']}",
             f"+refs/heads/{state['branch']}:refs/remotes/origin/{state['branch']}"], directory)
        expected = branch["commit"]["sha"]
        if git(["rev-parse", f"refs/remotes/origin/{state['branch']}"], directory) != expected:
            raise ValueError("Release branch changed during preparation; retry before backporting")
        common = git(["merge-base", f"refs/remotes/origin/{state['main']}", expected], directory)
        source_commits = set(git(["rev-list", f"{common}..refs/remotes/origin/{state['main']}"], directory).splitlines())
        merged_pulls = [item for item in merged_pulls if item["commit"] in source_commits]
        for item in merged_pulls:
            check_files(api, repository, item["number"])
        worktree = pathlib.Path(output) / "candidate"
        git(["worktree", "add", "--detach", str(worktree), expected], directory)
        environment["GIT_COMMITTER_DATE"] = state["started_at"]
        history = git(["log", "--format=%H%x00%B%x00", f"refs/remotes/origin/{state['branch']}"], directory)
        backports = []
        for item in merged_pulls:
            if f"(cherry picked from commit {item['commit']})" in history:
                records = history.split("\x00")
                backport = next(
                    records[index - 1].strip()
                    for index in range(1, len(records), 2)
                    if f"(cherry picked from commit {item['commit']})" in records[index]
                )
                if f"This reverts commit {backport}." in history:
                    raise ValueError(f"Backport #{item['number']} was reverted; reconcile before releasing")
                continue
            ancestors = subprocess.run(["git", "merge-base", "--is-ancestor", item["commit"], expected], cwd=directory)
            if ancestors.returncode == 0:
                continue
            if ancestors.returncode != 1:
                raise ValueError("Could not establish dependency commit ancestry")
            cherry = subprocess.run(
                ["git", "cherry-pick", "-x", "-S", item["commit"]],
                cwd=worktree, env=environment, capture_output=True, text=True,
            )
            if cherry.returncode:
                if not git(["ls-files", "--unmerged"], worktree) and not git(["diff", "--cached", "--name-only"], worktree):
                    git(["cherry-pick", "--skip"], worktree, environment)
                else:
                    raise ValueError(f"Dependency backport #{item['number']} failed:\n{cherry.stdout}{cherry.stderr}")
            else:
                backports.append({**item, "backport_commit": git(["rev-parse", "HEAD"], worktree)})
        candidate = git(["rev-parse", "HEAD"], worktree)
        if candidate == expected and cycle_id == "monthly":
            state.update(status="skipped", reason="All dependency updates are already on the release branch")
            cycle.save(state)
            return {**state, "action": "complete"}
        state.update(
            status="ready", reason="", merged=merged_pulls, backports=backports, previous_head=expected,
            commit=candidate,
        )
        root_manifest = json.loads(run(["go", "mod", "edit", "-json"], cwd=worktree))
        if mode == "execute" and root_manifest.get("Replace"):
            raise ValueError("Production release manifests must not contain trial replacements")
        if not re.fullmatch(r"\d+\.\d+(?:\.\d+)?", root_manifest["Go"]):
            raise ValueError("Unsupported minimum Go version")
        state["minimum_go"] = ".".join(root_manifest["Go"].split(".")[:2])
        for name, producer in producers.items():
            for module_directory in modules(repository):
                manifest = json.loads(run(["go", "mod", "edit", "-json"], cwd=worktree / module_directory))
                module = f"github.com/notaryproject/{name}"
                if module_directory == "." or any(item["Path"] == module for item in manifest.get("Require") or []):
                    if requirement(manifest, module, producer["repository"], mode) != producer["version"]:
                        raise ValueError("Release backport does not consume the required public producer version")
        if mode == "execute" and authorization:
            from monthly_release_pr import advance
            state = advance(api, state, directory, environment, cycle)
            if state["action"] != "qualify":
                return state
        cycle.save(state)
        git(["update-ref", "refs/heads/monthly-patch-candidate", state["commit"]], directory)
        git(["bundle", "create", str(pathlib.Path(output) / "candidate.bundle"),
             "refs/heads/monthly-patch-candidate"], directory)
        return {**state, "action": "qualify"}


def import_candidate(directory, plan, bundle):
    validate_plan(plan)
    if not SHA.fullmatch(plan["commit"]):
        raise ValueError("Invalid qualified candidate commit")
    git(["fetch", str(bundle), "refs/heads/monthly-patch-candidate"], directory)
    if git(["rev-parse", "FETCH_HEAD"], directory) != plan["commit"]:
        raise ValueError("Candidate bundle does not match the prepared commit")
    git(["checkout", "--detach", plan["commit"]], directory)
    if git(["status", "--porcelain"], directory):
        raise ValueError("Candidate checkout is dirty")


def stage_tag(api, plan, directory):
    validate_plan(plan)
    policy(plan["repository"], plan["mode"], os.environ, api.request(f"repos/{plan['repository']}"))
    coordinator_authorization(api, plan["repository"], plan["mode"], plan["month"], plan)
    if plan["mode"] == "execute" and plan.get("controller_issue"):
        from monthly_release_pr import validate
        validate(api, plan)
    if plan["status"] != "ready":
        raise ValueError("Only a qualified ready candidate can be tagged")
    repository = plan["repository"]
    cycle = Cycle(api, repository, plan["month"], plan["mode"], plan.get("cycle_id", "monthly"))
    if not cycle.state or cycle.state.get("commit") != plan["commit"]:
        raise ValueError("Prepared cycle state changed before publication")
    if cycle.state["status"] != "ready" or cycle.state["tag"] != plan["tag"] or cycle.state["previous_head"] != plan["previous_head"]:
        raise ValueError("Cycle changed before tagging")
    baseline_plan(api, repository, plan["month"], plan["mode"], plan)
    if git(["rev-parse", "HEAD"], directory) != plan["commit"] or git(["status", "--porcelain"], directory):
        raise ValueError("Tagging requires the exact clean qualified candidate")
    if api.request(f"repos/{repository}/branches/{plan['branch']}")["commit"]["sha"] != plan["previous_head"]:
        raise ValueError("Release branch advanced during qualification")
    with tempfile.TemporaryDirectory(prefix="monthly-local-tag-") as signing:
        environment = signing_environment(signing, api)
        git(["tag", "-s", plan["tag"], plan["commit"], "-m",
             f"Monthly dependency patch {plan['month']}\nQualified commit: {plan['commit']}"], directory, environment)
    return plan


def tag_candidate(api, plan, directory):
    validate_plan(plan)
    policy(plan["repository"], plan["mode"], os.environ, api.request(f"repos/{plan['repository']}"))
    coordinator_authorization(api, plan["repository"], plan["mode"], plan["month"], plan)
    if plan["mode"] == "execute" and plan.get("controller_issue"):
        from monthly_release_pr import validate
        validate(api, plan)
    repository = plan["repository"]
    cycle = Cycle(api, repository, plan["month"], plan["mode"], plan.get("cycle_id", "monthly"))
    if not cycle.state or any(cycle.state.get(key) != plan.get(key) for key in ("commit", "tag", "previous_head", "status")) or plan["status"] != "ready":
        raise ValueError("Qualified cycle changed before remote tagging")
    baseline_plan(api, repository, plan["month"], plan["mode"], plan)
    if git(["rev-parse", "HEAD"], directory) != plan["commit"] or git(["status", "--porcelain", "--untracked-files=no"], directory):
        raise ValueError("Qualified source changed during packaging")
    if git(["rev-parse", f"{plan['tag']}^{{commit}}"], directory) != plan["commit"]:
        raise ValueError("Local release tag does not identify the qualified commit")
    ref = api.optional(f"repos/{repository}/git/ref/tags/{plan['tag']}")
    if ref:
        assert_tag(api, plan)
    else:
        with tempfile.TemporaryDirectory(prefix="monthly-tag-signing-") as signing:
            environment = signing_environment(signing, api)
            git(["-c", "push.followTags=false", "push", "--atomic",
                 f"--force-with-lease=refs/heads/{plan['branch']}:{plan['previous_head']}",
                 f"--force-with-lease=refs/tags/{plan['tag']}:",
                 "origin", f"{plan['previous_head'] if plan.get('release_backport') else plan['commit']}:refs/heads/{plan['branch']}",
                 f"refs/tags/{plan['tag']}:refs/tags/{plan['tag']}"], directory, environment)
        assert_tag(api, plan)
    plan["tag_object"] = assert_tag(api, plan)
    plan["status"] = "tagged"
    cycle.save(plan)
    return {**plan, "action": "package"}


def publish(api, plan, directory):
    """Stage a true draft only. Automation never makes a release public."""
    validate_plan(plan)
    policy(plan["repository"], plan["mode"], os.environ, api.request(f"repos/{plan['repository']}"))
    authorization = coordinator_authorization(api, plan["repository"], plan["mode"], plan["month"], plan)
    if authorization and authorization["snapshot"].get("verification_recovery") is not None:
        raise ValueError("Verification-only recovery cannot build or republish release assets")
    if plan["mode"] == "execute" and plan.get("controller_issue"):
        from monthly_release_pr import validate
        validate(api, plan)
    assert_tag(api, plan)
    cycle = Cycle(api, plan["repository"], plan["month"], plan["mode"], plan.get("cycle_id", "monthly"))
    if (not cycle.state or cycle.state["status"] not in {"tagged", "staging"}
            or any(cycle.state.get(key) != plan.get(key) for key in ("commit", "tag", "previous_head"))):
        raise ValueError("Cycle changed before recording draft release")
    assets = []
    for path in sorted(pathlib.Path(directory).iterdir()):
        if path.is_file() and (path.name.endswith((".tar.gz", ".zip", "checksums.txt"))):
            assets.append({"name": path.name, "sha256": hashlib.sha256(path.read_bytes()).hexdigest(), "size": path.stat().st_size})
    count = 7 if plan["repository"].endswith("/notation") else 2
    if len(assets) != count:
        raise ValueError("Unexpected release asset set")
    if {item["name"] for item in assets} != set(asset_names(plan["repository"].split("/")[1], plan["tag"])):
        raise ValueError("Unexpected release asset names")
    notes = (
        f"# {plan['repository'].split('/')[1]} {plan['tag']}\n\n"
        + f"{plan_marker(plan)}\n"
        + f"Qualified commit: {plan['commit']}\n"
        + f"Previous stable tag: {plan['baseline']}\n\n"
        + "Dependabot backports:\n"
        + "\n".join(f"* #{item['number']}: {item['commit']}" for item in plan["merged"])
        + "\n\nAsset verification manifest:\n```json\n" + json.dumps(assets, indent=2) + "\n```\n"
        + "\nQualification and review are recorded in the separate qualified draft review issue. "
        + "Do not publish unless that issue exists and its exact-asset checks passed.\n"
    )
    if cycle.state["status"] == "staging":
        plan = copy.deepcopy(cycle.state)
        recorded = [{key: item[key] for key in ("name", "sha256", "size")} for item in plan["assets"]]
        if recorded != assets:
            raise ValueError("Retry must retain the exact originally staged asset bytes")
        existing = api.request(f"repos/{plan['repository']}/releases/{plan['release_id']}")
        if (existing.get("id") != plan["release_id"] or existing.get("draft") is not True
                or existing.get("prerelease") is not False or existing.get("tag_name") != plan["tag"]
                or existing.get("body") != notes):
            raise ValueError("Recorded draft identity or qualification manifest changed")
    else:
        if any(item["tag_name"] == plan["tag"] for item in api.pages(f"repos/{plan['repository']}/releases")):
            raise ValueError("Unrecorded existing release requires reconciliation; never adopt or replace it")
        existing = api.request(f"repos/{plan['repository']}/releases", "POST", {
            "tag_name": plan["tag"], "target_commitish": plan["commit"], "name": plan["tag"],
            "body": notes, "draft": True, "prerelease": False, "make_latest": "false",
        })
        identifier = existing.get("id")
        if type(identifier) is not int or identifier <= 0 or existing.get("draft") is not True:
            raise ValueError("Draft creation was not confirmed")
        plan.update(status="staging", release_id=identifier, assets=assets)
        cycle.save(plan)
    remote = list(api.pages(f"repos/{plan['repository']}/releases/{plan['release_id']}/assets"))
    if len({item["name"] for item in remote}) != len(remote) or any(item["name"] not in {a["name"] for a in assets} for item in remote):
        raise ValueError("Unexpected or duplicate staged draft assets")
    for item in plan["assets"]:
        found = next((a for a in remote if a["name"] == item["name"]), None)
        if found is None:
            from urllib.parse import quote
            found = json.loads(run([
                "gh", "api", "--method", "POST",
                f"https://uploads.github.com/repos/{plan['repository']}/releases/{plan['release_id']}/assets?name={quote(item['name'], safe='')}",
                "-H", "Content-Type: application/octet-stream", "--input", str(pathlib.Path(directory) / item["name"]),
            ]))
        if (type(found.get("id")) is not int or found["id"] <= 0
                or ("id" in item and item["id"] != found["id"]) or found.get("name") != item["name"]
                or found.get("size") != item["size"] or found.get("digest") != f"sha256:{item['sha256']}"):
            raise ValueError("Draft uploaded asset differs from the qualified bytes")
        item["id"] = found["id"]
        cycle.save(plan)
    plan.update(status="draft", reason="Draft download/native verification is pending; publication is not authorized")
    check_release_assets(api, plan, True)
    cycle.save(plan)
    return {**plan, "action": "draft-check"}


def complete(api, plan):
    validate_plan(plan)
    policy(plan["repository"], plan["mode"], os.environ, api.request(f"repos/{plan['repository']}"))
    coordinator_authorization(api, plan["repository"], plan["mode"], plan["month"], plan)
    cycle = Cycle(api, plan["repository"], plan["month"], plan["mode"], plan.get("cycle_id", "monthly"))
    if (not cycle.state or cycle.state["status"] not in {"draft", "verifying"}
            or any(cycle.state.get(key) != plan.get(key) for key in ("status", "commit", "tag", "release_id", "assets"))):
        raise ValueError("No matching draft/public release awaiting successful verification")
    plan = copy.deepcopy(cycle.state)
    assert_tag(api, plan)
    if plan["status"] == "draft":
        check_release_assets(api, plan, True)
        run_id = os.environ.get("GITHUB_RUN_ID", "")
        if not run_id.isdecimal() or int(run_id) <= 0:
            raise ValueError("Qualified draft requires its native verification run")
        if not plan.get("draft_verification"):
            plan["draft_plan_id"] = plan["controller_plan_id"]
            plan["draft_verification"] = {
                "run_id": int(run_id), "worker_sha": os.environ["MONTHLY_PATCH_WORKER_SHA"],
                "platforms": ["linux_amd64", "darwin_arm64", "windows_amd64"],
                "evidence": f"https://github.com/{plan['repository']}/actions/runs/{run_id}",
                "artifacts": ["monthly-source-*", "monthly-build-evidence", "monthly-draft-*"],
            }
            cycle.save(plan)
        body = draft_review_body(plan)
        from notation_release_controller import HOST, controller_body, decode_issue
        import release_issue_topology as threads
        actor = os.environ.get("MONTHLY_PATCH_ACTOR")
        root = (api.request(f"repos/{HOST}/issues/{plan['controller_issue']}")
                if plan.get("release_issue") else None)
        state = decode_issue(root, actor) if root else None
        if plan.get("release_issue") and (state is None or not state.get("release_issues")):
            raise ValueError("Qualified draft release issue topology is missing")
        if state is not None and state.get("release_issues"):
            if root.get("body") != controller_body(state) or state["plan"]["plan_id"] != plan["controller_plan_id"]:
                raise ValueError("Qualified draft plan issue changed")
            trusted_editor(api, HOST, root)
            number = state["release_issues"][plan["repository"]]
            if plan["repository"] != HOST:
                threads.checked_issue(api, actor, plan["repository"], number,
                                      threads.scope_body(state["plan"], plan["repository"], state["release_issues"]))
            main = api.request(f"repos/{plan['repository']}/branches/main")["commit"]["sha"]
            from notation_release_controller import repository_content
            from monthly_release_vote import parse_repo_maintainers
            roster = parse_repo_maintainers(repository_content(api, plan["repository"], "MAINTAINERS", main))
            body = " ".join(f"@{login}" for login in sorted(roster)) + "\n\n" + body
            identity = f"draft-ready-{plan['repository']}:{plan['tag']}:{plan['release_id']}"
            comment = threads.comment_once(api, actor, plan["repository"], number, identity, body)
            plan.update(draft_review_issue=number, draft_review_comment=comment["id"],
                        draft_review_body=comment["body"], status="awaiting-publication",
                        reason="Qualified draft awaits maintainer Publish release action")
            cycle.save(plan)
            check_draft_review(api, plan)
            return {**plan, "action": "waiting"}
        marker = body.split("\n", 1)[0]
        reviews = [issue for issue in api.pages(f"repos/{plan['repository']}/issues?state=all")
                   if marker in (issue.get("body") or "")]
        if len(reviews) > 1:
            raise ValueError("Duplicate draft review issues require reconciliation")
        if reviews:
            issue = reviews[0]
            plan["draft_review_issue"] = issue["number"]
            check_draft_review(api, plan)
        else:
            issue = api.request(f"repos/{plan['repository']}/issues", "POST", {
                "title": f"Review qualified draft: {plan['tag']} ({plan['month']})",
                "body": body, "state": "open",
            })
            plan["draft_review_issue"] = issue["number"]
            check_draft_review(api, plan)
        plan.update(status="awaiting-publication", reason="Qualified draft awaits maintainer Publish release action")
        cycle.save(plan)
        return {**plan, "action": "waiting"}
    check_public_assets(api, plan)
    plan.update(status="published", reason="Source, artifacts, and public downloaded-package checks passed")
    cycle.save(plan)
    return {**plan, "action": "complete"}


def dispatch_consumers(api, repository, mode, tag, notified=None):
    owner, producer = repository.split("/")
    consumers = [f"{owner}/{name}" for name, dependencies in PROJECTS.items() if producer in dependencies]
    notified = list(notified or [])
    if len(set(notified)) != len(notified) or any(target not in consumers for target in notified):
        raise ValueError("Invalid downstream notification progress")
    for target in consumers:
        if target in notified:
            continue
        api.request(f"repos/{target}/dispatches", "POST", {
            "event_type": "dependency-patch-upstream",
            "client_payload": {"producer": repository, "tag": tag, "mode": mode},
        })
        notified.append(target)
        yield list(notified)


def notify(api, plan):
    validate_plan(plan)
    policy(plan["repository"], plan["mode"], os.environ, api.request(f"repos/{plan['repository']}"))
    cycle = Cycle(api, plan["repository"], plan["month"], plan["mode"], plan.get("cycle_id", "monthly"))
    if not cycle.state or cycle.state["status"] != "published" or any(cycle.state.get(key) != plan.get(key) for key in ("tag", "commit")):
        raise ValueError("Only a successfully verified public cycle can notify consumers")
    assert_tag(api, cycle.state)
    check_public_assets(api, cycle.state)
    for notified in dispatch_consumers(api, plan["repository"], plan["mode"], plan["tag"], cycle.state.get("notified")):
        cycle.state["notified"] = notified
        cycle.save(cycle.state)
    return {**cycle.state, "action": "complete"}


def announce(api, repository, mode, tag):
    if not REPOSITORY.fullmatch(repository) or mode != "execute":
        raise ValueError("Invalid release notification repository or mode")
    metadata = api.request(f"repos/{repository}")
    if mode == "execute":
        if repository.split("/")[0] != "notaryproject" or os.environ.get("MONTHLY_PATCH_ENABLED") != "true":
            raise ValueError("Canonical release notifications are not enabled")
    if os.environ.get("MONTHLY_PATCH_TOKEN_READY") != "true" or not os.environ.get("MONTHLY_PATCH_ACTOR"):
        raise ValueError("Release notifications require the configured automation credential")
    if os.environ.get("GITHUB_REF") != f"refs/tags/{tag}":
        raise ValueError("Release notification must identify the triggering public tag")
    release = api.request(f"repos/{repository}/releases/tags/{tag}")
    if release["draft"] or release["tag_name"] != tag:
        raise ValueError("Release notification does not identify a public release")
    if "<!-- dependency-patch:" in (release.get("body") or ""):
        return {"action": "idle", "reason": "Automated release notification waits for successful package verification"}
    if mode != "execute" or release["prerelease"] or semver(tag) is None:
        return {"action": "idle", "reason": "Only ordinary stable releases need a publication notification"}
    notified = []
    for notified in dispatch_consumers(api, repository, mode, tag):
        pass
    return {"action": "notified", "repository": repository, "tag": tag, "notified": notified}


def failure_details(environment):
    outcomes = json.loads(environment.get("MONTHLY_PATCH_JOB_RESULTS", "{}"))
    allowed = {"prepare", "qualify", "package", "publish", "verify", "complete"}
    results = {"success", "failure", "cancelled", "skipped"}
    if (not isinstance(outcomes, dict) or any(name not in allowed for name in outcomes)
            or any(not isinstance(result, dict) or result.get("result") not in results
                   for result in outcomes.values())):
        raise ValueError("Invalid failed-stage evidence")
    run_id, attempt = environment.get("GITHUB_RUN_ID", ""), environment.get("GITHUB_RUN_ATTEMPT", "1")
    if not run_id.isdecimal() or int(run_id) <= 0 or not attempt.isdecimal() or int(attempt) <= 0:
        raise ValueError("Failure evidence requires its workflow run and attempt")
    return {"run_id": int(run_id), "run_attempt": int(attempt),
            "stages": [name for name, result in outcomes.items() if result["result"] == "failure"]}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=["prepare", "import", "stage-tag", "tag", "publish", "complete", "notify", "notify-controller", "announce", "failure"])
    parser.add_argument("--repository", default=os.environ.get("GITHUB_REPOSITORY", ""))
    parser.add_argument("--mode", choices=["dry-run", "execute"], default="dry-run")
    parser.add_argument("--month", default=datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m"))
    parser.add_argument("--directory", type=pathlib.Path, default=pathlib.Path.cwd())
    parser.add_argument("--output", type=pathlib.Path, required=True)
    parser.add_argument("--plan", type=pathlib.Path)
    parser.add_argument("--bundle", type=pathlib.Path)
    parser.add_argument("--start", action="store_true")
    parser.add_argument("--reconcile", action="store_true")
    parser.add_argument("--cycle-id", default="monthly")
    parser.add_argument("--tag")
    parser.add_argument("--reason", default="Workflow failed. Inspect its retained evidence and logs before retrying.")
    args = parser.parse_args()
    try:
        if not REPOSITORY.fullmatch(args.repository) or not MONTH.fullmatch(args.month):
            raise ValueError("Invalid repository/month")
        args.output.mkdir(parents=True, exist_ok=True)
        api = GitHub()
        if args.command == "failure":
            policy(args.repository, args.mode, os.environ, api.request(f"repos/{args.repository}"))
            cycle = Cycle(api, args.repository, args.month, args.mode, args.cycle_id)
            if cycle.state and cycle.state["status"] not in FINAL:
                coordinator_authorization(api, args.repository, args.mode, args.month, cycle.state)
                cycle.state["reason"] = args.reason
                cycle.state["failure_details"] = failure_details(os.environ)
                cycle.save(cycle.state)
            return 0
        if args.command == "prepare":
            result = prepare(api, args.repository, args.mode, args.month, args.directory, args.output, args.start, args.reconcile)
        elif args.command == "announce":
            result = announce(api, args.repository, args.mode, args.tag)
        elif args.command == "notify-controller":
            authorization = coordinator_authorization(api, args.repository, args.mode, args.month, notification=True)
            if authorization is None:
                raise ValueError("Controller notification requires a dispatched coordinator worker")
            from notation_release_controller import HOST
            api.request(f"repos/{HOST}/dispatches", "POST", {
                "event_type": "notation-release-progress",
                "client_payload": {"controller_issue": authorization["issue"], "plan_id": authorization["plan_id"]},
            })
            result = {"action": "notified", "controller_issue": authorization["issue"]}
        else:
            if args.plan is None:
                raise ValueError("An inspected candidate plan is required")
            plan = json.loads(args.plan.read_text())
            if plan["repository"] != args.repository:
                raise ValueError("Candidate belongs to another repository")
            if args.command == "import":
                if args.bundle is None:
                    raise ValueError("Candidate bundle is required")
                import_candidate(args.directory, plan, args.bundle)
                result = plan
            elif args.command == "tag":
                result = tag_candidate(api, plan, args.directory)
            elif args.command == "stage-tag":
                result = stage_tag(api, plan, args.directory)
            elif args.command == "publish":
                result = publish(api, plan, args.directory)
            elif args.command == "notify":
                result = notify(api, plan)
            else:
                result = complete(api, plan)
        (args.output / "plan.json").write_text(json.dumps(result, indent=2) + "\n")
        print(json.dumps(result, indent=2))
        if os.environ.get("GITHUB_OUTPUT"):
            with open(os.environ["GITHUB_OUTPUT"], "a", encoding="utf-8") as output:
                for field in ("action", "commit", "tag", "baseline", "minimum_go", "mode", "month", "cycle_id", "status"):
                    value = result.get(field, "")
                    if "\n" in str(value):
                        raise ValueError("Invalid multiline workflow output")
                    output.write(f"{field}={value}\n")
        return 1 if result.get("action") == "blocked" else 0
    except (ValueError, KeyError, OSError, RuntimeError, json.JSONDecodeError) as error:
        print(f"Monthly dependency release failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
