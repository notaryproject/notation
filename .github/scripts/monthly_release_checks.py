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

"""Shared strict source and shipped-asset gates for monthly releases."""

import hashlib
import datetime
import json
import os
import pathlib
import re
import shlex
import subprocess
import tarfile
import zipfile


PLATFORMS = ("darwin_amd64", "darwin_arm64", "linux_amd64",
             "linux_arm64", "linux_armv7", "windows_amd64")


def library_source_archive(source, project, tag, commit):
    archive = subprocess.run(
        ["git", "-c", "core.autocrlf=false", "archive", "--format=tar",
         f"--prefix={project}-{tag.removeprefix('v')}/", commit],
        cwd=source, capture_output=True,
    )
    if archive.returncode:
        raise ValueError(f"Source archive failed: {archive.stderr.decode('utf-8')}")
    return archive.stdout


def checked(command, evidence, cwd=None, environment=None, input_text=None):
    result = subprocess.run(
        [str(item) for item in command], cwd=cwd, env=environment, input=input_text,
        capture_output=True, text=True, encoding="utf-8",
    )
    evidence = pathlib.Path(evidence)
    evidence.parent.mkdir(parents=True, exist_ok=True)
    evidence.write_text(result.stdout + result.stderr, encoding="utf-8")
    print(result.stdout, end="")
    print(result.stderr, end="")
    if result.returncode:
        raise ValueError(f"Command failed ({result.returncode}): {command}; evidence: {evidence}")
    return result.stdout


