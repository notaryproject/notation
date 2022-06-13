# Notation CLI

The following spec outlines the notation CLI.
The CLI commands are what's currently available in [notation v0.7.1-alpha.1](https://github.com/notaryproject/notation/releases/tag/v0.7.1-alpha.1)


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