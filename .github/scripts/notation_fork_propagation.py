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

"""Open signed, plan-bound fork replacement PRs without bypassing normal CI."""

import argparse
import base64
import hashlib
import json
import os
import pathlib
import tempfile

import monthly_release as release
import notation_release_controller as controller


def module_paths(repository):
    return {f"{directory}/go.{extension}".removeprefix("./")
            for directory in release.modules(repository) for extension in ("mod", "sum")}


def source_head(api, plan, authorization):
    cycle = release.Cycle(api, plan["repository"], plan["month"], "rehearse")
    if cycle.state is None or cycle.state["status"] != "waiting":
        raise ValueError("Fork propagation requires a trusted waiting worker cycle")
    controller.worker_authorization(api, plan["repository"], plan["month"], cycle.state)
    head = api.request(f"repos/{plan['repository']}/branches/{plan['main']}")["commit"]["sha"]
    expected = authorization["snapshot"]["main_head"]
    if head != expected:
        ahead = controller.compare_commits(api, plan["repository"], expected, head)
        merged = cycle.state["merged"]
        checkpoint = next((index for index, item in enumerate(merged) if item["commit"] == expected), -1)
        approved = {item["commit"] for item in merged[checkpoint + 1:]}
        if ahead != approved:
            raise ValueError("Isolated main advanced outside the approved dependency merges")
    return head


def manifests(source, repository):
    return {directory: json.loads(release.run(["go", "mod", "edit", "-json"], source / directory))
            for directory in release.modules(repository)}


def prepare(api, plan, output):
    authorization = controller.worker_authorization(api, plan["repository"], plan["month"], plan)
    if authorization is None or plan["mode"] != "rehearse":
        raise ValueError("Replacement propagation is restricted to authorized fork rehearsals")
    producers = authorization["producers"]
    if not producers:
        return {"action": "none", "reason": "No planned producer update"}
    for pull in api.pulls(plan["repository"], plan["main"], "open"):
        if ((release.dependabot(pull, plan["main"]) and controller.scoped_pull(api, plan["repository"], pull, authorization))
                or controller.fork_propagation_pull(api, plan["repository"], pull, authorization)):
            return {"action": "none", "reason": "Waiting for existing approved dependency PR checks"}
    base = source_head(api, plan, authorization)
    if not release.producer_lag(api, plan["repository"], base, "rehearse", producers):
        return {"action": "none", "reason": "Planned producers are already consumed"}
    producer_modules = {name: api.manifest(item["repository"], item["version"], ".") for name, item in producers.items()}
    with tempfile.TemporaryDirectory(prefix="notation-fork-propagation-source-") as temporary:
        source = pathlib.Path(temporary)
        release.git(["init", "--quiet"], source)
        release.git(["remote", "add", "origin", f"https://github.com/{plan['repository']}.git"], source)
        release.git(["fetch", "--quiet", "--depth=1", "origin", base], source)
        release.git(["checkout", "--quiet", "--detach", "FETCH_HEAD"], source)
        if release.git(["rev-parse", "HEAD"], source) != base:
            raise ValueError("Fork propagation source changed")
        before = manifests(source, plan["repository"])
        original = {path: (source / path).read_text(encoding="utf-8") if (source / path).exists() else None
                    for path in module_paths(plan["repository"])}
        for directory in release.modules(plan["repository"]):
            for name, item in producers.items():
                module = f"github.com/notaryproject/{name}"
                if not any(entry["Path"] == module for entry in before[directory].get("Require") or []):
                    continue
                release.run(["go", "mod", "edit", f"-replace={module}=github.com/{item['repository']}@{item['version']}"],
                            source / directory)
            release.run(["go", "mod", "tidy"], source / directory)
        after = manifests(source, plan["repository"])
        current = release.run(["go", "env", "GOVERSION"]).strip()
        if not controller.producer_manifests(before, after, producers, producer_modules, current):
            raise ValueError("Propagation would change unrelated direct dependencies; finish their reviewed updates first")
        if release.producer_lag_from_manifests(plan["repository"], after, "rehearse", producers):
            raise ValueError("Propagation did not consume every planned producer")
        changed = set(release.git(["diff", "--name-only", "--no-renames"], source).splitlines())
        changed.update(release.git(["ls-files", "--others", "--exclude-standard"], source).splitlines())
        if not changed or not changed <= module_paths(plan["repository"]):
            raise ValueError("Propagation changed files outside module manifests")
        files = {path: (source / path).read_text(encoding="utf-8") for path in sorted(changed)}
        result = {"schema": 1, "action": "pull-request", "plan": plan, "base": base,
                  "original": {path: hashlib.sha256(value.encode()).hexdigest() if value is not None else None
                               for path, value in original.items()}, "files": files}
    output.mkdir(parents=True, exist_ok=True)
    return result


