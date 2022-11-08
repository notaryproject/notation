# notation sign

## Description

Use `notation sign` to sign artifacts.

Signs an OCI artifact that is stored in a registry. Always use a `digest` to identify an artifact. `Tags` are mutable, but `digests` uniquely and immutably identify artifacts. If a tag is used, notation resolves the tag to the `digest` before signing.

Upon successful signing, the generated signature is pushed to the registry and associated with the signed OCI artifact. The output message is printed out as following:

```text
Sign succeeded. Signature has been attached to <registry>/<repository>@<digest>.
```

If a `tag` is used to identify the OCI artifact, the output message is as following:

```test
Warning: Tag is used. Always use digest to identify the reference uniquely and immutably.
Resolve tag "<tag>" to digest "<digest>"
Sign succeeded. Signature has been attached to <registry>/<repository>@<digest>.
```

## Outline

```text
Sign artifacts

Usage:
  notation sign [flags] <reference>

Flags:
  -d, --debug                    print out debug output
  -e, --expiry duration          optional expiry that provides a "best by use" time for the artifact. The duration is specified in minutes(m) and/or hours(h). For example: 12h, 30m, 3h20m
  -h, --help                     help for sign
  -k, --key string               signing key name, for a key previously added to notation's key list.
  -p, --password string          password for registry operations (default to $NOTATION_PASSWORD if not specified)
      --plain-http               registry access via plain HTTP
      --plugin-config strings    {key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values
      --signature-format string  signature envelope format, options: 'jws', 'cose' (default "jws")
  -u, --username string          username for registry operations (default to $NOTATION_USERNAME if not specified)
```

## Usage

### Sign an OCI artifact stored in a registry using a remote key

```shell
# Prerequisites: 
# - A compliant signing plugin is installed in notation. See notation plugin documentation (https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details.
# - User creates keys and certificates in a 3rd party key provider (e.g. key vault, key management service). The signing plugin installed in previous step must support generating signatures using this key provider.

# Add a default signing key referencing the key identifier for the remote key, and the plugin associated with it.
notation key add --default --name <key_name> --plugin <plugin_name> --id <remote_key_id>

# sign an artifact stored in a registry using a remote key
notation sign <registry>/<repository>@<digest>
```

An example for a successful signing:

```shell
$ notation sign localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
Sign succeeded. Signature has been attached to localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

### Sign an OCI artifact using COSE signature format

```shell
# Prerequisites: 
# A default signing key is configured using CLI "notation key"

# Use option "--signature-format" to set the signature format to COSE.
$ notation sign --signature-format cose <registry>/<repository>@<digest>
```

### Sign an OCI artifact stored in a registry using the default signing key

```shell
# Prerequisites: 
# A default signing key is configured using CLI "notation key"

# Use a digest that uniquely and immutably identifies an OCI artifact.
$ notation sign <registry>/<repository>@<digest>
```

### Sign an OCI artifact stored in a registry and specify the signature expiry duration, for example 24 hours

```shell
notation sign --expiry 24h <registry>/<repository>@<digest>
```

### Sign an OCI artifact stored in a registry using a specified signing key

```shell
# List signing keys to get the key name
$ notation key list

# Sign a container image using the specified key name
$ notation sign --key <key_name> <registry>/<repository>@<digest>
```

### Sign an OCI artifact identified by a tag

```shell
# Prerequisites: 
# A default signing key is configured using CLI "notation key"

# Use a tag to identify a container image
$ notation sign <registry>/<repository>:<tag>
```

An example for a successful signing:

```shell
$ notation sign localhost:5000/net-monitor:v1
Warning: Tag is used. Always use digest to identify the reference uniquely and immutably.
Resolve tag "v1" to digest "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
Sign succeeded. Signature has been attached to localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

### Sign an OCI artifact with debug option

```shell
notation sign --debug <registry>/<repository>@<digest>
```

An example for a successful signing:

```shell
$ notation sign --debug localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
Use signature format jws.
Sign succeeded. Signature has been attached to localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```
