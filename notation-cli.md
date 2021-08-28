# Notation CLI

A set of commands the `notation` cli sill support.

## `notation` Root Commands

```bash
notation --help
NAME:
   notation - Commands for signing and verifying Artifacts stored within an OCI Artifact Registry

USAGE:
   notation command [command options] [arguments...]

COMMANDS:
   cert     Commands for managing certificates
   key      Commands for managing private keys
   config   Commands for configuring notation
   verify   Commands for verifying an artifacts signature integrity

OPTIONS:
   --help, -h  show help (default: false)
```

## `notation cert` Sub Commands

```
notation cert --help

NAME:
   notation cert - Commands for managing certificates

USAGE:
   notation cert command [command options] [arguments...]

COMMANDS:
   add, a      Commands for adding certificates
   remove, rm  Commands for removing certificates
   create      Create a self-signed certificate
   list, ls    List the concurrently configured certificates

OPTIONS:
   --help, -h  show help (default: false)
❯ notation cert ls
NAME                  PATH
wabbit-networks.io    /home/pat/.notary/keys/wabbit-networks.crt 
import-acme-rocket.io /home/pat/.notary/keys/import-acme-rockets.crt
```

### `notation cert add` Command

```
notation cert add --help
```
> TODO

### `notation cert remove` Command

```
notation cert remove --help
```
> TODO

### `notation cert create` Command

```
notation cert create --help
```
> TODO

### `notation cert` list Command

```
notation cert list --help
```
> TODO

## `notation key` Sub Commands

```
notation key --help

NAME:
   notation key - Commands for managing certificates

USAGE:
   notation cert command [command options] [arguments...]

COMMANDS:
   add, a      Commands for managing certificates
   remove, rm  Commands for managing private keys
   create      Create a self-signed certificate
   list, ls    List the concurrently configured certificates

OPTIONS:
   --help, -h  show help (default: false)
❯ notation cert ls
NAME                  PATH
wabbit-networks.io    /home/pat/.notary/keys/wabbit-networks.crt 
import-acme-rocket.io /home/pat/.notary/keys/import-acme-rockets.crt
```

## `notation verify` Sub Commands

```
notation verify --help

NAME:
   notation verify - Commands for verifying an artifacts signature integrity

USAGE:
   notation verify command [command options] [arguments...]
```
> TODO
