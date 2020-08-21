# TUF Integration Proposal

## Introduction

While developer-signed signatures guarantee the unforgeability of software packages, they are not good enough to defend attacks in the software update scenarios like rollback attacks, fast-forward attacks, indefinite freeze attacks, and many others. To resolve the security challenges in software update systems, Samuel et al. proposed the update framework (TUF) [[1][1]] and it was later enhanced by Kuppusamy et al. [[2][2], [3][3]].

Many organizations [[4][4], [5][5], [6][6], [7][7], [8][8], [9][9]] adopt TUF for their software repositories. PyPI [[9][9]] accepted to integrate TUF in the minimum security model [[10][10]] for security protection in continuous delivery of distributions. However, the adversaries are able to forge packages if PyPI is compromised. Therefore, the maximum security model [[11][11]] was proposed for end-to-end signing.

## Proposal

We extract the ideas in the Python Enhancement Proposals (PEPs) [[10][10], [11][11]] that

- TUF operated by organizations are used to defend *update related attacks*.
- Regular signatures signed by developed are used to defend *forgery*.

It is worth noting that the verification keys / public keys are managed by trusted PKIs, and not by TUF since each TUF trust collection includes an unique PKI.

Both techniques are standalone and orthogonal. Organizations and end users have options to choose either of them or both, depending on their security requirements or security model. If organizations or end users do not care about update scenarios, regular signatures are sufficient. If they care about software update, TUF is a better choice. If they care about the origin of the content and software update, they can choose both where trust can split and come from different parties. 

In the case of double security, the developer-signed signatures must be in the format of a JSON file for a TUF delegation role so that those signatures can seamlessly integrated to TUF as a real delegation role. For instance, Alice publishes a software package and signs it in the TUF delegation role format. Later, Alice publishes the TUF-formatted signature to a TUF trust collection operated by Bob. If Charlie downloads the package and wants to verify, Charlie can verify the TUF trust collection by Bob for update protection, and then verify the TUF-formatted signature with contemporary PKI for potential forgery.

Readers may find that the separation of trust is useful in signature movement and even in air-gapped network as a better solution than [[12][12]]. Continuing with the above example instance, Charlie can copy Alice's package and signature to his own possibly air-gapped registry, and register the signature as a delegation role in the TUF trusted collection operated by David in the same network. As a result, Charlie can verify the package without Bob. In this scenario, Charlie still trusts Alice for forgery protection. However, he moves the trust for update protection from Bob for David.

## Prototype

The `nv2` prototype with the `tuf` sub commands illustrates what a TUF-formatted signature, which is a JSON file for a delegation role, looks like.

```shell
NAME:
   nv2 tuf - TUF related commands

USAGE:
   nv2 tuf command [command options] [arguments...]

COMMANDS:
   sign     signs OCI Artifacts
   verify   verifies OCI Artifacts
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

Unlike Docker Content Trust [[6][6]], each organization owns a trust collection instead of per docker repository basis. We also use fully qualified image reference for the target name (e.g. `registry.acme-rockets.io/hello-world:v1`) instead of tag names (e.g. `v1`).

### Offline Signing

Offline signing is accomplished with the `nv2 tuf sign` command.

```shell
NAME:
   nv2 tuf sign - signs OCI Artifacts

