# notation policy

## Description

As part of signature verification workflow, users need to configure the trust policies to specify trusted identities that signed the artifacts, and the level of signature verification to use. For more details, see [trust policy specification and examples](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy).

The `notation policy` command provides a user-friendly way to manage trust policies. It allows users to import and export trust policies from/to a JSON file. To get started user can use following sample trust policy. In this sample, there are four policies configured for different requirements:

- The Policy named "wabbit-networks-images" is for verifying images signed by Wabbit Networks and stored in two repositories `registry.acme-rockets.io/software/net-monitor` and `registry.acme-rockets.io/software/net-logger`.

- Policy named "unsigned-image" is for skipping the verification on unsigned images stored in repository `registry.acme-rockets.io/software/unsigned/net-utils`.
- Policy "allow-expired-images" is for logging instead of failing expired images stored in repository `registry.acme-rockets.io/software/legacy/metrics`.
- Policy "global-policy-for-all-other-images" is for verifying any other images that signed by the ACME Rockets.
  
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
Manage trust policies for signature verification.

Usage:
  notation policy [command]

Available Commands:
  export    export trust policies to a JSON file
  import    import trust policies from a JSON file
  show      show trust policies

Flags:
  -h, --help   help for policy
```

### notation policy export

```text
Export trust policies to a JSON file

Usage:
  notation policy export [flags] <file_path>

Flags:
  -d, --debug     debug mode
      --force     override the supplied file, never prompt
  -h, --help      help for export
  -v, --verbose   verbose mode
```

### notation policy import

```text
Import trust policies from a JSON file

Usage:
  notation policy import [flags] <file_path>

Flags:
  -d, --debug     debug mode
      --force     override the existing policies, never prompt
  -h, --help      help for import
  -v, --verbose   verbose mode
```

### notation policy show

```text
Show trust policies

Usage:
  notation policy [flags] show

Flags:
  -d, --debug     debug mode
  -h, --help      help for show
  -v, --verbose   verbose mode
```

## Usage

### Import trust policies from a JSON file

```shell  
notation policy import ./my_policy.json
```

The trust policies in the JSON file will be validated according to [trust policy properties](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy-properties). A successful message should be printed out if trust policies are imported successfully. Error logs including the reason should be printed out if the importing fails.

Use `--force` flag to override existing policies without prompt.

### Export existing trust policies into a JSON file

```shell
notation policy export ./policy_exported.json
```

Upon successful execution, the existing trust policies are exported into a json file. 

Use `--force` flag to override supplied file without prompt.

### Show trust policies

```shell
notation policy show
```

Upon successful execution, the trust policies are printed out. If trust policies are not configured, users should receive an error message.
