# notation policy

## Description

Use `notation blob` command to sign, verify, and inspect signatures associated with arbitrary blobs. Notation can sign and verify any arbitrary bag of bits like zip files, documents, executables, etc. When a user signs a blob, `notation` produces a detached signature, which the user can transport/distribute in any medium that the user prefers along with the original blob. On the verification side, Notation can verify the blob's signature and assert that the blob has not been tampered with during its transmission. For more details, see [trust policy specification and examples](https://github.com/notaryproject/specifications/blob/main/specs/signing-and-verification-workflow.md#blob-signing-workflow).

## Outline

### notation blob command

```text
Sign, Inspect, and Verify signatures associates with arbitrary blobs.

Usage:
  notation blob [command]

Available Commands:
  sign      produce a detached signature for a given blob
  inspect   inspect a signature associated with a blob
  verify    verify a signature associated with a blob

Flags:
  -h, --help   help for blob
```

### notation blob sign

```text
Produce a detached signature for a given blob

Usage:
  notation blob sign [flags] -n my-blob-signature <blob_path>

Flags:
  -n,  --signature-name string      friendly name for the detached signature. Signature file will be written to the currently working directory with this name plus ".sig" plus signature format as the file extension
       --media-type string          optional media type of the blob (default: "application/octet-stream")
  -e,  --expiry duration            optional expiry that provides a "best by use" time for the blob. The duration is specified in minutes(m) and/or hours(h). For example: 12h, 30m, 3h20m
       --id string                  key id (required if --plugin is set). This is mutually exclusive with the --key flag
  -k,  --key string                 signing key name, for a key previously added to notation's key list. This is mutually exclusive with the --id and --plugin flags
       --plugin string              signing plugin name. This is mutually exclusive with the --key flag
       --plugin-config stringArray  {key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values.
       --signature-format string    signature envelope format, options: "jws", "cose" (default "jws")
  -m,  --user-metadata stringArray  {key}={value} pairs that are added to the signature payload
  -d,  --debug                      debug mode
  -v,  --verbose                    verbose mode
  -h,  --help                       help for sign
```

### notation blob inspect

```text
Inspect a signature associated with a blob

Usage:
  notation blob inspect [flags] <signature_path>

Flags:
  -o, --output string         output format, options: 'json', 'text' (default "text")
  -d, --debug                 debug mode
  -v, --verbose               verbose mode
  -h, --help                  help for inspect
```

### notation blob verify

```text
Verify a signature associated with a blob

Usage:
  notation blob verify [flags] --signature <signature_path> <blob_path>

Flags:
  -s, --signature path        location of the detached signature
      --media-type string     optional media type of the blob to verify
      --policy-scope string   optional policy scope to verify against. If not provided, notation verifies against wildcard policy if it exists.
  -m, --user-metadata stringArray   user defined {key}={value} pairs that must be present in the signature for successful verification if provided
  -o, --output string         output format, options: 'json', 'text' (default "text")
  -d, --debug                 debug mode
  -v, --verbose               verbose mode
  -h, --help                  help for inspect
```

## Usage

## Produce detached blob signatures

### Sign a blob by adding a new key

```shell
# Prerequisites:
# - A signing plugin is installed. See plugin documentation (https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details.
# - Configure the signing plugin as instructed by plugin vendor.

# Add a default signing key referencing the remote key identifier, and the plugin associated with it.
notation key add --default --name <key_name> --plugin <plugin_name> --id <remote_key_id>

# sign a blob
notation blob sign --signature-name my-blob-signature /tmp/my-blob.bin
```

An example for a successful signing:

```console
$ notation blob sign --signature-name my-blob-signature /tmp/my-blob.bin
Successfully signed /tmp/my-blob.bin
Signature written to ./my-blob-signature.sig.jws
```

### Sign a blob with on-demand remote key

```shell
notation blob sign --plugin <plugin_name> --id <remote_key_id> --signature-name my-blob-signature /tmp/my-blob.bin
```

### Sign a blob using COSE signature format

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# Use option "--signature-format" to set the signature format to COSE.
notation blob sign --signature-format cose --signature-name my-blob-signature /tmp/my-blob.bin
Successfully signed /tmp/my-blob.bin
Signature written to ./my-blob-signature.sig.cose
```

### Sign a blob using the default signing key

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

notation blob sign --signature-name my-blob-signature /tmp/my-blob.bin
```

### Sign a blob with user metadata

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# sign a blob and add user-metadata io.wabbit-networks.buildId=123 to the payload
notation blob sign --user-metadata io.wabbit-networks.buildId=123 --signature-name my-blob-signature /tmp/my-blob.bin

# sign a blob and add user-metadata io.wabbit-networks.buildId=123 and io.wabbit-networks.buildTime=1672944615 to the payload
notation blob sign --user-metadata io.wabbit-networks.buildId=123 --user-metadata io.wabbit-networks.buildTime=1672944615 --signature-name my-blob-signature /tmp/my-blob.bin
```

### Sign a blob with media type

```shell
notation blob sign --media-type <media type> --signature-name my-blob-signature /tmp/my-blob.bin
```

### Sign a blob and specify the signature expiry duration, for example 24 hours

```shell
notation blob sign --expiry 24h --signature-name my-blob-signature /tmp/my-blob.bin
```

### Sign a blob using a specified signing key

```shell
# List signing keys to get the key name
notation key list

