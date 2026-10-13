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

import json
import os
import pathlib
import subprocess
import textwrap
import unittest
from unittest.mock import mock_open, patch


class WorkerBootstrapReleaseTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        workflow = pathlib.Path(__file__).resolve().parents[1] / "workflows/monthly-patch-worker.yml"
        text = workflow.read_text(encoding="utf-8")
        bootstrap = text.split("        id: authorize-worker\n", 1)[1].split(
            "      - uses: actions/checkout@", 1
        )[0]
        if "GH_TOKEN: ${{ github.token }}" not in bootstrap or "secrets." in bootstrap:
            raise AssertionError("Bootstrap must use only the read-only workflow token")
        script = bootstrap.split("          python3 - <<'PY'\n", 1)[1].split(
            "\n          PY", 1
        )[0]
        cls.guard = compile(textwrap.dedent(script), str(workflow), "exec")

    def setUp(self):
        self.sha = "a" * 40
        self.entry = {
            "path": "notaryproject/notation/.github/workflows/monthly-patch-worker.yml@" + self.sha,
            "ref": self.sha,
            "sha": self.sha,
        }
        self.run = {
            "repository": {"full_name": "notaryproject/notation"},
            "head_branch": "main",
            "event": "workflow_dispatch",
            "head_sha": self.sha,
            "referenced_workflows": [self.entry],
        }
        self.environment = {
            "GH_TOKEN": "read-only-workflow-token",
            "GITHUB_REPOSITORY": "notaryproject/notation",
            "GITHUB_REF": "refs/heads/main",
            "GITHUB_EVENT_NAME": "workflow_dispatch",
            "GITHUB_SHA": self.sha,
            "GITHUB_RUN_ID": "123",
            "EXPECTED_WORKER_SHA": self.sha,
            "GITHUB_OUTPUT": "mocked-worker-output",
        }
        self.main = "d" * 40
        self.comparison = {"status": "ahead", "behind_by": 0}

    def execute_guard(self, accepted):
        def api(command, *, text):
            self.assertEqual(command[:2], ["gh", "api"])
            self.assertTrue(text)
            self.assertEqual(os.environ["GH_TOKEN"], "read-only-workflow-token")
            self.assertNotIn("MONTHLY_PATCH_TOKEN", os.environ)
            self.assertNotIn("MONTHLY_PATCH_SIGNING_KEY", os.environ)
            path = command[2]
            if path == f"repos/{self.environment['GITHUB_REPOSITORY']}/actions/runs/123":
                return json.dumps(self.run)
            if path == "repos/notaryproject/notation/branches/main":
                return json.dumps({"commit": {"sha": self.main}})
            self.assertEqual(path, f"repos/notaryproject/notation/compare/{self.environment['EXPECTED_WORKER_SHA']}...{self.main}")
            return json.dumps(self.comparison)

        output = mock_open()
        with patch.dict(os.environ, self.environment, clear=True), \
             patch("subprocess.check_output", side_effect=api) as request, \
             patch("builtins.open", output):
            if accepted:
                exec(self.guard, {})
                self.assertEqual(request.call_count, 3)
                output.assert_called_once_with("mocked-worker-output", "a", encoding="utf-8")
                output().write.assert_called_once_with(
                    f"worker_sha={self.environment['EXPECTED_WORKER_SHA']}\n"
                )
            else:
                with self.assertRaises(SystemExit):
                    exec(self.guard, {})
                output.assert_not_called()

    def test_matching_invoked_sha_is_accepted(self):
        self.execute_guard(True)

    def test_missing_reference_is_rejected(self):
        self.run["referenced_workflows"] = []
        self.execute_guard(False)

    def test_duplicate_reference_is_rejected(self):
        self.run["referenced_workflows"] = [self.entry, self.entry]
        self.execute_guard(False)

    def test_extra_moving_reference_cannot_hide_behind_matching_pin(self):
        self.run["referenced_workflows"] = [
            self.entry, {**self.entry, "path": self.entry["path"].split("@")[0] + "@main"},
        ]
        self.execute_guard(False)

    def test_unmerged_or_divergent_central_revision_is_rejected_before_checkout(self):
        for status, behind in (("behind", 1), ("diverged", 1), ("ahead", 1)):
            self.comparison = {"status": status, "behind_by": behind}
            self.execute_guard(False)

    def test_missing_ancestry_evidence_is_rejected(self):
        self.comparison = {}
        self.execute_guard(False)

    def test_cli_cannot_select_a_worker_commit_different_from_its_caller(self):
        self.environment["GITHUB_SHA"] = "b" * 40
        self.run["head_sha"] = "b" * 40
        self.execute_guard(False)

    def test_api_failure_cannot_authorize_a_checkout(self):
        output = mock_open()
        with patch.dict(os.environ, self.environment, clear=True), \
             patch("subprocess.check_output", side_effect=subprocess.CalledProcessError(1, ["gh", "api"])), \
             patch("builtins.open", output), self.assertRaises(subprocess.CalledProcessError):
            exec(self.guard, {})
        output.assert_not_called()

    def test_mismatched_reference_sha_is_rejected(self):
        self.entry["sha"] = "c" * 40
        self.execute_guard(False)

    def test_wrong_reference_repository_is_rejected(self):
        self.entry["path"] = "attacker/notation/.github/workflows/monthly-patch-worker.yml@main"
        self.execute_guard(False)

    def test_reference_without_ref_suffix_is_rejected(self):
        self.entry["path"] = "notaryproject/notation/.github/workflows/monthly-patch-worker.yml@"
        self.execute_guard(False)

    def test_moving_main_reference_is_rejected_even_with_matching_sha(self):
        self.entry["path"] = "notaryproject/notation/.github/workflows/monthly-patch-worker.yml@main"
        self.entry["ref"] = "refs/heads/main"
        self.execute_guard(False)

    def test_moving_full_main_reference_is_rejected_even_with_matching_sha(self):
        self.entry["path"] = (
            "notaryproject/notation/.github/workflows/monthly-patch-worker.yml@refs/heads/main"
        )
        self.entry["ref"] = "refs/heads/main"
        self.execute_guard(False)

    def test_noncanonical_caller_is_rejected(self):
        self.environment["GITHUB_REPOSITORY"] = "attacker/notation"
        self.execute_guard(False)

    def test_non_main_caller_ref_is_rejected(self):
        self.environment["GITHUB_REF"] = "refs/heads/other"
        self.execute_guard(False)

    def test_non_dispatch_caller_event_is_rejected(self):
        self.environment["GITHUB_EVENT_NAME"] = "schedule"
        self.execute_guard(False)

    def test_non_sha_worker_input_is_rejected(self):
        self.environment["EXPECTED_WORKER_SHA"] = "main"
        self.execute_guard(False)

    def test_mismatched_run_branch_is_rejected(self):
        self.run["head_branch"] = "other"
        self.execute_guard(False)

    def test_mismatched_run_event_is_rejected(self):
        self.run["event"] = "schedule"
        self.execute_guard(False)

    def test_mismatched_run_head_sha_is_rejected(self):
        self.run["head_sha"] = "c" * 40
        self.execute_guard(False)

    def test_mismatched_run_repository_is_rejected(self):
        self.run["repository"]["full_name"] = "attacker/notation"
        self.execute_guard(False)

    def test_local_cli_reusable_reference_uses_caller_sha(self):
        self.environment["EXPECTED_WORKER_SHA"] = self.environment["GITHUB_SHA"]
        self.environment["GITHUB_WORKFLOW_REF"] = (
            "notaryproject/notation/.github/workflows/monthly-patch-release.yml@refs/heads/main"
        )
        self.entry.update({
            "path": (
                "notaryproject/notation/.github/workflows/monthly-patch-worker.yml@"
                + self.environment["GITHUB_SHA"]
            ),
            "ref": self.environment["GITHUB_SHA"],
            "sha": self.environment["GITHUB_SHA"],
        })
        self.execute_guard(True)

    def test_literal_remote_library_reference_uses_central_sha_not_caller_sha(self):
        self.environment["GITHUB_REPOSITORY"] = "notaryproject/notation-core-go"
        self.environment["GITHUB_WORKFLOW_REF"] = (
            "notaryproject/notation-core-go/.github/workflows/monthly-patch-release.yml@refs/heads/main"
        )
        self.run["repository"]["full_name"] = self.environment["GITHUB_REPOSITORY"]
        self.environment["GITHUB_SHA"] = "b" * 40
        self.run["head_sha"] = "b" * 40
        self.assertNotEqual(self.environment["GITHUB_SHA"], self.environment["EXPECTED_WORKER_SHA"])
        self.execute_guard(True)


if __name__ == "__main__":
    unittest.main()
