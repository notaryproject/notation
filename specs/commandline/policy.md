# notation policy

## Description

Use `notation policy` command to add/update/show/delete trust policies. As part of signature verification workflow, user needs to configure the trust policies to specify trusted identities that sign the artifacts, and the level of signature verification to use. For more details, see [trust policy spec](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#trust-policy).

An example of `trustpolicy.json`:

```json
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

## Outline

### notation policy command

```text
Manage trust policies for signature verification.

Usage:
  notation policy [command]

Available Commands:
  add           Add trust policies
  delete        Delete trust policies
  list          List trust policies
  update        Update trust policies

Flags:
  -h, --help   help for policy
```

### notation policy add

```text
Add trust policies.

Usage:
  notation policy add (--file <policy_file_path> | --name <policy_name>) [flags]

Flags:
      --custom-level       stringArray   {key}={value} pairs that represent a custom level to override existing verification level
  -f  --file               string        path to a trust policy file in JSON format
  -n  --name               string        name of the trust policy statement
  -h, --help                             help for add
      --scope              stringArray   repository URIs to which the policy applies (default ["*"])
      --trust-store        stringArray   trust stores in format "<trust_store_type>:<trust_store_name>"
      --verification-level string        verification level, options: "strict", "permissive", "audit", "skip" (default "strict")
      --x509-cert          stringArray   paths to x509 certificate file that certificate subject is retrieved from
      --x509-id            stringArray   trust identities, user trusted x509 certificate subjects
```

### notation policy delete

```text
Delete trust policies. User cannot delete all the trust policies, at least one trust policy should be configured for signature verification.

Usage:
  notation policy delete [flags] <policy_name>...

Flags:
  -h, --help   help for delete
```

### notation policy list

```text
List trust policies by names

Usage:
  notation policy list [flags]

Aliases:
  list, ls

Flags:
      --details                 list the details of trust policies
  -h, --help                    help for list
  -n, --name        string      name of the trust policy statement
      --reference   string      reference to the artifact
      --scope       string      repository URI, e.g. localhost:5000/namespace/repo_name
```

### notation policy update

```text
Update the existing trust policies.

Usage:
  notation policy update (--file <policy_file_path> | --name <policy_name>) [flags]

Flags:
      --custom-level       stringArray   {key}={value} pairs that represent a custom level to override existing verification level
  -f  --file               string        path to a trust policy file in JSON format
  -n  --name               string        name of the trust policy statement
  -h, --help               help for add
      --scope              stringArray   repository URIs to which the policy applies (default ["*"])
      --trust-store        stringArray   trust stores in format "<trust_store_type>:<trust_store_name>"
      --verification-level string        verification level, options: "strict", "permissive", "audit", "skip" (default "strict")
      --x509-cert          stringArray   paths to x509 certificate file that certificate subject is retrieved from
      --x509-id            stringArray   trust identities, user trusted x509 certificate subjects
```

## Usage

### Add a trust policy by configuring the properties from command line

The following table shows how to configure properties for a trust policy and default values.

| Property Name       | Necessity | flag to configure                    | Default value |
| ------------------- | --------- | ------------------------------------ | ------------- |
| `name`              | Required  | `--name`                             | N/A           |
| `trustStores`       | Required  | `--trust-store`                      | N/A           |
| `registryScopes`    | Optional  | `--scope`                            | `["*"]`       |
| `level`             | Optional  | `--verification-level`               | `"strict"`    |
| `override`          | Optional  | `--custom-level`                     | nil           |
| `trustedIdentities` | Optional  | `--x509-id` or `--x509-cert` or both | `["*"]`       |

```shell  
notation policy add --name wabbit-network-dev --scope "dev.wabbitnetworks.io/net-monitor" --trust-store "ca:wabbit-network-dev" --x509-id "C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io" --verification-level "strict"
```

The execution of `add` fails in any of below cases:

- The trust policy name exists.
- More than one trust policy that uses a global scope, that is, the value of `registryScopes` is `["*"]`.
- The values of `--x509-id` or `--x509cert` overlap. For example, the following two identity values are overlapping:
  - "C=US, ST=WA, O=wabbit-network.io, OU=org1"
  - "C=US, ST=WA, O=wabbit-network.io"

Upon successful execution, the added trust policy is printed out. For example:

In text format

```text
name: "wabbit-networks-dev"
registryScopes: ["dev.wabbitnetworks.io/net-monitor"]
level: "strict"
trustStores: ["ca:wabbit-networks-dev"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"]
```

In json format

```json
{
  "name": "wabbit-networks-dev",                              
  "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],                                  
  "signatureVerification": {                                  
      "level": "strict"
  },
  "trustStores": [ "ca:wabbit-networks-dev" ],                
  "trustedIdentities": [                                      
      "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
  ]
}
```

Users can also use `notation policy list` to confirm the trust policies are added.

### Add a trust policy with default values

Based on the above table, users can leveraging default values when adding a trust policy. Note that the default value of  `registrySopes` is `["*"]` called global scope, which means this policy applies to all the artifacts. There can only be one trust policy that uses a global scope.

```shell
notation policy add --name wabbit-network-dev --trust-store "ca:wabbit-network-dev"
```

Upon successful execution, the added trust policy is printed out. For example:

In text format

```text
name: "wabbit-networks-dev"
registryScopes: ["*"]
level: "strict"
trustStores: ["ca:wabbit-networks-dev"]
trustedIdentities: ["*"]
```

In json format

```json
{
  "name": "wabbit-networks-dev",                              
  "registryScopes": [ "*" ],                                  
  "signatureVerification": {                                  
      "level": "strict"
  },
  "trustStores": [ "ca:wabbit-networks-dev" ],                
  "trustedIdentities": [                                      
      "*"
  ]
}
```

Users can also use `notation policy list` to confirm the trust policies are added.

### Add trust policies from a JSON file

Users can add trust policies from a JSON file. This is specially useful when add multiple trust polices.

Create a JSON file that includes the trust polices. For example, a file named `mytrustpolicy.json` with the following content:

```json
{
    "version": "1.0",
    "trustPolicies": [     
        // First policy                                             
        {
            "name": "wabbit-networks-dev",                              
            "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],  
            "signatureVerification": {                                  
                "level": "strict"
                "override" : {
                     "expiry" : "log",
                     "authenticity": "log"
                }
            },
            "trustStores": [ "ca:wabbit-networks-dev" ],
            "trustedIdentities": [
                "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
            ]
        },
        // Second policy
        {},
        // Third policy
        {}
    ]
}
```

Then run the following command to add trust policies

```shell
notation policy add --file mytrustpolicy.json
```

The execution fails in one of below cases:

- Version doesn't support. Currently only `1.0` version is supported.
- The trust policy name exists.
- More than one trust policy that uses a global scope, that is, the value of `registryScopes` is `["*"]`.
- The values of `trustIdentities` overlap. For example, the following two identity values are overlapping:
  - "C=US, ST=WA, O=wabbit-network.io, OU=org1"
  - "C=US, ST=WA, O=wabbit-network.io"

Upon successful execution, the added trust policy is printed out. Users can also use `notation policy list` to confirm the trust policies are added.

### Add a trust policy by using certificate files for trust identities

If users specify the certificate files for `trustedIdentities` property, notation retrieves the subjects from the certificates.

```shell
notation policy add --name wabbit-network-dev --trust-store "ca:wabbit-network-dev" --x509-cert "./wabbit-network-dev.crt"
```

Upon successful execution, the added trust policy is printed out. For example:

In text format

```text
name: "wabbit-networks-dev"
registryScopes: ["*"]
level: "strict"
trustStores: ["ca:wabbit-networks-dev"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"]
```

In json format

```json
{
  "name": "wabbit-networks-dev",                              
  "registryScopes": [ "*" ],                                  
  "signatureVerification": {                                  
      "level": "strict"
  },
  "trustStores": [ "ca:wabbit-networks-dev" ],                
  "trustedIdentities": [                                      
      "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
  ]
}
```

### Add a trust policy with custom verification level

Users can override the default verification level by using flag `--custom-level`.

```shell
# customize the verification level based on the default verification level `strict`
notation policy add --name wabbit-network-dev --trust-store "ca:wabbit-network-dev" --custom-level "expiry=log,authenticity=log"
```

Upon successful execution, the added trust policy is printed out. For example:

In text format

```text
name: "wabbit-networks-dev"
registryScopes: ["*"]
level: "strict"
override: ["expiry=log", "authenticity=log"]
trustStores: ["ca:wabbit-networks-dev"]
trustedIdentities: ["*"]
```

In json format

```json
{
  "name": "wabbit-networks-dev",                              
  "registryScopes": [ "*" ],                                  
  "signatureVerification": {                                  
      "level": "strict"
      "override": {
          "expiry" : "log",
          "authenticity": "log"
      }
  },
  "trustStores": [ "ca:wabbit-networks-dev" ],                
  "trustedIdentities": [                                      
      "*"
  ]
}
```

### Update a trust policy by updating the properties from command line

`notation policy update` command shares the same flags with `notation policy add` command to update the properties for a trust policy.

```shell
notation policy update --scope "dev-2.wabbitnetworks.io/net-monitor" --trust-store "ca:wabbit-network-dev-2" --x509-id "C=US, ST=WA, L=Seattle, O=Example, OU=Dev-2, CN=wabbit-networks.io" --verification-level "permissive" wabbit-network-dev
```

The execution of `update` fails in one of below cases:

- The policy name doesn't exist.
- More than one trust policy that uses a global scope, that is, the value of `registryScopes` is `["*"]`.
- The values of `--x509-id` or `--x509cert` overlap. For example, the following two identity values are overlapping:
  - "C=US, ST=WA, O=wabbit-network.io, OU=org1"
  - "C=US, ST=WA, O=wabbit-network.io"

Upon successful execution, the updated trust policy is printed out. For example:

In text format

```text
name: "wabbit-networks-dev"
registryScopes: ["dev-2.wabbitnetworks.io/net-monitor"]
level: "permissive"
trustStores: ["ca:wabbit-networks-dev-2"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev-2, CN=wabbit-networks.io"]
```

In json format

```json
{
  "name": "wabbit-networks-dev",                              
  "registryScopes": [ "dev-2.wabbitnetworks.io/net-monitor" ],                                  
  "signatureVerification": {                                  
      "level": "permissive"
  },
  "trustStores": [ "ca:wabbit-networks-dev-2" ],                
  "trustedIdentities": [                                      
      "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev-2, CN=wabbit-networks.io"
  ]
}
```

### Update trust policies from a JSON file

Users can update trust policies from a JSON file. This is specially useful when update multiple trust polices.

Create a JSON file that includes the trust polices with updated values. Then run the following command to update trust policies.

```shell
notation policy update --file mytrustpolicy.json
```

The execution fails in one of below cases:

- Version doesn't support. Currently only `1.0` version is supported.
- The trust policy name doesn't exists.
- More than one trust policy that uses a global scope, that is, the value of `registryScopes` is `["*"]`.
- The values of `trustIdentities` overlap. For example, the following two identity values are overlapping:
  - "C=US, ST=WA, O=wabbit-network.io, OU=org1"
  - "C=US, ST=WA, O=wabbit-network.io"

Upon successful execution, the updated trust policies are printed out. Users can also use `notation policy list` to confirm the trust policies are added.

### List all the trust policies by names

The output is a list of trust policy names.

```shell
notation policy list
```

An example of output messages:

in text format:

```text
name: wabbit-network-dev
registryScopes: ["dev.wabbitnetworks.io/net-monitor"]

name: wabbit-network-prod
registryScopes: ["prod.wabbitnetworks.io/net-monitor"]
```

in JSON format:

```json
{
    "result" : true,
    "trustPolicies": [
        {
            "name": "wabbit-networks-dev",
            "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],
        },
        {
            "name": "wabbit-networks-prod",
            "registryScopes": [ "prod.wabbitnetworks.io/net-monitor" ],
        }
    ]
}
```

### List all the trust policies with details

```shell
notation policy list --details
```

An example of output messages:

In text format:

```text
version: "1.0"

name: "wabbit-networks-dev"
registryScopes: ["dev.wabbitnetworks.io/net-monitor"]
level: "strict"
override: ["expiry=log", "authenticity=log"]
trustStores: ["ca:wabbit-networks-dev"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"]

name: "wabbit-networks-prod"
registryScopes: ["prod.wabbitnetworks.io/net-monitor"]
level: "permissive"
trustStores: ["ca:wabbit-networks-prod"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"]
```

In JSON format:

```json
{
    "version": "1.0",                                                  
    "trustPolicies": [
        {
            "name": "wabbit-networks-dev",                              
            "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],
            "signatureVerification": {
                "level": "strict"
                "override" : {
                     "expiry" : "log",
                     "authenticity": "log"
                }
            },
            "trustStores": [ "ca:wabbit-networks-dev" ],
            "trustedIdentities": [
                "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
            ]
        },
        {
            "name": "wabbit-networks-prod",
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

### List details of one trust policy

```shell
notation policy list --name "wabbit-networks-prod"
```

An example of output messages:

In text format:

```text
name: "wabbit-networks-prod"
registryScopes: ["prod.wabbitnetworks.io/net-monitor"]
level: "permissive"
trustStores: ["ca:wabbit-networks-prod"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"]
```

In JSON format:

```json
{
    "name": "wabbit-networks-prod",
    "registryScopes": [ "prod.wabbitnetworks.io/net-monitor" ],       
    "signatureVerification": {                                
        "level": "permissive"
    },
    "trustStores": [ "ca:wabbit-networks-prod" ],                  
    "trustedIdentities": [                                    
        "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"
    ]
}
```

### List trust policies for verifying specified artifact

```shell
notation policy list --reference prod.wabbitnetworks.io/net-monitor@sha256:xxx
```

An example of output messages:

In text format:

```text
name: "wabbit-networks-prod"
registryScopes: ["prod.wabbitnetworks.io/net-monitor"]
level: "permissive"
trustStores: ["ca:wabbit-networks-prod"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"]
```

In JSON format:

```json
{
    "name": "wabbit-networks-prod",
    "registryScopes": [ "prod.wabbitnetworks.io/net-monitor" ],       
    "signatureVerification": {                                
        "level": "permissive"
    },
    "trustStores": [ "ca:wabbit-networks-prod" ],                  
    "trustedIdentities": [                                    
        "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"
    ]
}
```

Note users may configure only one trust policy with `registryScopes` value as `["*"]`. Here is an example of output:

In text format:

```text
name: "wabbit-networks-prod"
registryScopes: ["*"]
level: "strict"
trustStores: ["ca:wabbit-networks-prod"]
trustedIdentities: ["*"]
```

In JSON format:

```json
{
    "name": "wabbit-networks-prod",
    "registryScopes": [ "*" ],       
    "signatureVerification": {                                
        "level": "strict"
    },
    "trustStores": [ "ca:wabbit-networks-prod" ],                  
    "trustedIdentities": [                                    
        "*"
    ]
}
```

### List trust policies for verifying artifacts in specified repository

```shell
notation policy list --scope prod.wabbitnetworks.io/net-monitor
```

An example of output messages:

In text format:

```text
name: "wabbit-networks-prod"
registryScopes: ["prod.wabbitnetworks.io/net-monitor"]
level: "permissive"
trustStores: ["ca:wabbit-networks-prod"]
trustedIdentities: ["x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"]
```

In JSON format:

```json
{
    "name": "wabbit-networks-prod",
    "registryScopes": [ "prod.wabbitnetworks.io/net-monitor" ],       
    "signatureVerification": {                                
        "level": "permissive"
    },
    "trustStores": [ "ca:wabbit-networks-prod" ],                  
    "trustedIdentities": [                                    
        "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"
    ]
}
```

### Delete trust policies

Users cannot delete all the trust policies, at least one trust policy should be kept for signature verification. Deletion MUST fail if users intend to delete all the trust policies.

```shell
# Delete one trust policy
notation policy delete wabbit-network-dev

# Delete multiple trust policies
notation policy delete wabbit-network-dev wabbit-network-prod
```

An example of output messages:

In text format:

```text
Successfully deleted the following trust policy(s):

name: "wabbit-networks-dev"
registryScopes: ["dev.wabbitnetworks.io/net-monitor"]

name: "wabbit-networks-prod"
registryScopes: ["prod.wabbitnetworks.io/net-monitor"]
```

In JSON format:

```json
{
    "result" : true,
    "trustPolicies": [
        {
            "name": "wabbit-networks-dev",
            "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],
        },
        {
            "name": "wabbit-networks-prod",
            "registryScopes": [ "prod.wabbitnetworks.io/net-monitor" ],
        }
    ]
}
```
