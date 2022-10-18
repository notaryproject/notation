# notation logout

## Description

Use `notation logout` to log out from an OCI registry.

## Outline

```text
Log out from an OCI registry

Usage:
  notation logout [flags] [server]

Flags:
      --all           log out from all logged in registries
  -h, --help          help for logout
      --plain-http    registry access via plain HTTP
```

## Usage

### Log out from an OCI registry

```shell
notation logout registry.example.com
```

### Log out from all logged in registries

```shell
notation logout --all
```
