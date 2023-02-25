# notation policy

## Description

As part of signature verification workflow, users need to configure the trust policies to specify trusted identities that sign the artifacts, and the level of signature verification to use. For more details, see [trust policy spec](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy).

An example of trust policy configuration file:

```jsonc
{
    "version": "1.0",                                                   // version info
    "trustPolicies": [                                                  // list of trust policy statements
        {
            "name": "wabbit-networks-dev",                              // Name of the first policy statement
            "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],  // The registry artifacts to which the policy applies
            "signatureVerification": {                                  // The level of verification - strict, permissive, audit, skip
                "level": "strict"
                "override" : {
                     "expiry" : "log",
                     "authenticity": "log"
                }
            },
            "trustStores": [ "ca:wabbit-networks-dev" ],                // The trust stores that contains the X.509 certificates
            "trustedIdentities": [                                      // Identities that are trusted to sign the artifact.
                "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
            ]
        },
        {
            "name": "wabbit-networks-prod",                             // Name of the second policy statement.
            "registryScopes": [ "prod.wabbitnetworks.io/net-monitor" ],       
            "signatureVerification": {                                
                "level": "permissive"
            },
            "trustStores": [ "ca:wabbit-networks-prod" ],                  
            "trustedIdentities": [                                    
                "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"
            ]
        }
    ]
}
```

The goal of `notation policy` command is to provide a user-friendly CLI for users to manage trust policies without the knowledge of trust policy configuration name, directory path and property names. Two high level use cases as follows:

1. Users can import trust policies from a JSON file, and export trust policies into a JSON file.
2. Users can add/view/update/delete trust policies and properties without editing the trust policy configuration file.

A phased approach is adopted to achieve the goal. The first use case will be implemented in phase-1. A trust policy template file will be provided for users to get started. This specification only covers phase-1.

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
  -h, --help    help for export
```

### notation policy import

```text
Import trust policies from a JSON file

Usage:
  notation policy import [flags] <file_path>

Flags:
  -h, --help    help for import
```

### notation policy show

```text
Show trust policies

Usage:
  notation policy [flags] show

Flags:
  -h, --help    help for show
```

## Usage

### Import trust policies from a JSON file

```shell  
notation policy import ./my_policy.json
```

The trust policies in the JSON file will be validated according to [trust policy properties](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy-properties). Upon successful import, the trust policies are printed out.

### Export trust policies into a JSON file

```shell
notation policy export ./policy_exported.json
```

For phase-1, to update trust policies, users need to export the trust policies to a file first, update the file, and import the file again.

### Show trust policies

```shell
notation policy show
```

Upon successful execution, the trust policies are printed out. If trust policies are not configured, users should receive a warning message.
