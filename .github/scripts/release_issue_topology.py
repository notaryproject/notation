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

"""One human release thread per repository, separate from technical cycle state."""
import hashlib
import html
import re

PROJECTS = {"notation-core-go": (), "notation-go": ("notation-core-go",),
            "notation": ("notation-core-go", "notation-go")}


def text(value):
    escaped = html.escape(" ".join(str(value).split())[:1000], quote=True)
    return re.sub(r"[\\`*_{}\[\]()#+.!|@]", lambda match: f"&#{ord(match[0])};", escaped)


def body_hash(body):
    return hashlib.sha256(body.encode()).hexdigest()


def repositories(host):
    if not re.fullmatch(r"[A-Za-z0-9-]+/notation", host):
        raise ValueError("Invalid release issue host")
    owner = host.split("/")[0]
    return tuple(f"{owner}/{name}" for name in PROJECTS)


def validate_routes(host, routes, complete=True):
    expected = set(repositories(host))
    if (not isinstance(routes, dict) or not set(routes) <= expected
            or (complete and set(routes) != expected)
            or any(type(value) is not int or value <= 0 for value in routes.values())):
        raise ValueError("Invalid release issue routing")
    return routes


def issue_url(repository, number):
    if not re.fullmatch(r"[A-Za-z0-9-]+/(notation-core-go|notation-go|notation)", repository):
        raise ValueError("Invalid release issue repository")
    if type(number) is not int or number <= 0:
        raise ValueError("Invalid release issue number")
    return f"https://github.com/{repository}/issues/{number}"


def marker(plan, repository):
    if repository not in repositories(plan["host"]) or not re.fullmatch(r"\d{4}-\d{2}", plan["month"]):
        raise ValueError("Invalid release scope identity")
    return f"<!-- notation-release-scope:v1:{repository}:{plan['month']} -->"


def dependency_links(plan, repository, routes):
    validate_routes(plan["host"], routes)
    owner, name = repository.split("/")
    if repository not in routes:
        raise ValueError("Missing release issue route")
    dependencies = []
    for dependency in PROJECTS[name]:
        producer = f"{owner}/{dependency}"
        dependencies.append(f"[{producer}#{routes[producer]}]({issue_url(producer, routes[producer])})")
    return "Dependencies: " + ("; ".join(dependencies) if dependencies else "None.") + "\n"


def scope_body(plan, repository, routes):
    validate_routes(plan["host"], routes)
    snapshot = plan["snapshots"][repository]
    lines = [marker(plan, repository), f"# {repository} patch release: {plan['month']}", "",
             f"Current version: `{snapshot['baseline']}`. Proposed patch: `{snapshot['tag']}`.", "",
             "## Changes", "",
             "The scope merges qualifying Dependabot PRs and cuts a patch release.", ""]
    source_links = []
    for field, revision, label in (("pulls", "head", plan.get("scope_label", "Proposed HEAD")),
                                   ("merged", "commit", "Original source merge")):
        for item in snapshot[field]:
            number, sha = item["number"], item[revision]
            if (type(number) is not int or number <= 0 or not isinstance(sha, str)
                    or not re.fullmatch(r"[0-9a-f]{40}", sha)):
                raise ValueError("Invalid release scope PR identity")
            additional = []
            for commit in item.get("commits", []):
                if not re.fullmatch(r"[0-9a-f]{40}", commit["sha"]):
                    raise ValueError("Invalid proposed commit")
                if commit["sha"] != sha:
                    additional.append(f"[`{commit['sha'][:12]}`](https://github.com/{repository}/commit/{commit['sha']}) "
                                      f"{text(commit.get('subject', ''))}")
            changes = item.get("dependency_changes", [])
            change_text = "; ".join(f"`{text(change['module'])}` `{text(change['before'])}` to `{text(change['after'])}`"
                                    for change in changes)
            lines.append(f"- [#{number}](https://github.com/{repository}/pull/{number}): "
                         f"{change_text or text(item.get('title', 'Dependabot dependency update'))}.")
            source_links.append(f"- PR #{number} {label}: [`{sha[:12]}`]"
                                f"(https://github.com/{repository}/commit/{sha})"
                                f"{'; additional proposed: ' + '; '.join(additional) if additional else ''}.")
    if not snapshot["pulls"] and not snapshot["merged"]:
        lines.append("No known source Dependabot PR; exact planned producer adoption may still be required.")
    if PROJECTS[repository.split("/")[1]]:
        versions = [f"`{name}` `{plan['snapshots'][repository.split('/')[0] + '/' + name]['tag']}`"
                    for name in PROJECTS[repository.split("/")[1]]]
        lines += ["", "Planned library adoption: " + ", ".join(versions) + "."]
    lines += ["", dependency_links(plan, repository, routes)]
    if source_links:
        lines += ["", "<details>", "<summary>Exact source revisions</summary>", "", *source_links, "", "</details>"]
    proposals = plan.get("bugfix_proposals", {}).get(repository, {}).get("proposals", [])
    candidates = [item for item in proposals if item["classification"] in {"candidate_bugfix", "needs_review"}
                  and item["applicability"]["status"] != "not_applicable"
                  and not any("already_present" in blocker or "revert" in blocker for blocker in item["blockers"])]
    if candidates:
        lines += ["", "## Proposed bug-fix backports", "",
                  "These candidates are not included or executable. Applicability and separate scope approval are required.", ""]
        for item in candidates:
            pr = (f"[#{item['pr_number']}](https://github.com/{repository}/pull/{item['pr_number']})"
                  if item["pr_number"] else "No verified PR")
            lines.append(f"- {pr}, [`{item['source_sha'][:12]}`]"
                         f"(https://github.com/{repository}/commit/{item['source_sha']}): "
                         f"{text(item['rationale'])} Applicability: {text(item['applicability']['status'])}.")
    lines += ["", "## Status", "",
              "Required CI and reviews must pass. Maintainer action: review this scope and linked dependencies. "
              "Blockers and qualified draft-ready requests appear in comments; publication remains manual.", ""]
    return "\n".join(lines)


