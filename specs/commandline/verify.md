# notation verify

## Description

Use `notation verify` command to verify signatures associated with the artifact. Signature verification succeeds if verification succeeds for at least one of the signatures associated with the artifact. Upon successful verification, the output message is printed out as follows:

```text
Successfully verified signature for <registry>/<repository>@<digest>
```

Tags are mutable and a tag reference can point to a different artifact than that was signed referred by the same tag. If a `tag` is used to identify the OCI artifact, the output message is as follows:

```text
Resolved artifact tag `<tag>` to digest `<digest>` before verification.
Warning: The resolved digest may not point to the same signed artifact, since tags are mutable.
Successfully verified signature for <registry>/<repository>@<digest>
```

A signature can have user defined metadata. If the signature for the OCI artifact contains any metadata, the output message is as follows:

```text
Successfully verified signature for <registry>/<repository>@<digest>

The artifact was signed with the following user metadata.

KEY    VALUE
<key>  <value>
```

## Outline

```text
Verify signatures associated with the artifact.

Usage:
  notation verify [flags] <reference>

Flags:
  -h,  --help                    help for verify
  -o,  --output string           output format, options: 'json', 'text' (default "text")
  -p,  --password string         password for registry operations (default to $NOTATION_PASSWORD if not specified)
       --plain-http              registry access via plain HTTP
       --plugin-config strings   {key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values
  -u,  --username string         username for registry operations (default to $NOTATION_USERNAME if not specified)
  -m,  --user-metadata strings   user defined {key}={value} pairs that must be present in the signature for successful verification if provided
```

## Usage

Pre-requisite: User needs to configure trust store and trust policy properly before using `notation verify` command.

### Configure Trust Store

Use `notation certificate` command to configure trust stores.

### Configure Trust Policy

Users who consume signed artifact from a registry use the trust policy to specify trusted identities which sign the artifacts, and level of signature verification to use. The trust policy is a JSON document. User needs to create a file named `trustpolicy.json` under `{NOTATION_CONFIG}`. See [Notation Directory Structure](https://notaryproject.dev/docs/tutorials/directory-structure/) for `{NOTATION_CONFIG}`.

An example of `trustpolicy.json`:

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            // Policy for all artifacts, from any registry location.
            "name": "wabbit-networks-images",                         // Name of the policy.
            "registryScopes": [ "localhost:5000/net-monitor" ],       // The registry artifacts to which the policy applies.
            "signatureVerification": {                                // The level of verification - strict, permissive, audit, skip.
                "level": "strict"
            },
            "trustStores": [ "ca:wabbit-networks" ],                  // The trust stores that contains the X.509 trusted roots.
            "trustedIdentities": [                                    // Identities that are trusted to sign the artifact.
                "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Finance, CN=SecureBuilder"
            ]
        }
    ]
}
```

For a Linux user, store file `trustpolicy.json` under directory `${HOME}/.config/notation/`.

For a MacOS user, store file `trustpolicy.json` under directory `${HOME}/Library/Application Support/notation/`.

For a Windows user, store file `trustpolicy.json` under directory `%USERPROFILE%\AppData\Roaming\notation\`.

Example values on trust policy properties:

| Property name         | Value                                                                                      | Meaning                                                                                                                                                            |
| ----------------------|--------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name                  | "wabbit-networks-images"                                                                   | The name of the policy is "wabbit-networks-images".                                                                                                                |
| registryScopes        | "localhost:5000/net-monitor"                                                               | The policy only applies to artifacts stored in repository `localhost:5000/net-monitor`.                                                                            |
| registryScopes        | "localhost:5000/net-monitor", "localhost:5000/nginx"                                       | The policy applies to artifacts stored in two repositories: `localhost:5000/net-monitor` and `localhost:5000/nginx`.                                               |
| registryScopes        | "*"                                                                                        | The policy applies to all the artifacts stored in any repositories.                                                                                                |
| signatureVerification | "level": "strict"                                                                          | Signature verification is performed at strict level, which enforces all validations: `integrity`, `authenticity`, `authentic timestamp`, `expiry` and `revocation`.|
| signatureVerification | "level": "permissive"                                                                      | The permissive level enforces most validations, but will only logs failures for `revocation` and `expiry`.                                                         |
| signatureVerification | "level": "audit"                                                                           | The audit level only enforces signature `integrity` if a signature is present. Failure of all other validations are only logged.                                   |
| signatureVerification | "level": "skip"                                                                            | The skip level does not fetch signatures for artifacts and does not perform any signature verification.                                                            |
| trustStores           | "ca:wabbit-networks"                                                                       | Specify the trust store that uses the format {trust-store-type}:{named-store}. The trust store is added using `notation certificate add` command.                  |
| trustStores           | "ca:wabbit-networks", "ca:rocket-networks"                                                 | Specify two trust stores, each of which contains the trusted roots against which signatures are verified.                                                          |
| trustedIdentities     | "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Finance, CN=SecureBuilder" | User only trusts the identity with specific subject. User can use `notation certificate show` command to get the `subject` info.                                   |
| trustedIdentities     | "*"                                                                                        | User trusts any identity (signing certificate) issued by the CA(s) in trust stores.                                                                                |

User can configure multiple trust policies for different scenarios. See [Trust Policy Schema and properties](https://github.com/notaryproject/notaryproject/blob/main/specs/trust-store-trust-policy.md#trust-policy) for details.

### Verify signatures on an OCI artifact stored in a registry

Configure trust store and trust policy properly before using `notation verify` command.

```shell

# Prerequisites: Signatures are stored in a registry referencing the signed OCI artifact
# Configure trust store by adding a certificate file into trust store named "wabbit-network" of type "ca"
notation certificate add --type ca --store wabbit-networks wabbit-networks.crt

# Create a JSON file named "trustpolicy.json" under directory "{NOTATION_CONFIG}".

# Verify signatures on the supplied OCI artifact identified by the digest
notation verify localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

An example of output messages for a successful verification:

```text
Successfully verified signature for localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

### Verify signatures on an OCI artifact with user metadata

Use the `--user-metadata` flag to verify that provided key-value pairs are present in the payload of the valid signature.

```shell
# Verify signatures on the supplied OCI artifact identified by the digest and verify that io.wabbit-networks.buildId=123 is present in the signed payload
notation verify --user-metadata io.wabbit-networks.buildId=123 localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

An example of output messages for a successful verification:

```text
Successfully verified signature for localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9

The artifact is signed with the following user metadata.

KEY                         VALUE
io.wabbit-networks.buildId  123
```

An example of output messages for an unsuccessful verification:

```text
Error: signature verification failed: unable to find specified metadata in any signatures
```

### Verify signatures on an OCI artifact identified by a tag

A tag is resolved to a digest first before verification.

```shell
# Prerequisites: Signatures are stored in a registry referencing the signed OCI artifact
# Verify signatures on an OCI artifact identified by the tag
notation verify localhost:5000/net-monitor:v1
```

An example of output messages for a successful verification:

```text
Resolved artifact tag `v1` to digest `sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9` before verification.
Warning: The resolved digest may not point to the same signed artifact, since tags are mutable.
Successfully verified signature for localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

### Verify signatures on an OCI artifact with json output

Use the `--output` flag to format successful verification output in json.

```shell
notation verify --output json localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

An example of output messages for a successful verification:

```text
{
    "reference": "localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
    "userMetadata": {
        "io.wabbit-networks.buildId": "123"
    }
}
```