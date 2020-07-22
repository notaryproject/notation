# Notary V2 (nv2) - Prototype

`nv2` is a command line tool for signing and verifying manifest-based artifacts or images. This implementation supports `gpg` and `x509` signing mechanisms.

## Prerequisites 

### Build and Install

This plugin requires [golang](https://golang.org/dl/) with version `>= 1.14`.

To build and install, run

```shell
go install github.com/notaryproject/nv2/cmd/nv2
```

To build and install to an optional path, run

```shell
go build -o path/to/the/target ./cmd/nv2
```

Next, install optional components:

- Install [GnuPG](https://gnupg.org/) for `gpg`/`pgp` signing, and key management.
- Install [docker-generate](https://github.com/shizhMSFT/docker-generate) for local Docker manifest generation and local signing.
- Install [OpenSSL](https://www.openssl.org/) for key generation.

### Key Generation

#### GnuPG Key Generation

To generate a `gpg` key, run

```shell
gpg --gen-key
```

By default, all keys sit in the directory `~/.gnupg`. If the `gpg` version is `>= 2.1`, key export is required after key generation

```shell
# Update to legacy public key ring 
[ ! -f ~/.gnupg/pubring.gpg ] && gpg --export > ~/.gnupg/pubring.gpg

# Export legacy secret key ring
gpg --export-secret-keys > ~/.gnupg/secring.gpg
```

until the issue [golang/go#29082](https://github.com/golang/go/issues/29082) is resolved.

#### Self-signed Certificate Generation

To generate a `x509` self-signed certificate key pair `key.pem` and `cert.pem`, run

```shell
openssl req -x509 -sha256 -nodes -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365
```

When generating the certificate, make sure that the Common Name (`CN`) is set properly in the `Subject` field. The Common Name will be verified against the claimed original references.

## CLI Overview

The two major commands of `nv2` are

- `nv2 sign`

  ```
  NAME:
     nv2 sign - signs artifacts or images
  
  USAGE:
     nv2 sign [command options] [<scheme://reference>]
  
  OPTIONS:
     --method value, -m value     siging method
     --key value, -k value        siging key file [x509]
     --cert value, -c value       siging cert [x509]
     --key-ring value             gpg public key ring file [gpg] (default: "/home/demo/.gnupg/secring.gpg")
     --identity value, -i value   signer identity [gpg]
     --expiry value, -e value     expire duration (default: 0s)
     --reference value, -r value  original references
     --output value, -o value     write signature to a specific path
     --username value, -u value   username for generic remote access
     --password value, -p value   password for generic remote access
     --insecure                   enable insecure remote access (default: false)
     --help, -h                   show help (default: false)
  ```

- `nv2 verify`

  ```
  NAME:
     nv2 verify - verifies artifacts or images
  
  USAGE:
     nv2 verify [command options] [<scheme://reference>]
  
  OPTIONS:
     --signature value, -s value, -f value  signature file
     --cert value, -c value                 certs for verification [x509]
     --ca-cert value                        CA certs for verification [x509]
     --key-ring value                       gpg public key ring file [gpg] (default: "/home/demo/.gnupg/pubring.gpg")
     --disable-gpg                          disable GPG for verification [gpg] (default: false)
     --username value, -u value             username for generic remote access
     --password value, -p value             password for generic remote access
     --insecure                             enable insecure remote access (default: false)
     --help, -h                             show help (default: false)
  ```

## Offline Signing and Verification

Signing and verification offline can be accomplished by the `nv2 sign` command and the `nv2 verify` command.
In this section, examples are provided for a tour of `nv2` signing and verification.

Since signing and verification are based on manifests, [docker-generate](https://github.com/shizhMSFT/docker-generate) is used to generate the manifest, which is exactly the same manifest as the `docker push` produces.

```shell
docker build -t example .
docker generate manifest example > example.json
```

The above commands build the image `example:latest` based on the local context, and then generate its manifest file `example.json`.

### Signing using GnuPG

To sign the manifest `example.json` using the GnuPG key identified by the identity name `Demo User`, run

```
$ nv2 sign -m gpg -i "Demo User" -r example.registry.io/example:latest -e 8760h file:example.json
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

where the optional option `-r` declares the original reference, and the optional option `-e` specifies the expiry time (`8760h = 365 days`). On successful signing, `nv2` prints out the `sha256` digest of the manifest, and writes the `nv2` signature JSON file `<digest>.nv2` to the working directory. If the file name is not desired, option `-o` can be specified for the alternative file name.

In this example, the signature file name is `3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55.nv2`. The formatted signature file is

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

where the claims `exp`, `nbf`, `iat`, `iss` are specified by [RFC 7519](https://tools.ietf.org/html/rfc7519), and all those claims will be verified against the GnuPG signature `sig`.

**NB** It is also possible to read local manifest file via an absolute path.

```shell
nv2 sign -m gpg -i "Demo User" file:///home/demo/example.json
```

### Signing using `x509`

To sign the manifest `example.json` using the key `key.pem` from the `x509` certificate `cert.pem` with the Common Name `example.registry.io`, run

```shell
nv2 sign -m x509 -k key.pem -c cert.pem -r example.registry.io/example:latest -o example.nv2 file:example.json
```

The formatted signature file `example.nv2` is

```json
{
  "signed": {
    "iat": 1595257070,
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
      "typ": "x509",
      "sig": "PnY2vpFJV0fayfGOAAxkokthImq932W8XutYCjGLgvBSqdzGM6VgbJhTgXGeettYv5S7A/FO6e319TxEFmx3ogf1bneOUOGDRCdEte+MupDhAISDkiN42Ktci18qFh7MlcR2DXFos5qux0H3Rrc5Rd6Hi4BTTTwHBjsbnNkN1aXuYmyrJZgYmlHBzfdbaDJRcNMo1RAX+j+BWsNZDv+Ae2dtcnoYc2gK5YC2YuNAsvtP4PpR0jtygpCDZjItdVNsJGMwB3dXHUes7Z88IX8hIKlEOt9qv4sq2iOBTju2zvzk4R/pCjUkbD6dOb+t2uyayXbvyAJbi/cEzsfCdwrXjg==",
      "alg": "RS256",
      "x5c": [
        "MIIDpzCCAo+gAwIBAgIUb6xLgtw1gaM45RnNL9PPhGgvjtEwDQYJKoZIhvcNAQELBQAwYzELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEcMBoGA1UEAwwTZXhhbXBsZS5yZWdpc3RyeS5pbzAeFw0yMDA3MjAxNDU1MjJaFw0yMTA3MjAxNDU1MjJaMGMxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxHDAaBgNVBAMME2V4YW1wbGUucmVnaXN0cnkuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDYW9yzMDsdVpZVU5bTj64iww0AMDGY0qptRaD5ivJGE3RA1WVS/M6gWsOmIHxINiURvuwyG77iUPGq/dXUXjbkypWLcdyEpgpsQRss/lMKjYVayi11nOe6L9nv/1vJ7FPuaBLZJLUIX7+6+/GwQCIe5SmbFmEERNPZ24HdTA+q5jAynYJqJQAx1ReUXNu8jKMo9ZPq787VJIK8eiLn4gty/JfZ0VyobFHaCClbVp+nvfv6IeV+34pFcnPX0UaA4b0zerQIYfkaAAu5pQcR7W5KNQgMR0HIMdvw7Kkuzx30pwJA3i8X49D9nyalyW41wWRnNe8emAjgFkMFXMqlxNmxAgMBAAGjUzBRMB0GA1UdDgQWBBTrHl7XtUeE0biwJngTM2DtVz2LqTAfBgNVHSMEGDAWgBTrHl7XtUeE0biwJngTM2DtVz2LqTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBAua2m0+i920V+KhzmAukJaEh1CkOXr1nFRN2eMuNO4H5l1NsMhtR2XXQkhap9GfSnS0BSIvJ9WbDWm7YBJFXo0zD9pbnlLEbnVtegDzUtEf+0yydKatTc+ClGVM+Cugrbbc7Jzb+hauh6WodYxUAMLUL7Ld4ae7x17VlpgQtRSMELJVrDXaabQXT7sY2pSomFBY5/3NnCJGUOLX0XLRU9dgjHqx1ARWeiJpvH/hV9w2o0jAM+W/vKJHXi4gz1StFLRv4C66cZbMH3yX7d4tlLB7V54ZU0jkRUOcWKFC9Cn4dRrs2dEjYgHRTuk2G3dcqxUCwWCaquuhjk1koi9xYA"
      ]
    }
  ]
}
```

If the embedded cert chain `x5c` is not desired, it can be replaced by a key ID `kid` by omitting the `-c` option.

```shell
nv2 sign -m x509 -k key.pem -r example.registry.io/example:latest -o example.nv2 file:example.json
```

The formatted resulted signature file `example.nv2` is

```json
{
  "signed": {
    "iat": 1595259542,
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
      "typ": "x509",
      "sig": "CIas/ACj5bI0aQHuQCGFRK5I7wKAFltide2a/7u5h5g5xIthbeDjGKUL8JNV9r1Bl2TlCQjyv8695eq8jpe4nlyWWkdf4S+79njLkvWhJiUakLHq4KV1gFUy1dUKSOLRA1YGiS30q0ZKUOiUUdiEF+OUqGc4bHvtrL9ByHA8QBffYvBHSqnzowu/yTwwmX9QvnGwh4ic4Hi4YhJpPwbIYvmcuiXtSgqj/oo2d+aVc+uj9QYp0/ETVl3h7HFZ5XjGB4SxxF77TxqsghpyojMOxf8bT8KxR7V05I1Acy6jmyXyh1pliF9ENdmvHQgSEbtXaWs+8tqkdZd+Y6BxpUA2tQ==",
      "alg": "RS256",
      "kid": "L7YO:TIUS:TSSY:DV6I:HOU4:YAIC:5HLB:JR7Y:W2EK:XU7W:L27M:YYHY"
    }
  ]
}
```

where the claims  `alg`, `x5c`, `kid` are specified by [RFC 7515](https://tools.ietf.org/html/rfc7515),

### Verifying

To verify a manifest `example.json` with a signature file `example.nv2`, run

```shell
nv2 verify -f example.nv2 file://example.json
```

Since the manifest was signed by a self-signed certificate, that certificate `cert.pem` is required to be provided to `nv2`.

```
$ nv2 verify -f example.nv2 -c cert.pem file:example.json
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

On successful verification, the `sha256` digest of the manifest is printed. Otherwise, `nv2` prints error messages and returns non-zero values.

The command `nv2 verify` takes care of all signing methods. Since the original references of a manifest signed using `gpg` does not imply that it is signed by the domain owner, we should disable the `gpg` verification by setting the `--disable-gpg` option.

```
$ nv2 verify -f gpg.nv2 --disable-gpg file:example.json
2020/07/20 23:54:35 verification failure: unknown signature type
```

## Remote Manifests

With `nv2`, it is also possible to sign and verify a manifest or a manifest list in a remote registry where the registry can be a docker registry or an OCI registry.

### Docker Registry

Here is an example to sign and verify the image `hello-world` in DockerHub, i.e. `docker.io/library/hello-world:latest`, using `gpg`.

```
$ nv2 sign -m gpg -i demo -o docker.nv2 docker://docker.io/library/hello-world:latest
sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202
$ nv2 verify -f docker.nv2 docker://docker.io/library/hello-world:latest
sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202
```

It is possible to use `digest` in the reference. For instance, 

```
docker.io/library/hello-world@sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202
```

If neither `tag` nor `digest` is specified, the default tag `latest` is used.

### OCI Registry

OCI registry works the same as Docker but with the scheme `oci`.

```
$ nv2 sign -m gpg -i demo -o oci.nv2 oci://docker.io/library/hello-world:latest
sha256:0ebe6f409b373c8baf39879fccee6cae5e718003ec3167ded7d54cb2b5da2946
$ nv2 verify -f oci.nv2 oci://docker.io/library/hello-world:latest
sha256:0ebe6f409b373c8baf39879fccee6cae5e718003ec3167ded7d54cb2b5da2946
```

**Note** The digest of the OCI manifest is different from the Docker manifest for the same image since their format is different. Therefore, the signer should be careful with the manifest type when signing.

### Insecure Registries

To sign and verify images from insecure registries accessed via `HTTP`, such as `localhost`, the option `--insecure` is required.

```
$ docker tag example localhost:5000/example
$ docker push localhost:5000/example
The push refers to repository [localhost:5000/example]
50644c29ef5a: Pushed
latest: digest: sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55 size: 528
$ nv2 verify -f gpg.nv2 --insecure docker://localhost:5000/example
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