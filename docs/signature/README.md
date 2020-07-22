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

  - **`manifests`** *array of objects*

    This REQUIRED property references manifests presented to notary for certifying.

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

    Implementations SHOULD support the following types

    - `gpg`: [GnuPG](https://www.gnupg.org/) detached signatures.

    Implementations MAY support the following types

    - `tuf`: [The update framework](https://theupdateframework.io/).

  - **`sig`** *string*

    This REQUIRED property provides the base64-encoded signature binary of the specified signature type.

  - **`iss`** *string*

    This REQUIRED property for the `gpg` type indicates the name of the signer / issuer. Implementations MUST verify the issuer name against the user ID of the `gpg` signature.

  - **`alg`** *string*

    This REQUIRED property for the `x509` type identifies the cryptographic algorithm used to sign the content. This field is based on [RFC 7515 Section 4.1.1](https://tools.ietf.org/html/rfc7515#section-4.1.1).

  - **`x5c`** *array of strings*

    This OPTIONAL property for the `x509` type contains the X.509 public key certificate or certificate chain corresponding to the key used to digitally sign the content. The certificates are encoded in base64. This field is based on [RFC 7515 Section 4.1.6](https://tools.ietf.org/html/rfc7515#section-4.1.6).

  - **`kid`** *string*

    This OPTIONAL property for the `x509` type is a hint (key ID) indicating which key was used to sign the content. This field is based on [RFC 7515 Section 4.1.4](https://tools.ietf.org/html/rfc7515#section-4.1.4).

## Example Signatures

Example showing a formatted `gpg` signature file [examples/gpg.nv2.json](examples/gpg.nv2.json):

```json
{
    "signed": {
        "exp": 1626938668,
        "nbf": 1595402668,
        "iat": 1595402668,
        "manifests": [
            {
                "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
                "size": 528,
                "references": [
                    "registry.example.com/example:latest",
                    "registry.example.com/example:v1.0"
                ]
            }
        ]
    },
    "signatures": [
        {
            "typ": "gpg",
            "sig": "wsDcBAABCAAQBQJfF+msCRDvXc1GQtqlQgAAJwUMAAZdQdDJCCoHl8VXyseeU2WB7/1Ip+Ei++C/ZFtA4ncsifdi28B4FQlAjOPbIPlIsldl7KtL6aMloHiQTm/sBl+aEys4Z2/xTSu+5//jcUeWwtDEiSur2K2w3F7RmDWhGFSjgXvlkPMt7iaCqy6dEPvrLSYXRgBAVnUEdtS/L/ANMSupt+FZh2AISyWL6TZKOKVcxKSiJ0SR72L7DYE1E6edBPsPHivc485qwRljvjG9q8WwWusvZM4OjBLaddn7d83+R4YQNqGBp8RGvEGiw9oWzu3f+2MCeT5USQWFcIr+KQHJi4R/0cqKGQ9TarUS1vIKSiasmnqufVCi2Ucb+5sj8oaI7/DIyCxYiv0lX1pJE1j/yuS1XtDVzn7J1enkuP9TgiRNSzjZJUc5rLa3IwyuXGaJOtUJm60ma5WU/LoUe1sqC4jpQ2nU4UNHH14KnoeElJzE1WknmmrGck2ewx0yiln7wCrwKQ5dC0kS8suJBoZD7Ms7SAwDMbHL5oj1fg==",
            "iss": "Demo User \u003cdemo@example.com\u003e"
        }
    ]
}
```

Example showing a formatted `x509` signature file [examples/x509_x5c.nv2.json](examples/x509_x5c.nv2.json) with certificates provided by `x5c`:

```json
{
    "signed": {
        "exp": 1626938793,
        "nbf": 1595402793,
        "iat": 1595402793,
        "manifests": [
            {
                "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
                "size": 528,
                "references": [
                    "registry.example.com/example:latest",
                    "registry.example.com/example:v1.0"
                ]
            }
        ]
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "uFKaCyQ4MtVHemfLVq5gYZyeiClS20tksXzP7hhpeqqjCNK9DiHnoDpkq91sutLqd1o6RCxpfFVuGXy20oqRu1/ZoXXAVC3y7lS6z/wqJ4VDBKSj/H6xyYn7pH3GE8GHR6kjFPqrGsl/OS4yYH2oNXEm9W8Pju2wC381+FCgf4LNf7k6u2Uf4Fb0/Fl40qzvr0m2Fv5pXtRY+wdJctqJb+t408VcXJkNj0U7xoOe0zUr3l1A6xLYqjd0ZY08JBQ8FQul0Vpxrmg0Xdtwd/wEolvia48lxD1x7yphW5bFvJOTd62rOJgd4uI7jYJF3ZLmwjY+geMk5e6Wkp5OyXGjXw==",
            "alg": "RS256",
            "x5c": [
                "MIIDmzCCAoOgAwIBAgIUFSzsIT4/pKtGzywuZWWE7ydiLBIwDQYJKoZIhvcNAQELBQAwXTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEWMBQGA1UEAwwNKi5leGFtcGxlLmNvbTAeFw0yMDA3MjIwMzA2MTBaFw0yMTA3MjIwMzA2MTBaMF0xCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFjAUBgNVBAMMDSouZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDM0MNLy/f1SyRM0ZQu3AtJnCU3O5x8nnOeV1mySmZNr2SCqR8+jENAoKE5FrrSi2ffMnFPP/7DqGnbb9+b1nD9ucFNsI1iW7IrF/GlqOM7jJhUMNnOyatz8mddtQgXr3SZ9bigbc/lxuVGacvi64DewoWzMFr4ZMGq8ik7aDnHryUDwXJFE+KGNbsReO1ePqKmPiLvkLG4sBTqeTuCk+Grrr5t1COujwuFWfhMjmRfq34QGqUZ3SHJYXPzOAxgV3fCmBP9IgHuSv/b1udx5Htf1BV7WlARtXfE21xuA6FM1Gq0pANUhcRF39KJRu4/RBZBmAxg7ces8hrZWTQ4LTo/AgMBAAGjUzBRMB0GA1UdDgQWBBR2pI+c2dexlOZCXLy84Baqu8NR8DAfBgNVHSMEGDAWgBR2pI+c2dexlOZCXLy84Baqu8NR8DAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCH2tChjmvs6/2acw+cJYWkEExdXMEyjdUvqEIcs7W7Ce32My7RcMtJxybtqjV+UVghEVUzq1pNf0Dt5FhFkC6BDHnHv2SIO9jq2TvfDUcJgMMgwSZdSaISmxk+iFD9Cll+RU8KgeoYSnwojOixTksyeBRi5rePdO5smz/n4Bd4ToluKaw42tdWhF4SMgx2Y1nlyHpFlkdUYtJ6D8rOvbVRGQaxo8Td3mWCWPMBYcGvjwO9ESCP1JAK+Z6WXD46JWilsIUd3Y+0NrfvOYKUdhLWuz9LrQ5060qi1pHfYBOTAbyXfnW97EB3TAuMtqBBe6h3VNw00c1p7qrilE1Of9uN"
            ]
        }
    ]
}
```

Example showing a formatted `x509` signature file [examples/x509_kid.nv2.json](examples/x509_kid.nv2.json) with certificates referenced by `kid`:

```json
{
    "signed": {
        "exp": 1626938803,
        "nbf": 1595402803,
        "iat": 1595402803,
        "manifests": [
            {
                "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
                "size": 528,
                "references": [
                    "registry.example.com/example:latest",
                    "registry.example.com/example:v1.0"
                ]
            }
        ]
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "JQWZ9/H1oQyuBxyYsPaKE7Xh4+U0uITmPwRpPOBNFOxe0qnIxmkyQD0g/W5eQRt1Jwa+2hn35EamqERmdT6ji2f/6haqfIwcSjjaiDu1q1sXGDQhk+ZVzOCCcqRaFNV0fPRwaVMwxeizTUy9ENe1ksZqAPI1SCyzSr6pAa5xKeoJXFUToPjjMm1VMzwj9qwphGk8sXhSqCAt9P9/PV50pxuWU1Dbe+y6M6ZlnET2YIswBze3EjloROQtKniy87Xb2ZwJp81R0XUbWRk5LqiJVT9jDN8/RMDBvMj8eymrjbcb/F3TugvZ99jkkEVjk6tH+dvXu9HbS9HtGh0KRO1XQw==",
            "alg": "RS256",
            "kid": "SE4Z:F3CT:DZ64:ONJX:6CRE:PTD2:Z755:DG7W:TSUI:I5GZ:RFKR:JCHY"
        }
    ]
}
```