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

"""Validate read-only assessments and authorize the ordered release DAG."""

import argparse
import base64
import copy
import datetime
import hashlib
import html
import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile
import urllib.parse

import monthly_release as release
from monthly_release_checks import scan_messages

OWNER = "notaryproject"
HOST = f"{OWNER}/notation"
MODE = "execute"
REPOSITORIES = tuple(f"{OWNER}/{name}" for name in release.PROJECTS)
POLICY_PATH = ".github/monthly-patch-policy.yml"
AGENT_PATH = ".github/workflows/notation-release-agent.lock.yml"
WORKER_PATH = ".github/workflows/monthly-patch-release.yml"
SHARED_PATH = ".github/workflows/monthly-patch-worker.yml"
DIGEST = re.compile(r"[0-9a-f]{64}$")
DECISIONS = {"release", "skip", "defer"}
NODE_STATUSES = {"pending", "dispatching", "running", "failed", "published", "skipped", "deferred"}
TERMINAL = {"published", "skipped"}


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"Duplicate JSON field: {key}")
        result[key] = value
    return result


def load_json(text):
    return json.loads(text, object_pairs_hook=unique_object)


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def utc_now():
    return datetime.datetime.now(datetime.timezone.utc).isoformat()


def positive_id(value):
    if type(value) is not int or value <= 0:
        raise ValueError("Expected a positive integer identifier")
    return value


def repository_metadata(api, repository):
    if repository not in REPOSITORIES:
        raise ValueError("Coordinator repository scope is fixed")
    metadata = api.request(f"repos/{repository}")
    if metadata.get("fork") or metadata.get("private") or metadata.get("default_branch") != "main":
        raise ValueError(f"Expected a public canonical repository with trusted default main: {repository}")
    return metadata


def repository_content(api, repository, path, commit):
    item = api.optional(f"repos/{repository}/contents/{path}?ref={commit}")
    if item is None:
        return None
    if item.get("encoding") != "base64":
        raise ValueError(f"Cannot inspect trusted worker installation: {path}")
    return base64.b64decode(item["content"], validate=False).decode("utf-8")


def worker_binding(api, repository, commit):
    """Bind the dispatch caller to its literal, immutable central implementation."""
    text = repository_content(api, repository, WORKER_PATH, commit)
    if text is None:
        return None
    sections = re.findall(r"(?m)^jobs:\s*\n([\s\S]*)", text)
    if len(sections) != 1:
        raise ValueError("Missing or duplicate caller jobs")
    section = sections[0]
    if re.findall(r"(?m)^  ([A-Za-z_][A-Za-z0-9_-]*):", section) != ["release"]:
        raise ValueError("Release caller must contain only the release reusable job")
    references = re.findall(r"(?m)^    uses: ([^\n#]+)$", section)
    pins = re.findall(r"(?m)^      worker_sha: ([^\n#]+)$", section)
    if len(references) != 1 or len(pins) != 1 or re.search(r"(?m)^\s*(?:<<:|[&*!])", text):
        raise ValueError("Ambiguous caller shared worker binding")
    reference, pin = references[0].strip(), pins[0].strip()
    if repository == HOST:
        if reference != "./" + SHARED_PATH or pin != "${{ github.sha }}":
            raise ValueError("CLI caller must use the same-commit relative shared worker")
        sha = commit
    else:
        match = re.fullmatch(re.escape(HOST + "/" + SHARED_PATH + "@") + r"([0-9a-f]{40})", reference or "")
        if match is None or pin != match[1]:
            raise ValueError("Library caller must bind uses and worker_sha to the same full central SHA")
        sha = pin
    if not release.SHA.fullmatch(sha):
        raise ValueError("Invalid shared worker revision")
    return {"worker_repository": HOST, "worker_path": SHARED_PATH, "worker_sha": sha,
            "worker_digest": automation_digest(api, HOST, sha)}


def installed_worker(api, repository, commit, binding=None):
    binding = binding or worker_binding(api, repository, commit)
    if binding is None:
        return False
    files = {
        SHARED_PATH: ("MONTHLY_PATCH_COORDINATOR_REQUIRED: 'true'", "  workflow_call:"),
        ".github/scripts/notation_release_controller.py": ("def worker_authorization(",),
        ".github/scripts/monthly_release.py": ("coordinator_authorization(",),
        ".github/scripts/monthly_release_pr.py": ("def advance(", "def validate("),
        ".github/scripts/monthly_release_vote.py": ("def validate_vote(",),
        ".github/scripts/qualify_monthly_candidate.py": ("def qualify(",),
        ".github/scripts/build_monthly_assets.py": ("def ",),
        ".github/scripts/test_monthly_public_release.py": ("def ",),
    }
    for path, required in files.items():
        content = repository_content(api, HOST, path, binding["worker_sha"])
        if content is None:
            return False
        if any(token not in content for token in required):
            return False
    return True


def validate_worker_binding(snapshot):
    if (snapshot.get("worker_repository") != HOST or snapshot.get("worker_path") != SHARED_PATH
            or not release.SHA.fullmatch(snapshot.get("worker_sha", ""))
            or not DIGEST.fullmatch(snapshot.get("worker_digest", ""))):
        raise ValueError("Missing or invalid central shared worker identity")


def trusted_shared_revision(api, binding):
    main = api.request(f"repos/{HOST}/branches/main")["commit"]["sha"]
    comparison = api.request(f"repos/{HOST}/compare/{binding['worker_sha']}...{main}")
    if comparison.get("status") not in {"identical", "ahead"} or comparison.get("behind_by") != 0:
        raise ValueError("Shared worker revision is not reachable from reviewed central main")
    if binding["worker_digest"] != automation_digest(api, HOST, main):
        raise ValueError("Central worker code changed; explicit reassessment is required")


def observed_worker(api, repository, run_id, sha):
    run = api.request(f"repos/{repository}/actions/runs/{positive_id(run_id)}")
    references = [item for item in run.get("referenced_workflows", [])
                  if item.get("path", "").split("@", 1)[0] == HOST + "/" + SHARED_PATH]
    if (len(references) != 1 or references[0].get("sha") != sha
            or references[0].get("path") != HOST + "/" + SHARED_PATH + "@" + sha):
        raise ValueError("Observed run does not invoke the authorized shared worker revision")
    return run


def automation_digest(api, repository, commit):
    tree = api.request(f"repos/{repository}/git/trees/{commit}?recursive=1")
    if tree.get("truncated") or not isinstance(tree.get("tree"), list):
        raise ValueError("Incomplete automation tree")
    entries = {
        item["path"]: {"sha": item["sha"], "mode": item["mode"], "type": item["type"]}
        for item in tree["tree"]
        if item["path"].startswith((".github/scripts/", ".github/workflows/", ".github/aw/"))
        or item["path"] in {POLICY_PATH, ".github/monthly-goreleaser.yml", ".github/licenserc.yml", ".gitattributes",
                           "MAINTAINERS", "RELEASE_CHECKLIST.md"}
    }
    if WORKER_PATH not in entries or any(item["type"] != "blob" or item["mode"] not in {"100644", "100755"} for item in entries.values()):
        raise ValueError("Missing or unsafe automation files")
    return digest(entries)


def source_checkpoint(api, repository, snapshot, month):
    current = api.request(f"repos/{repository}/branches/{snapshot['main']}")["commit"]["sha"]
    if current == snapshot["main_head"]:
        return current
    candidate = release.Cycle(api, repository, month, MODE).state
    if candidate is None:
        raise ValueError("Source changed without a trusted worker cycle")
    authorized = {item["commit"] for item in candidate.get("merged", [])}
    ahead = compare_commits(api, repository, snapshot["main_head"], current)
    if not ahead or not ahead <= authorized:
        raise ValueError("Source advanced outside recorded worker dependency merges")
    authorization = {"snapshot": snapshot, "producers": candidate.get("producers", {})}
    for item in candidate.get("merged", []):
        if item["commit"] not in ahead:
            continue
        pull = api.request(f"repos/{repository}/pulls/{item['number']}")
        if not release.dependabot(pull, snapshot["main"]) or not scoped_pull(api, repository, pull, authorization, candidate["merged"]):
            raise ValueError("Source checkpoint contains an untrusted dependency merge")
        release.check_files(api, repository, item["number"])
    return current


def assessed_automation(api, repository, snapshot, commit):
    if automation_digest(api, repository, commit) != snapshot.get("automation_digest"):
        raise ValueError("Release worker code changed after assessment; explicit reassessment is required")
    return commit


def assessed_worker(api, repository, snapshot, caller_sha):
    binding = worker_binding(api, repository, caller_sha)
    if binding is None or any(binding[key] != snapshot.get(key) for key in binding):
        # A CLI source dependency merge may advance its commit without changing automation.
        if not (repository == HOST and binding is not None
                and binding["worker_digest"] == snapshot.get("worker_digest")
                and binding["worker_repository"] == snapshot.get("worker_repository")
                and binding["worker_path"] == snapshot.get("worker_path")):
            raise ValueError("Shared worker pin changed; explicit reassessment is required")
    return binding


def load_policy(api, commit):
    item = api.request(f"repos/{HOST}/contents/{POLICY_PATH}?ref={commit}")
    if item.get("encoding") != "base64":
        raise ValueError("Missing maintainer-owned release policy")
    policy = load_json(base64.b64decode(item["content"]).decode())
    expected = {
        "schema": 1, "repositories": list(REPOSITORIES),
        "dependency_versions": "same-major-minor-nondecreasing",
        "producer_versions": "exact-planned-stable-patch",
        "go_requirement": "unchanged-for-routine-updates", "reachable_vulnerabilities": "block",
        "optional_backports": "excluded", "workflow_updates": "explicit-approval",
        "maximum_assessment_age_hours": 24,
        "plan_approval": "active-cli-maintainer-comment",
        "publication": "maintainer-publishes-qualified-draft",
        "release_votes": {"notation-core-go": "branch-rules", "notation-go": "repo-maintainer-majority",
                          "notation": "repo-maintainer-majority"},
    }
    if policy != expected or type(policy.get("schema")) is not int or type(policy.get("maximum_assessment_age_hours")) is not int:
        raise ValueError("Unsupported monthly-release policy; reviewed implementation is required")
    return policy


