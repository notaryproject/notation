# notation login

## Description

Use `notation login` to log in to an OCI registry.

## Outline

```text
Log in to an OCI registry

Usage:
  notation login <server> [flags]

Flags:
  -h, --help              help for login
  -p, --password string   Password for registry operations (default to $NOTATION_PASSWORD if not specified)
      --password-stdin    Take the password from stdin
  -u, --username string   Username for registry operations (default to $NOTATION_USERNAME if not specified)
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
