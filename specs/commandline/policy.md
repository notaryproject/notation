# notation policy

## Description

Use `notation policy` command to add/update/show/delete trust policies. As part of signature verification workflow, user needs to configure the trust policies to specify trusted identities that sign the artifacts, and the level of signature verification to use. For more details, see [trust policy spec](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#trust-policy).

An example of `trustpolicy.json`:

```jsonc
{
    "version": "1.0",                                                   // version info
    "trustPolicies": [
        {
            "name": "wabbit-networks-dev",                              // Name of the 1st policy
            "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],  // The repository list that policy applies to
            "signatureVerification": {                                  // The level of verification - strict, permissive, audit, skip
                "level": "strict"
                "override" : {
                     "expiry" : "log"
                }
            },
            "trustStores": [ "ca:wabbit-networks-dev" ],                // The trust stores that contains the X.509 certificates
            "trustedIdentities": [                                      // Identities that are trusted to sign the artifact.
                "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
            ]
        },
        {
            "name": "wabbit-networks-prod",                             // Name of the 2nd policy.
            "registryScopes": [ "prod.wabbitnetworks.io/net-monitor" ],       
            "signatureVerification": {                                
                "level": "strict"
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
  add           Add trust policies.
  delete        Delete trust policies.
  list          List trust policies.
  update        Update trust policies

Flags:
  -h, --help   help for policy
```

### notation policy add

```text
Add trust policies.

Usage:
  notation policy add [flags] <policy_name>

Flags:
      --cert-file          string   optional certificate filepath
      --custom-level       string   optional custom level to based existing verification level, e.g. "authenticity=log,expiry=log"
  -h, --help                        help for add
      --id                 string   optional trust identity aka certificate subject info
      --scope              string   optional repository URI (default "*")
      --trust-store        string   required trust store in format "<trust_store_type>:<trust_store_name>", e.g. "ca:my_trust_store"
      --verification-level string   optional verification level, options: "strict", "permissive", "audit", "skip" (default "strict")
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
      --details                  optional list the details of trust policies
  -h, --help                     help for list
      --id          string       optional trust identity aka certificate subject info
      --name        string       optional trust policy name
      --reference   string       optional reference to the artifact
      --repo        string       optional repository URI, e.g. localhost:5000/namespace/repo_name
      --trust-store string       optional trust store in format "<trust_store_type>:<trust_store_name>", e.g. "ca:my_trust_store"
```

### notation policy update

```text
Update the existing trust policies.

Usage:
  notation policy update [flags] <policy_name>

Flags:
      --cert-file          string   optional certificate filepath
      --custom-level       string   optional custom level to based existing verification level, e.g. "authenticity=log,expiry=log"
  -h, --help                        help for add
      --id                 string   optional trust identity aka certificate subject info
      --scope              string   optional repository URI
      --trust-store        string   optional trust store in format "<trust_store_type>:<trust_store_name>", e.g. "ca:my_trust_store"
      --verification-level string   optional verification level, options: strict, permissive, audit, skip
```

## Usage

### Add a trust policy with minimal configurations

```shell  
notation policy add --trust-store ca:wabbit-network-dev wabbit-network-dev
```

The execution of `add` fails in one of below cases:

- The policy name exists.
- There is already a policy with registry scope configured with value `*`.

Upon successful execution, the added trust policy is printed out. For example:

In text format

```text
"name": "wabbit-networks-dev"
"registryScopes": "*"
"level": "strict"
"trustStores": "ca:wabbit-networks-dev"
"trustedIdentities": "*"
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

### Add a trust policy by configuring all the properties

```shell
  notation policy add --scope dev.wabbitnetworks.io/net-monitor --trust-store ca:wabbit-network-dev --id "C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io" --verification-level "audit" wabbit-network-dev
```

Upon successful execution, the added trust policy is printed out. For example:

In text format

```text
"name": "wabbit-networks-dev"
"registryScopes": "dev.wabbitnetworks.io/net-monitor"
"level": "audit"
"trustStores": "ca:wabbit-networks-dev"
"trustedIdentities": "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
```

In json format

```json
{
  "name": "wabbit-networks-dev",                              
  "registryScopes": [ "dev.wabbitnetworks.io/net-monitor" ],                                  
  "signatureVerification": {                                  
      "level": "audit"
  },
  "trustStores": [ "ca:wabbit-networks-dev" ],                
  "trustedIdentities": [                                      
      "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
  ]
}
```

### Add a trust policy by using certificate files for identities

In this case, users specify the certificate files, notation retrieves the subject info from the certificates.

```shell
  notation policy add --scope dev.wabbitnetworks.io/net-monitor --trust-store ca:wabbit-network-dev --cert-file wabbit-network-dev.crt wabbit-network-dev
```

In text format

```text
"name": "wabbit-networks-dev"
"registryScopes": "dev.wabbitnetworks.io/net-monitor"
"level": "strict"
"trustStores": "ca:wabbit-networks-dev"
"trustedIdentities": "x509.subject: C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io"
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

### Add a trust policy with verification level

### Add a trust policy with custom level

### Update the registry scopes for a trust policy

```shell
notation policy update --name wabbit-network-build --scope localhost:5000/build/net-monitor --scope localhost:5000/build/nginx
```

### update the verification level for a trust policy

```shell
notation policy update --name wabbit-network-build --verification-level "audit" --override ""
```

### Update the trust stores for a trust policy

```shell
notation policy update --name wabbit-network-build --ts ca:wabbit-network-dev --ts ca:wabbit-network-prod
```

### Update the trust identities for a trust policy by setting specified x509 subjects

```shell
notation policy update --name wabbit-network-build --id "C=US, ST=WA, L=Seattle, O=Example, OU=Dev, CN=wabbit-networks.io" --id "C=US, ST=WA, L=Seattle, O=Example, OU=Prod, CN=wabbit-networks.io"
```

### Update the trust identities for a trust policy by setting specified certificate name in the trust stores

```shell
notation policy update --name wabbit-network-build --cert wabbit-network-build.crt
```

### List all the trust policies by names

The output is a list of trust policy names.

```shell
notation policy list
```

An example of output messages:

```text
wabbit-network-dev
wabbit-network-prod
```

### List all the trust policies with details (TODO)

```shell
notation policy list --details
```

An example of output messages:

```text
name: wabbit-network-dev

wabbit-network-prod
```

### List trust policies for verifying specified artifact

```shell
  notation policy list --ref localhost:5000/net-monitor@sha256:xxx
```

### List trust policies for verifying artifacts in specified repository

```shell
  notation policy list --repo localhost:5000/net-monitor
```

### List trust policies with specified trust store configured

```shell
  notation policy list --ts ca:wabbit-network
```

### List trust policies with specified trust identity configured

```shell
  notation policy list --ti "CN=SecureBuilder, C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Marketing"
```

### Delete trust policies

Users cannot delete all the trust policies, at least one trust policy should be kept for signature verification. Deletion SHOULD fail if users intend to delete all the trust policies.

```shell
# Delete one trust policy
notation policy delete wabbit-network-dev

# Delete multiple trust policies
notation policy delete wabbit-network-dev wabbit-network-prod
```