def routine_pull(api, repository, pull):
    if not release.dependabot(pull, "main") or (pull["head"].get("repo") or {}).get("full_name") != repository:
        raise ValueError("Automatic authority is limited to same-repository Dependabot PRs")
    release.check_files(api, repository, pull["number"])
    paths = {f"{directory}/go.{extension}".removeprefix("./")
             for directory in release.modules(repository) for extension in ("mod", "sum")}
    if any(item["filename"] not in paths for item in api.pages(f"repos/{repository}/pulls/{pull['number']}/files")):
        raise ValueError("Workflow updates require explicit actor approval and reassessment")
    for directory in release.modules(repository):
        before = api.manifest(repository, pull["base"]["sha"], directory)
        after = api.manifest(repository, pull["head"]["sha"], directory)
        if before.get("Replace") or after.get("Replace"):
            raise ValueError("Routine updates cannot contain replacements")
        for field in set(before) | set(after):
            if field != "Require" and before.get(field) != after.get(field):
                raise ValueError("Routine updates cannot change module or Go requirements")
        original = {item["Path"]: item for item in before.get("Require") or []}
        updated = {item["Path"]: item for item in after.get("Require") or []}
        if original.keys() != updated.keys():
            raise ValueError("Added or removed dependencies require explicit actor approval")
        for path, previous in original.items():
            following = updated[path]
            old, new = release.semver(previous["Version"]), release.semver(following["Version"])
            if (old is None or new is None or old[:2] != new[:2] or new < old
                    or previous.get("Indirect") != following.get("Indirect")):
                raise ValueError("Routine dependency versions must remain on the same stable major/minor")


def review_scope(api, plan):
    commit = api.request(f"repos/{HOST}/branches/main")["commit"]["sha"]
    policy = load_policy(api, commit)
    age = datetime.datetime.now(datetime.timezone.utc) - datetime.datetime.fromisoformat(plan["collected_at"])
    if age < datetime.timedelta(0) or age > datetime.timedelta(hours=policy["maximum_assessment_age_hours"]):
        raise ValueError("Automatic assessment is stale")
    for repository, decision in plan["decisions"].items():
        snapshot = plan["snapshots"][repository]
        if decision["decision"] == "defer":
            raise ValueError("Deferrals require maintainer resolution")
        if any(result["reachable"] for result in snapshot["scans"].values()):
            raise ValueError("Reachable findings require resolution and a new assessment")
        for item in snapshot["pulls"] + snapshot["merged"]:
            pull = api.request(f"repos/{repository}/pulls/{item['number']}")
            if "head" in item and pull["head"]["sha"] != item["head"]:
                raise ValueError("Assessed PR head changed")
            if "commit" in item and pull.get("merge_commit_sha") != item["commit"]:
                raise ValueError("Assessed merge changed")
            routine_pull(api, repository, pull)
    return {"kind": "pending", "policy_sha256": digest(policy), "policy_commit": commit}


def authorize_routine(api, plan):
    raise ValueError("Automatic cycle authorization is disabled; a maintainer must approve the plan issue")


def plan_comment(plan):
    return f"/approve-plan {plan['plan_id']}"


def safe_title(value):
    text = html.escape(" ".join(str(value).split())[:400], quote=True)
    return re.sub(r"[\\`*_{}\[\]()#+.!|@]", lambda match: f"&#{ord(match[0])};", text)


def release_scope(plan):
    lines = ["## Release scope", "",
             "| Repository | Baseline to proposed patch | Decision | Planned producer requirement |",
             "|---|---|---|---|"]
    for repository in REPOSITORIES:
        snapshot = plan["snapshots"][repository]
        producers = []
        for name in release.PROJECTS[repository.split("/")[1]]:
            source = f"{OWNER}/{name}"
            producer = plan["snapshots"][source]
            decision = plan["decisions"][source]["decision"]
            producers.append(f"`{name}` exactly `{producer['tag']}` after maintainer publication and public verification"
                             if decision == "release" else f"`{name}` has no new authorized version ({decision})")
        decision = "Proposed, not qualified" if plan.get("preview") else plan["decisions"][repository]["decision"]
        lines.append(f"| `{repository}` | `{snapshot['baseline']}` to `{snapshot['tag']}` | "
                     f"{decision} | {'; '.join(producers) or 'None'} |")
    lines += ["", "### Dependabot source commits", "",
              "| Repository | Dependabot PR/change | Exact source revision(s) | CI/review status |",
              "|---|---|---|---|"]
    count = 0
    for repository in REPOSITORIES:
        snapshot = plan["snapshots"][repository]
        for field, revision, label in (("pulls", "head", "Open PR proposed HEAD (not yet merged)"),
                                       ("merged", "commit", "Already merged source commit to backport")):
            for item in snapshot[field]:
                number = positive_id(item["number"])
                sha = item[revision]
                if not release.SHA.fullmatch(sha):
                    raise ValueError("Invalid scope commit identity")
                title = safe_title(item.get("title", "Dependabot dependency update"))
                revisions = [f"{label}: [`{sha}`](https://github.com/{repository}/commit/{sha})"]
                for commit in item.get("commits", []):
                    if not release.SHA.fullmatch(commit["sha"]):
                        raise ValueError("Invalid proposed PR commit")
                    if commit["sha"] != sha:
                        revisions.append(f"Additional proposed commit: [`{commit['sha']}`]"
                                         f"(https://github.com/{repository}/commit/{commit['sha']}) {safe_title(commit.get('subject', ''))}")
                lines.append(f"| `{repository}` | [#{number}](https://github.com/{repository}/pull/{number}) {title} | "
                             f"{'<br>'.join(revisions)} | {safe_title(item.get('review_status', 'Required checks and reviews must pass'))} |")
                count += 1
    if not count:
        lines.append("| All repositories | No existing Dependabot source PRs in this assessment | No source commit invented | Not qualified |")
    lines += ["", "Future producer-adoption Dependabot PRs, merge commits, and stable-branch backport commits "
              "do not exist yet and are not listed as known commits. Workers wait for exact planned producer versions "
              "and qualifying PRs. Signed Go/CLI version-metadata vote commits are separate release-governance "
              "changes, not Dependabot source commits. Optional CVE/bug-fix backports remain outside automatic scope.", ""]
    return "\n".join(lines)


def controller_body(state):
    instructions = (
        "# Maintainer plan review\n\n"
        "No worker execution is authorized until an active CLI repo-level maintainer posts "
        "the exact, unedited comment below in this issue. Closing the issue, editing a checklist, "
        "the AI assessment, and the legacy automatic-approval flag do not approve this plan.\n\n"
        f"```text\n{plan_comment(state['plan'])}\n```\n\n"
        f"Assessment: https://github.com/{HOST}/actions/runs/{state['source_run']}\n\n"
        "Review the complete decisions, source/dependency heads, worker pins, policy, and "
        "reserved versions below. Reassessment requires a new approval. Editing or deleting "
        "the approving comment revokes remaining execution. Publication is manual after "
        "each repository's qualified draft review; this approval never publishes a release.\n\n"
    )
    blocker = f"Execution blocker: {safe_title(state['review_blocker'])}\n\n" if state.get("review_blocker") else ""
    return (f"{marker(state['month'])}\n{instructions}{release_scope(state['plan'])}\n{blocker}Outcome: **{state['status']}**.\n\n"
            f"```json\n{json.dumps(state, indent=2)}\n```\n")


def pull_scope(api, repository, number):
    commits = [{"sha": item["sha"], "subject": item["commit"]["message"].split("\n")[0][:400]}
               for item in api.pages(f"repos/{repository}/pulls/{number}/commits")]
    if not commits or any(not release.SHA.fullmatch(item["sha"]) for item in commits):
        raise ValueError("Incomplete or invalid proposed PR commit list")
    files = [item["filename"] for item in api.pages(f"repos/{repository}/pulls/{number}/files")]
    return {"commits": commits, "review_status": "Workflow changes need explicit actor review; CI/reviews must pass"
            if any(path.startswith(".github/workflows/") for path in files) else "CI and required reviews must pass"}


