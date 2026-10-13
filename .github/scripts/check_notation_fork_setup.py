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

"""Check all fork release prerequisites using metadata, never secret values."""

import argparse
import base64
import datetime
import json
import pathlib

import monthly_release as release
import notation_release_controller as controller


def named_metadata(api, path, field):
    items = []
    for page in range(1, 101):
        result = api.request(f"{path}?per_page=100&page={page}")
        batch = result[field]
        items.extend(batch)
        if len(items) == result["total_count"]:
            return {item["name"]: item for item in items}
        if len(batch) < 100:
            raise ValueError("Incomplete setup metadata response")
    raise ValueError("Setup metadata exceeds the bounded pagination limit")


def check(api, month, dependabot_confirmed=False):
    result = {"schema": 1, "month": month, "repositories": {}, "ready": True}
    actors = {}
    for repository in controller.REPOSITORIES:
        metadata = controller.fork_metadata(api, repository)
        baseline = release.baseline_plan(api, repository, month, "rehearse")
        automation = api.request(f"repos/{repository}/branches/main")["commit"]["sha"]
        blockers = []
        if not metadata["has_issues"]:
            blockers.append("Enable Issues")
        if not dependabot_confirmed:
            blockers.append("Enable Dependabot version updates in fork Settings > Advanced Security and confirm all three forks")
        if not controller.installed_worker(api, repository, automation):
            blockers.append("Install the current coordinated worker and fork propagation helpers on main")
        for branch in (baseline["main"], baseline["branch"]):
            ref = api.optional(f"repos/{repository}/branches/{branch}")
            if ref is None:
                blockers.append(f"Create reviewed branch {branch}")
                continue
            commit = ref["commit"]["sha"]
            ci = api.optional(f"repos/{repository}/contents/.github/workflows/notation-fork-ci.yml?ref={commit}")
            if ci is None:
                blockers.append(f"Install fork dependency CI on {branch}")
            else:
                if ci.get("encoding") != "base64":
                    raise ValueError("Cannot inspect fork dependency CI")
                content = base64.b64decode(ci["content"]).decode("utf-8")
                if "qualify_monthly_candidate.py" not in content or "  pull_request:" not in content:
                    blockers.append(f"Use qualification-backed PR CI on {branch}")
            if repository == controller.HOST:
                try:
                    release.check_publisher_guard(api, repository, commit)
                except ValueError as error:
                    blockers.append(f"{branch}: {error}")
            for directory in release.modules(repository):
                manifest = api.manifest(repository, commit, directory)
                for producer in release.PROJECTS[repository.split("/")[1]]:
                    module = f"github.com/notaryproject/{producer}"
                    if directory != "." and not any(item["Path"] == module for item in manifest.get("Require") or []):
                        continue
                    try:
                        release.requirement(manifest, module, f"yizha1/{producer}", "rehearse")
                    except ValueError as error:
                        blockers.append(f"{branch}:{directory}: {error}")
        dependabot = api.optional(f"repos/{repository}/contents/.github/dependabot.yml?ref={automation}")
        if dependabot is None:
            blockers.append("Install the fork Dependabot configuration on main")
        else:
            if dependabot.get("encoding") != "base64":
                raise ValueError("Cannot inspect Dependabot setup")
            if f"target-branch: {baseline['main']}" not in base64.b64decode(dependabot["content"]).decode("utf-8"):
                blockers.append("Configure Dependabot to target monthly-patch-test-main")
        secrets = named_metadata(api, f"repos/{repository}/actions/secrets", "secrets")
        for name in ("MONTHLY_PATCH_TOKEN", "MONTHLY_PATCH_SIGNING_KEY"):
            if name not in secrets:
                blockers.append(f"Install Actions secret {name}")
        if repository == controller.HOST and "COPILOT_GITHUB_TOKEN" not in secrets:
            blockers.append("Install Actions secret COPILOT_GITHUB_TOKEN")
        variables = named_metadata(api, f"repos/{repository}/actions/variables", "variables")
        actors[repository] = variables.get("MONTHLY_PATCH_ACTOR", {}).get("value", "")
        for name in ("MONTHLY_PATCH_ACTOR", "MONTHLY_PATCH_SIGNER_LOGIN", "MONTHLY_PATCH_SIGNER_EMAIL"):
            if not variables.get(name, {}).get("value"):
                blockers.append(f"Set Actions variable {name}")
        signer = variables.get("MONTHLY_PATCH_SIGNER_LOGIN", {}).get("value", "")
        if actors[repository] and signer and actors[repository] != signer:
            blockers.append("Use the actor's signing identity for fork producer PRs")
        enabled = ["MONTHLY_PATCH_REHEARSAL_ENABLED"]
        if repository == controller.HOST:
            enabled.append("NOTATION_RELEASE_COORDINATOR_ENABLED")
        for name in enabled:
            if variables.get(name, {}).get("value") != "true":
                blockers.append(f"Set {name}=true after setup review")
        result["repositories"][repository] = {"automation_sha": automation, "baseline": baseline["baseline"], "blockers": blockers}
    host_actor = actors[controller.HOST]
    for repository, state in result["repositories"].items():
        if host_actor and actors[repository] and actors[repository] != host_actor:
            state["blockers"].append("Use the same MONTHLY_PATCH_ACTOR as the CLI coordinator")
    result["ready"] = all(not state["blockers"] for state in result["repositories"].values())
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--month", default=datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m"))
    parser.add_argument("--output", type=pathlib.Path)
    parser.add_argument("--confirm-fork-dependabot-enabled", action="store_true",
                        help="Owner confirmation that Dependabot version updates are enabled in all three fork settings")
    args = parser.parse_args()
    result = check(release.GitHub(), args.month, args.confirm_fork_dependabot_enabled)
    if args.output:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    for repository, state in result["repositories"].items():
        print(f"{repository}: {'READY' if not state['blockers'] else 'BLOCKED'}")
        for blocker in state["blockers"]:
            print("  " + blocker)
    raise SystemExit(0 if result["ready"] else 1)