USAGE:
   nv2 tuf sign [command options] [<scheme://reference>]

OPTIONS:
   --key value, -k value                  signing key file
   --cert value, -c value                 signing cert
   --signature value, -s value, -f value  base signature file
   --reference value, -r value            original references
   --expiry value, -e value               expire duration (default: 0s)
   --output value, -o value               write signature to a specific path
   --media-type value                     specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
   --username value, -u value             username for generic remote access
   --password value, -p value             password for generic remote access
   --insecure                             enable insecure remote access (default: false)
   --help, -h                             show help (default: false)
```

To sign the manifest `hello-world_v1-manifest.json` using the key `key.key` from the `x509` certificate `cert.crt` with the Common Name `registry.acme-rockets.io`, run

```shell
nv2 tuf sign \
  -k key.key \
  -c cert.crt \
  -r registry.acme-rockets.io/hello-world:v1 \
  -e 8670h \
  -o hello-world.signature.config.json \
  file:hello-world_v1-manifest.json
```

The formatted x509 signature: `hello-world.signature.config.json` is:

``` json
{
    "signed": {
        "_type": "Targets",
        "delegations": {
            "keys": {},
            "roles": []
        },
        "expires": "2021-08-16T15:57:23.4190247Z",
        "targets": {
            "registry.acme-rockets.io/hello-world:v1": {
                "custom": {
                    "accessedAt": "2020-08-20T09:57:23.4189973Z",
                    "mediaType": "application/vnd.docker.distribution.manifest.v2+json"
                },
                "hashes": {
                    "sha256": "JKdJAKTnSe8xKV5aq95wk+MkSxGVgr1uZLGojHHEENA="
                },
                "length": 3056
            }
        },
        "version": 1
    },
    "signatures": [
        {
            "keyid": "cd0e842b294d61bb73a03260b4228069a5c3d6af86842360cc53c35e58d90868",
            "method": "rsapss",
            "sig": "aJ9iR72Mgej55Ds2t5nFAvcHnSB//tefp7eXvmFWTTQYklmwZ4oPrgii5DhA+rNNDPKAA7z1y861+b5MhbET8Fd4tmQSsoay5lBHVrt2MtpEMYgXD/sqlrRWvQmgvLn+ibWfhfW1MkqOPQ14Y4iz8JIC4UK+c1xc5KWbykVGgyTHM0/JEe5rq/iVwq6rhurpn1rGjV4mbCoFLYMhpKVWuguu2Pj73ertJ+VxCCacNgbsC2DOJIg277lbZB9e/YmWufy5ZTKbc9ECwF4uLRsx2qVHXei3eop3nFpGzTt1cWRTLkKllpdW9mEnpp1GxFHAE1UYXP3E7xdVAL8VZQWIMA==",
            "x5c": [
                "MIID4DCCAsigAwIBAgIUJVfYMmyHvQ9u0TbluizZ3BtgATEwDQYJKoZIhvcNAQELBQAwbTEhMB8GA1UEAwwYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwHhcNMjAwODIwMDI0MDU2WhcNMjEwODIwMDI0MDU2WjBtMSEwHwYDVQQDDBhyZWdpc3RyeS5hY21lLXJvY2tldHMuaW8xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAN4qEOEICopQ9t1BSrh4Iu3f4LRMWI0p+P9Js92nhCQusn3wcEIddtb3J2eSy0d1MwTykXEMvSp3OU3+yJIm7BW3ZrCD+UjU+Je8JcDCXKZpKesLb8i42UAk7J35zWE/nRs9uzGQWelZTCJHBM1NVSnP4QqcGF2VkNwtsti7NbL4f+AunBMZvrK4p3kUwh92FoDgXen6+vrnqHb3MA8uJBa5pVsPOvcga13TWaYRfMm/nxM6xGvIwly2QpWfdJrC58aqEiRQAfUtfuD/MTYEQ0PL/fOpHJw/FcHLt0vkja1ggMexATinhvOPMhcJbL2JPliS1ExFtCjx2D+mtHjt9rECAwEAAaN4MHYwHQYDVR0OBBYEFME2lTAI0lfw6/2/SNNSydlMqWNzMB8GA1UdIwQYMBaAFME2lTAI0lfw6/2/SNNSydlMqWNzMA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RBBwwGoIYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMA0GCSqGSIb3DQEBCwUAA4IBAQB6HIusgEwUaQxcO9jm+iX4rgnGeDYT9qNdxtSL/tz7zzPTx2gPDSEzZinhFO+0AnH0kleiaMfJXFedvpX8xofP0zWNDgXAabZa1JR9HV42OmxMg/gBm3lpSQtnNoraqy6N88ot9xpRA0FQ/gAnGdRakrK7oeljDNpz3ay6ZgBqz3MpYIKkHL6dpvmQ4BbGEHfjLX1j/bC397XzapOcFqqhekc3Nk7vH51TheqQRIW1nI4BRo/guf6zjfxGcskTM4winCd/fk0F7XMlOddWleeg1vI+i/1TKV0p03aN23JZNUAt7MeZlz4nieaIGGFNinwOEsIRIXAZ65IMZwaLsCgY"
            ]
        }
    ]
}
```

If the embedded cert chain `x5c` is not desired, which is strongly not recommended, it can be removed by omitting the `-c` option. In this case, it is exactly a TUF metadata for a delegation role. Since the certificate is not provided, the key ID is purely calculated by the public key information extracted from the provided private key.

```shell
nv2 tuf sign \
  -k key.key \
  -r registry.acme-rockets.io/hello-world:v1 \
  -e 8670h \
  -o hello-world.signature.config.json \
  file:hello-world_v1-manifest.json
```

The formatted x509, without the `x5c` chain signature: `hello-world.signature.config.json` is:


```json
{
    "signed": {
        "_type": "Targets",
        "delegations": {
            "keys": {},
            "roles": []
        },
        "expires": "2021-08-16T16:00:10.3222917Z",
        "targets": {
            "registry.acme-rockets.io/hello-world:v1": {
                "custom": {
                    "accessedAt": "2020-08-20T10:00:10.3222657Z",
                    "mediaType": "application/vnd.docker.distribution.manifest.v2+json"
                },
                "hashes": {
                    "sha256": "JKdJAKTnSe8xKV5aq95wk+MkSxGVgr1uZLGojHHEENA="
                },
                "length": 3056
            }
        },
        "version": 1
    },
    "signatures": [
        {
            "keyid": "b18f65d69cfce50b03fe92f159726aa48f1c75b62e23f53b4ac4fda9d0739c10",
            "method": "rsapss",
            "sig": "tjrMGfrz0TaZWWsvTwbwKGQRHCdUpuS+CfP7TJPxh7oRrMgh3I74MGSy3w8S2DsAoMGipWBQ1vcWxCGPO49k8+ND3T10GvU66J7fkkSbesamiG3WCfZBR+6mNUr5fzu6gNAHop9twaSstnOPtXgTW5gLcf7geD4XHTCkzYgY7ejoFUiFUD+oBQCdIS7gj7rEgT2Wi0mYWjrt7tbsoDf26nDODnOmUhxrgBJ1iYqn8h8ZWG01iXmrRl7txUC4DINrxPQZQh8o7AP7R8M6hcDW12B+acWxt2Smb16FxhhUTNmZ3nSRL6Uyz9NqgtUftmdd4KjHYGOiKIez7ktSoITocQ=="
        }
    ]
}
```

It is also possible to add targets, refresh the expiry time, and auto increment the version number by providing the previous state.

```shell
nv2 tuf sign \
  -k key.key \
  -c cert.crt \
  -r registry.acme-rockets.io/hello-world:latest \
  -e 8670h \
  -f hello-world.signature.config.json \
  -o hello-world.signature.config.json \
  file:hello-world_v1-manifest.json
```

The formatted x509 signature: `hello-world.signature.config.json` is:

``` json
{
    "signed": {
        "_type": "Targets",
        "delegations": {
            "keys": {},
            "roles": []
        },
        "expires": "2021-08-16T16:06:04.8466197Z",
        "targets": {
            "registry.acme-rockets.io/hello-world:latest": {
                "custom": {
                    "accessedAt": "2020-08-20T10:06:04.8460942Z",
                    "mediaType": "application/vnd.docker.distribution.manifest.v2+json"
                },
                "hashes": {
                    "sha256": "JKdJAKTnSe8xKV5aq95wk+MkSxGVgr1uZLGojHHEENA="
                },
                "length": 3056
            },
            "registry.acme-rockets.io/hello-world:v1": {
                "custom": {
                    "accessedAt": "2020-08-20T10:06:02.3743791Z",
                    "mediaType": "application/vnd.docker.distribution.manifest.v2+json"
                },
                "hashes": {
                    "sha256": "JKdJAKTnSe8xKV5aq95wk+MkSxGVgr1uZLGojHHEENA="
                },
                "length": 3056
            }
        },
        "version": 2
    },
    "signatures": [
        {
            "keyid": "cd0e842b294d61bb73a03260b4228069a5c3d6af86842360cc53c35e58d90868",
            "method": "rsapss",
            "sig": "juUQtXgu9fFdEXk20mOk7XtfC4bniPt6EmoEFe5w1shBFmZLtsxeXRR9nW9+n2jmF1zTT/wk50Y9q14+nc3DA0EKlAnvW+ulUOfWSqEN6t7K6wHesJlkK3uTKa/TMW3pGTwtdMFUU823H7eJKxFR8n4lINuAGmakUk/Cr766dRdvOv85FepFwvM0ZwuwZ3aaDNGOUtwWjHLTgCNwljDym1sQ4sc96N8igQ+lG7wb0WNGMhZzTGPBDypr4lDO03A8s4AcD13j/uCUHWG0cVN3TBgPEZmqrmJLJ7iCsCO4ieJQLiGZKVee3QOBi1F0JgbNQ1LEeHtNegikDkPa2GOxgw==",
            "x5c": [
                "MIID4DCCAsigAwIBAgIUJVfYMmyHvQ9u0TbluizZ3BtgATEwDQYJKoZIhvcNAQELBQAwbTEhMB8GA1UEAwwYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwHhcNMjAwODIwMDI0MDU2WhcNMjEwODIwMDI0MDU2WjBtMSEwHwYDVQQDDBhyZWdpc3RyeS5hY21lLXJvY2tldHMuaW8xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAN4qEOEICopQ9t1BSrh4Iu3f4LRMWI0p+P9Js92nhCQusn3wcEIddtb3J2eSy0d1MwTykXEMvSp3OU3+yJIm7BW3ZrCD+UjU+Je8JcDCXKZpKesLb8i42UAk7J35zWE/nRs9uzGQWelZTCJHBM1NVSnP4QqcGF2VkNwtsti7NbL4f+AunBMZvrK4p3kUwh92FoDgXen6+vrnqHb3MA8uJBa5pVsPOvcga13TWaYRfMm/nxM6xGvIwly2QpWfdJrC58aqEiRQAfUtfuD/MTYEQ0PL/fOpHJw/FcHLt0vkja1ggMexATinhvOPMhcJbL2JPliS1ExFtCjx2D+mtHjt9rECAwEAAaN4MHYwHQYDVR0OBBYEFME2lTAI0lfw6/2/SNNSydlMqWNzMB8GA1UdIwQYMBaAFME2lTAI0lfw6/2/SNNSydlMqWNzMA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RBBwwGoIYcmVnaXN0cnkuYWNtZS1yb2NrZXRzLmlvMA0GCSqGSIb3DQEBCwUAA4IBAQB6HIusgEwUaQxcO9jm+iX4rgnGeDYT9qNdxtSL/tz7zzPTx2gPDSEzZinhFO+0AnH0kleiaMfJXFedvpX8xofP0zWNDgXAabZa1JR9HV42OmxMg/gBm3lpSQtnNoraqy6N88ot9xpRA0FQ/gAnGdRakrK7oeljDNpz3ay6ZgBqz3MpYIKkHL6dpvmQ4BbGEHfjLX1j/bC397XzapOcFqqhekc3Nk7vH51TheqQRIW1nI4BRo/guf6zjfxGcskTM4winCd/fk0F7XMlOddWleeg1vI+i/1TKV0p03aN23JZNUAt7MeZlz4nieaIGGFNinwOEsIRIXAZ65IMZwaLsCgY"
            ]
        }
    ]
}
```

### Offline Verification

Offline verification without TUF for update protection can be accomplished with the `nv2 tuf verify` command.

```shell
NAME:
   nv2 tuf verify - verifies OCI Artifacts

USAGE:
   nv2 tuf verify [command options] [<scheme://reference>]

OPTIONS:
   --signature value, -s value, -f value  signature file
   --cert value, -c value                 certs for verification
   --ca-cert value                        CA certs for verification
   --min-version value, -m value          min version of the signature (default: 0)
   --reference value, -r value            original reference
   --media-type value                     specify the media type of the manifest read from file or stdin (default: "application/vnd.docker.distribution.manifest.v2+json")
   --username value, -u value             username for generic remote access
   --password value, -p value             password for generic remote access
   --insecure                             enable insecure remote access (default: false)
   --help, -h                             show help (default: false)
```

To verify a manifest `hello-world_v1-manifest.json` with a signature file `hello-world.signature.config.json`, run

```shell
nv2 tuf verify \
  -f hello-world.signature.config.json \
  -c cert.crt \
  file:hello-world_v1-manifest.json
```

Since the manifest was signed by a self-signed certificate, that certificate `cert.crt` is required to be provided to `nv2`.

If the cert isn't self-signed, you can omit the `-c` parameter.

``` shell
nv2 tuf verify \
  -f hello-world.signature.config.json \
  file:hello-world_v1-manifest.json

sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

On successful verification, the `sha256` digest of the manifest is printed. Otherwise, `nv2` prints error messages and returns non-zero values.

### Working with TUF

**NOTE** This section is for information only and might not be implemented.

To make the TUF-formatted signature working in a trust collection, a delegation role should be created with the certificate / public key associated with the signing key. This operation can be done using the Docker Notary CLI:

```shell
notary delegation add $gun targets/example example.crt --paths "<fully_qualified_image_references>" --publish
```

Once the delegation role is registered, the TUF-formatted signature can be uploaded to the remote notary server with server-managed snapshot role:

```shell
curl -H "Authorization: bearer $token" \
  -F 'upload=@"</path/to/signature>";filename=targets/example.json' \
  "https://$notaryserver/v2/$gun/_trust/tuf/"
```

Later, consumers should be able to verify the TUF trust collection as normal. After the TUF verification against the TUF imbedded PKI is done, the consumers should verify the signature again with `nv2 tuf verify` against the contemporary PKI.



[1]: https://doi.org/10.1145/1866307.1866315	"Survivable key compromise in software update systems"
[2]: https://dl.acm.org/doi/10.5555/2930611.2930648	"Diplomat: using delegations to protect community repositories"
[3]: https://dl.acm.org/doi/10.5555/3154690.3154754	"Mercury: bandwidth-effective prevention of rollback attacks against community repositories"
[4]: https://theupdateframework.io/ "The Update Framework"
[5]: https://github.com/theupdateframework/notary   "The Notary Project"
[6]: https://docs.docker.com/engine/security/trust/content_trust/   "Docker Content Trust"
[7]: https://docs.microsoft.com/en-us/azure/container-registry/container-registry-content-trust "Content trust in Azure Container Registry"
[8]: https://github.com/cnabio/cnab-spec/blob/cnab-security-1.0.0/300-CNAB-security.md	"Cloud Native Application Bundles Security (CNAB-Sec) 1.0.0 GA"
[9]: https://pypi.org/  "PyPI"
[10]: https://www.python.org/dev/peps/pep-0458/  "PEP 458 -- Secure PyPI downloads with signed repository metadata"
[11]: https://www.python.org/dev/peps/pep-0480/  "PEP 480 -- Surviving a Compromise of PyPI: The Maximum Security Model"
[12]: https://github.com/cnabio/cnab-spec/blob/cnab-security-1.0.0/805-airgap.md#cnab-security	"CNAB Security in Disconnected Scenarios"
