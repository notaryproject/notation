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

import hashlib
import io
import json
import pathlib
import os
import subprocess
import tarfile
import tempfile
import unittest
from unittest.mock import patch

import monthly_release_checks as checks
from qualify_monthly_candidate import license_config


def scanner_messages(mode="binary", findings=True):
    messages = [
        {"config": {"protocol_version": "v1.0.0", "scanner_name": "govulncheck",
                    "scanner_version": "v1.8.0", "scan_level": "symbol",
                    "scan_mode": mode, "db": "https://vuln.go.dev"}},
        {"SBOM": {"modules": [{"path": "example.com/test", "version": "v1.0.0"}]}},
    ]
    if findings:
        messages += [
            {"osv": {"id": "GO-2026-1234"}},
            {"finding": {"osv": "GO-2026-1234", "trace": [
                {"module": "example.com/test", "package": "example.com/test", "function": "Verify"}]}},
        ]
    return messages


def stream(messages):
    return "\n".join(json.dumps(item) for item in messages)


def metadata(tag="v1.3.1", commit="a" * 40, flags=""):
    return f"""notation: go1.27.1
        path github.com/notaryproject/notation/cmd/notation
        mod github.com/notaryproject/notation {tag}
        build -ldflags="-w {flags} -X github.com/notaryproject/notation/internal/version.Version={tag.removeprefix('v')} -X github.com/notaryproject/notation/internal/version.GitCommit={commit}"
        build CGO_ENABLED=0
        build GOOS=darwin
        build GOARCH=arm64
"""


