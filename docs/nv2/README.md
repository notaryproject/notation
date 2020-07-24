# Notary V2 (nv2) - Prototype

`nv2` is a command line tool for signing and verifying [OCI Artifacts]. This implementation supports `x509` and `gpg` signing mechanisms.

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
     --method value, -m     siging method
     --key value, -k        siging key file [x509]
     --cert value, -c       siging cert [x509]
     --key-ring value       gpg public key ring file [gpg] (default: "/home/demo/.gnupg/secring.gpg")
     --identity value, -i   signer identity [gpg]
     --expiry value, -e     expire duration (default: 0s)
     --reference value, -r  original references
     --output value, -o     write signature to a specific path
     --username value, -u   username for generic remote access
     --password value, -p   password for generic remote access
     --insecure             enable insecure remote access (default: false)
     --help, -h             show help (default: false)
  ```

Signing and verification are based on [OCI manifests], [docker-generate](https://github.com/shizhMSFT/docker-generate) is used to generate the manifest, which is exactly the same manifest as the `docker push` produces.

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
  docker generate manifest hello-world:v1 > hello-world_v1-manifest.json
  ```

### Signing using `x509`

To sign the manifest `hello-world_v1-manifest.json` using the key `key.pem` from the `x509` certificate `cert.pem` with the Common Name `example.registry.io`, run

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
        "iat": 1595456071,
        "manifests": [
            {
                "digest": "sha256:407a722870b09ef1c037b3bd9d1e6fa828a1c64964ba8c292a8ebe4dcf3bde56",
                "size": 3056,
                "references": [
                    "registry.acme-rockets.io/hello-world:v1"
                ]
            }
        ]
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "BhGxUd+4pKkRvoVQFTx3XJ7P4IZxDHKFma6nIJEr3NehE53p5XBr03SCRQW4sa8Wr6IdRRXBVxXixy/QtdWKcXa6NjOP7b6reM8exJDOd6j9N/y/oH76MDONyibfGU8iA7zY0k6oqdLM7+pNlFv3V3eEGhpMx4ryVr7yUbg4g0swQr6TSdbUyKJGxVncg0RJuTZmeQ2VV+/uGGaN/ZkbYkmogK1Ji/8JvIjp14+99p/I2t388oqVTI9n8UUD0dm8F/7UMRzvbKfb23DTyFwZatLBXo4OP4zAWU1T+Zwp5urnqtJI/IU8x7qzC/1noNWGBEvK+/nd0avHRtTao+CtdA==",
            "alg": "RS256",
            "x5c": [
                "MIIDoDCCAoigAwIBAgIJAITsiynTSlpWMA0GCSqGSIb3DQEBCwUAMGUxCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApXYXNoaW5ndG9uMRAwDgYDVQQHDAdTZWF0dGxlMRUwEwYDVQQKDAxBQ01FIFJvY2tldHMxGDAWBgNVBAMMD2FjbWUtcm9ja2V0cy5pbzAeFw0yMDA3MjIyMjAxMjZaFw0yMTA3MjIyMjAxMjZaMGUxCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApXYXNoaW5ndG9uMRAwDgYDVQQHDAdTZWF0dGxlMRUwEwYDVQQKDAxBQ01FIFJvY2tldHMxGDAWBgNVBAMMD2FjbWUtcm9ja2V0cy5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAM5l9zOgzubTl/iLquIrVjNgM/7ZlUabsrisPtjN9d05T6FQQS8jYRuJN+XpTU6dWSP6AGf2bJCdX7TI04i/1Uci9TmzQrbp5aOCnIOIhOfX9W1TJ/7RMCw7BsROL8TVVDnMKJ8zde09svCZFDDzFpAbK0vYUnFb1+orlZ3wuALRw9VIxkZDBGrVE0UDqtnGbhw95V13Fiw4XMXN34bS/0alLnSOkTMMZbEXku54H4uNi9orcJ+rLvlvkFw2dQeSHmmHEqHnZkdQxs5HAky/4K2Eq/1DQhVi7Bg/YNC5IrNpw0picn5jqe3l8zjLpdUsVxgYN1G85DDqPDreah+EmCcCAwEAAaNTMFEwHQYDVR0OBBYEFE7L1GPDbahQusbLw3RPldzX5f0LMB8GA1UdIwQYMBaAFE7L1GPDbahQusbLw3RPldzX5f0LMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAJ7Ir7NgAmfaIhX1+oeGhz1hI+eY8EMSilGts+jIS8BOaBBrzbkVcnC4GKMnHQfXDNQUBW+gRML+ju2elSCrteZbbQU6UaO2er6pWXMQTFVw/nkK6lacNHGTGXbZOoKViAcRZkpVwdjKxmAhLDJcQJGwO+NKWf5WEo82HvwgaINvoEooe+NluN0CugQGdUhgJ+EkYx7jTa7XRpReH0aIsklzvjoakPBhCJ1xQ6VL3WV6zZCYwYUYVwpAMS8Gzo3aUhUwPS1W4mioRDqvJ8fKSkttNi8+N+pU65tKtAyRCvfl9KtJHQrgEqPwQYQ9bKt1H/7RI7oI4WoQ55iSgAU4Wyw="
            ]
        }
    ]
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
        "iat": 1595456231,
        "manifests": [
            {
                "digest": "sha256:407a722870b09ef1c037b3bd9d1e6fa828a1c64964ba8c292a8ebe4dcf3bde56",
                "size": 3056,
                "references": [
                    "acme-rockets.io/hello-world:v1"
                ]
            }
        ]
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "FyRRcAGi1qd0IHT8XoRh0vlGSkE4rjYpJzYEjodRQ2aUO0O/bBIzjV86UxqLnb1Y/GMU817YXeqHqDlySLWYoGKg+/aJGJDqbQpWIxUr6hhjGaxBYDZTt2ayzMAu5X/GNCx0vRLKl5dOOsgTO53QbuKEf4F3xxvQJv3rXHnObJbPSPCavxzs5TNRLepYEZzW1Mp5nkZT4l32/7QLnwwzTsJYGOMTmhGZ7O5LB/eeViKmwBJHXpNzd4rytFXccKlPuyUakSKgsPdTjEvY5UbFpH568wG21HXDQivz6qdST9eSVob2yUx7WV7z+2S2GfmiMZ30BMtKs4Jx1uPOY3Hk8g==",
            "alg": "RS256",
            "kid": "2MKO:CS4G:GP3F:HELH:TUI2:5YSX:NJNU:3O2N:LYM4:FBHC:T7NN:OM5A"
        }
    ]
}
```

Within the signature, the claims  `alg`, `x5c`, `kid` are specified by [RFC 7515](https://tools.ietf.org/html/rfc7515)

### Signing using GnuPG

To sign the manifest `example.json` using the GnuPG key identified by the identity name `Demo User`, run

``` shell
nv2 sign -m gpg \
  -i "Demo User" \
  -r registry.acme-rockets.io/hello-world:v1 \
  -e 8760h \
  -o hello-world.signature.config.json \
  file:hello-world_v1-manifest.json
