# notation policy

## Description

As part of signature verification workflow of signed OCI artifacts, users need to configure trust policy configuration file to specify trusted identities that signed the artifacts, the level of signature verification to use and other settings. For more details, see [OCI trust policy specification and examples](https://github.com/notaryproject/specifications/blob/main/specs/trust-store-trust-policy.md#oci-trust-policy).

The `notation policy` command provides a user-friendly way to manage trust policies for signed OCI images. It allows users to show trust policy configuration, import/export a trust policy configuration file from/to a JSON file. Users who want to manage trust policies for signed arbitrary blobs, please refer to `notation blob policy` command.

To get started, user can refer to the following trust policy configuration sample `trustpolicy.json` that is applicable for verifying signed OCI artifacts using `notation verify` command. In this sample, there are four policies configured for different requirements:

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
Manage trust policy configuration for OCI artifact signature verification.

Usage:
  notation policy [command]

Available Commands:
  import    import trust policy configuration from a JSON file
  show      show trust policy configuration

Flags:
  -h, --help   help for policy
```

### notation policy import

```text
Import OCI trust policy configuration from a JSON file

Usage:
  notation policy import [flags] <file_path>

Flags:
      --force     override the existing trust policy configuration, never prompt
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
