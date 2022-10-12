# notation verify

## Description

Use `notation verify` command to verify signatures on an artifact. Signature verification succeeds if verification succeeds for at least one of the signatures associated with the artifact. The digest of the supplied artifact is returned upon successful verification. It is recommended that this digest reference be used to pull the artifact subsequently, as registry tags may be mutable, and a tag reference can point to a different artifact that what was verified.

## Outline

```text
Verify signatures associated with the artifact.

Usage:
  notation verify [flags] <reference>

Flags:
  -h, --help                    help for verify
  -p, --password string         Password for registry operations (default to $NOTATION_PASSWORD if not specified)
      --plugin-config strings   {key}={value} pairs that are passed as is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values
  -u, --username string         Username for registry operations (default to $NOTATION_USERNAME if not specified)
```

## Usage

Pre-requisite: User needs to configure trust store and trust policy properly before using `notation verify` command.

### Configure Trust Store

Use `notation certificate` command to configure trust stores.

### Configure Trust Policy

Users who consume signed artifact from a registry use the trust policy to specify trusted identities which sign the artifacts, and level of signature verification to use. The trust policy is a JSON document. User needs to create a file named `trustpolicy.json` under `{NOTATION_CONFIG}`. See [Notation Directory Structure](https://github.com/notaryproject/notation/blob/main/specs/directory.md) for `{NOTATION_CONFIG}`.

An example of `trustpolicy.json`:

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            // Policy for all artifacts, from any registry location.
            "name": "wabbit-networks-images",  // Name of the policy.
            "registryScopes": [ "*" ],         // The registry artifacts to which the policy applies.
            "signatureVerification": {         // The level of verification - strict, permissive, audit, skip.
                "level": "strict"
            },
            "trustStores": [ "ca:wabbit-networks" ], // The trust stores that contains the X.509 trusted roots.
            "trustedIdentities": [                   // Identities that are trusted to sign the artifact.
                "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Finance, CN=SecureBuilder"
            ]
        }
    ]
}
```

In this example, only one policy is configured with the name `wabbit-networks-images`. With the value of property `registryScopes` set to `*`, this policy applies to all artifacts from any registry location. User can configure multiple trust policies for different scenarios. See [Trust Policy Schema and properties](https://github.com/notaryproject/notaryproject/blob/main/trust-store-trust-policy-specification.md#trust-policy) for details.

### Verify signatures on a container image stored in a registry (Neither trust store nor trust policy is configured yet)

```shell

# Prerequisites: Signatures are stored in a registry referencing the signed container image

# Configure trust store by adding a certificate file into trust store named "wabbit-network" of type "ca"
notation certificate add --type ca --store wabbit-networks wabbit-networks.crt

# Configure trust policy by creating a JSON document named "trustpolicy.json" under directory "{NOTATION_CONFIG}"
# Example on Linux
cat <<EOF > $HOME/.config/notation/trustpolicy.json
{
    "version": "1.0",
    "trustPolicies": [
        {
            "name": "wabbit-networks-images",   // Name of the policy.
            "registryScopes": [ "registry.wabbit-networks.io/software/net-monitor" ],          // The registry artifacts to which the policy applies.
            "signatureVerification": {          // The level of verification - strict, permissive, audit, skip.
                "level" : "strict" 
            },
            "trustStores": [ "ca:wabbit-networks" ], // The trust stores that contains the X.509 trusted roots.
            "trustedIdentities": [                   // Identities that are trusted to sign the artifact.
                "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Finance, CN=SecureBuilder"
            ]
        }
    ]
}
EOF

# Verify signatures on a container image 
notation verify registry.wabbit-networks.io/software/net-monitor:v1
```

### Verify signatures on an OCI artifact stored in a registry (Trust store and trust policy are configured properly)

```shell
# Prerequisites: Signatures are stored in a registry referencing the signed OCI artifact

# Verify signatures on an OCI artifact identified by the digest
notation verify registry.wabbit-networks.io/software/net-monitor@sha256:abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234
```
