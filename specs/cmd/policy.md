# notation policy

## Description

As part of signature verification workflow of signed OCI artifacts, users need to configure trust policy configuration to specify trusted identities that signed the artifacts, the level of signature verification to use and other settings. For more details, see [OCI trust policy specification and examples](https://github.com/notaryproject/specifications/blob/main/specs/trust-store-trust-policy.md#oci-trust-policy).

The `notation policy` command provides a user-friendly way to manage trust policies for signed OCI images. It allows users to show trust policy configuration, import/export trust policy configuration from/to a JSON file. Users who want to manage trust policies for signed arbitrary blobs, please refer to `notation blob policy` command.

To get started, user can refer to the following trust policy configuration sample `trustpolicy.oci.json` that is applicable for verifying signed OCI artifacts using `notation verify` command. In this sample, there are four policies configured for different requirements:

- The Policy named "wabbit-networks-images" is for verifying OCI artifacts signed by Wabbit Networks and stored in two repositories `registry.acme-rockets.io/software/net-monitor` and `registry.acme-rockets.io/software/net-logger`.
- Policy named "unsigned-image" is for skipping the verification on unsigned OCI artifacts stored in repository `registry.acme-rockets.io/software/unsigned/net-utils`.
- Policy "allow-expired-images" is for logging instead of failing expired OCI artifacts stored in repository `registry.acme-rockets.io/software/legacy/metrics`.
- Policy "global-policy-for-all-other-images" is for verifying any other OCI artifacts that signed by the ACME Rockets.

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            "name": "wabbit-networks-images",
            "registryScopes": [
              "registry.acme-rockets.io/software/net-monitor",
              "registry.acme-rockets.io/software/net-logger"
            ],
            "signatureVerification": {
                "level": "strict"
            },
            "trustStores": [
              "ca:wabbit-networks",
            ],
            "trustedIdentities": [
                "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Security Tools"
            ]
        },
        {
            "name": "unsigned-image",
            "registryScopes": [ "registry.acme-rockets.io/software/unsigned/net-utils" ],
            "signatureVerification": {
              "level" : "skip"
            }
        },
        {
            "name": "allow-expired-images",
            "registryScopes": [ "registry.acme-rockets.io/software/legacy/metrics" ],
            "signatureVerification": {
              "level" : "strict",
              "override" : {
                "expiry" : "log"
              }
            },
            "trustStores": ["ca:acme-rockets"],
            "trustedIdentities": ["*"]
        },
        {
            "name": "global-policy-for-all-other-images",
            "registryScopes": [ "*" ],
            "signatureVerification": {
                "level": "strict"
            },
            "trustStores": [
              "ca:acme-rockets"
            ],
            "trustedIdentities": [
                "x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io, CN=SecureBuilder"
            ]
        }
    ]
}
```

## Outline

### notation policy command

```text
Manage OCI trust policy configuration for OCI artifact signature verification.

Usage:
  notation policy [command]

Available Commands:
  import    import OCI trust policy configuration from a JSON file
  init      initialize OCI trust policy configuration
  show      show OCI trust policy configuration

Flags:
  -h, --help   help for policy
```

### notation policy init

```text
Initialize OCI trust policy configuration.

Usage:
  notation policy init [flags] --name <policy_name> --trust-store "<store_type>:<store_name>" --trusted-identity "<trusted_identity>"

Flags:
      --force                          override the existing OCI trust policy configuration, never prompt (default --force=false)
  -h, --help                           help for init
  -n, --name                           name of the OCI trust policy
      --registry-scope stringArray     registry scope the policy applies to, e.g. "registry.acme-rockets.io/software/net-monitor"; defaults to "*" for all registries
      --trust-store stringArray        trust store in the format "<store_type>:<store_name>"
      --trusted-identity stringArray   trusted identity, use the format "x509.subject:<subject_of_signing_certificate>" for x509 CA scheme and "<signing_authority_identity>" for x509 signingAuthority scheme
```

### notation policy import

```text
Import OCI trust policy configuration from a JSON file.

