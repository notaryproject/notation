# notation list

## Description

Use `notation list` to list all the signatures associated with signed artifact.

`Tags` are mutable, but `Digests` uniquely and immutably identify an artifact. If a tag is used to identify a signed artifact, notation resolves the tag to the `digest` first.

Upon successful execution, both the digest of the signed artifact and the digests of signatures manifest associated with signed artifact are printed out as following:

```shell
<registry>/<repository>@<digest>
└── application/vnd.cncf.notary.v2.signature
    ├──<digest_of_signature_manifest>
    └──<digest_of_signature_manifest>
```

## Outline

```text
List all the signatures associated with signed artifact

Usage:
  notation list [flags] <reference>

Aliases:
  list, ls

Flags:
  -h, --help              help for list
  -p, --password string   password for registry operations (default to $NOTATION_PASSWORD if not specified)
      --plain-http        registry access via plain HTTP
  -u, --username string   username for registry operations (default to $NOTATION_USERNAME if not specified)
```

## Usage

### List all the signatures of the signed container image

```text
notation list <registry>/<repository>:<tag>
```

An example output:

```shell
localhost:5000/net-monitor:v1
└── application/vnd.cncf.notary.v2.signature
    ├── sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    └── sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
```
