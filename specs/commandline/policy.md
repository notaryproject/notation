# notation policy

## Description

As part of signature verification workflow of signed OCI artifacts or blobs, users need to configure the trust policy configuration file to specify trusted identities that signed the artifacts, the level of signature verification to use and other settings. For more details, see [trust policy specification and examples](https://github.com/notaryproject/specifications/blob/main/specs/trust-store-trust-policy.md#trust-policy).

The `notation policy` command provides a user-friendly way to manage trust policies. It allows users to show trust policy configuration, import/export a trust policy configuration file from/to a JSON file. To get started, user can refer to the following trust policy configuration sample. In this sample, there are four policies configured for different requirements:

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

Policy language version 1.1 added support for verifying signatures associated with blob artifacts. User can use `scopes` field as a Policy selector string to decide which Policy gets applied to which blob. The `--policy-scope` argument provided in `notation blob verify` command will dictate which Policy gets picked from the policy configuration file and applied for verification. To get started with verifying blob signatures, users can refer to the following trust policy configuration sample. In this sample, there are three policies configured for different requirements:

- The Policy named "blob-verification-policy" is for verifying blob artifacts signed by Wabbit Networks and scoped to `blob-verification-selector`.
- Policy named "skip-blob-verification-policy" is for skipping verification on blob artifacts scoped to `skip-blob-verification-selector`.
- Policy "wildcard-blob-verification-policy" is for auditing verification results when user wants to apply a wildcard policy by not providing `--policy-scope` argument in `notation blob verify` command.

```jsonc
{
    "version": "1.1",
    "trustPolicies": [
        {
            "name": "blob-verification-policy",
            "scopes": [ 
              "blob:blob-verification-selector"
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
            "name": "skip-blob-verification-policy",
            "scopes": [ "blob:skip-blob-verification-selector" ],
            "signatureVerification": {
              "level" : "skip" 
            }
        },
        {
            "name": "wildcard-blob-verification-policy",
            "scopes": [ "blob:*" ],
            "signatureVerification": {
              "level" : "audit"
            },
            "trustStores": ["ca:acme-rockets"],
            "trustedIdentities": ["*"]
        }
    ]
}
```

Note: Policy language version 1.1 renamed the field `registryScopes` from version 1.0 to `scopes`. The new field accepts values with prefixes `oci` or `blob` to limit a scope value to either OCI signature verification or Blob signature verification. While scope values with `blob` prefix can be of free-form text, values with `oci` prefix must be valid OCI references. `notation` supports both policy language versions 1.0 and 1.1. However, `notation` rejects policy configuration files with mixed terminology i.e. both `registryScopes` and `scopes` defined in a single configuration file. Users migrating from 1.0 to 1.1 can simply rename `registryScopes` to `scopes` and prefix the values with `oci`.
Below is a sample Policy configuration file that verifies OCI artifacts using `scopes` field.

```jsonc
{
    "version": "1.1",
    "trustPolicies": [
        {
            "name": "wabbit-networks-images",
            "scopes": [ 
              "oci:registry.acme-rockets.io/software/net-monitor",
              "oci:registry.acme-rockets.io/software/net-logger" 
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
            "scopes": [ "oci:registry.acme-rockets.io/software/unsigned/net-utils" ],
            "signatureVerification": {
              "level" : "skip" 
            }
        },
        {
            "name": "allow-expired-images",
            "scopes": [ "oci:registry.acme-rockets.io/software/legacy/metrics" ],
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
            "scopes": [ "oci:*" ],       
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
Manage trust policy configuration for signature verification.

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
Import trust policy configuration from a JSON file

Usage:
  notation policy import [flags] <file_path>

Flags:
      --force     override the existing trust policy configuration, never prompt
  -h, --help      help for import
```

### notation policy show

```text
Show trust policy configuration

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

The trust policy configuration in the JSON file should be validated according to [trust policy properties](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy-properties). A successful message should be printed out if trust policy configuration are imported successfully. Error logs including the reason should be printed out if the importing fails.

If there is an existing trust policy configuration, prompt for users to confirm whether discarding existing configuration or not. Users can use `--force` flag to discard existing trust policy configuration without prompt.

### Show trust policies

Use the following command to show trust policy configuration:

```shell
notation policy show
```

Upon successful execution, the trust policy configuration are printed out to standard output. If trust policy is not configured or is malformed, users should receive an error message via standard error output, and a tip to import trust policy configuration from a JSON file.

### Export trust policy configuration into a JSON file

Users can redirect the output of command `notation policy show` to a JSON file.

```shell
notation policy show > ./trust_policy.json
```

### Update trust policy configuration

The steps to update trust policy configuration:

1. Export trust policy configuration into a JSON file.

   ```shell
   notation policy show > ./trust_policy.json
   ```

2. Edit the exported JSON file "trust_policy.json", update trust policy configuration and save the file.
3. Import trust policy configuration from the file.

   ```shell
   notation policy import ./trust_policy.json
   ```
