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

import base64
import copy
import pathlib
import re
import unittest
from unittest.mock import patch

import notation_release_controller as controller
from test_upstream_authorization import UpstreamAPI, SOURCE, AUTOMATION, CORE, CLI


def caller(reference, pin):
    return f"jobs:\n  release:\n    uses: {reference}\n    with:\n      worker_sha: {pin}\n"


class SharedReleaseTests(unittest.TestCase):
    def setUp(self):
        self.api = UpstreamAPI()

    def bind(self, repository, text):
        with patch.object(controller, "repository_content", return_value=text):
            return controller.worker_binding(self.api, repository, SOURCE)

    def test_cli_relative_call_binds_shared_code_to_exact_caller_commit(self):
        binding = self.bind(CLI, caller("./" + controller.SHARED_PATH, "${{ github.sha }}"))
        self.assertEqual(binding["worker_sha"], SOURCE)
        self.assertEqual(binding["worker_repository"], CLI)
        self.assertEqual(binding["worker_digest"], AUTOMATION)

    def test_library_literal_call_binds_same_full_sha_for_execution_and_checkout(self):
        pin = "f" * 40
        binding = self.bind(CORE, caller(CLI + "/" + controller.SHARED_PATH + "@" + pin, pin))
        self.assertEqual(binding["worker_sha"], pin)
        self.assertNotEqual(binding["worker_sha"], SOURCE)

    def test_moving_mismatched_or_arbitrary_worker_identity_is_rejected(self):
        good = CLI + "/" + controller.SHARED_PATH + "@" + SOURCE
        for reference, pin in (
            (good, "b" * 40), (good.replace(SOURCE, "main"), SOURCE),
            (good.replace("notaryproject", "attacker"), SOURCE),
            (good.replace("monthly-patch-worker", "other"), SOURCE),
            ("./" + controller.SHARED_PATH, SOURCE),
            (good, "${{ github.sha }}"),
        ):
            with self.subTest(reference=reference, pin=pin), self.assertRaises(ValueError):
                self.bind(CORE, caller(reference, pin))
        with self.assertRaises(ValueError):
            self.bind(CLI, caller(good, SOURCE))

    def test_duplicate_bindings_or_extra_caller_job_fail_closed(self):
        text = caller(CLI + "/" + controller.SHARED_PATH + "@" + SOURCE, SOURCE)
        for bad in (text + "      worker_sha: " + SOURCE + "\n",
                    text + "  malicious:\n    runs-on: ubuntu-latest\n",
                    text + "    uses: ./other.yml\n"):
            with self.subTest(bad=bad), self.assertRaises(ValueError):
                self.bind(CORE, bad)

    def test_fork_network_resolvable_but_unmerged_shared_revision_is_rejected(self):
        binding = controller.worker_binding(self.api, CORE, SOURCE)
        original = self.api.request
        for status, behind in (("behind", 1), ("diverged", 1)):
            def request(path, status=status, behind=behind):
                return {"status": status, "behind_by": behind} if "/compare/" in path else original(path)
            with patch.object(self.api, "request", side_effect=request), self.assertRaisesRegex(ValueError, "reachable"):
                controller.trusted_shared_revision(self.api, binding)
        controller.trusted_shared_revision(self.api, binding)

    def test_default_main_changed_shared_automation_rejects_old_pin(self):
        binding = controller.worker_binding(self.api, CORE, SOURCE)
        with patch.object(controller, "automation_digest", return_value="f" * 64), \
             self.assertRaisesRegex(ValueError, "Central worker code changed"):
            controller.trusted_shared_revision(self.api, binding)

    def test_actual_reusable_run_reference_is_required_and_sha_pinned(self):
        valid = {"path": CLI + "/" + controller.SHARED_PATH + "@" + SOURCE, "sha": SOURCE}
        for references in ([], [dict(valid, sha="f" * 40)], [valid, valid],
                           [dict(valid, path="attacker/notation/" + controller.SHARED_PATH + "@" + SOURCE)]):
            with patch.object(self.api, "request", return_value={"referenced_workflows": references}), \
                 self.assertRaisesRegex(ValueError, "Observed"):
                controller.observed_worker(self.api, CLI, 123, SOURCE)
        for suffix in ("main", "refs/heads/main"):
            with patch.object(self.api, "request", return_value={
                    "referenced_workflows": [dict(valid, path=CLI + "/" + controller.SHARED_PATH + "@" + suffix)]}), \
                 self.assertRaises(ValueError):
                controller.observed_worker(self.api, CLI, 123, SOURCE)
        with patch.object(self.api, "request", return_value={"referenced_workflows": [valid]}):
            controller.observed_worker(self.api, CLI, 123, SOURCE)

    def test_changed_library_pin_does_not_resume_on_unchanged_caller_sha(self):
        snapshot = controller.worker_binding(self.api, CORE, SOURCE)
        changed = copy.deepcopy(snapshot)
        changed["worker_sha"] = "b" * 40
        with patch.object(controller, "worker_binding", return_value=changed), \
             self.assertRaisesRegex(ValueError, "pin changed"):
            controller.assessed_worker(self.api, CORE, snapshot, SOURCE)

    def test_same_content_cli_dependency_merge_retains_automation_authority(self):
        snapshot = controller.worker_binding(self.api, CLI, SOURCE)
        changed = {**snapshot, "worker_sha": "b" * 40}
        with patch.object(controller, "worker_binding", return_value=changed):
            self.assertEqual(controller.assessed_worker(self.api, CLI, snapshot, "b" * 40), changed)

    def test_caller_and_worker_have_no_implicit_secret_inheritance_or_collision(self):
        root = pathlib.Path(__file__).parent.parent / "workflows"
        worker = (root / "monthly-patch-worker.yml").read_text()
        dispatch = (root / "monthly-patch-release.yml").read_text()
        self.assertIn("  workflow_call:", worker)
        self.assertNotIn("  workflow_dispatch:", worker)
        self.assertIn("uses: ./.github/workflows/monthly-patch-worker.yml", dispatch)
        self.assertIn("worker_sha: ${{ github.sha }}", dispatch)
        self.assertNotIn("secrets: inherit", dispatch)
        self.assertNotIn("\nenv:", dispatch)
        self.assertNotIn("\nconcurrency:", dispatch)
        self.assertIn("referenced_workflows", worker)
        self.assertNotIn("GITHUB_WORKFLOW_REF", worker)
        for name, next_name in (("qualify", "package"), ("package", "publish"), ("verify", "complete")):
            section = worker.split(f"  {name}:\n", 1)[1].split(f"  {next_name}:\n", 1)[0]
            self.assertNotIn("secrets.", section)
        for section in worker.split("with:\n")[1:]:
            settings = section.split("      -", 1)[0]
            if "path: automation" in settings:
                self.assertIn("repository: notaryproject/notation", settings)
                self.assertTrue("ref: ${{ inputs.worker_sha }}" in settings
                                or "ref: ${{ needs.prepare.outputs.worker_sha }}" in settings
                                or "ref: ${{ steps.authorize-worker.outputs.worker_sha }}" in settings)


if __name__ == "__main__":
    unittest.main()
