# notation blob

## Description

Use `notation blob` command to sign, verify, and inspect signatures associated with arbitrary blobs. Notation can sign and verify any arbitrary bag of bits like zip files, documents, executables, etc. When a user signs a blob, `notation` produces a detached signature, which the user can transport/distribute using any medium that the user prefers along with the original blob. On the verification side, Notation can verify the blob's signature and assert that the blob has not been tampered with during its transmission. 

Users can use `notation blob policy` command to manage trust policies for verifying a blob signature. The `notation blob policy` command provides a user-friendly way to manage trust policies for signed blobs. It allows users to show trust policy configuration, import/export a trust policy configuration file from/to a JSON file. For more details, see [blob trust policy specification and examples](https://github.com/notaryproject/specifications/blob/main/specs/trust-store-trust-policy.md#blob-trust-policy).

The sample trust policy file (`trustpolicy.blob.json`) for verifying signed blobs is shown below. This sample trust policy file, contains three different statements for different usecases:

- The Policy named "wabbit-networks-policy" is for verifying blob artifacts signed by Wabbit Networks.
- Policy named "skip-verification-policy" is for skipping verification on blob artifacts.
- Policy "global-verification-policy" is for auditing verification results when user does not provide `--policy-name` argument in `notation blob verify` command.

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            "name": "wabbit-networks-policy",
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
            "name": "skip-verification-policy",
            "signatureVerification": {
              "level" : "skip" 
            }
        },
        {
            "name": "global-verification-policy",
            "globalPolicy": true,
            "signatureVerification": {
              "level" : "audit"
            },
            "trustStores": ["ca:acme-rockets"],
            "trustedIdentities": ["*"]
        }
    ]
}
```

## Outline

### notation blob command

```text
Sign, inspect, and verify signatures and configure trust policies.

Usage:
  notation blob [command]

Available Commands:
  inspect   inspect a signature associated with a blob
  policy    manage trust policy configuration for signed blobs
  sign      produce a detached signature for a given blob
  verify    verify a signature associated with a blob

Flags:
  -h, --help   help for blob
```

### notation blob sign

```text
Produce a signature for a given blob. A detached signature file will be written to the currently working directory with blob file name + ".sig" + signature format as the file extension. For example, signature file name for "myBlob.bin" will be "myBlob.bin.sig.jws" for JWS signature format or "myBlob.bin.sig.cose" for COSE signature format.

Usage:
  notation blob sign [flags] <blob_path>

Flags:
       --signature-directory string optional path where the blob signature needs to be placed (default: currently working directory) 
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

### notation blob policy

```text
Manage trust policy configuration for arbitrary blob signature verification.

Usage:
  notation blob policy [command]

Available Commands:
  import    import trust policy configuration from a JSON file
  show      show trust policy configuration

Flags:
  -h, --help   help for policy
```

### notation blob policy import

```text
Import blob trust policy configuration from a JSON file

Usage:
  notation blob policy import [flags] <file_path>

Flags:    
      --force     override the existing trust policy configuration, never prompt
  -h, --help      help for import
```

### notation blob policy show

```text
Show blob trust policy configuration

Usage:
  notation blob policy show [flags]

Flags:
  -h, --help      help for show
```

### notation blob verify

```text
Verify a signature associated with a blob

Usage:
  notation blob verify [flags] --signature <signature_path> <blob_path>

Flags:
      --signature string      location of the blob signature file
      --media-type string     optional media type of the blob to verify
      --policy-name string    optional policy name to verify against. If not provided, notation verifies against the global policy if it exists.
  -m, --user-metadata stringArray   user defined {key}={value} pairs that must be present in the signature for successful verification if provided
      --plugin-config stringArray   {key}={value} pairs that are passed as it is to a plugin, if the verification is associated with a verification plugin, refer plugin documentation to set appropriate values
  -o, --output string         output format, options: 'json', 'text' (default "text")
  -d, --debug                 debug mode
  -v, --verbose               verbose mode
  -h, --help                  help for inspect
```

