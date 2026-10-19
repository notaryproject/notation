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

"""Qualify the exact candidate with existing module, race, E2E and license checks."""

import argparse
import json
import pathlib
import re

from monthly_release import modules
from monthly_release_checks import checked, scan


def license_config(source, module_directories):
    config = (source / ".github/licenserc.yml").read_text(encoding="utf-8").rstrip() + "\n"
    pattern = r"(?m)(^dependency:\n  files:\n)(?:    - [^\n]+\n)+"
    if len(re.findall(pattern, config)) != 1:
        raise ValueError("Repository license policy needs an explicit dependency.files list")
    files = "".join(f"    - {json.dumps(str(source / directory / 'go.mod'))}\n" for directory in module_directories)
    return re.sub(pattern, lambda match: match[1] + files, config)


def qualify(source, repository, evidence):
    source, evidence = source.resolve(), evidence.resolve()
    if evidence == source or source in evidence.parents:
        raise ValueError("Evidence must be outside the qualified checkout")
    evidence.mkdir(parents=True, exist_ok=True)
    module_directories = modules(repository)
    root = json.loads(checked(["go", "mod", "edit", "-json"], evidence / "manifest.json", source))
    version = checked(["go", "version"], evidence / "compiler.txt", source)
    compiler = re.search(r"go(\d+)\.(\d+)(?:\.(\d+))?", version)
    if not compiler:
        raise ValueError("Cannot identify qualification compiler")
    current = tuple(int(part or 0) for part in compiler.groups())
    for directory in module_directories:
        name = "root" if directory == "." else directory.replace("/", "-")
        module = source / directory
        manifest = json.loads(checked(["go", "mod", "edit", "-json"], evidence / f"{name}-manifest.json", module))
        if repository.startswith("notaryproject/") and manifest.get("Replace"):
            raise ValueError(f"Production release manifest contains replacements: {directory}/go.mod")
        minimum = tuple(int(part) for part in manifest["Go"].split("."))
        minimum += (0,) * (3 - len(minimum))
        root_minimum = tuple(int(part) for part in root["Go"].split("."))
        root_minimum += (0,) * (3 - len(root_minimum))
        if current < minimum or root_minimum < minimum:
            raise ValueError("Root minimum Go must cover every tested module minimum")
        for command in (["go", "mod", "download"], ["go", "mod", "verify"], ["go", "mod", "tidy"], ["go", "vet", "./..."]):
            checked(command, evidence / f"{name}-{'-'.join(command[1:3])}.log", module)
        if directory != "test/e2e":
            checked(["go", "test", "-race", "-covermode=atomic",
                     f"-coverprofile={evidence / (name + '.out')}", "./..."],
                    evidence / f"{name}-tests.log", module)
            checked(["go", "tool", "cover", f"-func={evidence / (name + '.out')}"],
                    evidence / f"{name}-coverage.txt", module)
        scan(["govulncheck", "-format=json", "-test", "./..."], evidence / f"{name}-scan", "source", module)
    checked(["git", "diff", "--exit-code", "--", "go.mod", "go.sum", "**/go.mod", "**/go.sum"], evidence / "module-integrity.log", source)
    config = evidence / "licenses.yml"
    config.write_text(license_config(source, module_directories), encoding="utf-8")
    checked(["license-eye", "-c", str(config), "dependency", "check", "--weak-compatible=true"], evidence / "licenses.log", source)
    if repository.endswith("/notation"):
        checked(["make", "e2e"], evidence / "source-e2e.log", source)
    checked(["git", "diff", "--exit-code", "--", "go.mod", "go.sum", "**/go.mod", "**/go.sum"], evidence / "final-integrity.log", source)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=pathlib.Path)
    parser.add_argument("--repository", required=True)
    parser.add_argument("--evidence", type=pathlib.Path, required=True)
    args = parser.parse_args()
    qualify(args.source, args.repository, args.evidence)
