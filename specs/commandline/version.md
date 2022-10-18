# notation version

## Description

Use `notation version` to print notation version information.

Upon successful execution, the following information is printed.

```text
Notation: Notary v2, A tool to sign, store, and verify artifacts.

Version:     <MAJOR.MINOR.PATCH>
Go version:  go<MAJOR.MINOR.PATCH>
Git commit:  <long_hash>
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

Version:     1.0.0
Go Version:  go1.19.2
Git commit:  1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a
```