def dependency_changes(files):
    """Extract only explicit go.mod version-line changes as review data, never authority."""
    changes = []
    for item in files:
        if item.get("filename", "").split("/")[-1] != "go.mod":
            continue
        before, after = {}, {}
        for line in item.get("patch", "").splitlines():
            match = re.fullmatch(r"([+-])\s*(?:require\s+)?([A-Za-z0-9._~/-]+)\s+(v[^\s]+)(?:\s+// indirect)?", line)
            if match:
                (before if match[1] == "-" else after)[match[2]] = match[3]
        for module in sorted(before.keys() & after.keys()):
            if before[module] != after[module]:
                change = {"module": module, "before": before[module], "after": after[module]}
                if change not in changes:
                    changes.append(change)
    return changes


def placeholder(plan, repository):
    return (f"{marker(plan, repository)}\nRelease issue wiring is incomplete; execution is not authorized.\n"
            f"Plan: `{plan['plan_id']}`.\n")


def checked_issue(api, actor, repository, number, expected):
    issue = api.request(f"repos/{repository}/issues/{number}")
    if ("pull_request" in issue or issue.get("user", {}).get("login") != actor
            or issue.get("body") != expected or issue.get("state") != "open"):
        raise ValueError("Release scope issue changed or has an untrusted author")
    editor = api.issue_editor(repository, number)
    if editor not in (None, actor):
        raise ValueError("Release scope issue has an untrusted editor")
    return issue


def ensure(api, actor, plan, root_number, routes=None, previous_hashes=None):
    """Create Core, then Go, then wire all links; exact placeholders make retries safe."""
    host = plan["host"]
    routes = dict(routes or {})
    validate_routes(host, routes, complete=False)
    if host in routes and routes[host] != root_number:
        raise ValueError("CLI release issue changed")
    routes[host] = root_number
    issues = {}
    for repository in repositories(host):
        if repository == host:
            continue
        matches = [issue for issue in api.pages(f"repos/{repository}/issues?state=all")
                   if marker(plan, repository) in (issue.get("body") or "")]
        if len(matches) > 1:
            raise ValueError("Duplicate repository release issues require reconciliation")
        if matches:
            issue = matches[0]
            if repository in routes and routes[repository] != issue["number"]:
                raise ValueError("Repository release route changed")
        elif repository in routes:
            raise ValueError("Recorded repository release issue is missing")
        else:
            issue = api.request(f"repos/{repository}/issues", "POST", {
                "title": f"Patch release scope: {plan['month']} ({repository.split('/')[1]})",
                "body": placeholder(plan, repository),
            })
        if (type(issue.get("number")) is not int or issue["number"] <= 0
                or issue.get("user", {}).get("login") != actor or "pull_request" in issue
                or issue.get("state") != "open"):
            raise ValueError("Repository release issue creation was not confirmed")
        if api.issue_editor(repository, issue["number"]) not in (None, actor):
            raise ValueError("Repository release issue has an untrusted editor")
        routes[repository], issues[repository] = issue["number"], issue
    hashes = {}
    for repository, issue in issues.items():
        expected = scope_body(plan, repository, routes)
        before = issue.get("body", "")
        if (before not in (expected, placeholder(plan, repository))
                and body_hash(before) != (previous_hashes or {}).get(repository)):
            raise ValueError("Existing repository release scope does not match the reviewed plan")
        if before != expected:
            api.request(f"repos/{repository}/issues/{issue['number']}", "PATCH", {"body": expected})
        checked_issue(api, actor, repository, issue["number"], expected)
        hashes[repository] = body_hash(expected)
    return routes, hashes


def comment_once(api, actor, repository, number, identity, body):
    tag = f"<!-- notation-release-notice:v1:{identity} -->"
    expected = f"{tag}\n{body}"
    matches = [item for item in api.pages(f"repos/{repository}/issues/{number}/comments")
               if tag in (item.get("body") or "") and item.get("user", {}).get("login") == actor]
    if len(matches) > 1:
        raise ValueError("Duplicate release notifications require reconciliation")
    if matches:
        if matches[0].get("body") != expected:
            raise ValueError("Release notification receipt changed")
        return matches[0]
    item = api.request(f"repos/{repository}/issues/{number}/comments", "POST", {"body": expected})
    if (item.get("body") != expected or item.get("user", {}).get("login") != actor
            or type(item.get("id")) is not int or item["id"] <= 0):
        raise ValueError("Release notification was not confirmed")
    return item
