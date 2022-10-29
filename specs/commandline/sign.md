# notation sign

## Description

Use `notation sign` to sign artifacts.

Signs an OCI artifact that is stored in a registry. Upon successful signing, the generated signature is pushed to the registry with the digest of the OCI artifact returned.

## Outline

```text
Sign artifacts

Usage:
  notation sign [flags] <reference>

Flags:
  -e, --expiry duration          optional expiry that provides a "best by use" time for the artifact. The duration is specified in minutes(m), hours(h) or days(d). For example: 30d, 12h, 30m, 1d3h20m
  -h, --help                     help for sign
  -k, --key string               signing key name, for a key previously added to notation's key list.
  -p, --password string          password for registry operations (default to $NOTATION_PASSWORD if not specified)
      --plain-http               registry access via plain HTTP
      --plugin-config strings    {key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values
      --signature-format string  signature envelope format, options: 'jws', 'cose' (default "jws")
  -u, --username string          username for registry operations (default to $NOTATION_USERNAME if not specified)
```

## Usage

### Sign a container image

```shell
# Add a key which uses a local private key and certificate, and make it a default signing key
notation key add --default --name <key_name> <key_path> <cert_path>

# Or change the default signing key to an existing signing key
notation key update --default <key_name>

# Sign a container image using the default signing key
notation sign <registry>/<repository>:<tag>

# Or using container image digests instead of tags. A container image digest uniquely and immutably identifies a container image.
notation sign <registry>/<repository>@<digest>
```

### Sign a container image using a remote key

```shell
# Prerequisites: 
# - A compliant signing plugin is installed in notation. See notation plugin documentation (https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details.
# - User creates keys and certificates in a 3rd party key provider (e.g. key vault, key management service). The signing plugin installed in previous step must support generating signatures using this key provider.

# Add a default signing key referencing the key identifier for the remote key, and the plugin associated with it.
notation key add --default --name <key_name> --plugin <plugin_name> --id <remote_key_id>

# sign a container image using a remote key
notation sign <registry>/<repository>:<tag>
```

### Sign an OCI artifact using the default signing key

```shell
# Prerequisites: 
# A default signing key is configured using CLI "notation key"

# Use a digest that uniquely and immutably identifies an OCI artifact.
notation sign <registry>/<repository>@<digest>
```

### Sign a container image and specify the signature expiry duration, for example 1 day

```shell
notation sign --expiry 1d <registry>/<repository>:<tag>
```

### Sign a container image using a specified signing key

```shell
# List signing keys to get the key name
notation key list

# Sign a container image using the specified key name
notation sign --key <key_name> <registry>/<repository>:<tag>
```
