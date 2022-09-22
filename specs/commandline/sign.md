# notation sign
## Description
Use `notation sign` to sign artifacts.

If the container image being signed is stored in the registry, upon successful signing, the generated signature will be pushed to the registry and the digest of the container image will be returned.
## Outline
```console
Sign artifacts

Usage:
  notation sign <reference> [flags]

Flags:
      --cert-file string        Location of file containing signing(leaf) certificate and certificate chain
  -e, --expiry duration         Expire duration in seconds, minutes or hours
  -h, --help                    Help for sign
  -k, --key string              Signing key name
      --key-file string         Signing key file
  -p, --password string         Password for registry operations (default from $NOTATION_PASSWORD)
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
notation sign <image>
```
### Sign a container image using a remote key
```console
# Prerequisites: 
# - A Key Vault plugin is installed in notation
# - User creates keys and certificates in a Key Vault
# Add a default signing key referencing the key stored in the Key Vault
notation key add -n <key name> --plugin <plugin name> --id <remote key id> --default

# sign a container image using a remote key
notation sign <image>
```
### Sign a container image and specify the signature expiry duration, for example 24 hours
```console
notation sign <image> --expiry 24h
```
### Sign a container image using a specified signing key
```console
# List signing keys to get the key name
notation key list

# Sign a container image using the specified key name
notation sign <image> --key <key name>
```
### Sign a container image using a local key and certificate which are not added in the signing key list
```console
notation sign <image> --key-file <key path> --cert-file <cert path>
```