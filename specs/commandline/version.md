# notation version

## Description

Use `notation version` to print notation version information.

Upon successful execution, the following information is printed.

```text
Notation: Notary v2, A tool to sign, store, and verify artifacts.

Version:    <version_in_format_vX.Y.Z[suffix]>
Go version: <version in format goX.Y.Z>
Git commit: <commit_id_in_7_characters>
```

## Outline

```text
Print the notation version information

Usage:
  notation version [flags]

Flags:
  -h, --help          Help for version

```

## Usage

### Print notation version information

```shell
notation version
```

An example output:

```text
Notation: Notary v2, A tool to sign, store, and verify artifacts.

Version:    v1.0.0-rc.1
Go version: go1.19.1
Git commit: abcd123
```
