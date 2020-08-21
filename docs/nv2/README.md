# Notary V2 (nv2) - Prototype

`nv2` is a command line tool for signing and verifying [OCI Artifacts]. This implementation supports `x509` signing mechanisms.

## Table of Contents

- [Prerequisites](#prerequisites)
- [CLI Overview](#cli-overview)
- [Offline signing & verification](#offline-signing-and-verification)

## Prerequisites

### Build and Install

This binary requires [golang](https://golang.org/dl/) with version `>= 1.15`.

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
  -addext "subjectAltName=DNS:registry.example.com" \
  -keyout example.key \
  -out example.crt
```

When generating the certificate, make sure that the Common Name (`CN`) in the `Subject` field and the Subject Alternative Name (`subjectAltName`) are set properly.
The Common Name (`go < 1.15`) or the Subject Alternative Name (`go >= 1.15`) will be verified against the registry name within the signature.

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
   --reference value, -r value  original references
   --expiry value, -e value     expire duration (default: 0s)
   --output value, -o value     write signature to a specific path
   --media-type value           specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
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

To sign the manifest `hello-world_v1-manifest.json` using the key `key.key` from the `x509` certificate `cert.crt` with the Common Name `registry.acme-rockets.io`, run

```shell
nv2 sign --method x509 \
  -k key.key \
  -c cert.crt \
  -r registry.acme-rockets.io/hello-world:v1 \
  -o hello-world.signature.config.jwt \
  file:hello-world_v1-manifest.json
```

The formatted x509 signature: `hello-world.signature.config.jwt` is:

``` json
{
    "alg": "RS256",
    "typ": "x509",
    "x5c": [
        "MIID4DCCAsigAwIBAgIUJVfYMmyHvQ9u0TbluizZ3BtgATEwDQYJKoZIhvcNAQELBQAwbTEhMB8GA1UEAwwYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwHhcNMjAwODIwMDI0MDU2WhcNMjEwODIwMDI0MDU2WjBtMSEwHwYDVQQDDBhyZWdpc3RyeS5hY21lLXJvY2tldHMuaW8xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAN4qEOEICopQ9t1BSrh4Iu3f4LRMWI0p+P9Js92nhCQusn3wcEIddtb3J2eSy0d1MwTykXEMvSp3OU3+yJIm7BW3ZrCD+UjU+Je8JcDCXKZpKesLb8i42UAk7J35zWE/nRs9uzGQWelZTCJHBM1NVSnP4QqcGF2VkNwtsti7NbL4f+AunBMZvrK4p3kUwh92FoDgXen6+vrnqHb3MA8uJBa5pVsPOvcga13TWaYRfMm/nxM6xGvIwly2QpWfdJrC58aqEiRQAfUtfuD/MTYEQ0PL/fOpHJw/FcHLt0vkja1ggMexATinhvOPMhcJbL2JPliS1ExFtCjx2D+mtHjt9rECAwEAAaN4MHYwHQYDVR0OBBYEFME2lTAI0lfw6/2/SNNSydlMqWNzMB8GA1UdIwQYMBaAFME2lTAI0lfw6/2/SNNSydlMqWNzMA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RBBwwGoIYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMA0GCSqGSIb3DQEBCwUAA4IBAQB6HIusgEwUaQxcO9jm+iX4rgnGeDYT9qNdxtSL/tz7zzPTx2gPDSEzZinhFO+0AnH0kleiaMfJXFedvpX8xofP0zWNDgXAabZa1JR9HV42OmxMg/gBm3lpSQtnNoraqy6N88ot9xpRA0FQ/gAnGdRakrK7oeljDNpz3ay6ZgBqz3MpYIKkHL6dpvmQ4BbGEHfjLX1j/bC397XzapOcFqqhekc3Nk7vH51TheqQRIW1nI4BRo/guf6zjfxGcskTM4winCd/fk0F7XMlOddWleeg1vI+i/1TKV0p03aN23JZNUAt7MeZlz4nieaIGGFNinwOEsIRIXAZ65IMZwaLsCgY"
    ]
}.{
    "digest": "sha256:24a74900a4e749ef31295e5aabde7093e3244b119582bd6e64b1a88c71c410d0",
    "iat": 1597893535,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "references": [
        "registry.acme-rockets.io/hello-world:v1"
    ],
    "size": 3056
}.[Signature]
```

If the embedded cert chain `x5c` is not desired, it can be replaced by a key ID `kid` by omitting the `-c` option.

```shell
nv2 sign -m x509 \
  -k key.key \
  -r registry.acme-rockets.io/hello-world:v1 \
  -o hello-world.signature.config.jwt \
  file:hello-world_v1-manifest.json
```

The formatted x509, without the `x5c` chain signature: `hello-world.signature.config.jwt` is:


```json
{
    "alg": "RS256",
    "kid": "JF3F:UG7I:NCNR:3TCC:XNW3:3ZIW:S77O:O2PT:QXC3:IQ5X:GMMS:CUYB",
    "typ": "x509"
}.{
    "digest": "sha256:24a74900a4e749ef31295e5aabde7093e3244b119582bd6e64b1a88c71c410d0",
    "iat": 1597893598,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "references": [
        "registry.acme-rockets.io/hello-world:v1"
    ],
    "size": 3056
}.[Signature]
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
   --media-type value                     specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
   --username value, -u value             username for generic remote access
   --password value, -p value             password for generic remote access
   --insecure                             enable insecure remote access (default: false)
   --help, -h                             show help (default: false)
```

To verify a manifest `hello-world_v1-manifest.json` with a signature file `hello-world.signature.config.jwt`, run

```shell
nv2 verify \
  -f hello-world.signature.config.jwt \
  -c cert.crt \
  file:hello-world_v1-manifest.json
```

Since the manifest was signed by a self-signed certificate, that certificate `cert.crt` is required to be provided to `nv2`.

If the cert isn't self-signed, you can omit the `-c` parameter.

``` shell
nv2 verify \
  -f hello-world.signature.config.jwt \
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
  -o hello-world_latest.signature.config.jwt \
  docker://docker.io/library/hello-world:latest

sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202

nv2 verify \
  -c cert.crt \
  -f hello-world_latest.signature.config.jwt \
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
  -o hello-world_latest.signature.config.jwt \
  oci://docker.io/library/hello-world:latest

sha256:0ebe6f409b373c8baf39879fccee6cae5e718003ec3167ded7d54cb2b5da2946

nv2 verify \
  -c cert.crt \
  -f hello-world_latest.signature.config.jwt \
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
digest=$(nv2 verify -f hello-world_latest.signature.config.jwt -c cert.crt docker://docker.io/library/hello-world:latest)
if [ $? -eq 0 ]; then
    docker pull docker.io/library/hello-world@$digest
fi
```