def publish(api, prepared, output):
    if prepared.get("action") == "none":
        return prepared
    if set(prepared) != {"schema", "action", "plan", "base", "original", "files"} or type(prepared["schema"]) is not int or prepared["schema"] != 1:
        raise ValueError("Invalid propagation artifact")
    plan = prepared["plan"]
    authorization = controller.worker_authorization(api, plan["repository"], plan["month"], plan)
    if authorization is None or plan["mode"] != "rehearse" or source_head(api, plan, authorization) != prepared["base"]:
        raise ValueError("Propagation source or authorization changed")
    release.policy(plan["repository"], "rehearse", os.environ, api.request(f"repos/{plan['repository']}"))
    files, allowed = prepared["files"], module_paths(plan["repository"])
    if (not files or not set(files) <= allowed or set(prepared["original"]) != allowed
            or any(not isinstance(value, str) for value in files.values())
            or sum(len(value.encode()) for value in files.values()) > 2_000_000):
        raise ValueError("Invalid or oversized propagation file set")
    branch = controller.propagation_branch(plan["month"], authorization["plan_id"])
    for pull in api.pulls(plan["repository"], plan["main"], "open"):
        if pull["head"].get("ref") == branch:
            if not controller.fork_propagation_pull(api, plan["repository"], pull, authorization):
                raise ValueError("Existing propagation branch has an untrusted PR")
            return {"action": "waiting", "pull_request": pull["number"], "url": pull["html_url"]}
    with tempfile.TemporaryDirectory(prefix="notation-fork-propagation-publish-") as temporary:
        source = pathlib.Path(temporary) / "source"
        source.mkdir()
        release.git(["init", "--quiet"], source)
        release.git(["remote", "add", "origin", f"https://github.com/{plan['repository']}.git"], source)
        release.git(["fetch", "--quiet", "--depth=1", "origin", prepared["base"]], source)
        release.git(["checkout", "--quiet", "--detach", "FETCH_HEAD"], source)
        if release.git(["rev-parse", "HEAD"], source) != prepared["base"]:
            raise ValueError("Propagation publish checkout changed")
        before = manifests(source, plan["repository"])
        for path in sorted(allowed):
            file = source / path
            actual = hashlib.sha256(file.read_bytes()).hexdigest() if file.exists() else None
            if actual != prepared["original"][path]:
                raise ValueError("Propagation artifact does not match its immutable source")
        for path, content in files.items():
            (source / path).write_text(content, encoding="utf-8")
        after = manifests(source, plan["repository"])
        producer_modules = {name: api.manifest(item["repository"], item["version"], ".")
                            for name, item in authorization["producers"].items()}
        if not controller.producer_manifests(before, after, authorization["producers"], producer_modules,
                                             release.run(["go", "env", "GOVERSION"]).strip()):
            raise ValueError("Propagation artifact escapes its planned producer scope")
        if release.producer_lag_from_manifests(plan["repository"], after, "rehearse", authorization["producers"]):
            raise ValueError("Propagation artifact does not consume all planned producers")
        with tempfile.TemporaryDirectory(prefix="notation-fork-propagation-signing-") as signing:
            environment = release.signing_environment(signing, api)
            environment["GIT_AUTHOR_DATE"] = environment["GIT_COMMITTER_DATE"] = plan["started_at"]
            release.git(["add", "--", *sorted(files)], source, environment)
            release.git(["commit", "-S", "-m", "Update approved Notation fork producer versions",
                         "-m", "Co-authored-by: Copilot App <223556219+Copilot@users.noreply.github.com>"], source, environment)
            head = release.git(["rev-parse", "HEAD"], source)
            existing = api.optional(f"repos/{plan['repository']}/git/ref/heads/{branch}")
            if existing and existing["object"]["sha"] != head:
                raise ValueError("Propagation branch exists at a different immutable commit")
            if not existing:
                release.git(["-c", "push.followTags=false", "push", "--atomic",
                             f"--force-with-lease=refs/heads/{branch}:", "origin", f"{head}:refs/heads/{branch}"],
                            source, environment)
    receipt = {"schema": 1, "repository": plan["repository"], "month": plan["month"],
               "controller_issue": authorization["issue"], "plan_id": authorization["plan_id"],
               "producers": authorization["producers"], "base": prepared["base"], "head": head}
    marker = f"<!-- notation-fork-propagation:{plan['month']}:{authorization['plan_id']} -->"
    pull = api.request(f"repos/{plan['repository']}/pulls", "POST", {
        "title": "Update approved Notation fork producer versions", "head": branch, "base": plan["main"],
        "maintainer_can_modify": False,
        "body": marker + "\n\nFork rehearsal only. Normal successful CI and required reviews must precede merging.\n\n```json\n"
                + json.dumps(receipt, indent=2) + "\n```\n",
    })
    if pull["user"]["login"] != os.environ["MONTHLY_PATCH_ACTOR"]:
        raise ValueError("Propagation PR was created by an unexpected actor")
    return {"action": "waiting", "pull_request": pull["number"], "url": pull["html_url"]}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("prepare", "publish"))
    parser.add_argument("--plan", type=pathlib.Path)
    parser.add_argument("--prepared", type=pathlib.Path)
    parser.add_argument("--output", type=pathlib.Path, required=True)
    args = parser.parse_args()
    api = release.GitHub()
    if args.command == "prepare":
        if args.plan is None:
            parser.error("prepare requires --plan")
        result = prepare(api, controller.load_json(args.plan.read_text()), args.output)
    else:
        if args.prepared is None:
            parser.error("publish requires --prepared")
        result = publish(api, controller.load_json(args.prepared.read_text()), args.output)
    args.output.mkdir(parents=True, exist_ok=True)
    (args.output / "propagation.json").write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({key: value for key, value in result.items() if key not in {"files", "original", "plan"}}, indent=2))
