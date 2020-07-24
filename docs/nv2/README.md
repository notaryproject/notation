# Notary V2 (nv2) - Prototype

`nv2` is a command line tool for signing and verifying manifest-based artifacts or images. This implementation supports `x509` signing mechanism.

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

- Install [docker-generate](https://github.com/shizhMSFT/docker-generate) for local Docker manifest generation and local signing.
- Install [OpenSSL](https://www.openssl.org/) for key generation.

### Key Generation

#### Self-signed Certificate Generation

To generate a `x509` self-signed certificate key pair `key.pem` and `cert.pem`, run

```shell
openssl req -x509 -sha256 -nodes -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365
```

When generating the certificate, make sure that the Common Name (`CN`) is set properly in the `Subject` field. The Common Name will be verified against the claimed original references.

## CLI Overview

The two major commands of `nv2` are

```
NAME:
   nv2 sign - signs artifacts or images

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

```
NAME:
   nv2 verify - verifies artifacts or images

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

## Offline Signing and Verification

Signing and verification offline can be accomplished by the `nv2 sign` command and the `nv2 verify` command.
In this section, examples are provided for a tour of `nv2` signing and verification.

Since signing and verification are based on manifests, [docker-generate](https://github.com/shizhMSFT/docker-generate) is used to generate the manifest, which is exactly the same manifest as the `docker push` produces.

```shell
docker build -t example .
docker generate manifest example > example.json
```

The above commands build the image `example:latest` based on the local context, and then generate its manifest file `example.json`.

### Signing using `x509`

To sign the manifest `example.json` using the key `key.pem` from the `x509` certificate `cert.pem` with the Common Name `example.registry.io`, run

```
$ nv2 sign -m x509 -k key.pem -c cert.pem -r example.registry.io/example:latest -e 8760h file:example.json
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

where the optional option `-r` declares the original reference, and the optional option `-e` specifies the expiry time (`8760h = 365 days`). On successful signing, `nv2` prints out the `sha256` digest of the manifest, and writes the `nv2` signature JSON file `<digest>.nv2` to the working directory. If the file name is not desired, option `-o` can be specified for the alternative file name.

In this example, the signature file name is `3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55.nv2`. The formatted signature file is

```json
{
    "signed": {
        "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
        "size": 528,
        "references": [
            "example.registry.io/example:latest"
        ],
        "exp": 1627097957,
        "nbf": 1595561957,
        "iat": 1595561957
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "o0pytd+Ck8ePDoFlqD76W8N4AiRf4MEqKGRbo/IvOvhxaE8PIzwcoIiLGgbeYUq9K6GRiEZlBiBtkc+OVf4Ld8AN4o1EMrqrmZXrZKB/l0xzs00Yei6Z0UzsdGi1FFtJzFq9BofDz7tK/0122iPu7u5riPXHexKiuPbfxhXd4hpktiXs/qHy9bMGhmlQIyup5IZ3EUMlEIlQchOZuj/xSxIzBaxYuTxFJOGSMviGZWI7hFrs2bisDdN/FQWjisPpUoA+S2NWLkpOCkFI2uSzFKTGKte6GkV2ygnGAU9cXfVIDBMxPBeDgA/vej3bA9263kY8NrXM8z1Kz4OJR11jrA==",
            "alg": "RS256",
            "x5c": [
                "MIIDHTCCAgWgAwIBAgIUGrRoPvh96XNsYpgYEGzbT5X+weQwDQYJKoZIhvcNAQELBQAwHjEcMBoGA1UEAwwTZXhhbXBsZS5yZWdpc3RyeS5pbzAeFw0yMDA3MjQwMzM2MjNaFw0yMTA3MjQwMzM2MjNaMB4xHDAaBgNVBAMME2V4YW1wbGUucmVnaXN0cnkuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC36/XRfhNMyM0kxuOlEleX28MGdzevJH7KWH4qZlswDnroVoAqydbyJ64X3w1wQHIuNA0fv2Nzz42QuHQ/30yjUH+6flLX0NbBE7dXFpQX+A+gA9fh5B1Iv8xtOA4GbcRe7tQqSwOzv4aOBMsfWONSkAp98gS50gSUxlyFOJB6+OQu/7qOtJp2adeqJl3NRcN/cZRHjf8wuJ1lW5bmlw0cqr0o0djO4LtFHUfLOel8ssuxZ0r4KJPrsIQnrqrdCcQks+0HMJcnCUGON+lCnjmZHy1Wnb2GJ2muC9dgvcAo/S9ACXTMTt/Y2+Pd8cvfJdGiqV+TzfMBElOBalizWmWnAgMBAAGjUzBRMB0GA1UdDgQWBBQ5jPzDjKFkIzWqw+bWXRQnIE0BDTAfBgNVHSMEGDAWgBQ5jPzDjKFkIzWqw+bWXRQnIE0BDTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQAYWp48KRbJbUMWntg7OKDIeh23HE+YOLXXh2Zp8sB9hbMyg/GVvJLcqYokct3l2lFVszykw3s/eACkTne5sCktKewP63xQ58ZnYYg4iXFULbRHsltlVPniMjCKI/sb0C/Q/w6kynxfBlsD62ETVrgKGn1KJJCIaNqcTLvSS7yQ4Kb7abP5caRNxjsWz89j+MxE7F4uNLeLvFOMGSSwAbR2pWetvgvt7i/unGBEw2NP7Xqy9KUJsqBhfYlfmX4qPWWRVNeN/r4UrkYVpP+uYhvYWTNjDoYcgdXr08EkpkBsDWIUiDVpJEQBzmt8L9Oalf5lwqGGrsM0xf47H0G/PFD1"
            ]
        }
    ]
}
```

The detailed signature specification is [available](../signature/README.md).

**Note** It is also possible to read local manifest file via an absolute path.

```shell
nv2 sign -m x509 -k key.pem -c cert.pem file:///home/demo/example.json
```

If the embedded cert chain `x5c` is not desired, it can be replaced by a key ID `kid` by omitting the `-c` option.

```shell
nv2 sign -m x509 -k key.pem -r example.registry.io/example:latest -o example.nv2 file:example.json
```

The formatted resulted signature file `example.nv2` is

```json
{
    "signed": {
        "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
        "size": 528,
        "references": [
            "example.registry.io/example:latest"
        ],
        "iat": 1595562116
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "rEVuMlvfrXjhPx+LgWSsRDyGrWiOuSG74l7ex4ekPwxIzetWe5q8/i8VGwnhAUQvHnq/TgFz9AoRjJECLUskJXrNTeDTcIrCudTZIri+urQkW7dj/mOt53Iq8X3tYGHfFX060I95WtPFwLxbO018THQbIrbhFoK9Lwotrk/k+9gMGwq5tZ6JMjfiHV61Ne1OKli5WZlpCjIjsScwmATMUyIc7or/c7w70L058QR6vZjZaDoMzXOYv6uiusj5wKCaySbsaNz07Gmx9DYRkbbDt7l5YjB4eC+gaH08Yrrbq/yWHyHlbR4EnKOJw5Ki1y4QKLRKWFjhw3tXdWacnD//7g==",
            "alg": "RS256",
            "kid": "DD2E:DW7J:OVJK:KCZR:2PGY:SYC5:WFJF:RMMV:FH6W:VGYM:2WW4:7ZGC"
        }
    ]
}
```

### Verifying

To verify a manifest `example.json` with a signature file `example.nv2`, run

```shell
nv2 verify -f example.nv2 file:example.json
```

Since the manifest was signed by a self-signed certificate, that certificate `cert.pem` is required to be provided to `nv2`.

```
$ nv2 verify -f example.nv2 -c cert.pem file:example.json
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

On successful verification, the `sha256` digest of the manifest is printed. Otherwise, `nv2` prints error messages and returns non-zero values.

The command `nv2 verify` takes care of all signing methods.

## Remote Manifests

With `nv2`, it is also possible to sign and verify a manifest or a manifest list in a remote registry where the registry can be a docker registry or an OCI registry.

### Docker Registry

Here is an example to sign and verify the image `hello-world` in DockerHub, i.e. `docker.io/library/hello-world:latest`, using `x509`.

```
$ nv2 sign -m x509 -k key.pem -o docker.nv2 docker://docker.io/library/hello-world:latest
sha256:49a1c8800c94df04e9658809b006fd8a686cab8028d33cfba2cc049724254202
$ nv2 verify -f docker.nv2 -c cert.pem docker://docker.io/library/hello-world:latest
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
$ nv2 sign -m x509 -k key.pem -o oci.nv2 oci://docker.io/library/hello-world:latest
sha256:0ebe6f409b373c8baf39879fccee6cae5e718003ec3167ded7d54cb2b5da2946
$ nv2 verify -f oci.nv2 -c cert.pem oci://docker.io/library/hello-world:latest
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
$ nv2 verify -f example.nv2 --insecure docker://localhost:5000/example
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

### Secure Image Pulling

Since the tag might be changed during the verification process, it is required to pull by digest after verification.

```shell
digest=$(nv2 verify -f docker.nv2 -c cert.pem docker://docker.io/library/hello-world:latest)
if [ $? -eq 0 ]; then
    docker pull docker.io/library/hello-world@$digest
fi
```