# Sign a container image using the specified key name
notation blob sign --key <key_name> --signature-name my-blob-signature /tmp/my-blob.bin
```

## Inspect detached blob signatures

### Display details of the given detached blob signature and its associated certificate properties


```text
notation blob inspect [flags] /tmp/my-blob-signature.sig.jws
```

### Inspect the given detached blob signature

```shell
# Prerequisites: Signatures is produced by notation blob sign command
notation blob inspect /tmp/my-blob-signature.sig.jws
```

An example output:
```shell
Inspecting /tmp/my-blob-signature.sig.jws
/tmp/my-blob-signature.sig.jws
└── application/octet-stream
    ├── sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signature format: jws
        ├── signed attributes
        │   ├── content type: application/vnd.cncf.notary.payload.v1+json
        │   ├── signing scheme: notary.signingAuthority.x509
        │   ├── signing time: Fri Jun 23 22:04:01 2023
        │   ├── expiry: Sat Jun 29 22:04:01 2024
        │   └── io.cncf.notary.verificationPlugin: com.example.nv2plugin
        ├── unsigned attributes
        │   ├── io.cncf.notary.timestampSignature: <Base64(TimeStampToken)>
        │   └── io.cncf.notary.signingAgent: notation/1.0.0
        ├── certificates
        │   ├── SHA256 fingerprint: b13a843be16b1f461f08d61c14f3eab7d87c073570da077217541a7eb31c084d
        │   │   ├── issued to: wabbit-com Software
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   ├── SHA256 fingerprint: 4b9fa61d5aed0fabbc7cb8fe2efd049da57957ed44f2b98f7863ce18effd3b89
        │   │   ├── issued to: wabbit-com Software Code Signing PCA 2010
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   └── SHA256 fingerprint: ea3939548ad0c0a86f164ab8b97858854238c797f30bddeba6cb28688f3f6536
        │       ├── issued to: wabbit-com Software Root Certificate Authority
        │       ├── issued by: wabbit-com Software Root Certificate Authority
        │       └── expiry: Sat Jun 23 22:04:01 2035
        └── signed artifact
            ├── media type: application/vnd.oci.image.manifest.v1+json
            ├── digest: sha256:b94d27b9934d3e08a52e52d7da7fac484efe37a5380ee9088f7ace2efcde9
            └── size: 16724
```

### Inspect the given detached blob signature with JSON Output

```shell
notation blob inspect -o json /tmp/my-blob-signature.sig.jws
```

## Verify detached blob signatures
The `notation blob verify` command can be used to verify blob signatures. In order to verify signatures, user will need to setup a policy configuration file with Policies scoped to blobs. Below are three examples of how a policy configuration file can be setup for verifying blob signatures.

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

### Verify the detached signature of a blob

Configure trust store and trust policy properly before using `notation blob verify` command.

```shell

# Prerequisites: Signature is produced on the filesystem from `notation blob sign` command.
# Configure trust store by adding a certificate file into trust store named "wabbit-network" of type "ca"
notation certificate add --type ca --store wabbit-networks wabbit-networks.crt

# Create a JSON file named "trustpolicy.json" under directory "{NOTATION_CONFIG}".

# Verify the detached signature
notation blob verify --signature /tmp/my-blob-signature.sig.jws /tmp/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature /tmp/my-blob-signature.sig.jws
```

### Verify the signature with user metadata

Use the `--user-metadata` flag to verify that provided key-value pairs are present in the payload of the valid signature.

```shell
# Verify the signature and verify that io.wabbit-networks.buildId=123 is present in the signed payload
notation blob verify --user-metadata io.wabbit-networks.buildId=123 --signature /tmp/my-blob-signature.sig.jws /tmp/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature /tmp/my-blob-signature.sig.jws

The blob signature is having the following user metadata.

KEY                         VALUE
io.wabbit-networks.buildId  123
```

An example of output messages for an unsuccessful verification:

```text
Error: signature verification failed: unable to find specified metadata in the given signature
```

### Verify the signature with media type

Use the `--media-type` flag to verify that signature is for the provided media-type.

```shell
# Verify the signature and verify that application/my-media-octet-stream is the media type
notation blob verify --media-type application/my-media-octet-stream --signature /tmp/my-blob-signature.sig.jws /tmp/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature /tmp/my-blob-signature.sig.jws

The blob is of media type `application/my-media-octet-stream`.

```

An example of output messages for an unsuccessful verification:

```text
Error: signature verification failed: The blob is not of media type `application/my-media-octet-stream`.
```

### Verify the signature using a policy scope

Use the `--policy-scope` flag to select a Policy scope to verify the signature against.

```shell
notation blob verify --policy-scope blob-verification-selector --signature /tmp/my-blob-signature.sig.jws /tmp/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature /tmp/my-blob-signature.sig.jws using policy scope `blob-verification-selector`

```
An example of output messages for an unsuccessful verification:

```text
Error: signature verification failed for Policy scope `blob-verification-selector`
```