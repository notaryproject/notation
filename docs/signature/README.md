# Notary V2 Signature Specification

This section defines the signature file, which is in JSON format with no whitespaces. Its JSON schema is available at [schema.json](schema.json).

# Signature

A Notary v2 signature is clear-signed signature of manifest metadata, including but not limited to

- [OCI Image Index](https://github.com/opencontainers/image-spec/blob/master/image-index.md)
- [OCI Image Manifest](https://github.com/opencontainers/image-spec/blob/master/manifest.md)
- [Docker Image Manifest List](https://docs.docker.com/registry/spec/manifest-v2-2/#manifest-list)
- [Docker Image Manifest](https://docs.docker.com/registry/spec/manifest-v2-2/#image-manifest)

## *Signature* Property Descriptions

- **`signed`** *object*

  This REQUIRED property provides the signed content.

  - **`iat`** *integer*

    This OPTIONAL property identities the time at which the manifests were presented to the notary. This field is based on [RFC 7519 Section 4.1.6](https://tools.ietf.org/html/rfc7519#section-4.1.6). When used, it does not imply the issue time of any signature in the `signatures` property.

  - **`nbf`** *integer*

    This OPTIONAL property identifies the time before which the  signed content MUST NOT be accepted for processing. This field is based on [RFC 7519 Section 4.1.5](https://tools.ietf.org/html/rfc7519#section-4.1.5).

  - **`exp`** *integer*

    This OPTIONAL property identifies the expiration time on or after which the signed content MUST NOT be accepted for processing. This field is based on [RFC 7519 Section 4.1.4](https://tools.ietf.org/html/rfc7519#section-4.1.4).

  - **`digest`** *string*

    This REQUIRED property is the *digest* of the target manifest, conforming to the requirements outlined in [Digests](https://github.com/opencontainers/image-spec/blob/master/descriptor.md#digests). If the actual content is fetched according to the *digest*, implementations MUST verify the content against the *digest*.

  - **`size`** *integer*

    This REQUIRED property is the *size* of the target manifest. If the actual content is fetched according the *digest*, implementations MUST verify the content against the *size*.

  - **`references`** *array of strings*

    This OPTIONAL property claims the manifest references of its origin. The format of the value MUST matches the [*reference* grammar](https://github.com/docker/distribution/blob/master/reference/reference.go). With used, the `x509` signatures are valid only if the domain names of all references match the Common Name (`CN`) in the `Subject` field of the certificate.

- **`signatures`** *array of objects*

  This REQUIRED property provides the signatures of the signed content. The entire signature file is valid if any signature in `signatures` is valid. The `signature` object is influenced by JSON Web Signature (JWS) at [RFC 7515](https://tools.ietf.org/html/rfc7515).

  - **`typ`** *string*

    This REQUIRED property identifies the signature type. Implementations MUST support at least the following types

    - `x509`: X.509 public key certificates. Implementations MUST verify that the certificate of the signing key has the `digitalSignature` `Key Usage` extension ([RFC 5280 Section 4.2.1.3](https://tools.ietf.org/html/rfc5280#section-4.2.1.3)).

    Implementations MAY support the following types

    - `tuf`: [The update framework](https://theupdateframework.io/).

  - **`sig`** *string*

    This REQUIRED property provides the base64-encoded signature binary of the specified signature type.

  - **`alg`** *string*

    This REQUIRED property for the `x509` type identifies the cryptographic algorithm used to sign the content. This field is based on [RFC 7515 Section 4.1.1](https://tools.ietf.org/html/rfc7515#section-4.1.1).

  - **`x5c`** *array of strings*

    This OPTIONAL property for the `x509` type contains the X.509 public key certificate or certificate chain corresponding to the key used to digitally sign the content. The certificates are encoded in base64. This field is based on [RFC 7515 Section 4.1.6](https://tools.ietf.org/html/rfc7515#section-4.1.6).

  - **`kid`** *string*

    This OPTIONAL property for the `x509` type is a hint (key ID) indicating which key was used to sign the content. This field is based on [RFC 7515 Section 4.1.4](https://tools.ietf.org/html/rfc7515#section-4.1.4).

## Example Signatures

Example showing a formatted `x509` signature file [examples/x509_x5c.nv2.json](examples/x509_x5c.nv2.json) with certificates provided by `x5c`:

```json
{
    "signed": {
        "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
        "size": 528,
        "references": [
            "registry.example.com/example:latest",
            "registry.example.com/example:v1.0"
        ],
        "exp": 1626772975,
        "nbf": 1595560975,
        "iat": 1595560975
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "bmVxnJIBDAP9TVDWdpQwfI5OkdDo75wB+yigjJbFa74fhNmzm7GIu9jCnkmcwBLMO+XuUkTbP+5eFYC+N3B+nhrsD4ci4dsVGJY8XBP0YCTc513hO+LjxP5ITP4gd0DZsSU7eSgnqo+Yd7DO0ZpL3YRAOWB+ZS1EHYwcm0VGL0YmItrbIf6irPMpVyfrkjKTywtQfLSZ2K3KPM4OKM9aTaDXy+MeSpnN22xDyeor/eUSi59OOUILfWGTIYsC0jpryE/dNN6dHej1K+AmG/wCnlmexUNvY7WN1YsqQKiti9NM6nfsGcNsjP4D9DF+spoGtoE3ZLjrUqcePQ0v2mZGtQ==",
            "alg": "RS256",
            "x5c": [
                "MIIDmzCCAoOgAwIBAgIUCaGfGGjg9hCjxS2KncvpdlQZ6kUwDQYJKoZIhvcNAQELBQAwXTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEWMBQGA1UEAwwNKi5leGFtcGxlLmNvbTAeFw0yMDA3MjQwMzE2NDBaFw0yMTA3MjQwMzE2NDBaMF0xCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFjAUBgNVBAMMDSouZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDVgGfJ3gb41R+WfDFfSck6HkoJHTMpbGcowmP8gCmVU3E3NMF6y3LrqaYM8UZAJSSe28dmkhdWvlBeR6UhUpa/f2HxsOY3w/lgBiVfB2rrRegC6WEcqkWNh4JewOLOwEjvdjPiaaCZpgIvHbyEiT3hJPRTGOfNvoeidXSOLiEpAF10HRO0OXO+A06LyiY2qfB5HrCOrmu/GNSch1oICrW6gJwQY5JxSULIRTcZV7496rLtKfw/DnkWoqc2JJUIslS2IvqdmrylOtWXUeErDKrGF/TtC476av5ssecCf4nGTbSSiu+eW85xug3Urgh54Ei0ztIfxdta2frWGg+lQ6CZAgMBAAGjUzBRMB0GA1UdDgQWBBR/2xw5wUZHuuiZJAWOHRB5s+hynjAfBgNVHSMEGDAWgBR/2xw5wUZHuuiZJAWOHRB5s+hynjAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQAu3hJCaGvPddR3dPJCXW09BPd2gz1d1QB32jEhUrfB4YO+N7ib3ck4i743rC7AvuznwVlkw6FIaJTvcdwwDanWZhOtXOy1wcVXYFFn3HeI+pYSpqSuYMo/wsdaFEPI0D20R+BeM8FOAQuC30Ve9lr6E0xO0vDZZ/ZqVgL6J2yHDQVJGaXbLIRkv19gOh5IkktmpUCnTFpgj+EhHZJIhVpSR1IUh7FPVslwnUsjgOxWdKJn3Vcbep+Gw4zdP1XMTOStagEjaic0cg6Ls4Av/9YNGcwTfzzzoAvCWekTuf+LcvAa2JJvJnXl6azwLUFHTM664yThG+3MgiMkDhd7sVn1"
            ]
        }
    ]
}
```

Example showing a formatted `x509` signature file [examples/x509_kid.nv2.json](examples/x509_kid.nv2.json) with certificates referenced by `kid`:

```json
{
    "signed": {
        "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
        "size": 528,
        "references": [
            "registry.example.com/example:latest",
            "registry.example.com/example:v1.0"
        ],
        "exp": 1626773010,
        "nbf": 1595561010,
        "iat": 1595561010
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "i1pEMz7UUU0RWGefQOlhcTxL8f+oh4FvQRPGJbyUa7/oGb7xNXKLon3wlmulOrw9MnLRbr/j9jC8nnZm3wuIU4JQf5qL/bk80Atid/VylNdvJAr4aBNu0qMsoEkeQ1N6LAM4JsFMX0Q/T5vfqzJc+S3+GU1oJMGTmvT0lpRdD4sn8EMX2L/+VIuziAgQRHnFv2HNUYYbLdPIo2GQ6gCCfhNnak/PGznFQzxP7QXfkdkLJ18WSu0X9zD156EF1MlYxk9Hz+WuiaOo/P69V2UxGFeIyKHPQ7Q8eGlFkUypBW66HlDK62O1+jxXNtNW1zB9UBEqBiEb+vTEQllms/94Cg==",
            "alg": "RS256",
            "kid": "6WR3:WGKO:JOQM:6SNH:MD7N:EVFT:XDXH:SOA4:36CN:D6FJ:ZEQ5:3C46"
        }
    ]
}
```