def scan_messages(output, mode):
    def unique(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError("Duplicate scanner JSON field")
            result[key] = value
        return result
    decoder = json.JSONDecoder(object_pairs_hook=unique)
    messages = []
    while output.strip():
        message, end = decoder.raw_decode(output.lstrip())
        if not isinstance(message, dict) or len(message) != 1:
            raise ValueError("Invalid scanner JSON message")
        key, = message
        if key not in {"config", "progress", "SBOM", "osv", "finding"} or not isinstance(message[key], dict):
            raise ValueError("Unknown or invalid scanner message")
        messages.append(message)
        output = output.lstrip()[end:]
    configs = [message["config"] for message in messages if "config" in message]
    expected = {
        "protocol_version": "v1.0.0", "scanner_name": "govulncheck",
        "scanner_version": "v1.8.0", "scan_level": "symbol",
        "scan_mode": mode, "db": "https://vuln.go.dev",
    }
    if not messages or "config" not in messages[0] or len(configs) != 1 or any(configs[0].get(key) != value for key, value in expected.items()):
        raise ValueError("Unexpected scanner configuration")
    sboms = [message["SBOM"] for message in messages if "SBOM" in message]
    if len(sboms) != 1 or not isinstance(sboms[0].get("modules"), list) or not sboms[0]["modules"]:
        raise ValueError("Missing scanner SBOM")
    advisories = {}
    for message in messages:
        if "osv" in message:
            entry = message["osv"]
            identifier = entry.get("id", "")
            if not re.fullmatch(r"GO-\d{4}-\d+", identifier):
                raise ValueError("Invalid advisory identifier")
            if identifier in advisories and advisories[identifier] != entry:
                raise ValueError("Conflicting advisory records")
            advisories[identifier] = entry
    actionable, informational = set(), set()
    for message in messages:
        if "finding" not in message:
            continue
        finding = message["finding"]
        identifier, trace = finding.get("osv"), finding.get("trace")
        if identifier not in advisories or not isinstance(trace, list) or not trace or not all(isinstance(frame, dict) for frame in trace):
            raise ValueError("Finding has no valid advisory or trace")
        if mode == "binary" and len(trace) != 1:
            raise ValueError("Unexpected binary finding trace")
        first = trace[0]
        if not isinstance(first.get("module"), str) or not first["module"]:
            raise ValueError("Finding has no module")
        if "function" in first:
            if not first["function"] or not first.get("package"):
                raise ValueError("Symbol finding has no function/package")
            actionable.add(identifier)
        else:
            informational.add(identifier)
    return actionable, informational - actionable, advisories


def accepted_findings(actionable, advisories):
    raw = os.environ.get("MONTHLY_PATCH_ADVISORY_DISPOSITIONS", "")
    if not raw:
        return set()
    if os.environ.get("GITHUB_REPOSITORY", "").startswith("notaryproject/"):
        raise ValueError("Canonical releases cannot use fork advisory dispositions; resolve reachable findings")
    policies = json.loads(raw)
    if not isinstance(policies, list):
        raise ValueError("Monthly advisory dispositions must be an explicit JSON list")
    tag = os.environ.get("MONTHLY_RELEASE_TAG", "")
    match = re.fullmatch(r"v(\d+)\.(\d+)\.\d+(?:-monthly-test\.\d{6})?", tag)
    if not match:
        raise ValueError("Advisory dispositions require the exact monthly release tag context")
    accepted = set()
    for policy in policies:
        if not isinstance(policy, dict) or not all(isinstance(policy.get(field), str) and policy[field].strip()
                                                  for field in ("advisory", "repository", "line", "owner", "expires", "modified", "reason", "reference")):
            raise ValueError("Invalid reviewed advisory disposition")
        if policy["repository"] != os.environ.get("GITHUB_REPOSITORY") or policy["line"] != f"{match[1]}.{match[2]}":
            raise ValueError("Advisory disposition belongs to another repository or release line")
        if datetime.datetime.now(datetime.timezone.utc).date() >= datetime.date.fromisoformat(policy["expires"]):
            raise ValueError("Reviewed advisory disposition has expired")
        identifier = policy["advisory"]
        if not re.fullmatch(r"GO-\d{4}-\d+", identifier) or identifier in accepted:
            raise ValueError("Invalid or duplicate advisory disposition")
        if identifier not in actionable:
            continue
        entry = advisories[identifier]
        if entry.get("modified") != policy["modified"] or entry.get("withdrawn"):
            raise ValueError("Advisory metadata changed; renewed review is required")
        affected = entry.get("affected")
        if not isinstance(affected, list) or not affected:
            raise ValueError("Reviewed advisory has invalid affected-version metadata")
        if any("fixed" in event for module in affected for interval in module.get("ranges", []) for event in interval.get("events", [])):
            raise ValueError("Reviewed advisory now lists a fixed version; renew its disposition")
        accepted.add(identifier)
        print(f"Reviewed advisory disposition: {identifier}; owner={policy['owner']}; expires={policy['expires']}. {policy['reason']}")
    return accepted


def scan(command, evidence, mode, cwd=None):
    evidence = pathlib.Path(evidence)
    evidence.mkdir(parents=True, exist_ok=True)
    result = subprocess.run(command, cwd=cwd, capture_output=True, text=True, encoding="utf-8")
    (evidence / "scan.jsonstream").write_text(result.stdout, encoding="utf-8")
    (evidence / "scan.stderr.txt").write_text(result.stderr, encoding="utf-8")
    if result.returncode:
        raise ValueError(f"Scanner execution failed ({result.returncode}); see {evidence}")
    actionable, informational, advisories = scan_messages(result.stdout, mode)
    rendered = subprocess.run(
        ["govulncheck", "-mode=convert"], input=result.stdout,
        capture_output=True, text=True, encoding="utf-8",
    )
    (evidence / "scan.txt").write_text(rendered.stdout + rendered.stderr, encoding="utf-8")
    print(rendered.stdout, end="")
    accepted = accepted_findings(actionable, advisories)
    blocked = actionable - accepted
    summary = {"blocked": sorted(blocked), "accepted": sorted(accepted), "informational": sorted(informational)}
    (evidence / "summary.json").write_text(json.dumps(summary, indent=2) + "\n", encoding="utf-8")
    if rendered.returncode not in (0, 3) or (rendered.returncode == 3) != bool(actionable):
        raise ValueError("Scanner JSON/text findings disagree")
    if blocked:
        raise ValueError(f"Release blocked by reachable vulnerabilities: {', '.join(sorted(blocked))}")


def asset_names(project, tag):
    version = tag.removeprefix("v")
    if project == "notation":
        archives = [f"{project}_{version}_{platform}." + ("zip" if platform.startswith("windows") else "tar.gz") for platform in PLATFORMS]
    else:
        archives = [f"{project}_{version}_source.tar.gz"]
    return archives + [f"{project}_{version}_checksums.txt"]


def checksum_entries(path):
    entries = {}
    for line in pathlib.Path(path).read_text(encoding="utf-8").splitlines():
        match = re.fullmatch(r"([0-9a-f]{64})  (\S+)", line)
        if not match or pathlib.PurePosixPath(match[2]).name != match[2] or "\\" in match[2] or match[2] in entries:
            raise ValueError("Invalid or duplicate checksum entry")
        entries[match[2]] = match[1]
    if not entries:
        raise ValueError("Empty checksum manifest")
    return entries


def verify_checksums(directory, project, tag, platform=None):
    directory = pathlib.Path(directory)
    names = asset_names(project, tag)
    entries = checksum_entries(directory / names[-1])
    if set(entries) != set(names[:-1]):
        raise ValueError("Checksum manifest does not describe the exact release asset set")
    archives = names[:-1]
    if platform:
        if platform not in PLATFORMS or project != "notation":
            raise ValueError("Invalid native platform selection")
        archives = [name for name in archives if f"_{platform}." in name]
    for name in archives:
        if hashlib.sha256((directory / name).read_bytes()).hexdigest() != entries[name]:
            raise ValueError(f"Checksum mismatch: {name}")
    return archives


def extract_binary(archive, destination):
    archive, destination = pathlib.Path(archive), pathlib.Path(destination)
    binary_name = "notation.exe" if archive.suffix == ".zip" else "notation"
    if archive.suffix == ".zip":
        with zipfile.ZipFile(archive) as source:
            members = [item for item in source.infolist() if item.filename == binary_name and not item.is_dir()]
            if len(members) != 1:
                raise ValueError("Archive must contain exactly one root CLI binary")
            content = source.read(members[0])
    else:
        with tarfile.open(archive, "r:gz") as source:
            members = [item for item in source.getmembers() if item.name == binary_name and item.isfile()]
            if len(members) != 1:
                raise ValueError("Archive must contain exactly one regular root CLI binary")
            content = source.extractfile(members[0]).read()
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_bytes(content)
    destination.chmod(0o755)
    return destination


def verify_metadata(metadata, tag, commit, platform):
    tokens = shlex.split(metadata)
    flags = [item.removeprefix("-ldflags=") for item in tokens if item.startswith("-ldflags=")]
    if len(flags) != 1:
        raise ValueError("Missing or duplicate linker metadata")
    arguments = shlex.split(flags[0])
    if "-s" in arguments:
        raise ValueError("Stripped symbol tables do not qualify precise binary scans")
    assignments = []
    for index, argument in enumerate(arguments):
        if argument == "-X":
            if index + 1 == len(arguments):
                raise ValueError("Truncated linker assignment")
            assignments.append(arguments[index + 1])
        elif argument.startswith("-X="):
            assignments.append(argument[3:])
    for field, expected in (("Version", tag.removeprefix("v")), ("GitCommit", commit)):
        values = [value for assignment in assignments
                  for name, separator, value in [assignment.partition("=")]
                  if separator and name == f"github.com/notaryproject/notation/internal/version.{field}"]
        if values != [expected]:
            raise ValueError(f"Binary {field} does not match qualified candidate")
    os_name, arch = platform.split("_")
    expected = {"GOOS": os_name, "GOARCH": "arm" if arch == "armv7" else arch, "CGO_ENABLED": "0"}
    if arch == "armv7":
        expected["GOARM"] = "7"
    for name, value in expected.items():
        values = [item.split("=", 1)[1] for item in tokens if item.startswith(name + "=")]
        if len(values) != 1 or values[0].split(",")[0] != value:
            raise ValueError(f"Unexpected binary target setting: {name}")
    modules = re.findall(r"^\s*mod\s+(\S+)\s+(\S+)", metadata, re.MULTILINE)
    if modules != [("github.com/notaryproject/notation", tag)]:
        raise ValueError("Root-module version is not the exact release tag")


def packaged_e2e(source, binary, evidence):
    checked(["bash", "./run.sh", "zot", str(pathlib.Path(binary).resolve())],
            evidence, cwd=pathlib.Path(source) / "test/e2e")
