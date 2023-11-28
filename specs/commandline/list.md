# notation list

## Description

Use `notation list` to list all the signatures associated with a signed OCI artifact.

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
List all the signatures associated with a signed OCI artifact

Usage:
  notation list [flags] <reference>

Aliases:
  list, ls

Flags:
      --allow-referrers-api   [Experimental] use the Referrers API to list signatures, if not supported (returns 404), fallback to the Referrers tag schema
  -d, --debug                 debug mode
  -h, --help                  help for list
      --insecure-registry     use HTTP protocol while connecting to registries. Should be used only for testing
      --max-signatures int    maximum number of signatures to evaluate or examine (default 100)
      --oci-layout            [Experimental] list signatures stored in OCI image layout
  -p, --password string       password for registry operations (default to $NOTATION_PASSWORD if not specified)
  -u, --username string       username for registry operations (default to $NOTATION_USERNAME if not specified)
  -v, --verbose               verbose mode
```

## Usage

### List all the signatures of the signed container image

```shell
notation list <registry>/<repository>:<tag>
```

An example output:

```shell
localhost:5000/net-monitor@sha256:8456f085dd609fd12cdebc5f80b6f33f25f670a7a9a03c8fa750b8aee0c4d657
└── application/vnd.cncf.notary.signature
    ├── sha256:647039638efb22a021f59675c9449dd09956c981a44b82c1ff074513c2c9f273
    └── sha256:6bfb3c4fd485d6810f9656ddd4fb603f0c414c5f0b175ef90eeb4090ebd9bfa1
```

### [Experimental] List all the signatures associated with the image in OCI layout directory

The following example lists the signatures associated with the image in OCI layout directory named `hello-world`. To access this flag `--oci-layout` , set the environment variable `NOTATION_EXPERIMENTAL=1`.

Reference an image in OCI layout directory using tags:

```shell
export NOTATION_EXPERIMENTAL=1
# Assume OCI layout directory hello-world is under current path
notation list --oci-layout hello-world:v1
```

Reference an image in OCI layout directory using exact digest:

```shell
export NOTATION_EXPERIMENTAL=1
# Assume OCI layout directory hello-world is under current path
notation list --oci-layout hello-world@sha256:xxx
```

An example output:

```shell
hello-world@sha256:a08753c0c7bcdaaf5c2fdb375f68e860c34bffb146368982c201d41769e1763c
└── application/vnd.cncf.notary.signature
    ├── sha256:647039638efb22a021f59675c9449dd09956c981a44b82c1ff074513c2c9f273
    └── sha256:6bfb3c4fd485d6810f9656ddd4fb603f0c414c5f0b175ef90eeb4090ebd9bfa1
```
