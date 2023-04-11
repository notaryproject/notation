# notation list

## Description

Use `notation list` to list all the signatures associated with signed artifact.

`Tags` are mutable, but `Digests` uniquely and immutably identify an artifact. If a tag is used to identify a signed artifact, notation resolves the tag to the `digest` first.

Upon successful execution, both the digest of the signed artifact and the digests of signatures manifest associated with signed artifact are printed out as following:

```shell
<registry>/<repository>@<digest>
└── application/vnd.cncf.notary.signature
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
  -d, --debug             debug mode
  -h, --help              help for list
      --oci-layout        [Experimental] list signatures stored in OCI image layout
  -p, --password string   password for registry operations (default to $NOTATION_PASSWORD if not specified)
      --plain-http        registry access via plain HTTP
  -u, --username string   username for registry operations (default to $NOTATION_USERNAME if not specified)
  -v, --verbose           verbose mode
```

## Usage

### List all the signatures of the signed container image

```shell
notation list <registry>/<repository>:<tag>
```

An example output:

```shell
localhost:5000/net-monitor:v1
└── application/vnd.cncf.notary.signature
    ├── sha256:647039638efb22a021f59675c9449dd09956c981a44b82c1ff074513c2c9f273
    └── sha256:6bfb3c4fd485d6810f9656ddd4fb603f0c414c5f0b175ef90eeb4090ebd9bfa1
```

### [Experimental] List all the signatures associated with the image in OCI layout directory

The following example lists the signatures associated with the image in OCI layout directory named `hello-world`. To access this flag `--oci-layout` , set the environment variable `NOTATION_EXPERIMENTAL=1`.

Reference an image in OCI layout directory using tags:

```shell
NOTATION_EXPERIMENTAL=1 notation list --oci-layout hello-world:v1
```

Reference an image in OCI layout directory using exact digest:

```shell
NOTATION_EXPERIMENTAL=1 notation list --oci-layout hello-world@sha256:xxx
```

An example output:

```shell
hello-world@sha256:a08753c0c7bcdaaf5c2fdb375f68e860c34bffb146368982c201d41769e1763c
└── application/vnd.cncf.notary.signature
    ├── sha256:647039638efb22a021f59675c9449dd09956c981a44b82c1ff074513c2c9f273
    └── sha256:6bfb3c4fd485d6810f9656ddd4fb603f0c414c5f0b175ef90eeb4090ebd9bfa1
```
