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

"""Download public assets, verify provenance, and test the actual shipped package."""

import argparse
import gzip
import hashlib
import json
import os
import pathlib
import urllib.error
import urllib.parse
import urllib.request

from monthly_cli_smoke import smoke
from monthly_release import GitHub, assert_tag, check_public_assets, check_release_assets, git, validate_plan
from monthly_release_checks import (
    asset_names, checked, extract_binary, library_source_archive, packaged_e2e, scan,
    verify_checksums, verify_metadata,
)


class NoCredentialRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, request, response, code, message, headers, newurl):
        return None


def download_draft(repository, identifier):
    token = os.environ.get("GH_TOKEN", "")
    if not token or type(identifier) is not int or identifier <= 0:
        raise ValueError("Draft asset download requires its recorded ID and read-only workflow token")
    request = urllib.request.Request(
        f"https://api.github.com/repos/{repository}/releases/assets/{identifier}",
        headers={"Authorization": f"Bearer {token}", "Accept": "application/octet-stream",
                 "X-GitHub-Api-Version": "2022-11-28"},
    )
    try:
        with urllib.request.build_opener(NoCredentialRedirect()).open(request, timeout=120) as response:
            return response.read()
    except urllib.error.HTTPError as error:
        if error.code != 302:
            raise ValueError(f"Read-only draft asset access failed (HTTP {error.code}); no write-token fallback") from None
        location = error.headers.get("Location", "")
        target = urllib.parse.urlsplit(location)
        if (target.scheme != "https" or target.hostname != "release-assets.githubusercontent.com"
                or target.username or target.password or target.port not in {None, 443}):
            raise ValueError("Untrusted draft asset redirect") from None
    # Signed storage redirects receive no Authorization header or token.
    with urllib.request.urlopen(location, timeout=120) as response:
        return response.read()


def verify(plan, source, platform, evidence):
    validate_plan(plan)
    draft = plan["status"] == "draft"
    if plan["status"] not in {"draft", "verifying"}:
        raise ValueError("Only an awaiting-verification release can complete this gate")
    source, evidence = source.resolve(), evidence.resolve()
    evidence.mkdir(parents=True, exist_ok=True)
    api = GitHub()
    assert_tag(api, plan)
    check_release_assets(api, plan, True) if draft else check_public_assets(api, plan)
    if git(["rev-parse", "HEAD"], source) != plan["commit"]:
        raise ValueError("Downloaded release tests must use its qualified source fixtures")
    project = plan["repository"].split("/")[1]
    names = asset_names(project, plan["tag"])
    wanted = [name for name in names if name.endswith("checksums.txt") or project != "notation" or f"_{platform}." in name]
    recorded = {item["name"]: item for item in plan["assets"]}
    if set(recorded) != set(names):
        raise ValueError("Publication manifest has an unexpected asset set")
    download = evidence / "download"
    download.mkdir(exist_ok=True)
    for name in wanted:
        expected = recorded[name]
        if draft:
            content = download_draft(plan["repository"], expected["id"])
        else:
            url = f"https://github.com/{plan['repository']}/releases/download/{plan['tag']}/{name}"
            # No credentials: this gate also establishes public download availability.
            with urllib.request.urlopen(url, timeout=120) as response:
                content = response.read()
        if len(content) != expected["size"] or hashlib.sha256(content).hexdigest() != expected["sha256"]:
            raise ValueError(f"Downloaded bytes differ from qualified asset: {name}")
        (download / name).write_bytes(content)
    archives = verify_checksums(download, project, plan["tag"], platform if project == "notation" else None)
    if project == "notation":
        binary = extract_binary(download / archives[0], evidence / ("notation.exe" if platform.startswith("windows") else "notation"))
        metadata = checked(["go", "version", "-m", binary], evidence / "metadata.txt")
        verify_metadata(metadata, plan["tag"], plan["commit"], platform)
        scan(["govulncheck", "-mode=binary", "-format=json", str(binary)], evidence / "native-scan", "binary")
        smoke(binary, source, plan["tag"], plan["commit"], evidence / "functional")
        if platform == "linux_amd64":
            packaged_e2e(source, binary, evidence / "packaged-e2e.log")
    else:
        expected = library_source_archive(source, project, plan["tag"], plan["commit"])
        if gzip.decompress((download / archives[0]).read_bytes()) != expected:
            raise ValueError("Public library archive does not match the qualified source")
    (evidence / "verified.json").write_text(json.dumps({
        "repository": plan["repository"], "tag": plan["tag"],
        "commit": plan["commit"], "platform": platform, "assets": wanted,
        "release_id": plan["release_id"], "public_downloads": not draft, "draft_downloads": draft, "passed": True,
    }, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--plan", type=pathlib.Path, required=True)
    parser.add_argument("--source", type=pathlib.Path, required=True)
    parser.add_argument("--platform", required=True)
    parser.add_argument("--evidence", type=pathlib.Path, required=True)
    args = parser.parse_args()
    verify(json.loads(args.plan.read_text(encoding="utf-8")), args.source, args.platform, args.evidence)
