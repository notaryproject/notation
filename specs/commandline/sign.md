# notation sign

## Description

Use `notation sign` to sign artifacts.

Signs an OCI artifact stored in the registry. Always sign artifact using digest(`@sha256:...`) rather than a tag(`:v1`) because tags are mutable and a tag reference can point to a different artifact than the one signed. If a tag is used, notation resolves the tag to the `digest` before signing.

Upon successful signing, the generated signature is pushed to the registry and associated with the signed OCI artifact. The output message is printed out as following:

```text
Successfully signed <registry>/<repository>@<digest>
```

If a `tag` is used to identify the OCI artifact, the output message is as following:

```test
Warning: Always sign the artifact using digest(`@sha256:...`) rather than a tag(`:<tag>`) because tags are mutable and a tag reference can point to a different artifact than the one signed.
Successfully signed <registry>/<repository>@<digest>
```

## Outline

```text
Sign artifacts

Usage:
  notation sign [flags] <reference>

Flags:
  -d,  --debug                      debug mode
  -e,  --expiry duration            optional expiry that provides a "best by use" time for the artifact. The duration is specified in minutes(m) and/or hours(h). For example: 12h, 30m, 3h20m
  -h,  --help                       help for sign
       --id string                  key id (required if --plugin is set). This is mutually exclusive with the --key flag
  -k,  --key string                 signing key name, for a key previously added to notation's key list. This is mutually exclusive with the --id and --plugin flags
  -p,  --password string            password for registry operations (default to $NOTATION_PASSWORD if not specified)
       --plain-http                 registry access via plain HTTP
       --plugin string              signing plugin name. This is mutually exclusive with the --key flag
       --plugin-config stringArray  {key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values.
       --signature-format string    signature envelope format, options: "jws", "cose" (default "jws")
       --signature-manifest string  manifest type for signature, options: "image", "artifact" (default "artifact")
  -u,  --username string            username for registry operations (default to $NOTATION_USERNAME if not specified)
  -m,  --user-metadata stringArray  {key}={value} pairs that are added to the signature payload
  -v,  --verbose                    verbose mode
```

## Use OCI image manifest to store signatures

By default, Notation uses [OCI artifact manifest][oci-artifact-manifest] to store signatures in registries. For registries that don't support `OCI artifact` or [Referrers API][oci-referers-api] is not enabled, users SHOULD use flag `--signature-manifest image` to force Notation to store the signatures using [OCI image manifest][oci-image-spec].

Note that there is no deterministic way to determine whether a registry supports `OCI artifact` or not. The following response status contained in error messages MAY indicate that the registry doesn't support `OCI artifact`.

- Response status `400 BAD Request` with error code `MANIFEST_INVALID` or `UNSUPPORTED`

### Set config property for OCI image manifest

OCI image manifest requires additional property `config` of type `descriptor`, which is not required by OCI artifact manifest. Notation creates a default config descriptor for the user if flag `--signature-manifest image` is used.

Notation uses empty JSON object `{}` as the default configuration content, and thus the default `config` property is fixed, as following:

```json
"config": {
    "mediaType": "application/vnd.cncf.notary.signature",
    "size": 2,
    "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
}
```

## Usage

### Sign an OCI artifact by adding new key

```shell
# Prerequisites:
# - A signing plugin is installed. See plugin documentation (https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details.
# - Configure the signing plugin as instructed by plugin vendor.

# Add a default signing key referencing the remote key identifier, and the plugin associated with it.
notation key add --default --name <key_name> --plugin <plugin_name> --id <remote_key_id>

# sign an artifact stored in a registry
notation sign <registry>/<repository>@<digest>
```

An example for a successful signing:

```console
$ notation sign localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
Successfully signed localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

### Sign an OCI artifact with on-demand remote key

```shell
notation sign --plugin <plugin_name> --id <remote_key_id> <registry>/<repository>@<digest>
```

### Sign an OCI artifact using COSE signature format

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# Use option "--signature-format" to set the signature format to COSE.
notation sign --signature-format cose <registry>/<repository>@<digest>
```

### Sign an OCI artifact stored in a registry using the default signing key

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# Use a digest that uniquely and immutably identifies an OCI artifact.
notation sign <registry>/<repository>@<digest>
```

### Sign an OCI Artifact with user metadata

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# sign an artifact stored in a registry and add user-metadata io.wabbit-networks.buildId=123 to the payload
notation sign --user-metadata io.wabbit-networks.buildId=123 <registry>/<repository>@<digest>

# sign an artifact stored in a registry and add user-metadata io.wabbit-networks.buildId=123 and io.wabbit-networks.buildTime=1672944615 to the payload
notation sign --user-metadata io.wabbit-networks.buildId=123 --user-metadata io.wabbit-networks.buildTime=1672944615 <registry>/<repository>@<digest>
```

### Sign an OCI artifact stored in a registry and specify the signature expiry duration, for example 24 hours

```shell
notation sign --expiry 24h <registry>/<repository>@<digest>
```

### Sign an OCI artifact stored in a registry using a specified signing key

```shell
# List signing keys to get the key name
notation key list

# Sign a container image using the specified key name
notation sign --key <key_name> <registry>/<repository>@<digest>
```

### Sign an OCI artifact identified by a tag

```shell
# Prerequisites:
# A default signing key is configured using CLI "notation key"

# Use a tag to identify a container image
notation sign <registry>/<repository>:<tag>
```

An example for a successful signing:

```console
$ notation sign localhost:5000/net-monitor:v1
Warning: Always sign the artifact using digest(`@sha256:...`) rather than a tag(`:v1`) because tags are mutable and a tag reference can point to a different artifact than the one signed.
Successfully signed localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

### Sign an artifact and store the signature using OCI image manifest

```shell
notation sign --signature-manifest image <registry>/<repository>@<digest>
```

[oci-artifact-manifest]: https://github.com/opencontainers/image-spec/blob/v1.1.0-rc2/artifact.md
[oci-image-spec]: https://github.com/opencontainers/image-spec/blob/v1.1.0-rc2/spec.md
[oci-referers-api]: https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#listing-referrers