class MonthlyReleaseCheckTests(unittest.TestCase):
    def test_source_archive_matches_exactly_despite_windows_git_crlf_conversion(self):
        with tempfile.TemporaryDirectory() as temporary:
            source = pathlib.Path(temporary)
            subprocess.run(["git", "init", "--quiet", str(source)], check=True)
            (source / "go.mod").write_bytes(b"module example.com/library\n\ngo 1.26\n")
            subprocess.run(["git", "add", "go.mod"], cwd=source, check=True)
            subprocess.run(["git", "-c", "commit.gpgsign=false", "-c", "user.name=Test",
                            "-c", "user.email=test@example.com", "commit", "--quiet", "-m", "Fixture"],
                           cwd=source, check=True)
            commit = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=source, text=True).strip()
            command = ["git", "archive", "--format=tar", "--prefix=notation-core-go-1.3.1/", commit]
            expected = subprocess.check_output(["git", "-c", "core.autocrlf=false", *command[1:]], cwd=source)
            subprocess.run(["git", "config", "core.autocrlf", "true"], cwd=source, check=True)
            self.assertNotEqual(subprocess.check_output(command, cwd=source), expected)
            self.assertEqual(checks.library_source_archive(source, "notation-core-go", "v1.3.1", commit), expected)
            self.assertEqual(subprocess.check_output(["git", "config", "core.autocrlf"],
                                                    cwd=source, text=True).strip(), "true")
            with self.assertRaisesRegex(ValueError, "Source archive failed"):
                checks.library_source_archive(source, "notation-core-go", "v1.3.1", "missing-ref")

    def test_asset_names_preserve_six_standard_platform_archives_and_checksum(self):
        names = checks.asset_names("notation", "v1.3.1")
        self.assertEqual(len(names), 7)
        self.assertIn("notation_1.3.1_windows_amd64.zip", names)
        self.assertIn("notation_1.3.1_linux_armv7.tar.gz", names)
        self.assertEqual(checks.asset_names("notation-go", "v1.3.1"), [
            "notation-go_1.3.1_source.tar.gz", "notation-go_1.3.1_checksums.txt"])

    def test_scanner_parses_json_findings_even_when_json_exit_is_zero(self):
        self.assertEqual(checks.scan_messages(stream(scanner_messages()), "binary")[:2], ({"GO-2026-1234"}, set()))
        messages = scanner_messages()
        messages.insert(3, messages[2])
        self.assertEqual(checks.scan_messages(stream(messages), "binary")[0], {"GO-2026-1234"})
        messages[-1]["finding"]["trace"][0].pop("function")
        self.assertEqual(checks.scan_messages(stream(messages), "binary")[:2], (set(), {"GO-2026-1234"}))

    def test_scanner_rejects_missing_duplicate_config_sbom_and_unknown_messages(self):
        cases = [
            [], scanner_messages()[1:], scanner_messages() + [scanner_messages()[0]],
            scanner_messages() + [{"unknown": {}}], scanner_messages()[:1],
            scanner_messages() + [{"osv": {"id": "GO-2026-1234", "modified": "changed"}}],
            scanner_messages() + [{"finding": {"osv": "GO-2026-9999", "trace": []}}],
        ]
        for messages in cases:
            with self.subTest(messages=messages), self.assertRaises(ValueError):
                checks.scan_messages(stream(messages), "binary")
        with self.assertRaises(ValueError):
            checks.scan_messages(stream(scanner_messages(mode="source")), "binary")
        with self.assertRaises(ValueError):
            checks.scan_messages('{"config":{},"config":{}}', "binary")

    def test_binary_gate_retains_evidence_and_blocks_without_trial_exceptions(self):
        scanned = subprocess.CompletedProcess([], 0, stream(scanner_messages()), "")
        rendered = subprocess.CompletedProcess([], 3, "Vulnerability", "")
        with tempfile.TemporaryDirectory() as temporary, patch.object(checks.subprocess, "run", side_effect=[scanned, rendered]), self.assertRaisesRegex(ValueError, "reachable vulnerabilities"):
            checks.scan(["govulncheck"], temporary, "binary")
            self.fail("A zero JSON exit must not hide vulnerabilities")
        scanned = subprocess.CompletedProcess([], 0, stream(scanner_messages(findings=False)), "")
        with tempfile.TemporaryDirectory() as temporary, patch.object(checks.subprocess, "run", side_effect=[scanned, rendered]), self.assertRaisesRegex(ValueError, "disagree"):
            checks.scan(["govulncheck"], temporary, "binary")

    def test_metadata_requires_exact_version_commit_symbols_and_root_module(self):
        checks.verify_metadata(metadata(), "v1.3.1", "a" * 40, "darwin_arm64")
        for content in [
            metadata(commit="b" * 40), metadata(flags="-s"),
            metadata().replace("mod github.com/notaryproject/notation v1.3.1", "mod github.com/notaryproject/notation (devel)"),
            metadata().replace("GOOS=darwin", "GOOS=linux"),
            metadata().replace("CGO_ENABLED=0", "CGO_ENABLED=1"),
            metadata(flags="-X github.com/notaryproject/notation/internal/version.Version=1.3.1"),
        ]:
            with self.subTest(content=content), self.assertRaises(ValueError):
                checks.verify_metadata(content, "v1.3.1", "a" * 40, "darwin_arm64")

    def test_checksums_reject_duplicate_escape_invalid_hash_and_missing_platforms(self):
        with tempfile.TemporaryDirectory() as temporary:
            path = pathlib.Path(temporary) / "checksums.txt"
            digest = "a" * 64
            for content in [
                "", f"{digest}  ../archive.tar.gz\n", f"{digest}  a\\b\n",
                "bad  archive.tar.gz\n", f"{digest}  archive.tar.gz\n{digest}  archive.tar.gz\n",
            ]:
                path.write_text(content)
                with self.subTest(content=content), self.assertRaises(ValueError):
                    checks.checksum_entries(path)
            path = pathlib.Path(temporary) / "notation_1.3.1_checksums.txt"
            path.write_text(f"{digest}  notation_1.3.1_darwin_arm64.tar.gz\n")
            with self.assertRaisesRegex(ValueError, "exact release"):
                checks.verify_checksums(temporary, "notation", "v1.3.1", "darwin_arm64")

    def test_archive_extraction_does_not_extract_paths_or_follow_symlinks(self):
        with tempfile.TemporaryDirectory() as temporary:
            root = pathlib.Path(temporary)
            archive = root / "test.tar.gz"
            with tarfile.open(archive, "w:gz") as target:
                entry = tarfile.TarInfo("notation")
                entry.size = 4
                target.addfile(entry, io.BytesIO(b"test"))
                escape = tarfile.TarInfo("../escaped")
                escape.size = 5
                target.addfile(escape, io.BytesIO(b"other"))
            output = checks.extract_binary(archive, root / "binary")
            self.assertEqual(output.read_bytes(), b"test")
            self.assertFalse((root.parent / "escaped").exists())
            with tarfile.open(archive, "w:gz") as target:
                entry = tarfile.TarInfo("notation")
                entry.type = tarfile.SYMTYPE
                entry.linkname = "../escaped"
                target.addfile(entry)
            with self.assertRaises(ValueError):
                checks.extract_binary(archive, root / "binary")

    def test_all_module_licensing_preserves_repository_policy(self):
        with tempfile.TemporaryDirectory() as temporary:
            source = pathlib.Path(temporary)
            (source / ".github").mkdir()
            (source / ".github/licenserc.yml").write_text(
                "header:\n  comment: on-failure\n\ndependency:\n  files:\n    - ../go.mod\n  licenses:\n    - Apache-2.0\n")
            config = license_config(source, (".", "test/e2e", "test/e2e/plugin"))
            self.assertIn("comment: on-failure", config)
            self.assertIn("licenses:\n    - Apache-2.0", config)
            self.assertIn(str(source / "test/e2e/plugin/go.mod"), config)

    def test_packaged_e2e_passes_downloaded_binary_without_make_or_rebuild(self):
        with patch.object(checks, "checked") as command:
            checks.packaged_e2e(pathlib.Path("/source"), pathlib.Path("/download/notation"), pathlib.Path("/evidence.log"))
        self.assertEqual(command.call_args.args[0], ["bash", "./run.sh", "zot", "/download/notation"])
        self.assertEqual(command.call_args.kwargs["cwd"], pathlib.Path("/source/test/e2e"))

    def test_no_repository_context_can_waive_reachable_findings(self):
        for repository in ("", "notaryproject/notation", "untrusted/notation"):
            with patch.dict(os.environ, {"GITHUB_REPOSITORY": repository, "MONTHLY_PATCH_ADVISORY_DISPOSITIONS": "[]"}), \
                 self.assertRaises(ValueError):
                checks.accepted_findings({"GO-2026-1234"}, {})
        with patch.dict(os.environ, {"MONTHLY_PATCH_ADVISORY_DISPOSITIONS": ""}):
            self.assertEqual(checks.accepted_findings({"GO-2026-1234"}, {}), set())

    def test_canonical_releases_cannot_reuse_fork_advisory_dispositions(self):
        with patch.dict(os.environ, {
            "GITHUB_REPOSITORY": "notaryproject/notation",
            "MONTHLY_PATCH_ADVISORY_DISPOSITIONS": "[]",
        }), self.assertRaisesRegex(ValueError, "Canonical"):
            checks.accepted_findings(set(), {})

    def test_workflow_gates_remote_tagging_on_binary_checks_and_public_completion(self):
        workflow = pathlib.Path(__file__).parent.parent / "workflows/monthly-patch-worker.yml"
        content = workflow.read_text()
        self.assertLess(content.index("build_monthly_assets.py"), content.index("monthly_release.py tag"))
        self.assertIn("needs.verify.result == 'success'", content)
        self.assertIn("fail-fast: false", content)
        self.assertIn("  contents: read", content)
        self.assertNotIn("pull_request_target:", content)
        self.assertIn("inputs.mode == 'execute' && secrets.MONTHLY_PATCH_TOKEN", content)
        self.assertIn("inputs.mode == 'execute' && secrets.MONTHLY_PATCH_SIGNING_KEY", content)
        self.assertNotIn("  schedule:", content)
        self.assertNotIn("dependency-patch-upstream", content)
        self.assertNotIn("--reconcile", content)
        self.assertIn("MONTHLY_PATCH_COORDINATOR_REQUIRED: 'true'", content)
        self.assertIn("monthly_release.py notify-controller", content)
        self.assertIn("--cycle-id", content)


if __name__ == "__main__":
    unittest.main()
