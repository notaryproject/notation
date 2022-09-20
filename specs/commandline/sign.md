---
title: "notation sign"
description: "The notation sign command description and usage"
keywords: "notation, sign"
---
# notation sign
## Description
Use `notation sign` to sign artifacts.

If the signing artifact is a container image stored in a registry, the signature is pushed to the registry by default after signing successfully, and the digest of the container image is returned.
## Outline
```console
$ notation sign --help
Sign artifacts

Usage:
  notation sign [reference] [flags]

Flags:
      --cert-file string        signing certificate file
  -e, --expiry duration         expire duration
  -h, --help                    help for sign
  -k, --key string              signing key name
      --key-file string         signing key file
  -l, --local                   reference is a local file
      --media-type string       specify the media type of the manifest read from file or stdin (default "application/vnd.docker.distribution.manifest.v2+json")
  -o, --output string           write signature to a specific path
  -p, --password string         Password for registry operations (default from $NOTATION_PASSWORD)
  -c, --pluginConfig string     list of comma-separated {key}={value} pairs that are passed as is to the plugin, refer plugin documentation to set appropriate values
      --push                    push after successful signing (default true)
  -r, --reference string        original reference
  -u, --username string         Username for registry operations (default from $NOTATION_USERNAME)

Global Flags:
      --plain-http   Registry access via plain HTTP
```
## Usage
### sign a container image with a local key and certificate
```console
notation sign <image> --key-file <key path> --cert-file <cert path> 
```
### sign a container image using a key name
```console
# Add a key with a key name referencing signing key file
notation key add -n <key name> <key path> <cert path> 

# sign a container image using a key name
notation sign <image> --key <key name>
```
### sign a container image with key and certificate stored in a Key Vault
```console
# Pre-condition: 
# - A Key Vault plugin is installed in notation
# - User creates keys and certificates in a Key vault
# Add the key with a key name referencing the key stored in Key Vault
notation key add -n <key name> --plugin <plugin name> --id <key id>

# sign a container image using a key name
notation sign <image> --key <key name>
```
### store signature in a local file
```console
# disable auto push and store signature in a specified file
notation sign <image> --key <key name> --push false -o <signature file>
```
### sign a local file and store signature in a specified file
```console
notation sign -l <local file> --key <key name> -o <signature file>
```