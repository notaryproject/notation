# Notation CLI

The following spec outlines the notation CLI.
The CLI commands are what's currently available in [notation v0.7.1-alpha.1](https://github.com/notaryproject/notation/releases/tag/v0.7.1-alpha.1). The CLI experience in alpha.1 does not represent the final user experience, and CLI commands may have breaking changes before RC release as the CLI experience is finalized.

## Table of Contents
- [notation](#notation): command group for signing and verification operations
- [certificate](#certificate): Manage certificates used for verification
- [key](#key): Manage keys used for signing
- [list](#list): List signatures from remote
- [login](#login): Provide credentials for authenticated registry operations
- [plugin](#plugin): Manage KMS plugins
- [sign](#sign): Signs artifacts
- [verify](#verify): Verifies OCI Artifacts

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
   certificate, cert  Manage certificates used for verification
   key                Manage keys used for signing
   list, ls           List signatures from remote
   login              Provide credentials for authenticated registry operations   
   plugin             Manage KMS plugins
   sign               Signs artifacts
   verify             Verifies OCI Artifacts
   help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
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

## list

```console
notation list --help
NAME:
   notation list - List signatures from remote

USAGE:
   notation list [command options] <reference>

OPTIONS:
   --username value, -u value    Username for registry operations [$NOTATION_USERNAME]
   --password value, -p value    Password for registry operations [$NOTATION_PASSWORD]
   --help, -h                    show help (default: false)

GLOBAL ARGUMENTS
   --plain-http                  Registry access via plain HTTP (default: false)
```

## login

```console
notation login --help
NAME:
   notation login - Provides credentials for authenticated registry operations

USAGE:
   notation login [options] [server]

OPTIONS:
   --username value, -u value    Username for registry operations [$NOTATION_USERNAME]
   --password value, -p value    Password for registry operations [$NOTATION_PASSWORD]
   --password-stdin              Take the password from stdin
   --help, -h                    Show help (default: false)

POSITIONAL
  <server>                       The registry URL for authentication

GLOBAL ARGUMENTS
   --plain-http                  Registry access via plain HTTP (default: false)

EXAMPLES
# Login with provided username and password
notation login -u <user> -p <password> registry.example.com

# Login using $NOTATION_USERNAME $NOTATION_PASSWORD variables
notation login registry.example.com

NOTES
Once login is completed, then -u -p is no longer required for any notation commands against the registry server authenticated.
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
   --media-type value           specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
   --username value, -u value   Username for registry operations [$NOTATION_USERNAME]
   --password value, -p value   Password for registry operations [$NOTATION_PASSWORD]
   --help, -h                   show help (default: false)

GLOBAL ARGUMENTS
   --plain-http                 Registry access via plain HTTP (default: false)
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
   --media-type value                     specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
   --username value, -u value             Username for registry operations [$NOTATION_USERNAME]
   --password value, -p value             Password for registry operations [$NOTATION_PASSWORD]
   --help, -h                             show help (default: false)

GLOBAL ARGUMENTS
   --plain-http                           Registry access via plain HTTP (default: false)
```
