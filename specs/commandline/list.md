# notation list

## Description

Use `notation list` to list all the signatures of the signing artifact.

## Outline

```text
List all the signatures of signing artifacts

Usage:
  notation list <reference> [flags]

Aliases:
  list, ls

Flags:
  -h, --help              help for list
  -p, --password string   Password for registry operations (default to $NOTATION_PASSWORD if not specified)
  -u, --username string   Username for registry operations (default to $NOTATION_USERNAME if not specified)
```

## Usage

### List all the signatures of the signing container image

```text
notation list https://<registry>/<repository>:<tag>
```

Upon successful execution, the digests of signatures of signing container image are printed out.
