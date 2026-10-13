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

"""Build library source assets or gate all six GoReleaser CLI archives."""

import argparse
import gzip
import hashlib
import pathlib
import tempfile

from monthly_release import git
from monthly_release_checks import (
    PLATFORMS, asset_names, checked, extract_binary, library_source_archive, scan,
    verify_checksums, verify_metadata,
)


def build(source, repository, tag, commit, evidence):
    source, evidence = source.resolve(), evidence.resolve()
    evidence.mkdir(parents=True, exist_ok=True)
    if git(["rev-parse", "HEAD"], source) != commit or git(["rev-parse", f"{tag}^{{commit}}"], source) != commit:
        raise ValueError("Assets must be built from the exact qualified, tagged commit")
    project = repository.split("/")[1]
    names = asset_names(project, tag)
    dist = source / "dist"
    dist.mkdir(exist_ok=True)
    if project != "notation":
        archive = library_source_archive(source, project, tag, commit)
        (dist / names[0]).write_bytes(gzip.compress(archive, mtime=0))
        (dist / names[1]).write_text(f"{hashlib.sha256((dist / names[0]).read_bytes()).hexdigest()}  {names[0]}\n")
    archives = verify_checksums(dist, project, tag)
    if project == "notation":
        failures = []
        for platform, name in zip(PLATFORMS, archives):
            with tempfile.TemporaryDirectory(prefix="monthly-binary-") as temporary:
                binary = extract_binary(dist / name, pathlib.Path(temporary) / ("notation.exe" if platform.startswith("windows") else "notation"))
                try:
                    metadata = checked(["go", "version", "-m", binary], evidence / f"{platform}-metadata.txt")
                    verify_metadata(metadata, tag, commit, platform)
                    scan(["govulncheck", "-mode=binary", "-format=json", str(binary)], evidence / platform, "binary")
                except ValueError as error:
                    failures.append(f"{platform}: {error}")
        if failures:
            raise ValueError("\n".join(failures))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("source", type=pathlib.Path)
    parser.add_argument("--repository", required=True)
    parser.add_argument("--tag", required=True)
    parser.add_argument("--commit", required=True)
    parser.add_argument("--evidence", type=pathlib.Path, required=True)
    args = parser.parse_args()
    build(args.source, args.repository, args.tag, args.commit, args.evidence)
