# notation policy

## Description

As part of signature verification workflow, users need to configure the trust policies to specify trusted identities that signed the artifacts, and the level of signature verification to use. For more details, see [trust policy specification and examples](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy).

The `notation policy` command provides a user-friendly way to manage trust policies. It allows users to initialize trust policies with default values, import trust policies from a JSON file, and show trust policies. To get started user can use following sample trust policy. In this sample, there are four policies configured for different requirements:

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
  import    import trust policies from a JSON file
  init      initialize trust policies with default values
  show      show trust policies

Flags:
  -h, --help   help for policy
```

### notation policy import

```text
Import trust policies from a JSON file

Usage:
  notation policy import [flags] <file_path>

Flags:
  -d, --debug     debug mode
      --force     override the existing trust policies, never prompt
  -h, --help      help for import
  -v, --verbose   verbose mode
```

### notation policy init

```text
Initialize trust policies with default values

Usage:
  notation policy init [flags]

Flags:
  -d, --debug                      debug mode
      --force                      restore the trust policies to default values. Any existing trust polices will be discarded, never prompt
  -h, --help                       help for export
      --trust-store stringArray    specify the trust stores in format "<type>:<name>", e.g. "ca:my_store". If this flag is ignored, a default trust store "ca:default" is used
  -v, --verbose                    verbose mode
```

### notation policy show

```text
Show trust policies

Usage:
  notation policy show [flags]

Flags:
  -d, --debug     debug mode
  -h, --help      help for show
  -v, --verbose   verbose mode
```

## Usage

### Initialize trust policies with default values

```shell
notation policy init
```

Upon successful execution, trust policies with default values are created and printed out as following. Use command `notation cert add --type ca --store default <cert_file>` to add CA certificates to trust store `ca:default`. If users want to use different trust stores, refer to section [Update trust policies](#update-trust-policies) on how to do the update.

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            "name": "policy-by-init-command",
            "registryScopes": ["*"],
            "signatureVerification": {
                "level": "strict"
            },
            "trustStores": ["ca:default"],
            "trustedIdentities": ["*"]
        }
    ]
}

```

If there are existing trust policies configured and users still run `notation policy init` command, A prompt should be displayed asking for confirmation on whether restoring to default values or not. Use `--force` flag to discard any existing trust policies without prompt.

### Initialize trust policies with specified trust stores

```shell
notation policy init --ts "ca:my_store" --ts "ca:my_store_2"
```

Upon successful execution, trust policies with default values and specified trust stores are created and printed out as following:

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            "name": "policy-by-init-command",
            "registryScopes": ["*"],
            "signatureVerification": {
                "level": "strict"
            },
            "trustStores": ["ca:my_store", "ca:my_store_2"],
            "trustedIdentities": ["*"]
        }
    ]
}
```

Use `--force` flag to override existing policies without prompt.

### Show trust policies

```shell
notation policy show
```

Upon successful execution, the trust policies are printed out. If trust policies are not configured, users should receive an error message, and a tip to initialize trust policies using command `notation policy init`.

### Export trust policies into a JSON file

Users can redirect the output of command `notation policy show` to a JSON file.

```shell
notation policy show > ./trust_policy.json
```

### Import trust policies from a JSON file

```shell  
notation policy import ./my_policy.json
```

The trust policies in the JSON file will be validated according to [trust policy properties](https://github.com/notaryproject/notaryproject/blob/v1.0.0-rc.2/specs/trust-store-trust-policy.md#trust-policy-properties). A successful message should be printed out if trust policies are imported successfully. Error logs including the reason should be printed out if the importing fails.

Use `--force` flag to override existing policies without prompt.

### Update trust policies

The steps to update trust policies:

1. Export trust policies into a JSON file.

   ```shell
   notation policy show > ./trust_policy.json
   ```

2. Edit the exported JSON file "trust_policy.json", update trust policies and save the file.
3. Import trust policies from the file.

   ```shell
   notation policy import ./trust_policy.json
   ```
