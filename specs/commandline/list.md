# notation list

## Description

Use `notation list` to list all the signatures of the signed artifact.

## Outline

```text
List all the signatures of signed artifacts

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

### List all the signatures of the signed container image

```text
notation list <registry>/<repository>:<tag>
```

Upon successful execution, the digests of signatures of signed container image are printed out.
