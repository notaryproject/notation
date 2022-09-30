# notation sign

## Description

Use `notation sign` to sign artifacts.

Signs an OCI artifact that is stored in a registry. Upon successful signing, the generated signature is pushed to the registry with the digest of the OCI artifact returned.

## Outline

```text
Sign artifacts

Usage:
  notation sign <reference> [flags]

Flags:
      --cert-file string        Location of file containing a complete certificate chain for the signing key. Use this flag with '--key-file'.
  -e, --expiry duration         Optional expiry that provides a "best by use" time for the artifact. The duration is specified in minutes(m), hours(h) or days(d). For example: 30d, 12h, 30m, 1d3h20m
  -h, --help                    Help for sign
  -k, --key string              Signing key name, for a key previously added to notation's key list.
      --key-file string         Location of file containing signing key file. Use this flag with '--cert-file'.
  -p, --password string         Password or identity token for registry operations (default to $NOTATION_PASSWORD if not specified)
      --plugin-config strings   List of {key}={value} pairs that are passed as is to a plugin, if the key (--key) is associated with a signing plugin, refer plugin documentation to set appropriate values
  -u, --username string         Username for registry operations (default to $NOTATION_USERNAME if not specified)

Global Flags:
      --plain-http   Registry access via plain HTTP
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
notation sign <registry>/<repository>:<tag> --key <key_name>
```

### Sign a container image using a local key and certificate which are not added in the signing key list

```shell
notation sign <registry>/<repository>:<tag> --key-file <key_path> --cert-file <cert_path>
```