```

- `-r` declares the original registry reference - which must match the key.
- `-e` specifies the optional expiry time (`8760h = 365 days`)

On successful signing, `nv2` prints out the `sha256` digest of the manifest, and writes the `nv2` signature JSON file `<digest>.nv2` to the working directory. If the file name is not desired, option `-o` can be specified for the alternative file name.

The formatted signature file is:

```json
{
  "signed": {
    "exp": 1626792407,
    "nbf": 1595256407,
    "iat": 1595256407,
    "manifests": [
      {
        "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
        "size": 528,
        "references": [
          "example.registry.io/example:latest"
        ]
      }
    ]
  },
  "signatures": [
    {
      "typ": "gpg",
      "sig": "wsDcBAABCAAQBQJfFa5XCRBGnsnNNqPeHgAAlM0MAHFXZxyCgsxiGVat8YCRIhR7IoQe2scswGyvGGinYBy88EpKGFEAGO+Kt1frTQNW9kLYWmTw4EFctgMw+XxeDD/CI2rsMSluRh9h2t8xBsO9Ux+7eJoxSEsfU8Jc/YZWpGs/kJOGQ3ERjvPt+SCG0Y8tuNtjnzpV4Gz+8fLSlNZ7b3f+rd7nvvJuB8iWr+yojsCeWh/VGuibyqAXPKVxSrKgkmziyYK3O/0D3KhgyR+CMtjTXL5hP314Gpc415YyN82LC3L44okimN/+X3avX0vQkthiyVw+R+Vgmpa1qk1P/ySrs81yQgFBPBC7+m4n54TqsW46X/UlkQdfP/x5Jg3jUURKgQb0wSLvzbr7Jk1RiThlwjcLhM0VgRIUwbqcqjg/5UNvMRehD44PxQXRz5feZjER2awMyKqRZnImpm8Ub+hAjhqtLGYT34oU2lwctoObV4f4BzffY9kQ0x37PQ3V8aj8k6YFQZbB4vLgwtZdA2c1froVHyuRBUwLzSBevg==",
      "iss": "Demo User <demo@example.com>"
    }
  ]
}
```

The claims `exp`, `nbf`, `iat`, `iss` are specified by [RFC 7519](https://tools.ietf.org/html/rfc7519), and all those claims will be verified against the GnuPG signature `sig`.

**NB** It is also possible to read local manifest file via an absolute path.

```shell
nv2 sign -m gpg \
  -i "Demo User" \
  file:///home/demo/hello-world_v1-manifest.json
