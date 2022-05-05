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