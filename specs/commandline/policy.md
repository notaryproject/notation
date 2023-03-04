# notation policy

## Description

As part of signature verification workflow, users need to configure the trust policies to specify trusted identities that signed the artifacts, and the level of signature verification to use. For more details, see [trust policy specification and examples](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy).

The `notation policy` command provides a user-friendly way to manage trust policies. It allows users to import and export trust policies from/to a JSON file. Users can export a template file with trust policy configuration to start from scratch.

An example of template file:

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            // Policy for set of artifacts signed by Wabbit Networks
            // that are pulled from ACME Rockets repository
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
              "ca:wabbit-networks-ca2"
            ],
            "trustedIdentities": [
                "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Security Tools"
            ]
        },
        {
            // Exception policy for a single unsigned artifact pulled from
            // Wabbit Networks repository
            "name": "unsigned-image",
            "registryScopes": [ "registry.wabbit-networks.io/software/unsigned/net-utils" ],
            "signatureVerification": {
              "level" : "skip" 
            }
        },
        {
            // Policy that uses custom verification level to relax the strict verification.
            // It logs expiry and skips revocation check for a specific artifact.
            "name": "allow-expired-images",
            "registryScopes": [ "registry.acme-rockets.io/software/legacy/metrics" ],
            "signatureVerification": {
              "level" : "strict",
              "override" : {
                "expiry" : "log",
                "revocation" : "skip"
              }
            },
            "trustStores": ["ca:acme-rockets"],
            "trustedIdentities": ["*"]
        },
        {
            // Policy for all other artifacts signed by ACME Rockets
            "name": "global-policy-for-all-other-images",
            "registryScopes": [ "*" ],       
            "signatureVerification": {                                
                "level": "strict"
            },
            "trustStores": [ 
              "ca:acme-rockets-others"
            ],                  
            "trustedIdentities": [                                    
                "x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io, OU=Finance, CN=SecureBuilder",
                "x509.subject: C=US, ST=WA, L=Seattle, O=acme-rockets.io, OU=Hr, CN=SecureBuilder"
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
  -h, --help      help for export
      --template  export a template of trust policies
  -v, --verbose   verbose mode
```

### notation policy import

```text
Import trust policies from a JSON file

Usage:
  notation policy import [flags] <file_path>

Flags:
  -d, --debug     debug mode
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

### Export a template file of trust policies

The template file is for users to create trust policies from scratch. Users should update the trust policies according to own requirements before importing the template file.

```shell
notation policy export ./trustpolicy_template.json
```

### Export trust policies into a JSON file

```shell
notation policy export ./policy_exported.json
```

Upon successful execution, the trust policies are exported into a json file.

### Show trust policies

```shell
notation policy show
```

Upon successful execution, the trust policies are printed out. If trust policies are not configured, users should receive an error message.
