# notation sign

## Description

Use `notation sign` to sign artifacts stored in OCI compliant registries.

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

NOTE: This command is for signing OCI artifacts only. Use `notation blob sign` command for signing arbitrary blobs.

## Outline

```text
Sign artifacts

Usage:
  notation sign [flags] <reference>

Flags:
       --allow-referrers-api        [Experimental] use the Referrers API to store signatures in the registry, if not supported (returns 404), fallback to the Referrers tag schema
  -d,  --debug                      debug mode
  -e,  --expiry duration            optional expiry that provides a "best by use" time for the artifact. The duration is specified in minutes(m) and/or hours(h). For example: 12h, 30m, 3h20m
  -h,  --help                       help for sign
       --id string                  key id (required if --plugin is set). This is mutually exclusive with the --key flag
       --insecure-registry          use HTTP protocol while connecting to registries. Should be used only for testing
  -k,  --key string                 signing key name, for a key previously added to notation's key list. This is mutually exclusive with the --id and --plugin flags
       --oci-layout                 [Experimental] sign the artifact stored as OCI image layout
  -p,  --password string            password for registry operations (default to $NOTATION_PASSWORD if not specified)
       --plugin string              signing plugin name. This is mutually exclusive with the --key flag
       --plugin-config stringArray  {key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values.
       --signature-format string    signature envelope format, options: "jws", "cose" (default "jws")
  -u,  --username string            username for registry operations (default to $NOTATION_USERNAME if not specified)
  -m,  --user-metadata stringArray  {key}={value} pairs that are added to the signature payload
  -v,  --verbose                    verbose mode
```

### Set config property for OCI image manifest

Notation uses [OCI image manifest][oci-image-spec] to store signatures in registries. The empty JSON object `{}` is used as the default configuration content, and thus the `config` property is fixed, as following:

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
# - A signing plugin is installed. See plugin documentation (https://github.com/notaryproject/notaryproject/blob/v1.0.0/specs/plugin-extensibility.md) for more details.
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

### [Experimental] Sign container images stored in OCI layout directory

Container images can be stored in OCI image Layout defined in spec [OCI image layout][oci-image-layout]. It is a directory structure that contains files and folders. The OCI image layout could be a tarball or a directory in the filesystem. For example, a file named `hello-world.tar` or a directory named `hello-world`. Notation only supports signing images stored in OCI layout directory for now. Users can reference an image in the layout using either tags, or the exact digest. For example, use `hello-world:v1` or `hello-world@sha256xxx` to reference the image in OCI layout directory named `hello-world`.

Tools like `docker buildx` support building images stored in OCI image layout. The following example creates a tarball named `hello-world.tar` with tag `v1`. Please note that the digest can be retrieved in the output messages of `docker buildx build`.

```shell
docker buildx create --use
docker buildx build . -f Dockerfile -o type=oci,dest=hello-world.tar -t hello-world:v1
```

Users need to extract the tarball into a directory first, since Notation only support OCI layout directory for now. The following command creates the OCI layout directory.

```shell
mkdir hello-world
tar -xf hello-world.tar -C hello-world
```

Use flag `--oci-layout` to sign the image stored in OCI layout directory referenced by `hello-world@sha256xxx`. To access this flag `--oci-layout` , set the environment variable `NOTATION_EXPERIMENTAL=1`. For example:

```shell
export NOTATION_EXPERIMENTAL=1
# Assume OCI layout directory hello-world is under current path
notation sign --oci-layout hello-world@sha256:xxx
```

Upon successful signing, the signature is stored in the same layout directory and associated with the image. Use `notation list` command to list the signatures, for example:

```shell
export NOTATION_EXPERIMENTAL=1
# Assume OCI layout directory hello-world is under current path
notation list --oci-layout hello-world@sha256:xxx
```

[oci-artifact-manifest]: https://github.com/opencontainers/image-spec/blob/v1.1.0-rc2/artifact.md
[oci-image-spec]: https://github.com/opencontainers/image-spec/blob/v1.1.0-rc2/spec.md
[oci-referers-api]: https://github.com/opencontainers/distribution-spec/blob/v1.1.0-rc1/spec.md#listing-referrers
[oci-image-layout]: https://github.com/opencontainers/image-spec/blob/v1.1.0-rc2/image-layout.md