## Usage

## Produce blob signatures

### Sign a blob by adding a new key

```shell
# Prerequisites:
# - A signing plugin is installed. See plugin documentation (https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details.
# - Configure the signing plugin as instructed by plugin vendor.

# Add a default signing key referencing the remote key identifier, and the plugin associated with it.
notation key add --default --name <key_name> --plugin <plugin_name> --id <remote_key_id>

# sign a blob
notation blob sign /tmp/my-blob.bin
```

An example for a successful signing:

```console
$ notation blob sign /tmp/my-blob.bin
Successfully signed /tmp/my-blob.bin
Signature file written to /absolute/path/to/cwd/my-blob.bin.sig.jws
```

### Sign a blob by generating the signature in a particular directory
```console
$ notation blob sign --signature-directory /tmp/xyz/sigs /tmp/my-blob.bin
Successfully signed /tmp/my-blob.bin
Signature file written to /tmp/xyz/sigs/my-blob.bin.sig.jws
```

### Sign a blob using a relative path
```console
$ notation blob sign ./relative/path/my-blob.bin
Successfully signed ./relative/path/my-blob.bin
Signature file written to /absolute/path/to/cwd/my-blob.bin.sig.jws
```

### Sign a blob with a plugin

```shell
notation blob sign --plugin <plugin_name> --id <remote_key_id> /tmp/my-blob.bin
```

### Sign a blob using COSE signature format

```console
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# Use option "--signature-format" to set the signature format to COSE.
$ notation blob sign --signature-format cose /tmp/my-blob.bin
Successfully signed /tmp/my-blob.bin
Signature file written to /absolute/path/to/cwd/my-blob.bin.sig.cose
```

### Sign a blob using the default signing key

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

notation blob sign /tmp/my-blob.bin
```

### Sign a blob with user metadata

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# sign a blob and add user-metadata io.wabbit-networks.buildId=123 to the payload
notation blob sign --user-metadata io.wabbit-networks.buildId=123 /tmp/my-blob.bin

# sign a blob and add user-metadata io.wabbit-networks.buildId=123 and io.wabbit-networks.buildTime=1672944615 to the payload
notation blob sign --user-metadata io.wabbit-networks.buildId=123 --user-metadata io.wabbit-networks.buildTime=1672944615 /tmp/my-blob.bin
```

### Sign a blob and specify the media type for the blob

```shell
notation blob sign --media-type <media type> /tmp/my-blob.bin
```

### Sign a blob and specify the signature expiry duration, for example 24 hours

```shell
notation blob sign --expiry 24h /tmp/my-blob.bin
```

### Sign a blob using a specified signing key

```shell
# List signing keys to get the key name
notation key list

# Sign a container image using the specified key name
notation blob sign --key <key_name> /tmp/my-blob.bin
```

## Inspect blob signatures

### Display details of the given blob signature and its associated certificate properties


```text
notation blob inspect [flags] /tmp/my-blob.bin.sig.jws
```

### Inspect the given blob signature

```shell
# Prerequisites: Signatures is produced by notation blob sign command
notation blob inspect /tmp/my-blob.bin.sig.jws
```

An example output:
```shell
/tmp/my-blob.bin.sig.jws
    ├── signature algorithm: RSASSA-PSS-SHA-256
    ├── signature envelope type: jws
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
        ├── media type: application/octet-stream
        ├── digest: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
        └── size: 16724
```

### Inspect the given blob signature with JSON Output

```shell
notation blob inspect -o json /tmp/my-blob.bin.sig.jws
```

## Import/Export trust policy configuration files

### Import blob trust policy configuration from a JSON file

An example of import trust policy configuration from a JSON file:

```shell  
notation blob policy import ./my_policy.json
```

