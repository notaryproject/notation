# notation sign
## Description
Use `notation sign` to sign artifacts.

Signs a container artifact that is stored in a registry. Upon successful signing, the generated signature is pushed to the registry and the digest of the container image is returned.
## Outline
```console
Sign artifacts

Usage:
  notation sign <reference> [flags]

Flags:
      --cert-file string        Location of file containing signing(leaf) certificate and certificate chain. Use this flag with flag '--key-file' together.
  -e, --expiry duration         Optional expiry that provides a “best by use” time for the artifact. The duration is specified in seconds, minutes or hours.
  -h, --help                    Help for sign
  -k, --key string              Signing key name, for a key previously added to notation's key list.
      --key-file string         Signing key file. Use this flag with flag '--cert-file' together.
  -p, --password string         Password for registry operations (default to $NOTATION_PASSWORD if not specified)
  -c, --pluginConfig string     List of comma-separated {key}={value} pairs that are passed as is to the plugin, refer plugin documentation to set appropriate values
  -u, --username string         Username for registry operations (default from $NOTATION_USERNAME)

Global Flags:
      --plain-http   Registry access via plain HTTP
```
## Usage
### Sign a container image
```console
# Add a key and make it a default signing key
notation key add -n <key name> <key path> <cert path> --default

# [Optional] Change a default signing key
notation key update <key name> --default

# Sign a container image using the default signing key
notation sign https://<registry>/<repository>:<tag>
```
### Sign a container image using a remote key
```console
# Prerequisites: 
# - A compliant signing plugin is installed in notation. See notation plugin documentation for more details (https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md).
# - User creates keys and certificates in a 3rd party key provider (e.g. key vault, key management service). The signing plugin installed in previous step must support generating signatures using this key provider.

# Add a default signing key referencing the key identifier for the remote key, and the plugin associated with it.
notation key add -n <key name> --plugin <plugin name> --id <remote key id> --default

# sign a container image using a remote key
notation sign https://<registry>/<repository>:<tag>
```
### Sign a container image and specify the signature expiry duration, for example 1 day
```console
notation sign https://<registry>/<repository>:<tag> --expiry 24h
```
### Sign a container image using a specified signing key
```console
# List signing keys to get the key name
notation key list

# Sign a container image using the specified key name
notation sign https://<registry>/<repository>:<tag> --key <key name>
```
### Sign a container image using a local key and certificate which are not added in the signing key list
```console
notation sign https://<registry>/<repository>:<tag> --key-file <key path> --cert-file <cert path>
```