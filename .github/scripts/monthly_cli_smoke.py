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

"""Native functional checks of an existing CLI, isolated from user configuration."""

import argparse
import json
import os
import pathlib
import shutil
import subprocess
import tempfile


def smoke(binary, source, tag, commit, evidence):
    binary, source, evidence = binary.resolve(), source.resolve(), evidence.resolve()
    evidence.mkdir(parents=True, exist_ok=True)
    transcript, results = [], []
    with tempfile.TemporaryDirectory(prefix="monthly-cli-smoke-") as temporary:
        root = pathlib.Path(temporary)
        home = root / "home"
        home.mkdir()
        environment = {
            **os.environ, "HOME": str(home), "USERPROFILE": str(home),
            "APPDATA": str(home / "appdata"), "LOCALAPPDATA": str(home / "localappdata"),
            "XDG_CONFIG_HOME": str(home / "config"), "XDG_CACHE_HOME": str(home / "cache"),
            "NOTATION_CONFIG": str(home / "config/notation"),
            "NOTATION_LIBEXEC": str(home / "config/notation"),
            "NOTATION_CACHE": str(home / "cache/notation"), "NOTATION_EXPERIMENTAL": "1",
        }

        def invoke(*command, success=True, contains=None):
            result = subprocess.run(
                [str(binary), *command], cwd=root, env=environment,
                capture_output=True, text=True, encoding="utf-8",
            )
            output = result.stdout + result.stderr
            transcript.append("$ notation " + " ".join(command) + "\n" + output)
            (evidence / "functional.log").write_text("\n".join(transcript), encoding="utf-8")
            if success != (result.returncode == 0):
                raise ValueError(f"Unexpected exit {result.returncode}: {command}\n{output}")
            if contains and contains not in output:
                raise ValueError(f"Missing expected output {contains!r}: {command}\n{output}")
            results.append({"command": list(command), "exit_code": result.returncode})
            return output

        version = invoke("version")
        for label, value in (("Version", tag.removeprefix("v")), ("Git commit", commit)):
            values = [line.split(":", 1)[1].strip() for line in version.splitlines() if line.startswith(label + ":")]
            if values != [value]:
                raise ValueError(f"Downloaded executable has incorrect {label}")
        invoke("--help", contains="verify")
        invoke("cert", "generate-test", "--default", "monthly-package-smoke", contains="mark as default signing key")
        invoke("key", "ls", contains="monthly-package-smoke")
        invoke("cert", "ls", contains="monthly-package-smoke")
        policy = {
            "version": "1.0", "trustPolicies": [{
                "name": "downloaded-package-smoke", "registryScopes": ["local/e2e"],
                "signatureVerification": {"level": "strict"},
                "trustStores": ["ca:monthly-package-smoke"],
                "trustedIdentities": ["x509.subject:C=US,ST=WA,O=Notary,CN=monthly-package-smoke"],
            }],
        }
        policy_path = root / "trustpolicy.json"
        policy_path.write_text(json.dumps(policy), encoding="utf-8")
        invoke("policy", "import", str(policy_path))
        invoke("policy", "show", contains="downloaded-package-smoke")
        fixture = source / "test/e2e/testdata/registry/oci_layout/e2e"
        manifest, = json.loads((fixture / "index.json").read_text(encoding="utf-8"))["manifests"]
        references = []
        for signature_format in ("jws", "cose"):
            layout = root / signature_format
            shutil.copytree(fixture, layout)
            reference = f"{layout}@{manifest['digest']}" if signature_format == "jws" else f"{layout}:{manifest['annotations']['org.opencontainers.image.ref.name']}"
            references.append(reference)
            invoke("list", "--oci-layout", reference, contains="has no associated signature")
            invoke("verify", "--oci-layout", "--scope", "local/e2e", reference,
                   success=False, contains="no signature")
            invoke("sign", "--oci-layout", "--signature-format", signature_format, reference, contains="Successfully signed")
            invoke("list", "--oci-layout", reference, contains="application/vnd.cncf.notary.signature")
            invoke("verify", "--oci-layout", "--scope", "local/e2e", reference, contains="Successfully verified signature")
        policy["trustPolicies"][0]["trustedIdentities"] = ["x509.subject:C=US,ST=WA,O=Notary,CN=untrusted-package-smoke"]
        policy_path.write_text(json.dumps(policy), encoding="utf-8")
        invoke("policy", "import", "--force", str(policy_path))
        for reference in references:
            invoke("verify", "--debug", "--oci-layout", "--scope", "local/e2e", reference,
                   success=False, contains="does not match the X.509 trusted identities")
    (evidence / "functional-summary.json").write_text(json.dumps({
        "tag": tag, "commit": commit, "commands": results,
        "jws": "passed", "cose": "passed", "unsigned_rejected": True,
        "untrusted_identity_rejected": True, "isolated_configuration": True,
        "temporary_keys_removed": True,
    }, indent=2) + "\n", encoding="utf-8")
    print(f"Downloaded CLI functional checks passed: {len(results)} commands")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("binary", type=pathlib.Path)
    parser.add_argument("source", type=pathlib.Path)
    parser.add_argument("--tag", required=True)
    parser.add_argument("--commit", required=True)
    parser.add_argument("--evidence", type=pathlib.Path, required=True)
    args = parser.parse_args()
    smoke(args.binary, args.source, args.tag, args.commit, args.evidence)