The trust policy configuration in the JSON file should be validated according to [trust policy properties](https://github.com/notaryproject/notaryproject/specs/trust-store-trust-policy.md#blob-trust-policy). A successful message should be printed out if trust policy configuration are imported successfully. Error logs including the reason should be printed out if the importing fails.

If there is an existing trust policy configuration, prompt for users to confirm whether discarding existing configuration or not. Users can use `--force` flag to discard existing trust policy configuration without prompt.

### Show blob trust policies

Use the following command to show trust policy configuration:

```shell
notation blob policy show
```

Upon successful execution, the trust policy configuration is printed out to standard output. If trust policy is not configured or is malformed, users should receive an error message via standard error output, and a tip to import trust policy configuration from a JSON file.

### Export blob trust policy configuration into a JSON file

Users can redirect the output of command `notation blob policy show` to a JSON file.

```shell
notation blob policy show > ./blob_trust_policy.json
```

### Update trust policy configuration

The steps to update blob trust policy configuration:

1. Export trust policy configuration into a JSON file.

   ```shell
   notation blob policy show > ./blob_trust_policy.json
   ```

2. Edit the exported JSON file "blob_trust_policy.json", update trust policy configuration and save the file.
3. Import trust policy configuration from the file.

   ```shell
   notation blob policy import ./blob_trust_policy.json
   ```

## Verify blob signatures
The `notation blob verify` command can be used to verify blob signatures. In order to verify signatures, user will need to setup a trust policy file `trustpolicy.blob.json` with Policies for blobs. Below are two examples of how a policy configuration file can be setup for verifying blob signatures.

- The Policy named "wabbit-networks-policy" is for verifying blob artifacts signed by Wabbit Networks.
- Policy  named "global-verification-policy" is for auditing verification results when user doesn't not provide `--policy-name` argument in `notation blob verify` command.

```jsonc
{
    "version": "1.0",
    "trustPolicies": [
        {
            "name": "wabbit-networks-policy",
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
            "name": "global-verification-policy",
            "globalPolicy": true,
            "signatureVerification": {
              "level" : "audit"
            },
            "trustStores": ["ca:acme-rockets"],
            "trustedIdentities": ["*"]
        }
    ]
}
```

### Verify the signature of a blob

Configure trust store and trust policy properly before using `notation blob verify` command.

```shell

# Prerequisites: Blob and its associated signature is present on the filesystem.
# Configure trust store by adding a certificate file into trust store named "wabbit-network" of type "ca"
notation certificate add --type ca --store wabbit-networks wabbit-networks.crt

# Setup the trust policy in a JSON file named "trustpolicy.blob.json" under directory "{NOTATION_CONFIG}".

# Verify the blob signature
notation blob verify --signature /tmp/my-blob.bin.sig.jws /tmp/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature /tmp/my-blob.bin.sig.jws
```

### Verify the signature with user metadata

Use the `--user-metadata` flag to verify that provided key-value pairs are present in the payload of the valid signature.

```shell
# Verify the signature and verify that io.wabbit-networks.buildId=123 is present in the signed payload
notation blob verify --user-metadata io.wabbit-networks.buildId=123 --signature /tmp/my-blob.bin.sig.jws /tmp/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature /tmp/my-blob.bin.sig.jws

The signature contains the following user metadata:

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
notation blob verify --media-type application/my-media-octet-stream --signature /tmp/my-blob.bin.sig.jws /tmp/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature /tmp/my-blob.bin.sig.jws

The blob is of media type `application/my-media-octet-stream`.

```

An example of output messages for an unsuccessful verification:

```text
Error: Signature verification failed due to a mismatch in the blob's media type 'application/xyz' and the expected type 'application/my-media-octet-stream'.
```

### Verify the signature using a policy name

Use the `--policy-name` flag to select a policy to verify the signature against.

```shell
notation blob verify --policy-name wabbit-networks-policy --signature ./sigs/my-blob.bin.sig.jws ./blobs/my-blob.bin
```

An example of output messages for a successful verification:

```text
Successfully verified signature ./sigs/my-blob.bin.sig.jws using policy `wabbit-networks-policy`

```
An example of output messages for an unsuccessful verification:

```text
Error: signature verification failed for policy `wabbit-networks-policy`
```