```

### Offline Verification

Notary v2 verification can be accomplished with the `nv2 verify` command.

```shell
NAME:
    nv2 verify - verifies artifacts or images

USAGE:
    nv2 verify [command options] [<scheme://reference>]

OPTIONS:
    --signature value, -s, -f  signature file
    --cert value, -c           certs for verification [x509]
    --ca-cert value            CA certs for verification [x509]
    --key-ring value           gpg public key ring file [gpg] (default: "/home/demo/.gnupg/pubring.gpg")
    --disable-gpg              disable GPG for verification [gpg] (default: false)
    --username value  -u       username for generic remote access
    --password value, -p       password for generic remote access
    --insecure                 enable insecure remote access (default: false)
    --help, -h                 show help (default: false)
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

The command `nv2 verify` takes care of all signing methods. Since the original references of a manifest signed using `gpg` does not imply that it is signed by the domain owner, we should disable the `gpg` verification by setting the `--disable-gpg` option.

``` shell
nv2 verify \
  -f hello-world.signature.config.json \
  --disable-gpg \
  file:hello-world_v1-manifest.json

2020/07/20 23:54:35 verification failure: unknown signature type
```

## Remote Manifests

With `nv2`, it is also possible to sign and verify a manifest or a manifest list in a remote registry where the registry can be a docker registry or an OCI registry.

### Docker Registry

Here is an example to sign and verify the image `hello-world` in DockerHub, i.e. `docker.io/library/hello-world:latest`, using `gpg`.

``` shell
nv2 sign -m gpg \
  -i demo \
  -o hello-world_latest. \
  docker://docker.io/library/hello-world:latest

sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202

nv2 verify -f docker.nv2 docker://docker.io/library/hello-world:latest
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
nv2 sign -m gpg \
  -i demo \
  -o oci.nv2 \
  oci://docker.io/library/hello-world:latest

sha256:0ebe6f409b373c8baf39879fccee6cae5e718003ec3167ded7d54cb2b5da2946

nv2 verify \
  -f oci.nv2 \
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

nv2 verify -f gpg.nv2 --insecure docker://localhost:5000/example
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

### Secure Image Pulling

Since the tag might be changed during the verification process, it is required to pull by digest after verification.

```shell
digest=$(nv2 verify -f docker.nv2 docker://docker.io/library/hello-world:latest)
if [ $? -eq 0 ]; then
    docker pull docker.io/library/hello-world@$digest
fi
```

[oci-artifacts]:    https://github.com/opencontainers/artifacts
[oci-manifests]:    https://github.com/opencontainers/image-spec/blob/master/manifest.md