Usage:
  notation policy import [flags] <file_path>

Flags:
      --force     override the existing OCI trust policy configuration without prompt
  -h, --help      help for import
```

### notation policy show

```text
Show OCI trust policy configuration

Usage:
  notation policy show [flags]

Flags:
  -h, --help      help for show
```

## Usage

### Initialize trust policy configuration

New users often don't have a trust policy yet and hand-writing `trustpolicy.oci.json` from the specification is error prone. `notation policy init` scaffolds a valid single-statement policy from flags so the file is ready for `notation verify` right away.

```shell
notation policy init --name "wabbit-networks-images" --trust-store "ca:wabbit-networks" --trusted-identity "x509.subject:C=US,ST=WA,O=wabbit-networks.io"
```

The statement is generated with `signatureVerification.level` set to `strict`. Sample output for a successful initialization:

```jsonc
{
  "version": "1.0",
  "trustPolicies": [
    {
      "name": "wabbit-networks-images",
      "registryScopes": [ "*" ],
      "signatureVerification": {
        "level": "strict"
      },
      "trustStores": [ "ca:wabbit-networks" ],
      "trustedIdentities": [
        "x509.subject:C=US,ST=WA,O=wabbit-networks.io"
      ]
    }
  ]
}
```

The `--trust-store` and `--trusted-identity` flags can be repeated to configure multiple trust stores or trusted identities. To trust any identity, set `--trusted-identity` to `"*"`; this is not recommended for production and cannot be combined with other values.

Unlike a blob trust policy, an OCI trust policy statement is bound to registry scopes rather than a global flag. By default the generated statement uses the wildcard scope `"*"` so it applies to every registry. Use `--registry-scope` (repeatable) to pin it to specific repositories instead:

```shell
notation policy init --name "wabbit-networks-images" --registry-scope "registry.acme-rockets.io/software/net-monitor" --trust-store "ca:wabbit-networks" --trusted-identity "x509.subject:C=US,ST=WA,O=wabbit-networks.io"
```

The generated policy is validated according to [trust policy properties](https://github.com/notaryproject/notaryproject/specs/trust-store-trust-policy.md#trust-policy-properties) before it is written; if validation fails no file is written and the reason is printed to standard error.

If a trust policy configuration already exists, the command prompts for confirmation before overwriting it. Use the `--force` flag to overwrite without a prompt:

```shell
notation policy init --force --name "wabbit-networks-images" --trust-store "ca:wabbit-networks" --trusted-identity "x509.subject:C=US,ST=WA,O=wabbit-networks.io"
```

### Import trust policy configuration from a JSON file

An example of import trust policy configuration from a JSON file:

```shell
notation policy import ./my_policy.json
```

The trust policy configuration in the JSON file should be validated according to [trust policy properties](https://github.com/notaryproject/notaryproject/specs/trust-store-trust-policy.md#trust-policy-properties). A successful message should be printed out if trust policy configuration are imported successfully. Error logs including the reason should be printed out if the importing fails.

If there is an existing trust policy configuration, prompt for users to confirm whether discarding existing configuration or not. Users can use `--force` flag to discard existing trust policy configuration without prompt.

### Show trust policies

Use the following command to show trust policy configuration:

```shell
notation policy show
```

Upon successful execution, the trust policy configuration is printed out to standard output. If trust policy is not configured or is malformed, users should receive an error message via standard error output, and a tip to import trust policy configuration from a JSON file.

### Export OCI trust policy configuration into a JSON file

Users can redirect the output of command `notation policy show` to a JSON file.

```shell
notation policy show > ./oci_trust_policy.json
```

### Update trust policy configuration

The steps to update OCI trust policy configuration:

1. Export trust policy configuration into a JSON file.

   ```shell
   notation policy show > ./oci_trust_policy.json
   ```

2. Edit the exported JSON file "oci_trust_policy.json", update trust policy configuration and save the file.
3. Import trust policy configuration from the file.

   ```shell
   notation policy import ./oci_trust_policy.json
   ```
