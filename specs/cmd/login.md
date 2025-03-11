# notation login

## Description

Use `notation login` to log in to an OCI registry. Users can execute `notation login` multiple times to log in multiple registries.

## Outline

```text
Log in to an OCI registry

Usage:
  notation login [flags] <server>

Flags:
  -d, --debug               debug mode
  -h, --help                help for login
      --insecure-registry   use HTTP protocol while connecting to registries. Should be used only for testing
  -p, --password string     password for registry operations (default to $NOTATION_PASSWORD if not specified)
      --password-stdin      take the password from stdin
  -u, --username string     username for registry operations (default to $NOTATION_USERNAME if not specified)
  -v, --verbose             verbose mode
```

## Usage

### Log in with provided username and password

```shell
notation login -u <username> -p <password> registry.example.com
```

### Log in using $NOTATION_USERNAME $NOTATION_PASSWORD variables

```shell
# Prerequisites:
# set environment variable NOTATION_USERNAME and NOTATION_PASSWORD
notation login registry.example.com
```
