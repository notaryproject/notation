# Notation CLI

The following spec outlines the notation CLI.
The CLI commands are what's currently available in [notation v0.7.1-alpha.1](https://github.com/notaryproject/notation/releases/tag/v0.7.1-alpha.1). The CLI experience in alpha.1 does not represent the final user experience, and CLI commands may have breaking changes before RC release as the CLI experience is finalized.

## Table of Contents
- [notation](#notation)
- [sign](#sign): Signs artifacts
- [verify](#verify): Verifies OCI Artifacts
- [push](#push): Push signature to remote
- [pull](#pull): Pull signatures from remote
- [list](#list): List signatures from remote
- [certificate](#certificate): Manage certificates used for verification
- [key](#key): Manage keys used for signing
- [cache](#cache): Manage signature cache
- [plugin](#plugin): Manage KMS plugins

## notation

```bash
notation help
NAME:
   notation - Notation - Notary V2

USAGE:
   notation [global options] command [command options] [arguments...]

VERSION:
   0.0.0-SNAPSHOT-17c7607

AUTHOR:
   CNCF Notary Project

COMMANDS:
   sign               Signs artifacts
   verify             Verifies OCI Artifacts
   push               Push signature to remote
   pull               Pull signatures from remote
   list, ls           List signatures from remote
   certificate, cert  Manage certificates used for verification
   key                Manage keys used for signing
   cache              Manage signature cache
   plugin             Manage KMS plugins
   help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## sign

```console
 notation sign --help
NAME:
   notation sign - Signs artifacts

USAGE:
   notation sign [command options] <reference>

OPTIONS:
   --key value, -k value        signing key name
   --key-file value             signing key file
   --cert-file value            signing certificate file
   --timestamp value, -t value  timestamp the signed signature via the remote TSA
   --expiry value, -e value     expire duration (default: 0s)
   --reference value, -r value  original reference
   --local, -l                  reference is a local file (default: false)
   --output value, -o value     write signature to a specific path
   --push                       push after successful signing (default: true)
   --push-reference value       different remote to store signature
   --username value, -u value   username for generic remote access [$NOTATION_USERNAME]
   --password value, -p value   password for generic remote access [$NOTATION_PASSWORD]
   --plain-http                 remote access via plain HTTP (default: false)
   --media-type value           specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
   --help, -h                   show help (default: false)
```

## verify

```console
notation verify --help
NAME:
   notation verify - Verifies OCI Artifacts

USAGE:
   notation verify [command options] <reference>

OPTIONS:
   --signature value, -s value, -f value  signature files                     (accepts multiple inputs)
   --cert value, -c value                 certificate names for verification  (accepts multiple inputs)
   --cert-file value                      certificate files for verification  (accepts multiple inputs)
   --pull                                 pull remote signatures before verification (default: true)
   --local, -l                            reference is a local file (default: false)
   --username value, -u value             username for generic remote access [$NOTATION_USERNAME]
   --password value, -p value             password for generic remote access [$NOTATION_PASSWORD]
   --plain-http                           remote access via plain HTTP (default: false)
   --media-type value                     specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
   --help, -h                             show help (default: false)
```

## push

```console
notation push --help
NAME:
   notation push - Push signature to remote

USAGE:
   notation push [command options] <reference>

OPTIONS:
   --signature value, -s value, -f value  signature files  (accepts multiple inputs)
   --username value, -u value             username for generic remote access [$NOTATION_USERNAME]
   --password value, -p value             password for generic remote access [$NOTATION_PASSWORD]
   --plain-http                           remote access via plain HTTP (default: false)
   --help, -h                             show help (default: false)
```

## pull

```console
notation pull --help
NAME:
   notation pull - Pull signatures from remote

USAGE:
   notation pull [command options] <reference>

OPTIONS:
   --strict                    pull the signature without lookup the manifest (default: false)
   --output value, -o value    write signature to a specific path
   --username value, -u value  username for generic remote access [$NOTATION_USERNAME]
   --password value, -p value  password for generic remote access [$NOTATION_PASSWORD]
   --plain-http                remote access via plain HTTP (default: false)
   --help, -h                  show help (default: false)
```

## list

```console
notation list --help
NAME:
   notation list - List signatures from remote

USAGE:
   notation list [command options] <reference>

OPTIONS:
   --username value, -u value  username for generic remote access [$NOTATION_USERNAME]
   --password value, -p value  password for generic remote access [$NOTATION_PASSWORD]
   --plain-http                remote access via plain HTTP (default: false)
   --help, -h                  show help (default: false)
```

## certificate

```console
notation certificate --help
NAME:
   notation certificate - Manage certificates used for verification

USAGE:
   notation certificate command [command options] [arguments...]

COMMANDS:
   add            Add certificate to verification list
   list, ls       List certificates used for verification
   remove, rm     Remove certificate from the verification list
   generate-test  Generates a test RSA key and a corresponding self-signed certificate
   help, h        Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

## key

```console
notation key --help
NAME:
   notation key - Manage keys used for signing

USAGE:
   notation key command [command options] [arguments...]

COMMANDS:
   add          Add key to signing key list
   update, set  Update key in signing key list
   list, ls     List keys used for signing
   remove, rm   Remove key from signing key list
   help, h      Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```
## cache

```console
 notation cache --help
NAME:
   notation cache - Manage signature cache

USAGE:
   notation cache command [command options] [arguments...]

COMMANDS:
   list, ls    List signatures in cache
   prune       Prune signature from cache
   remove, rm  Remove signature from cache
   help, h     Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

## plugin

```console
notation plugin --help
NAME:
   notation plugin - Manage plugins

USAGE:
   notation plugin command [command options] [arguments...]

COMMANDS:
   list     List registered plugins
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```