def preview_scope(api, output, inventory_path=None):
    snapshots, decisions = {}, {}
    captured = load_json(inventory_path.read_text())["repositories"] if inventory_path else None
    for repository in REPOSITORIES:
        if captured is not None:
            node = captured[repository]
            baseline = node["latest_stable"]
            if isinstance(baseline, dict):
                baseline = baseline["tag_name"]
            version = release.semver(baseline)
            if version is None:
                raise ValueError("Preview requires a real stable baseline")
            snapshots[repository] = {
                "baseline": baseline, "tag": f"v{version[0]}.{version[1]}.{version[2] + 1}",
                "main_head": node["main_head"], "pulls": [], "merged": [],
            }
            for p in node["pulls"]:
                if p["author"] != "dependabot[bot]" or p["head_repository"] != repository or p["base"] != "main":
                    raise ValueError("Preview source is not a canonical Dependabot PR")
                snapshots[repository]["pulls"].append({
                    **{key: p[key] for key in ("number", "head", "title", "commits")},
                    "review_status": "Workflow changes need explicit actor review; CI/reviews are not qualified"
                    if any(path.startswith(".github/workflows/") for path in p["files"])
                    else "CI/reviews are not qualified in this format preview",
                })
            decisions[repository] = {"decision": "release" if snapshots[repository]["pulls"] else "skip"}
            continue
        repository_metadata(api, repository)
        baseline = release.baseline_plan(api, repository, datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m"), MODE)
        main = api.request(f"repos/{repository}/branches/main")["commit"]["sha"]
        stable = api.request(f"repos/{repository}/branches/{baseline['branch']}")["commit"]["sha"]
        ahead = compare_commits(api, repository, stable, main)
        pulls = [{"number": p["number"], "head": p["head"]["sha"], "title": p.get("title", ""),
                  **pull_scope(api, repository, p["number"])}
                 for p in api.pulls(repository, "main", "open") if release.dependabot(p, "main")]
        merged = [p for p in release.inventory(api, repository, "main") if p["commit"] in ahead]
        for item in merged:
            item["title"] = api.request(f"repos/{repository}/pulls/{item['number']}").get("title", "")
        snapshots[repository] = {**baseline, "pulls": pulls, "merged": merged, "main_head": main, "release_head": stable}
        decisions[repository] = {"decision": "release" if pulls or merged else "skip"}
    for repository in REPOSITORIES:
        if any(decisions[f"{OWNER}/{name}"]["decision"] == "release" for name in release.PROJECTS[repository.split("/")[1]]):
            decisions[repository]["decision"] = "release"
    data = {"snapshots": snapshots, "decisions": decisions, "preview": True}
    (output / "preview-inventory.json").write_text(json.dumps(data, indent=2) + "\n")
    revision = release.git(["rev-parse", "HEAD"], pathlib.Path(__file__).resolve().parents[2])
    (output / "plan-preview.md").write_text(
        "# FORMAT PREVIEW ONLY: no release authorization\n\n"
        f"Implementation revision: `{revision}`. Read-only live canonical Dependabot inventory; "
        "no source qualification, security verdict, or execution approval is represented. "
        "A copy of this preview in a fork cannot authorize canonical execution.\n\n" + release_scope(data))
    return {"writes": False, "preview": str(output / "plan-preview.md")}


def plan_roster(api, state):
    current = api.request(f"repos/{HOST}/branches/main")["commit"]["sha"]
    assessed_automation(api, HOST, state["plan"]["snapshots"][HOST], current)
    text = repository_content(api, HOST, "MAINTAINERS", current)
    if text is None:
        raise ValueError("Missing active CLI maintainer roster")
    from monthly_release_vote import parse_repo_maintainers
    return parse_repo_maintainers(text)


def approval_comment(api, issue, state, receipt=None):
    roster = plan_roster(api, state)
    if (issue.get("state") != "open" or issue.get("body") != controller_body(state)
            or issue.get("user", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR")):
        raise ValueError("Plan review issue body, author or open state changed")
    release.trusted_editor(api, HOST, issue)
    comments = ([api.request(f"repos/{HOST}/issues/comments/{positive_id(receipt['comment_id'])}")]
                if receipt else api.pages(f"repos/{HOST}/issues/{issue['number']}/comments"))
    for comment in comments:
        if (comment.get("body") != plan_comment(state["plan"])
                or comment.get("user", {}).get("type") != "User"
                or comment["user"].get("login", "").lower() not in roster
                or not comment.get("created_at") or comment.get("updated_at") != comment["created_at"]
                or comment.get("issue_url") != f"https://api.github.com/repos/{HOST}/issues/{issue['number']}"):
            if receipt:
                raise ValueError("Maintainer plan approval was edited, revoked or is no longer eligible")
            continue
        if comment["created_at"] < issue["created_at"]:
            raise ValueError("Plan approval predates the review issue")
        expected = {"kind": "maintainer", "comment_id": positive_id(comment["id"]),
                    "login": comment["user"]["login"], "created_at": comment["created_at"],
                    "policy_sha256": state["authorization"]["policy_sha256"],
                    "policy_commit": state["authorization"]["policy_commit"]}
        if receipt is not None and receipt != expected:
            raise ValueError("Maintainer plan approval receipt changed")
        return expected
    if receipt:
        raise ValueError("Maintainer plan approval is missing")
    return None


def require_plan_approval(api, issue, state):
    if state["authorization"].get("kind") != "maintainer":
        raise ValueError("Execution requires maintainer approval of the exact plan issue")
    approval_comment(api, issue, state, state["authorization"])
    current = api.request(f"repos/{HOST}/branches/main")["commit"]["sha"]
    if digest(load_policy(api, current)) != state["authorization"]["policy_sha256"]:
        raise ValueError("Maintainer policy changed; explicit reassessment is required")


def explicit_actor():
    if os.environ.get("GITHUB_EVENT_NAME") != "workflow_dispatch" or (
        not os.environ.get("MONTHLY_PATCH_ACTOR")
        or os.environ.get("GITHUB_ACTOR") != os.environ["MONTHLY_PATCH_ACTOR"]
    ):
        raise ValueError("Exceptional approval, retry and replan require the configured actor's manual dispatch")


def scan_source(repository, commit, directories, output):
    """Scan an immutable source tree without executing repository tests or hooks."""
    output.mkdir(parents=True, exist_ok=True)
    results = {}
    with tempfile.TemporaryDirectory(prefix="notation-release-assessment-") as temporary:
        tree = pathlib.Path(temporary)
        release.git(["init", "--quiet"], tree)
        release.git(["remote", "add", "origin", f"https://github.com/{repository}.git"], tree)
        release.git(["fetch", "--quiet", "--depth=1", "origin", commit], tree)
        release.git(["checkout", "--quiet", "--detach", "FETCH_HEAD"], tree)
        if release.git(["rev-parse", "HEAD"], tree) != commit:
            raise ValueError("Assessment source does not match its immutable commit")
        for index, directory in enumerate(directories):
            scanned = subprocess.run(
                ["govulncheck", "-format=json", "./..."], cwd=tree / directory,
                capture_output=True, text=True, encoding="utf-8",
            )
            (output / f"{index}.jsonstream").write_text(scanned.stdout, encoding="utf-8")
            (output / f"{index}.stderr.txt").write_text(scanned.stderr, encoding="utf-8")
            if scanned.returncode != 0:
                raise ValueError(f"Assessment scanner failed for {repository}:{directory}: {scanned.stderr}")
            reachable, informational, advisories = scan_messages(scanned.stdout, "source")
            results[directory] = {
                "reachable": sorted(reachable), "informational": sorted(informational),
                "advisories": {
                    key: {"summary": item.get("summary", ""), "modified": item.get("modified"),
                          "aliases": item.get("aliases", []),
                          "fixed": sorted({event["fixed"] for affected in item.get("affected", [])
                                           for ranges in affected.get("ranges", [])
                                           for event in ranges.get("events", []) if "fixed" in event})}
                    for key, item in advisories.items()
                },
            }
    return results


def compare_commits(api, repository, base, head):
    commits = []
    for page in range(1, 11):
        result = api.request(f"repos/{repository}/compare/{base}...{head}?per_page=100&page={page}")
        if result.get("status") not in {"ahead", "identical", "behind", "diverged"}:
            raise ValueError("Cannot establish main/release divergence")
        commits.extend(result["commits"])
        if len(commits) == result["ahead_by"]:
            if any(not release.SHA.fullmatch(item.get("sha", "")) for item in commits):
                raise ValueError("Invalid comparison commit")
            return {item["sha"] for item in commits}
        if len(result["commits"]) < 100:
            raise ValueError("Incomplete main/release comparison")
    raise ValueError("Assessment comparison exceeds 1,000 commits; manual reconciliation is required")


def collect(api, month, output):
    if not release.MONTH.fullmatch(month):
        raise ValueError("Invalid assessment month")
    inventory = {"schema": 1, "host": HOST, "month": month, "collected_at": utc_now(), "repositories": {}}
    for repository in REPOSITORIES:
        metadata = repository_metadata(api, repository)
        baseline = release.baseline_plan(api, repository, month, MODE)
        automation = api.request(f"repos/{repository}/branches/main")["commit"]["sha"]
        main = api.optional(f"repos/{repository}/branches/{baseline['main']}")
        branch = api.optional(f"repos/{repository}/branches/{baseline['branch']}")
        blockers = []
        if not metadata.get("has_issues"):
            blockers.append("Enable repository issues before writing release state")
        if main is None or branch is None:
            blockers.append("Reviewed source/release branches are missing")
        workflow = api.optional(f"repos/{repository}/actions/workflows/monthly-patch-release.yml")
        binding = worker_binding(api, repository, automation)
        if workflow is None or workflow.get("state") != "active" or not installed_worker(api, repository, automation, binding):
            blockers.append("Install and enable the complete coordinator-aware release worker on main")
        if binding is not None:
            trusted_shared_revision(api, binding)
        source = repository if branch else f"notaryproject/{repository.split('/')[1]}"
        commit = branch["commit"]["sha"] if branch else api.request(
            f"repos/{source}/commits/{baseline['baseline']}"
        )["sha"]
        if not release.SHA.fullmatch(automation) or not release.SHA.fullmatch(commit):
            raise ValueError("Invalid source or automation commit")
        source_ref = main["commit"]["sha"] if main else automation
        provenance = {}
        baseline_sha = api.request(f"repos/{repository}/commits/{baseline['baseline']}")["sha"]
        provenance = {
            "baseline_sha": baseline_sha,
            "baseline_tag_object": release.verified_tag(api, repository, baseline["baseline"], baseline_sha),
        }
        pulls = []
        for pull in api.pulls(repository, baseline["main"] if main else "main", "open"):
            if not release.dependabot(pull, baseline["main"] if main else "main"):
                continue
            release.check_files(api, repository, pull["number"])
            ready, reason = release.checks_ready(pull, api.checks(repository, pull["number"]))
            pulls.append({"number": pull["number"], "head": pull["head"]["sha"],
                          "title": (pull.get("title") or "")[:400], "body": (pull.get("body") or "")[:1600],
                          **pull_scope(api, repository, pull["number"]),
                          "ready": ready, "reason": reason})
            if pulls[-1]["commits"][-1]["sha"] != pull["head"]["sha"]:
                raise ValueError("Dependabot PR changed during commit inventory collection")
        merged = []
        if main and branch:
            ahead = compare_commits(api, repository, commit, source_ref)
            records = list(api.pages(f"repos/{repository}/commits?sha={commit}"))
            history = "\n".join(item["commit"]["message"] for item in records)
            for item in release.inventory(api, repository, baseline["main"]):
                if item["commit"] not in ahead:
                    continue
                backports = [record["sha"] for record in records
                             if f"(cherry picked from commit {item['commit']})" in record["commit"]["message"]]
                if backports:
                    if any(f"This reverts commit {backport}." in history for backport in backports):
                        blockers.append(f"Dependency backport #{item['number']} was reverted; manual review is required")
                    continue
                release.check_files(api, repository, item["number"])
                item["title"] = api.request(f"repos/{repository}/pulls/{item['number']}").get("title", "")[:400]
                merged.append(item)
        scans = scan_source(source, commit, release.modules(repository), output / repository.split("/")[1])
        manifests = {directory: api.manifest(source, commit, directory) for directory in release.modules(repository)}
        inventory["repositories"][repository] = {
            "baseline": baseline["baseline"], "main": baseline["main"], "branch": baseline["branch"],
            "tag": baseline["tag"], "main_head": main["commit"]["sha"] if main else None,
            "release_head": branch["commit"]["sha"] if branch else None,
            "automation_sha": automation, "workflow_id": workflow["id"] if workflow else None,
            **(binding or {key: "" for key in ("worker_repository", "worker_path", "worker_sha", "worker_digest")}),
            **({"automation_digest": automation_digest(api, repository, automation)}),
            **provenance,
            "source_repository": source, "source_commit": commit, "blockers": blockers,
            "pulls": pulls, "merged": merged, "manifests": manifests, "scans": scans,
        }
    return inventory


def evidence_ids(repository, node):
    return {f"{repository}:baseline", f"{repository}:setup", f"{repository}:dependencies",
            *(f"{repository}:pr:{item['number']}" for item in node["pulls"] + node["merged"]),
            *(f"{repository}:scan:{directory}" for directory in node["scans"])}


def validate_inventory(inventory):
    if type(inventory.get("schema")) is not int or inventory["schema"] != 1 or inventory.get("host") != HOST:
        raise ValueError("Unexpected assessment schema or host")
    if not release.MONTH.fullmatch(inventory.get("month", "")) or set(inventory.get("repositories", {})) != set(REPOSITORIES):
        raise ValueError("Incomplete assessment inventory")
    when = datetime.datetime.fromisoformat(inventory["collected_at"])
    if when.tzinfo is None:
        raise ValueError("Assessment timestamp must be timezone-aware")
    for repository, node in inventory["repositories"].items():
        release.validate_plan({
            "schema": 1, "repository": repository, "month": inventory["month"], "mode": MODE,
            **{key: node[key] for key in ("baseline", "main", "branch", "tag")}, "status": "merging",
        })
        for field in ("automation_sha", "source_commit"):
            if not release.SHA.fullmatch(node.get(field, "")):
                raise ValueError("Invalid inventory commit")
        for field in ("main_head", "release_head"):
            if node.get(field) is not None and not release.SHA.fullmatch(node[field]):
                raise ValueError("Invalid branch snapshot")
        if node["source_repository"] not in {repository, f"notaryproject/{repository.split('/')[1]}"}:
            raise ValueError("Assessment source escapes its project")
        if not isinstance(node["blockers"], list) or not all(isinstance(item, str) for item in node["blockers"]):
            raise ValueError("Invalid setup evidence")
        if set(node["manifests"]) != set(release.modules(repository)) or set(node["scans"]) != set(release.modules(repository)):
            raise ValueError("Incomplete module or vulnerability evidence")
        for result in node["scans"].values():
            for field in ("reachable", "informational"):
                if not isinstance(result[field], list) or any(not re.fullmatch(r"GO-\d{4}-\d+", item) for item in result[field]):
                    raise ValueError("Invalid vulnerability evidence")
        numbers = set()
        for kind, field in (("pulls", "head"), ("merged", "commit")):
            if not isinstance(node[kind], list):
                raise ValueError("Invalid dependency inventory")
            for item in node[kind]:
                number = positive_id(item["number"])
                if number in numbers or not release.SHA.fullmatch(item.get(field, "")):
                    raise ValueError("Duplicate or invalid dependency provenance")
                numbers.add(number)
                if kind == "pulls":
                    commits = item.get("commits")
                    if (not isinstance(commits, list) or not commits or commits[-1].get("sha") != item["head"]
                            or any(not release.SHA.fullmatch(commit.get("sha", "")) for commit in commits)
                            or len({commit["sha"] for commit in commits}) != len(commits)):
                        raise ValueError("Missing, duplicate or mismatched proposed PR commit list")
        if node["workflow_id"] is not None:
            positive_id(node["workflow_id"])
        if not DIGEST.fullmatch(node.get('automation_digest', '')):
            raise ValueError("Missing automation tree identity")
        if not node["blockers"]:
            validate_worker_binding(node)
        if any((not release.SHA.fullmatch(node.get(key, '')) for key in ('baseline_sha', 'baseline_tag_object'))):
            raise ValueError("Missing signed stable-baseline provenance")
    return inventory


def validate_assessment(inventory, assessment):
    validate_inventory(inventory)
    if set(assessment) != {"schema", "month", "inventory_sha256", "decisions"}:
        raise ValueError("Unexpected assessment fields")
    if type(assessment["schema"]) is not int or assessment["schema"] != 1 or assessment["month"] != inventory["month"] or assessment["inventory_sha256"] != digest(inventory):
        raise ValueError("Assessment is not bound to the collected inventory")
    if not isinstance(assessment["decisions"], dict) or set(assessment["decisions"]) != set(REPOSITORIES):
        raise ValueError("Agent must assess all three repositories exactly once")
    all_evidence = set().union(*(evidence_ids(repo, node) for repo, node in inventory["repositories"].items()))
    for repository, item in assessment["decisions"].items():
        if set(item) != {"decision", "reason", "evidence"} or item["decision"] not in DECISIONS:
            raise ValueError("Invalid release decision")
        if not isinstance(item["reason"], str) or not 1 <= len(item["reason"]) <= 2000:
            raise ValueError("Every decision needs a bounded rationale")
        if not isinstance(item["evidence"], list) or not 1 <= len(item["evidence"]) <= 30 or any(
            not isinstance(reference, str) or reference not in all_evidence for reference in item["evidence"]
        ):
            raise ValueError("Decision cites missing or fabricated evidence")
        if not set(item["evidence"]) & evidence_ids(repository, inventory["repositories"][repository]):
            raise ValueError("Decision needs evidence from its own repository")
        node = inventory["repositories"][repository]
        signals = bool(node["pulls"] or node["merged"] or any(result["reachable"] for result in node["scans"].values()))
        if item["decision"] == "skip" and (signals or node["blockers"]):
            raise ValueError("Known updates, CVEs or missing evidence cannot be skipped")
        if item["decision"] == "release" and node["blockers"]:
            raise ValueError("Repository setup is incomplete; defer rather than publish")
    decisions = copy.deepcopy(assessment["decisions"])
    affected = set()
    for repository in REPOSITORIES:
        node = inventory["repositories"][repository]
        upstream = [f"{OWNER}/{name}" for name in release.PROJECTS[repository.split("/")[1]]]
        required = any(producer in affected for producer in upstream)
        if required and decisions[repository]["decision"] == "skip":
            raise ValueError("A planned library update requires downstream inclusion")
        if decisions[repository]["decision"] == "release":
            if not (node["pulls"] or node["merged"] or required):
                raise ValueError("A CVE without a dependency fix requires deferral, not an empty patch")
            affected.add(repository)
        elif decisions[repository]["decision"] == "defer" and (
            node["pulls"] or node["merged"] or required or any(result["reachable"] for result in node["scans"].values())
        ):
            affected.add(repository)
    plan = {
        "schema": 1, "host": HOST, "month": inventory["month"], "inventory_sha256": digest(inventory),
        "collected_at": inventory["collected_at"], "decisions": decisions,
        "snapshots": {
            repo: {key: node[key] for key in ("baseline", "main", "branch", "tag", "main_head",
                                            "release_head", "automation_sha", "workflow_id", "pulls", "merged",
                                            "automation_digest", "scans", "baseline_sha", "baseline_tag_object",
                                            "worker_repository", "worker_path", "worker_sha", "worker_digest")}
            for repo, node in inventory["repositories"].items()
        },
    }
    # PR text helps the agent assess compatibility but is not executable authorization.
    for node in plan["snapshots"].values():
        node["pulls"] = [{"number": item["number"], "head": item["head"],
                          "title": item.get("title", "Dependabot dependency update"),
                          "commits": item.get("commits", []), "review_status": item.get("review_status", "CI/reviews required")}
                         for item in node["pulls"]]
    plan["plan_id"] = digest(plan)
    return plan


def marker(month):
    if not release.MONTH.fullmatch(month):
        raise ValueError("Invalid controller month")
    return f"<!-- notation-release-controller:{month}:{MODE} -->"


def validate_state(state):
    if type(state.get("schema")) is not int or state["schema"] != 1 or state.get("host") != HOST or state.get("mode") != MODE:
        raise ValueError("Invalid controller state or execution namespace")
    marker(state["month"])
    plan = state["plan"]
    if not DIGEST.fullmatch(plan.get("plan_id", "")) or digest({key: value for key, value in plan.items() if key != "plan_id"}) != plan["plan_id"]:
        raise ValueError("Approved plan changed")
    if plan.get("host") != HOST or type(plan.get("schema")) is not int or plan["schema"] != 1 or plan.get("month") != state["month"]:
        raise ValueError("Approved plan identity changed")
    if set(plan["decisions"]) != set(REPOSITORIES) or set(plan["snapshots"]) != set(REPOSITORIES) or set(state["nodes"]) != set(REPOSITORIES):
        raise ValueError("Controller repository scope changed")
    if state.get("status") not in {"review", "active", "completed"}:
        raise ValueError("Invalid controller status")
    authorization = state.get("authorization", {})
    if authorization.get("kind") not in {"pending", "maintainer"}:
        raise ValueError("Missing authorization provenance")
    if (
        not DIGEST.fullmatch(authorization.get("policy_sha256", ""))
        or not release.SHA.fullmatch(authorization.get("policy_commit", ""))
    ):
        raise ValueError("Invalid policy authorization")
    if (state["status"] == "review") != (authorization["kind"] == "pending"):
        raise ValueError("Unreviewed plan cannot execute")
    if authorization["kind"] == "maintainer":
        positive_id(authorization.get("comment_id"))
    for repository, node in state["nodes"].items():
        decision = plan["decisions"][repository]["decision"]
        if decision not in DECISIONS or node["status"] not in NODE_STATUSES or type(node["attempt"]) is not int or node["attempt"] < 0:
            raise ValueError("Invalid controller node")
        if decision == "skip" and node["status"] != "skipped" or decision == "defer" and node["status"] != "deferred":
            raise ValueError("Agent decision changed during execution")
        snapshot = plan["snapshots"][repository]
        if decision == "release":
            validate_worker_binding(snapshot)
        recovery = snapshot.get("verification_recovery")
        if recovery is not None and (
            not isinstance(recovery, dict) or set(recovery) != {"from_plan_id", "candidate_sha256"}
            or not all(isinstance(value, str) and DIGEST.fullmatch(value) for value in recovery.values())
            or decision != "release"
        ):
            raise ValueError("Invalid verification-only recovery receipt")
        release.validate_plan({"schema": 1, "repository": repository, "mode": MODE, "month": state["month"],
                               **{key: snapshot[key] for key in ("baseline", "main", "branch", "tag")}, "status": "merging"})
        if node.get("run_id") is not None:
            positive_id(node["run_id"])
        if node["status"] == "published" and (
            node.get("version") != snapshot["tag"] or not release.SHA.fullmatch(node.get("commit", ""))
        ):
            raise ValueError("Published producer identity changed")
    if state["status"] == "completed" and any(node["status"] not in TERMINAL for node in state["nodes"].values()):
        raise ValueError("Controller completed before its repositories")
    return state


def decode_issue(issue, actor):
    if "pull_request" in issue or not actor or issue.get("user", {}).get("login") != actor:
        return None
    body = issue.get("body") or ""
    if "<!-- notation-release-controller:" not in body:
        return None
    blocks = re.findall(r"```json\n(.*?)\n```", body, re.DOTALL)
    if len(blocks) != 1:
        raise ValueError("Malformed trusted controller record")
    state = validate_state(load_json(blocks[0]))
    if body.count(marker(state["month"])) != 1:
        raise ValueError("Controller marker changed")
    return state


class Controller:
    def __init__(self, api, actor):
        if not actor:
            raise ValueError("A configured controller actor is required")
        self.api, self.actor = api, actor

    def records(self):
        records = []
        for issue in self.api.pages(f"repos/{HOST}/issues?state=all"):
            state = decode_issue(issue, self.actor)
            if state is not None:
                release.trusted_editor(self.api, HOST, issue)
                records.append((issue, state))
        months = [state["month"] for _, state in records]
        if len(set(months)) != len(months) or sum(state["status"] in {"active", "review"} for _, state in records) > 1:
            raise ValueError("Duplicate or overlapping controller cycles require reconciliation")
        return records

    def report_failure(self, issue, state, repository, run, reason, stages=()):
        validate_state(state)
        current = self.api.request(f"repos/{HOST}/issues/{positive_id(issue['number'])}")
        if decode_issue(current, self.actor) != state or current.get("body") != controller_body(state):
            raise ValueError("Release issue changed before failure notification")
        release.trusted_editor(self.api, HOST, current)
        if repository not in REPOSITORIES:
            raise ValueError("Failure notification escapes the release repositories")
        run_id = positive_id(run["id"])
        attempt = positive_id(run.get("run_attempt", 1))
        allowed = {"prepare", "qualify", "package", "publish", "verify", "complete", "controller"}
        if (not isinstance(stages, (list, tuple)) or len(stages) > len(allowed)
                or any(not isinstance(stage, str) or stage not in allowed for stage in stages)):
            raise ValueError("Unexpected failure notification stage")
        if not isinstance(reason, str) or not 1 <= len(reason) <= 500:
            raise ValueError("Failure notifications need a bounded reason")
        stages = list(dict.fromkeys(stages))
        affected = set()
        for target in REPOSITORIES:
            if repository in [f"{OWNER}/{name}" for name in release.PROJECTS[target.split("/")[1]]] or any(
                f"{OWNER}/{name}" in affected for name in release.PROJECTS[target.split("/")[1]]
            ):
                affected.add(target)
        blocked = [
            target for target in REPOSITORIES
            if (target in affected or "controller" in stages)
            and state["nodes"][target]["status"] not in TERMINAL
        ]
        main = self.api.request(f"repos/{HOST}/branches/main")["commit"]["sha"]
        if not release.SHA.fullmatch(main):
            raise ValueError("Invalid maintainer notification roster revision")
        from monthly_release_vote import parse_repo_maintainers
        text = repository_content(self.api, HOST, "MAINTAINERS", main)
        if text is None:
            raise ValueError("Missing maintainer notification roster")
        roster = parse_repo_maintainers(text)
        marker = "<!-- notation-release-failure:" + digest({
            "plan": state["plan"]["plan_id"], "repository": repository,
            "run": run_id, "attempt": attempt,
        }) + " -->"
        body = (
            f"{marker}\n"
            + " ".join(f"@{login}" for login in sorted(roster))
            + f"\n\nRelease cycle **{state['month']}** is blocked.\n\n"
            + f"- Repository: `{repository}`\n"
            + f"- Failed stage: {', '.join(stages) if stages else 'release worker'}\n"
            + f"- Reason: {safe_title(reason)}\n"
            + f"- Logs and retained evidence: https://github.com/{repository}/actions/runs/{run_id}/attempts/{attempt}\n"
            + "- Blocked downstream work: " + (", ".join(f"`{target}`" for target in blocked) or "none") + "\n\n"
            + "Review the linked logs and resolve the failed gate before requesting a retry. "
            + "Do not approve past the failure or publish an unqualified draft. "
            + "The approved scope has not changed."
        )
        matches = [
            comment for comment in self.api.pages(f"repos/{HOST}/issues/{issue['number']}/comments")
            if comment.get("user", {}).get("login") == self.actor and marker in (comment.get("body") or "")
        ]
        if len(matches) > 1:
            raise ValueError("Duplicate release failure notifications require reconciliation")
        if matches:
            return matches[0]
        comment = self.api.request(f"repos/{HOST}/issues/{issue['number']}/comments", "POST", {"body": body})
        if (comment.get("user", {}).get("login") != self.actor
                or comment.get("body") != body or type(comment.get("id")) is not int):
            raise ValueError("Release failure notification was not confirmed")
        positive_id(comment["id"])
        return comment

    def save(self, issue, state):
        validate_state(state)
        body = controller_body(state)
        if len(body.encode()) > 60000:
            raise ValueError("Controller record is too large")
        payload = {"body": body, "state": "closed" if state["status"] == "completed" else "open"}
        if issue:
            result = self.api.request(f"repos/{HOST}/issues/{issue['number']}", "PATCH", payload)
        else:
            payload["title"] = f"Notation dependency release cycle: {state['month']}"
            result = self.api.request(f"repos/{HOST}/issues", "POST", payload)
        if decode_issue(result, self.actor) != state:
            raise ValueError("Controller state write was not confirmed by the trusted actor")
        return result

    def approve(self, plan, source_run, replan=False, recover_public_verification=False, authorization=None):
        if recover_public_verification and not replan:
            raise ValueError("Public verification recovery requires an explicit replan")
        records = self.records()
        for issue, state in records:
            if state["month"] == plan["month"]:
                if state["plan"]["plan_id"] != plan["plan_id"]:
                    if not replan:
                        raise ValueError("This month already has an immutable approved plan")
                    return self.replan(issue, state, plan, source_run, recover_public_verification)
                if recover_public_verification:
                    raise ValueError("Verification recovery requires a fresh changed assessment")
                return issue, state
            if state["status"] in {"active", "review"}:
                raise ValueError("Finish the active cycle before approving another month")
        if recover_public_verification:
            raise ValueError("Verification recovery requires an existing approved cycle")
        when = datetime.datetime.fromisoformat(plan["collected_at"])
        age = datetime.datetime.now(datetime.timezone.utc) - when
        if age < datetime.timedelta(0) or age > datetime.timedelta(days=7):
            raise ValueError("Assessment is stale; collect a fresh assessment before approval")
        for repository, item in plan["decisions"].items():
            if item["decision"] != "release":
                continue
            metadata = repository_metadata(self.api, repository)
            snapshot = plan["snapshots"][repository]
            if not metadata.get("has_issues"):
                raise ValueError("Worker cycle issues must be enabled before approval")
            for name, expected in ((snapshot["main"], snapshot["main_head"]),
                                   (snapshot["branch"], snapshot["release_head"]),
                                   ("main", snapshot["automation_sha"])):
                if self.api.request(f"repos/{repository}/branches/{name}")["commit"]["sha"] != expected:
                    raise ValueError("Assessed refs changed before approval")
        state = {
            "schema": 1, "host": HOST, "mode": MODE, "month": plan["month"],
            "plan": copy.deepcopy(plan), "source_run": positive_id(source_run), "status": "review",
            "authorization": {"kind": "pending", "policy_commit": plan["snapshots"][HOST]["automation_sha"],
                              "policy_sha256": digest(load_policy(self.api, plan["snapshots"][HOST]["automation_sha"]))},
            "nodes": {repo: {"status": {"release": "pending", "skip": "skipped", "defer": "deferred"}[item["decision"]],
                             "attempt": 0}
                      for repo, item in plan["decisions"].items()},
        }
        if authorization and authorization.get("blocker"):
            state["review_blocker"] = authorization["blocker"]
        return self.save(None, state), state

    def replan(self, issue, state, plan, source_run, recover_public_verification=False):
        if state["status"] not in {"active", "review"} or any(node["status"] in {"dispatching", "running"} for node in state["nodes"].values()):
            raise ValueError("Cannot revise a completed cycle or an in-flight worker")
        when = datetime.datetime.fromisoformat(plan["collected_at"])
        age = datetime.datetime.now(datetime.timezone.utc) - when
        if age < datetime.timedelta(0) or age > datetime.timedelta(days=7):
            raise ValueError("Replacement assessment is stale")
        updated = copy.deepcopy(plan)
        recovered = False
        for repository, node in state["nodes"].items():
            previous = state["plan"]["snapshots"][repository]
            if node["status"] in TERMINAL:
                if updated["decisions"][repository]["decision"] != "skip":
                    raise ValueError("Reassessment cannot release an already completed repository again")
                updated["decisions"][repository] = state["plan"]["decisions"][repository]
                updated["snapshots"][repository] = previous
                continue
            if self.api.optional(f"repos/{repository}/git/ref/tags/{previous['tag']}") is not None:
                if not recover_public_verification:
                    raise ValueError("An immutable public candidate must be recovered, not replanned")
                cycle = release.Cycle(self.api, repository, state["month"], MODE)
                candidate = cycle.state
                if (candidate is None or candidate["status"] not in {"draft", "awaiting-publication", "verifying"}
                        or candidate.get("controller_issue") != issue["number"]
                        or candidate.get("controller_plan_id") != state["plan"]["plan_id"]
                        or any(candidate.get(key) != previous[key] for key in ("baseline", "main", "branch", "tag"))):
                    raise ValueError("Recovery requires this approved cycle's exact public verification candidate")
                snapshot = updated["snapshots"][repository]
                branch_tip = candidate["previous_head"] if candidate.get("release_backport") else candidate["commit"]
                if (any(snapshot[key] != previous[key] for key in ("baseline", "main", "branch", "workflow_id"))
                        or snapshot["release_head"] != branch_tip):
                    raise ValueError("Recovery cannot change the public candidate or its release branch")
                for name, expected in ((snapshot["main"], snapshot["main_head"]), (snapshot["branch"], branch_tip),
                                       ("main", snapshot["automation_sha"])):
                    if self.api.request(f"repos/{repository}/branches/{name}")["commit"]["sha"] != expected:
                        raise ValueError("Verification recovery assessment refs are stale")
                release.assert_tag(self.api, candidate)
                if candidate["status"] == "verifying":
                    release.check_public_assets(self.api, candidate)
                else:
                    release.check_release_assets(self.api, candidate, True)
                updated["decisions"][repository] = state["plan"]["decisions"][repository]
                updated["snapshots"][repository] = {
                    **previous, "automation_sha": snapshot["automation_sha"], "release_head": branch_tip,
                    "automation_digest": snapshot["automation_digest"], "main_head": snapshot["main_head"],
                    **{key: snapshot[key] for key in ("worker_repository", "worker_path", "worker_sha", "worker_digest")},
                    "verification_recovery": {
                        "from_plan_id": state["plan"]["plan_id"],
                        "candidate_sha256": digest({key: candidate.get(key) for key in ("tag", "commit", "assets", "release_id")}),
                    },
                }
                recovered = True
                continue
            if updated["decisions"][repository]["decision"] != state["plan"]["decisions"][repository]["decision"]:
                raise ValueError("Reassessment may refresh refs, not silently change the release DAG")
            snapshot = updated["snapshots"][repository]
            if any(snapshot[key] != previous[key] for key in ("baseline", "main", "branch", "tag", "release_head")):
                raise ValueError("Reassessment changed the reserved patch or release baseline")
            for name, expected in ((snapshot["main"], snapshot["main_head"]), (snapshot["branch"], snapshot["release_head"]),
                                   ("main", snapshot["automation_sha"])):
                if self.api.request(f"repos/{repository}/branches/{name}")["commit"]["sha"] != expected:
                    raise ValueError("Replacement assessment refs are stale")
        if recover_public_verification and not recovered:
            raise ValueError("Verification recovery requires an existing immutable public candidate")
        previous_id = state["plan"]["plan_id"]
        updated["supersedes"] = previous_id
        updated["plan_id"] = digest({key: value for key, value in updated.items() if key != "plan_id"})
        revisions = state.setdefault("revisions", [])
        if len(revisions) >= 20:
            raise ValueError("Cycle revision limit reached")
        revisions.append({"from": previous_id, "to": updated["plan_id"], "source_run": positive_id(source_run)})
        for node in state["nodes"].values():
            if node["status"] not in TERMINAL:
                node["status"] = "pending"
                node.pop("run_id", None)
        state.update(plan=updated, source_run=source_run)
        state.pop("review_blocker", None)
        state["status"] = "review"
        state["authorization"] = {"kind": "pending", "policy_commit": updated["snapshots"][HOST]["automation_sha"],
                                  "policy_sha256": digest(load_policy(self.api, updated["snapshots"][HOST]["automation_sha"]))}
        return self.save(issue, state), state

    def accept_review(self, issue, state):
        validate_state(state)
        if state["status"] == "completed":
            return issue, state
        current_issue = self.api.request(f"repos/{HOST}/issues/{issue['number']}")
        if decode_issue(current_issue, self.actor) != state:
            raise ValueError("Plan issue changed before execution")
        if state["status"] == "review":
            receipt = approval_comment(self.api, current_issue, state)
            if receipt is None:
                return current_issue, state
            if any(item["decision"] == "defer" for item in state["plan"]["decisions"].values()):
                raise ValueError("Resolve deferrals and reassess before approving a release cycle")
            if state.get("review_blocker"):
                raise ValueError(f"Plan execution is blocked: {state['review_blocker']}")
            current = self.api.request(f"repos/{HOST}/branches/main")["commit"]["sha"]
            policy = load_policy(self.api, current)
            if digest(policy) != state["authorization"]["policy_sha256"]:
                raise ValueError("Maintainer policy changed; reassess before approval")
            age = datetime.datetime.now(datetime.timezone.utc) - datetime.datetime.fromisoformat(state["plan"]["collected_at"])
            if age < datetime.timedelta(0) or age > datetime.timedelta(hours=policy["maximum_assessment_age_hours"]):
                raise ValueError("Assessment is stale; reassess and approve the new exact plan")
            for repository, decision in state["plan"]["decisions"].items():
                if decision["decision"] == "release":
                    snapshot = state["plan"]["snapshots"][repository]
                    caller = self.api.request(f"repos/{repository}/branches/main")["commit"]["sha"]
                    assessed_automation(self.api, repository, snapshot, caller)
                    trusted_shared_revision(self.api, assessed_worker(self.api, repository, snapshot, caller))
                    if self.api.request(f"repos/{repository}/branches/{snapshot['branch']}")["commit"]["sha"] != snapshot["release_head"]:
                        raise ValueError("Stable branch changed before maintainer approval")
                    source_checkpoint(self.api, repository, snapshot, state["month"])
                    for item in snapshot["pulls"]:
                        if self.api.request(f"repos/{repository}/pulls/{item['number']}")["head"]["sha"] != item["head"]:
                            raise ValueError("Assessed PR head changed before maintainer approval")
            state.update(status="active", authorization=receipt)
            issue = self.save(current_issue, state)
        return issue, state

    def advance(self, issue, state, retry_failed=False):
        validate_state(state)
        if state["status"] == "completed":
            return issue, state
        issue, state = self.accept_review(issue, state)
        if state["status"] == "review":
            return issue, state
        require_plan_approval(self.api, issue, state)
        for repository in REPOSITORIES:
            node = state["nodes"][repository]
            if node["status"] in TERMINAL:
                continue
            producers = [f"{OWNER}/{name}" for name in release.PROJECTS[repository.split("/")[1]]]
            if node["status"] == "deferred" or any(state["nodes"][producer]["status"] not in TERMINAL for producer in producers):
                return issue, state
            snapshot = state["plan"]["snapshots"][repository]
            if node["status"] == "failed":
                if not retry_failed:
                    run = {"id": node["run_id"], "run_attempt": node.get("failure_run_attempt", 1)}
                    self.report_failure(
                        issue, state, repository, run, node["reason"], node.get("failure_stages", ()),
                    )
                    return issue, state
                node.pop("run_id", None)
                node["status"] = "pending"
            if node["status"] == "pending":
                metadata = repository_metadata(self.api, repository)
                dispatch_sha = assessed_automation(
                    self.api, repository, snapshot,
                    self.api.request(f"repos/{repository}/branches/{metadata['default_branch']}")["commit"]["sha"],
                )
                binding = assessed_worker(self.api, repository, snapshot, dispatch_sha)
                trusted_shared_revision(self.api, binding)
                source_checkpoint(self.api, repository, snapshot, state["month"])
                worker = self.api.request(f"repos/{repository}/actions/workflows/monthly-patch-release.yml")
                if worker["id"] != snapshot["workflow_id"] or worker["state"] != "active":
                    raise ValueError("Assessed release worker changed or is disabled")
                node.update(status="dispatching", attempt=node["attempt"] + 1, dispatched_at=utc_now(),
                            dispatch_sha=dispatch_sha, worker_sha=binding["worker_sha"])
                issue = self.save(issue, state)
                self.api.request(f"repos/{repository}/actions/workflows/{worker['id']}/dispatches", "POST", {
                    "ref": "main", "inputs": {
                        "mode": MODE, "month": state["month"], "controller_issue": str(issue["number"]),
                        "plan_id": state["plan"]["plan_id"], "attempt": str(node["attempt"]),
                    },
                })
                return issue, state
            if node["status"] == "dispatching":
                candidates = []
                query = urllib.parse.urlencode({"event": "workflow_dispatch", "created": ">=" + node["dispatched_at"]})
                for page in range(1, 11):
                    runs = self.api.request(
                        f"repos/{repository}/actions/workflows/{snapshot['workflow_id']}/runs?{query}&per_page=100&page={page}"
                    )["workflow_runs"]
                    candidates.extend(run for run in runs if run.get("display_title") == run_title(state["month"], state["plan"]["plan_id"], issue["number"], node["attempt"]))
                    if len(runs) < 100:
                        break
                else:
                    raise ValueError("Incomplete dispatched-run discovery")
                if len(candidates) > 1:
                    raise ValueError("Duplicate dispatched workers require reconciliation")
                if not candidates:
                    # Do not re-dispatch an ambiguous request after a runner/network failure.
                    return issue, state
                run = candidates[0]
                if run.get("actor", {}).get("login") != self.actor or run.get("head_sha") != node.get("dispatch_sha", snapshot["automation_sha"]) or run.get("event") != "workflow_dispatch":
                    raise ValueError("Worker provenance does not match the approved dispatch")
                node.update(status="running", run_id=positive_id(run["id"]))
                issue = self.save(issue, state)
            run = self.api.request(f"repos/{repository}/actions/runs/{node['run_id']}")
            if (run.get("actor", {}).get("login") != self.actor or run.get("workflow_id") != snapshot["workflow_id"]
                    or run.get("head_sha") != node.get("dispatch_sha", snapshot["automation_sha"]) or run.get("event") != "workflow_dispatch"
                    or run.get("display_title") != run_title(state["month"], state["plan"]["plan_id"], issue["number"], node["attempt"])):
                raise ValueError("Recorded worker identity changed")
            observed_worker(self.api, repository, run["id"], node["worker_sha"])
            cycle = release.Cycle(self.api, repository, state["month"], MODE)
            final = cycle.state is not None and cycle.state["status"] in release.FINAL
            if run['status'] != 'completed':
                return issue, state
            if run['status'] == 'completed' and run['conclusion'] != 'success':
                details = cycle.state.get("failure_details", {}) if cycle.state else {}
                stages = details.get("stages", ()) if details.get("run_id") == run["id"] else ()
                node.update(status="failed", reason=f"Worker {node['run_id']} concluded {run['conclusion']}",
                            failure_stages=list(stages), failure_run_attempt=positive_id(run.get("run_attempt", 1)))
                issue = self.save(issue, state)
                self.report_failure(issue, state, repository, run, node["reason"], stages)
                return issue, state
            if cycle.state is None:
                # A blocked setup must never look like a successfully skipped patch.
                node.update(status="failed", reason="Worker returned without a trusted release cycle",
                            failure_stages=["prepare"], failure_run_attempt=positive_id(run.get("run_attempt", 1)))
                issue = self.save(issue, state)
                self.report_failure(issue, state, repository, run, node["reason"], node["failure_stages"])
                return issue, state
            worker_plan = cycle.state
            if worker_plan.get("controller_plan_id") != state["plan"]["plan_id"] or worker_plan.get("controller_issue") != issue["number"]:
                raise ValueError("Worker cycle belongs to a different approved plan")
            if worker_plan["status"] not in release.FINAL:
                node.update(status="pending", reason=worker_plan.get("reason", "Worker is waiting for dependency checks"))
                node.pop("run_id", None)
                return self.save(issue, state), state
            if worker_plan["tag"] != snapshot["tag"]:
                raise ValueError("Worker changed its approved patch version")
            if worker_plan["status"] == "published":
                release.assert_tag(self.api, worker_plan)
                release.check_public_assets(self.api, worker_plan)
                expected = {producer.split("/")[1]: {"repository": producer, "version": state["nodes"][producer]["version"]}
                            for producer in producers if state["nodes"][producer]["status"] == "published"}
                if release.producer_lag(self.api, repository, worker_plan["commit"], MODE, expected):
                    raise ValueError("Published consumer does not include its planned producer versions")
                node.update(status="published", version=worker_plan["tag"], commit=worker_plan["commit"])
            else:
                if any(state["nodes"][producer]["status"] == "published" for producer in producers):
                    raise ValueError("A required consumer patch cannot be skipped after its producer released")
                node["status"] = "skipped"
            issue = self.save(issue, state)
        state["status"] = "completed"
        return self.save(issue, state), state


def run_title(month, plan_id, issue, attempt):
    return f"Monthly patch {month} {plan_id} #{issue} attempt {attempt}"


def worker_authorization(api, repository, month, plan=None, allow_previous=False, notification=False):
    if notification and plan is not None:
        raise ValueError("Controller progress notifications cannot authorize writing stages")
    if os.environ.get("MONTHLY_PATCH_COORDINATOR_REQUIRED") != "true":
        raise ValueError("Canonical workers always require a coordinator receipt")
    issue_number = os.environ.get("MONTHLY_PATCH_CONTROLLER_ISSUE", "")
    plan_id = os.environ.get("MONTHLY_PATCH_CONTROLLER_PLAN", "")
    attempt = os.environ.get("MONTHLY_PATCH_CONTROLLER_ATTEMPT", "")
    if not issue_number.isdecimal() or not attempt.isdecimal() or int(issue_number) <= 0 or int(attempt) <= 0 or not DIGEST.fullmatch(plan_id):
        raise ValueError("Writing workers require an approved coordinator receipt")
    repository_metadata(api, repository)
    issue = api.request(f"repos/{HOST}/issues/{issue_number}")
    release.trusted_editor(api, HOST, issue)
    state = decode_issue(issue, os.environ.get("MONTHLY_PATCH_ACTOR"))
    if state is None or state["month"] != month or state["status"] != "active" or state["plan"]["plan_id"] != plan_id:
        raise ValueError("Untrusted, stale or completed coordinator authorization")
    require_plan_approval(api, issue, state)
    node = state["nodes"][repository]
    if state["plan"]["decisions"][repository]["decision"] != "release" or node["status"] not in {"dispatching", "running"} or node["attempt"] != int(attempt):
        raise ValueError("Coordinator did not dispatch this worker attempt")
    snapshot = state["plan"]["snapshots"][repository]
    if os.environ.get("GITHUB_SHA") != node.get("dispatch_sha", snapshot["automation_sha"]) or os.environ.get("GITHUB_REF") != "refs/heads/main":
        raise ValueError("Writing worker must use the assessed default-branch code")
    if release.verified_tag(api, repository, snapshot["baseline"], snapshot["baseline_sha"]) != snapshot["baseline_tag_object"]:
        raise ValueError("Signed stable baseline changed")
    assessed_automation(api, repository, snapshot, os.environ["GITHUB_SHA"])
    binding = assessed_worker(api, repository, snapshot, os.environ["GITHUB_SHA"])
    if (binding["worker_sha"] != node.get("worker_sha")
            or os.environ.get("MONTHLY_PATCH_WORKER_SHA") != node["worker_sha"]):
        raise ValueError("Worker input revision does not match the dispatched shared implementation")
    run_id = os.environ.get("GITHUB_RUN_ID", "")
    if not run_id.isdecimal():
        raise ValueError("Writing worker requires its observed run identity")
    observed = observed_worker(api, repository, int(run_id), binding["worker_sha"])
    if (observed.get("actor", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR")
            or observed.get("head_sha") != node["dispatch_sha"]
            or observed.get("workflow_id") != snapshot["workflow_id"]
            or observed.get("event") != "workflow_dispatch"
            or observed.get("display_title") != run_title(month, plan_id, int(issue_number), int(attempt))):
        raise ValueError("Current worker run does not match the issued dispatch receipt")
    current = api.request(f"repos/{repository}/branches/main")["commit"]["sha"]
    assessed_automation(api, repository, snapshot, current)
    assessed_worker(api, repository, snapshot, current)
    trusted_shared_revision(api, binding)
    source_checkpoint(api, repository, snapshot, month)
    recovery = snapshot.get("verification_recovery")
    if recovery is not None:
        candidate = release.Cycle(api, repository, month, MODE).state
        if (candidate is None or (
                candidate["status"] not in {"draft", "awaiting-publication", "verifying"}
                and not (notification and candidate["status"] == "published"))
                or candidate.get("controller_issue") != int(issue_number)
                or candidate.get("controller_plan_id") not in {plan_id, recovery["from_plan_id"]}
                or candidate["tag"] != snapshot["tag"]
                or digest({key: candidate.get(key) for key in ("tag", "commit", "assets", "release_id")}) != recovery["candidate_sha256"]):
            raise ValueError("Verification recovery cannot change or rebuild its immutable public candidate")
        release.assert_tag(api, candidate)
        if candidate["status"] in {"draft", "awaiting-publication"}:
            release.check_release_assets(api, candidate, True)
        else:
            release.check_public_assets(api, candidate)
    producers = {}
    for name in release.PROJECTS[repository.split("/")[1]]:
        source = f"{OWNER}/{name}"
        upstream = state["nodes"][source]
        if upstream["status"] not in TERMINAL:
            raise ValueError("Producer verification has not completed")
        if upstream["status"] == "published":
            producer_snapshot = state["plan"]["snapshots"][source]
            producer_run = api.request(f"repos/{source}/actions/runs/{positive_id(upstream.get('run_id'))}")
            observed_worker(api, source, producer_run["id"], upstream["worker_sha"])
            if (producer_run.get("actor", {}).get("login") != os.environ.get("MONTHLY_PATCH_ACTOR")
                    or producer_run.get("workflow_id") != producer_snapshot["workflow_id"]
                    or producer_run.get("head_sha") != upstream.get("dispatch_sha")
                    or producer_run.get("event") != "workflow_dispatch"
                    or producer_run.get("conclusion") != "success"
                    or producer_run.get("display_title") != run_title(month, plan_id, int(issue_number), upstream["attempt"])):
                raise ValueError("Producer publication has no trusted successful worker provenance")
            candidate = release.Cycle(api, source, month, MODE).state
            if (not candidate or candidate["status"] != "published"
                    or candidate.get("controller_issue") != int(issue_number)
                    or candidate.get("controller_plan_id") != plan_id
                    or candidate["tag"] != upstream["version"] or candidate["commit"] != upstream["commit"]):
                raise ValueError("Producer cycle does not match the approved publication")
            release.assert_tag(api, candidate)
            release.check_public_assets(api, candidate)
            producers[name] = {"repository": source, "version": upstream["version"]}
    if plan is not None:
        if recovery is not None and (
            plan.get("status") not in {"draft", "awaiting-publication", "verifying"}
            or digest({key: plan.get(key) for key in ("tag", "commit", "assets", "release_id")}) != recovery["candidate_sha256"]
        ):
            raise ValueError("A recovered public candidate authorizes verification only")
        allowed_ids = {plan_id}
        if allow_previous and plan.get("status") in {"merging", "waiting", "voting", "ready"}:
            allowed_ids.update(item["from"] for item in state.get("revisions", []))
        if allow_previous and recovery is not None and plan.get("status") in {"draft", "awaiting-publication", "verifying"}:
            if digest({key: plan.get(key) for key in ("tag", "commit", "assets", "release_id")}) == recovery["candidate_sha256"]:
                allowed_ids.add(recovery["from_plan_id"])
        if (plan.get("mode") != MODE or plan.get("cycle_id", "monthly") != "monthly"
                or plan.get("controller_issue") != int(issue_number) or plan.get("controller_plan_id") not in allowed_ids
                or any(plan.get(key) != snapshot[key] for key in ("baseline", "main", "branch", "tag"))
                or plan.get("producers", {}) != producers):
            raise ValueError("Worker plan escapes its fixed coordinator authorization")
        if plan.get('status') != 'verifying' and plan.get('controller_plan_id') == plan_id:
            if plan.get("maintainer_roster_sha") != snapshot["automation_sha"]:
                raise ValueError("Worker maintainer roster does not match the assessed main")
    return {"issue": int(issue_number), "plan_id": plan_id, "snapshot": snapshot, "producers": producers}


def module_version(value):
    match = re.fullmatch(
        r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)"
        r"(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?",
        value if isinstance(value, str) else "",
    )
    if match is None:
        raise ValueError("Invalid producer module version")
    major, minor, patch, prerelease = match.groups()
    identifiers = []
    for identifier in prerelease.split(".") if prerelease is not None else ():
        if identifier.isdecimal():
            if len(identifier) > 1 and identifier.startswith("0"):
                raise ValueError("Invalid numeric producer prerelease identifier")
            identifiers.append((0, int(identifier)))
        else:
            identifiers.append((1, identifier))
    return int(major), int(minor), int(patch), prerelease is None, tuple(identifiers)


def producer_manifests(before_modules, after_modules, producers, producer_modules, current):
    changed = False
    allowed = {f"github.com/notaryproject/{name}" for name in producers}
    def go_version(value):
        if not isinstance(value, str) or not re.fullmatch(r"\d+\.\d+(?:\.\d+)?", value):
            raise ValueError("Invalid producer Go requirement")
        return tuple(map(int, value.split("."))) + (0,) * (3 - len(value.split(".")))
    producer_go = {name: go_version(manifest["Go"]) for name, manifest in producer_modules.items()}
    for directory, before in before_modules.items():
        after = after_modules[directory]
        changed_go = []
        changed_producers = []
        for name, item in producers.items():
            module = f"github.com/notaryproject/{name}"
            def dependency(manifest):
                return ([entry for entry in manifest.get("Require") or [] if entry["Path"] == module],
                        [entry for entry in manifest.get("Replace") or [] if entry["Old"]["Path"] == module])
            if dependency(before) != dependency(after):
                if release.requirement(after, module, item["repository"], MODE) != item["version"]:
                    return False
                changed = True
                changed_go.append(producer_go[name])
                changed_producers.append(name)
        floor = max([go_version(before["Go"]), *changed_go])
        if before["Go"] != after["Go"] and go_version(after["Go"]) != floor:
            return False
        if before.get("Toolchain") != after.get("Toolchain"):
            chain = after.get("Toolchain") or ""
            if not isinstance(chain, str) or not chain.startswith("go"):
                return False
            if not floor <= go_version(chain.removeprefix("go")) <= go_version(current.removeprefix("go")):
                return False
        required = {}
        for name in changed_producers:
            for entry in producer_modules[name].get("Require") or []:
                path, version = entry["Path"], entry["Version"]
                if path not in required or module_version(version) > module_version(required[path]):
                    required[path] = version
        original = {entry["Path"]: entry for entry in before.get("Require") or []}
        updated = {entry["Path"]: entry for entry in after.get("Require") or []}
        for path, previous in original.items():
            if path not in allowed and not previous.get("Indirect") and path in required:
                if (module_version(required[path]) > module_version(previous["Version"])
                        and updated.get(path, {}).get("Version") != required[path]):
                    return False
        def unrelated(manifest, normalize=False):
            result = copy.deepcopy(manifest)
            # MVS can change indirect requirements when a producer is updated.
            result["Require"] = [item for item in result.get("Require") or [] if item["Path"] not in allowed and not item.get("Indirect")]
            for entry in result["Require"]:
                path = entry["Path"]
                previous = original.get(path)
                if (normalize and previous is not None and path in required
                        and entry["Version"] == required[path]
                        and module_version(required[path]) > module_version(previous["Version"])):
                    entry["Version"] = previous["Version"]
            result["Replace"] = [item for item in result.get("Replace") or [] if item["Old"]["Path"] not in allowed]
            result.pop("Go", None)
            result.pop("Toolchain", None)
            return result
        if unrelated(before) != unrelated(after, normalize=True):
            return False
    return changed


def producer_pull(api, repository, pull, producers):
    """Permit new propagation PRs, but not unrelated changes hidden in grouped updates."""
    if not producers or (pull["head"].get("repo") or {}).get("full_name") != repository:
        return False
    files = list(api.pages(f"repos/{repository}/pulls/{pull['number']}/files"))
    paths = {f"{directory}/go.{extension}".removeprefix("./") for directory in release.modules(repository)
             for extension in ("mod", "sum")}
    if not files or any(item["filename"] not in paths or (item.get("previous_filename") and item["previous_filename"] not in paths) for item in files):
        return False
    before = {directory: api.manifest(repository, pull["base"]["sha"], directory) for directory in release.modules(repository)}
    after = {directory: api.manifest(repository, pull["head"]["sha"], directory) for directory in release.modules(repository)}
    producer_modules = {name: api.manifest(item["repository"], item["version"], ".") for name, item in producers.items()}
    current = release.run(["go", "env", "GOVERSION"]).strip()
    return producer_manifests(before, after, producers, producer_modules, current)


def scoped_pull(api, repository, pull, authorization, merged=()):
    snapshot = authorization["snapshot"]
    heads = {item["number"]: item["head"] for item in snapshot["pulls"]}
    commits = {item["number"]: item["commit"] for item in snapshot["merged"]}
    number = pull["number"]
    approved_merges = {item["number"]: item["commit"] for item in merged}
    if number in approved_merges and pull.get("merged_at"):
        if pull.get("merge_commit_sha") != approved_merges[number]:
            raise ValueError("Worker-approved dependency merge changed")
        return True
    if number in heads:
        if pull["head"]["sha"] != heads[number]:
            raise ValueError("An assessed Dependabot PR head changed; reassessment is required")
        return True
    if number in commits:
        if pull.get("merge_commit_sha") != commits[number]:
            raise ValueError("Assessed dependency backport provenance changed")
        return True
    if pull.get("user", {}).get("login") != "dependabot[bot]":
        return False
    return producer_pull(api, repository, pull, authorization["producers"])


def trusted_assessment(api, run_id, output):
    run_id = positive_id(run_id)
    run = api.request(f"repos/{HOST}/actions/runs/{run_id}")
    workflow = api.request(f"repos/{HOST}/actions/workflows/notation-release-agent.lock.yml")
    metadata = repository_metadata(api, HOST)
    if (run.get("workflow_id") != workflow["id"] or run.get("path") != AGENT_PATH
            or run.get("head_branch") != metadata["default_branch"] or run.get("head_repository", {}).get("full_name") != HOST
            or run.get("event") not in {"schedule", "workflow_dispatch"} or run.get("conclusion") != "success"):
        raise ValueError("Assessment must come from a successful trusted default-branch agent run")
    artifacts = api.request(f"repos/{HOST}/actions/runs/{run_id}/artifacts?per_page=100")
    if artifacts["total_count"] > 100:
        raise ValueError("Unexpected assessment artifact count")
    matches = [item for item in artifacts["artifacts"] if item["name"] == "notation-release-assessment" and not item["expired"]]
    if len(matches) != 1:
        raise ValueError("Missing or duplicate validated assessment artifact")
    artifact = matches[0]
    if artifact.get("workflow_run", {}).get("head_sha") != run["head_sha"]:
        raise ValueError("Assessment artifact provenance changed")
    release.run(["gh", "run", "download", str(run_id), "--repo", HOST, "--name", artifact["name"], "--dir", str(output)])
    envelope = load_json((output / "assessment.json").read_text())
    expected = validate_assessment(envelope["inventory"], envelope["assessment"])
    if expected != envelope["plan"]:
        raise ValueError("Validated assessment was modified")
    assessed_automation(api, HOST, expected["snapshots"][HOST], run["head_sha"])
    return expected


def notify_controller_failure(api):
    if (os.environ.get("GITHUB_REPOSITORY") != HOST or os.environ.get("GITHUB_REF") != "refs/heads/main"
            or os.environ.get("MONTHLY_PATCH_TOKEN_READY") != "true"
            or not os.environ.get("MONTHLY_PATCH_ACTOR")):
        raise ValueError("Failure notification requires the trusted controller and its scoped credential")
    actor = os.environ["MONTHLY_PATCH_ACTOR"]
    if api.request("user")["login"] != actor:
        raise ValueError("Failure notification credential belongs to another actor")
    run_id = os.environ.get("GITHUB_RUN_ID", "")
    if not run_id.isdecimal() or int(run_id) <= 0:
        raise ValueError("Failure notification requires the actual controller run")
    run = api.request(f"repos/{HOST}/actions/runs/{run_id}")
    workflow = api.request(f"repos/{HOST}/actions/workflows/notation-release-controller.yml")
    if (run.get("workflow_id") != workflow["id"]
            or run.get("path") != ".github/workflows/notation-release-controller.yml"
            or run.get("head_branch") != "main" or run.get("head_sha") != os.environ.get("GITHUB_SHA")
            or run.get("repository", {}).get("full_name") != HOST
            or run.get("event") != os.environ.get("GITHUB_EVENT_NAME")
            or run.get("event") not in {"schedule", "repository_dispatch", "workflow_dispatch"}):
        raise ValueError("Failure notification does not identify the trusted controller run")
    controller = Controller(api, actor)
    records = [(issue, state) for issue, state in controller.records() if state["status"] in {"active", "review"}]
    if not records:
        print("No trusted release issue exists; inspect the failed GitHub Actions run.", file=sys.stderr)
        return {"action": "no-release-issue", "run_id": int(run_id)}
    issue, state = records[0]
    controller.report_failure(
        issue, state, HOST, run, "Controller validation or execution failed; inspect the linked logs", ["controller"],
    )
    return {"action": "notified", "issue": issue["number"], "run_id": int(run_id)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=["collect", "validate", "control", "preview", "notify-failure"])
    parser.add_argument("--month", default=datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m"))
    parser.add_argument("--inventory", type=pathlib.Path)
    parser.add_argument("--agent-output", type=pathlib.Path)
    parser.add_argument("--assessment-run", type=int)
    parser.add_argument("--mode", choices=["dry-run", MODE], default="dry-run")
    parser.add_argument("--approve", action="store_true")
    parser.add_argument("--retry-failed", action="store_true")
    parser.add_argument("--replan", action="store_true")
    parser.add_argument("--recover-public-verification", action="store_true")
    parser.add_argument("--output", type=pathlib.Path, required=True)
    args = parser.parse_args()
    if args.recover_public_verification and not (args.assessment_run and args.approve and args.replan and args.mode == MODE):
        parser.error("Public verification recovery requires a fresh assessment, writing mode, approval and replan")
    try:
        args.output.mkdir(parents=True, exist_ok=True)
        api = release.GitHub()
        if args.command == "preview":
            result = preview_scope(api, args.output, args.inventory)
        elif args.command == "notify-failure":
            if args.mode != MODE:
                raise ValueError("Read-only dry runs cannot write failure comments")
            result = notify_controller_failure(api)
        elif args.command == "collect":
            inventory = collect(api, args.month, args.output / "scans")
            validate_inventory(inventory)
            (args.output / "inventory.json").write_text(json.dumps(inventory, indent=2) + "\n")
            (args.output / "inventory.sha256").write_text(digest(inventory) + "\n")
            result = {"inventory_sha256": digest(inventory), "month": args.month, "writes": False}
        elif args.command == "validate":
            inventory = load_json(args.inventory.read_text())
            items = load_json(args.agent_output.read_text())["items"]
            if len(items) != 1 or items[0].get("type") != "submit_release_plan":
                raise ValueError("Agent must submit exactly one structured assessment")
            assessment = load_json(items[0]["assessment_json"])
            plan = validate_assessment(inventory, assessment)
            result = {"inventory": inventory, "assessment": assessment, "plan": plan}
            (args.output / "assessment.json").write_text(json.dumps(result, indent=2) + "\n")
        else:
            if os.environ.get("GITHUB_REPOSITORY") != HOST or os.environ.get("GITHUB_REF") != "refs/heads/main":
                raise ValueError("Controller must run from trusted default main")
            if args.mode != "dry-run" and (
                os.environ.get("NOTATION_RELEASE_COORDINATOR_ENABLED") != "true"
                or not os.environ.get("MONTHLY_PATCH_ACTOR")
                or os.environ.get("MONTHLY_PATCH_TOKEN_READY") != "true"
            ):
                raise ValueError("Writing controller requires explicit opt-in and its scoped credential")
            if args.mode != "dry-run":
                identity = api.request("user")["login"]
                if identity != os.environ["MONTHLY_PATCH_ACTOR"]:
                    raise ValueError("Release credential does not belong to the configured actor")
                if args.approve or args.replan or args.retry_failed or args.recover_public_verification:
                    explicit_actor()
                if os.environ.get('MONTHLY_PATCH_ENABLED') != 'true':
                    raise ValueError("Canonical monthly release execution is disabled")
                controller = Controller(api, os.environ["MONTHLY_PATCH_ACTOR"])
                existing = controller.records()
                if args.assessment_run is None and not any(state["status"] in {"active", "review"} for _, state in existing):
                    runs = api.request(f"repos/{HOST}/actions/workflows/notation-release-agent.lock.yml/runs?status=success&per_page=1")["workflow_runs"]
                    if runs:
                        args.assessment_run = positive_id(runs[0]["id"])
            if args.assessment_run is not None:
                plan = trusted_assessment(api, args.assessment_run, args.output / "download")
                if args.mode == "dry-run":
                    result = {"plan": plan, "writes": False}
                else:
                    controller = Controller(api, os.environ["MONTHLY_PATCH_ACTOR"])
                    records = controller.records()
                    active = [(issue, state) for issue, state in records if state["status"] in {"active", "review"}]
                    completed = any(state["month"] == plan["month"] and state["status"] == "completed" for _, state in records)
                    if not args.approve and active:
                        issue, state = active[0]
                    elif not args.approve and completed:
                        result = {"reason": "Assessed cycle is already completed", "writes": False}
                        (args.output / "result.json").write_text(json.dumps(result, indent=2) + "\n")
                        return 0
                    else:
                        review = None
                        if not args.approve:
                            try:
                                review_scope(api, plan)
                            except ValueError as error:
                                review = {"blocker": str(error)}
                        issue, state = controller.approve(plan, args.assessment_run, args.replan,
                                                          args.recover_public_verification, review)
                    issue, state = controller.advance(issue, state, args.retry_failed)
                    result = {"issue": issue["number"], "state": state}
            elif args.mode == "dry-run":
                result = {"writes": False, "reason": "Supply an assessment run to preview its validated plan"}
            else:
                controller = Controller(api, os.environ["MONTHLY_PATCH_ACTOR"])
                records = [(issue, state) for issue, state in controller.records() if state["status"] in {"active", "review"}]
                if records:
                    issue, state = controller.advance(*records[0], retry_failed=args.retry_failed)
                    result = {"issue": issue["number"], "state": state}
                else:
                    result = {"reason": "No approved active cycle to resume"}
        (args.output / "result.json").write_text(json.dumps(result, indent=2) + "\n")
        print(json.dumps(result, indent=2))
        return 0
    except (ValueError, KeyError, TypeError, OSError, RuntimeError) as error:
        print(f"Notation release coordinator failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
