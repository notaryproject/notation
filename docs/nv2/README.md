# Notary V2 (nv2) - Prototype

`nv2` is a command line tool for signing and verifying [OCI Artifacts]. This implementation supports `x509` signing mechanisms.

## Table of Contents

- [Prerequisites](#prerequisites)
- [CLI Overview](#cli-overview)
- [Offline signing & verification](#offline-signing-and-verification)

## Prerequisites

### Build and Install

This plugin requires [golang](https://golang.org/dl/) with version `>= 1.14`.

To build and install, run

```shell
go install github.com/notaryproject/nv2/cmd/nv2
```

To build and install to an optional path, run

```shell
go build -o nv2 ./cmd/nv2
```

Next, install optional components:

- Install [docker-generate](https://github.com/shizhMSFT/docker-generate) for local Docker manifest generation and local signing.
- Install [OpenSSL](https://www.openssl.org/) for key generation.

### Self-signed certificate key generation

To generate a `x509` self-signed certificate key pair `example.key` and `example.crt`, run

``` shell
openssl req \
  -x509 \
  -sha256 \
  -nodes \
  -newkey rsa:2048 \
  -days 365 \
  -subj "/CN=registry.example.com/O=example inc/C=US/ST=Washington/L=Seattle" \
  -keyout example.key \
  -out example.crt
```

When generating the certificate, make sure that the Common Name (`CN`) is set properly in the `Subject` field. The Common Name will be verified against the registry name within the signature.

## Offline Signing

Offline signing is accomplished with the `nv2 sign` command.

### nv2 sign options

```shell
NAME:
   nv2 sign - signs OCI Artifacts

USAGE:
   nv2 sign [command options] [<scheme://reference>]

OPTIONS:
   --method value, -m value     signing method
   --key value, -k value        signing key file [x509]
   --cert value, -c value       signing cert [x509]
   --expiry value, -e value     expire duration (default: 0s)
   --reference value, -r value  original references
   --output value, -o value     write signature to a specific path
   --username value, -u value   username for generic remote access
   --password value, -p value   password for generic remote access
   --insecure                   enable insecure remote access (default: false)
   --help, -h                   show help (default: false)
```

Signing and verification are based on [OCI manifests](https://github.com/opencontainers/image-spec/blob/master/manifest.md), [docker-generate](https://github.com/shizhMSFT/docker-generate) is used to generate the manifest, which is exactly the same manifest as the `docker push` produces.

### Generating a manifest

Notary v2 signing is accomplished by signing the OCI manifest representing the artifact. When building docker images, the manifest is not generated until the image is pushed to a registry. To accomplish offline/local signing, the manifest must first exist.

- Build the hello-world image

  ``` shell
  docker build \
    -f Dockerfile.build \
    -t registry.acme-rockets.io/hello-world:v1 \
    https://github.com/docker-library/hello-world.git
  ```

- Generate a manifest, saving it as `hello-world_v1-manifest.json`

  ``` shell
  docker generate manifest registry.acme-rockets.io/hello-world:v1 > hello-world_v1-manifest.json
  ```

### Signing using `x509`

To sign the manifest `hello-world_v1-manifest.json` using the key `key.pem` from the `x509` certificate `cert.pem` with the Common Name `registry.acme-rockets.io`, run

```shell
nv2 sign --method x509 \
  -k key.key \
  -c cert.crt \
  -r registry.acme-rockets.io/hello-world:v1 \
  -o hello-world.signature.config.json \
  file:hello-world_v1-manifest.json
```

The formatted x509 signature: `hello-world.signature.config.json` is:

``` json
{
    "signed": {
        "digest": "sha256:5de47f48e0be1a9d41176a980728449a696fd4fcc37e9d99b8d26618c0f5bf51",
        "size": 3056,
        "references": [
            "registry.acme-rockets.io/hello-world:v1"
        ],
        "iat": 1596020554
    },
    "signature": {
        "typ": "x509",
        "sig": "vUNmuwrdHmcMyvG//eZQLjmIz2gnOUFNaL5Y5Jc3x1oaYu3nFnJxBEkB8232l0zBmV30sVUX2vjao0IDgLMv0Q7VWT2hiTutocgf+oRq88Jz/xKGvByGUWmVyYx9sMW6R+JHK/LlzthCLgDoYTjFD9qDTHf+AWnmRNPLv5nSYNQrVSxNH22jiO3CV/bNEQD8xoR7kZOdov6QzNw3rAP+XvlKxdf/D7vcYdR0D5T9G5xGa72aQSZmzXL/Zd2V7JQnxyJmw6PL3moU1i/8t8RK7LbsU6slvTScLUokFLZxzqCz8TcjujtaThyyxPF47ekx/HVsKW0mYXidpgCOfl+nqw==",
        "alg": "RS256",
        "x5c": [
            "MIIDJzCCAg+gAwIBAgIUMwVg7bpx8QmWaFzRcgpRFBN6JoQwDQYJKoZIhvcNAQELBQAwIzEhMB8GA1UEAwwYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMB4XDTIwMDcyOTExMDIzMloXDTIxMDcyOTExMDIzMlowIzEhMB8GA1UEAwwYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAx2mXqcXqkllwxj7S12WhVDsIu6y4ebZ/CwVwwime44yDcd0bcpdJExqIH/Qy6axQd/1zmLCHPeOXGFq48Ul0oS4Bawj1GEeLvB7VFvqB0KaBeAdxrZAvdKXCXIDH5qyFSGnOmvkja1BuR8XrH7tts5u56i+U3KEDBZg5tfx4cQuKKt0DfXZAL+4RZkNh1LoO77X0ThaBThFoRsg6aZA/cEpttoWmvnO6uUkK73oZEVgZNKGGIZZKzhUjnydRSTphp9GmZzbqUHlOiMvbzdtsQYC0qeQeNqua38HN93Ur3p+oH7oSrBWxX1Xlx933oVb+4G6h5oz0aZvMQ0G6gCLzjwIDAQABo1MwUTAdBgNVHQ4EFgQU8l2F7avSjFZ9TvnpHackunxSFcswHwYDVR0jBBgwFoAU8l2F7avSjFZ9TvnpHackunxSFcswDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAwECYhttcbCbqyi7DvOTHw5bixmxplbgD0AmMvE6Ci4P/MrooBququlkri/Jcp58GBaMjxItE4qVsaWwFCEvZEfP2xN4DAbr+rdrIFy9VYuwEIBs5l0ZLRH2H2N3HlqBzhYOjVzNlYfIqnqHUDip2VsUKqhcVFkCmb3cpJ1VNAgjQU2N60JUW28L0XrGyBctBIiicLvdP4NMhHP/hhN2vr2VGIyyo5XtP+QHFi/Uwa48BJ+c9bbVpXeghOMOPMeSJmJ2b/qlp95e/YHlSCfxDXyxZ70N2vBGecrc8ly4tD9KGLb9y3Q7RBgsagOFe7cGQ2db/t60AwTIxP0a9bIyJMg=="
        ]
    }
}
```

If the embedded cert chain `x5c` is not desired, it can be replaced by a key ID `kid` by omitting the `-c` option.

```shell
nv2 sign -m x509 \
  -k key.key \
  -r registry.acme-rockets.io/hello-world:v1 \
  -o hello-world.signature.config.json \
  file:hello-world_v1-manifest.json
```

The formatted x509, without the `x5c` chain signature: `hello-world.signature.config.json` is:


```json
{
    "signed": {
        "digest": "sha256:5de47f48e0be1a9d41176a980728449a696fd4fcc37e9d99b8d26618c0f5bf51",
        "size": 3056,
        "references": [
            "registry.acme-rockets.io/hello-world:v1"
        ],
        "iat": 1596020616
    },
    "signature": {
        "typ": "x509",
        "sig": "OyRPlwwsO5mYDxKkiNeTQlSl4WV8SOiQMCJv4i1+sx7uv6Pe8dHDaPt1SE5s64HzFvo6s26PrfiPYp4RphQOd/KvW2Hh03nS8ZByE4NWFOE6VLQcfNpScba6Q9vAzc3TnZrg1c9t992MGuec1oZB9pR77Ms7Jv/+gZd1qr6VPpA0A6+UucEbN6+pKRTiPRx5WkFXTkN0a4jmlJnev6MyBY3VI0EzjLI4nbCu9P05e4SK1dO0hXtD7aQCf2CCVKdYNHAMX4pNPTLxS3a5p4CFjV3oCbZO6cYT/5ZxgQrVV7vaGEI1MGCOEXS2KSI14zO6KlU1awtOQq3g04e03O+SVQ==",
        "alg": "RS256",
        "kid": "RQGT:OPJI:IABT:DFXB:52VS:FNOJ:4XBS:H4KY:WHGM:HQMC:WSMN:LKXM"
    }
}
```

The detailed signature specification is [available](../signature/README.md).

### Offline Verification

Notary v2 verification can be accomplished with the `nv2 verify` command.

```shell
NAME:
   nv2 verify - verifies OCI Artifacts

USAGE:
   nv2 verify [command options] [<scheme://reference>]

OPTIONS:
   --signature value, -s value, -f value  signature file
   --cert value, -c value                 certs for verification [x509]
   --ca-cert value                        CA certs for verification [x509]
   --username value, -u value             username for generic remote access
   --password value, -p value             password for generic remote access
   --insecure                             enable insecure remote access (default: false)
   --help, -h                             show help (default: false)
```

To verify a manifest `example.json` with a signature file `example.nv2`, run

Since the manifest was signed by a self-signed certificate, that certificate `cert.pem` is required to be provided to `nv2`.

```shell
nv2 verify \
  -f hello-world.signature.config.json \
  -c cert.crt \
  file:hello-world_v1-manifest.json
```

If the cert isn't self-signed, you can omit the `-c` parameter.

``` shell
nv2 verify \
  -f hello-world.signature.config.json \
  file:hello-world_v1-manifest.json

sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

On successful verification, the `sha256` digest of the manifest is printed. Otherwise, `nv2` prints error messages and returns non-zero values.

The command `nv2 verify` takes care of all signing methods.

## Remote Manifests

With `nv2`, it is also possible to sign and verify a manifest or a manifest list in a remote registry where the registry can be a docker registry or an OCI registry.

### Docker Registry

Here is an example to sign and verify the image `hello-world` in DockerHub, i.e. `docker.io/library/hello-world:latest`, using `x509`.

``` shell
nv2 sign -m x509 \
  -k key.key \
  -o hello-world_latest.signature.config.json \
  docker://docker.io/library/hello-world:latest

sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202

nv2 verify \
  -c cert.crt \
  -f hello-world_latest.signature.config.json \
  docker://docker.io/library/hello-world:latest
sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202
```

It is possible to use `digest` in the reference. For instance:

``` shell
docker.io/library/hello-world@sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202
```

If neither `tag` nor `digest` is specified, the default tag `latest` is used.

### OCI Registry

OCI registry works the same as Docker but with the scheme `oci`.


``` shell
nv2 sign -m x509 \
  -k key.key \
  -o hello-world_latest.signature.config.json \
  oci://docker.io/library/hello-world:latest

sha256:0ebe6f409b373c8baf39879fccee6cae5e718003ec3167ded7d54cb2b5da2946

nv2 verify \
  -c cert.crt \
  -f hello-world_latest.signature.config.json \
  oci://docker.io/library/hello-world:latest

sha256:0ebe6f409b373c8baf39879fccee6cae5e718003ec3167ded7d54cb2b5da2946
```

**Note** The digest of the OCI manifest is different from the Docker manifest for the same image since their format is different. Therefore, the signer should be careful with the manifest type when signing.

### Insecure Registries

To sign and verify images from insecure registries accessed via `HTTP`, such as `localhost`, the option `--insecure` is required.

``` shell
docker tag example localhost:5000/example
docker push localhost:5000/example
The push refers to repository [localhost:5000/example]
50644c29ef5a: Pushed
latest: digest: sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55 size: 528
nv2 verify -f example.nv2 --insecure docker://localhost:5000/example

sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

### Secure Image Pulling

Since the tag might be changed during the verification process, it is required to pull by digest after verification.

```shell
digest=$(nv2 verify -f hello-world_latest.signature.config.json -c cert.crt docker://docker.io/library/hello-world:latest)
if [ $? -eq 0 ]; then
    docker pull docker.io/library/hello-world@$digest
fi
```
