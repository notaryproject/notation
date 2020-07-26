# Notary V2 Signature Specification

This section defines the signature file, which is in JSON format with no whitespaces. Its JSON schema is available at [schema.json](schema.json).

## Signature Goals

- Offline signature creation
- Persistance within an [OCI Artifact][oci-artifacts] enabled, [distribution-spec][distribution-spec] based registry
- Artifact and signature copying within and across [OCI Artifact][oci-artifacts] enabled, [distribution-spec][distribution-spec] based registries
- Support public registry acquisition of content - where the public registry may host certified content as well as public, non-certified content
- Support private registries, where public content may be copied to, and new content originated within
- Air-gapped environments, where the originating registry of content is not accessable
- Multiple signatures per artifact, enabling the originating vendor signature, public registry certification and user/environment signatures
- Maintain the original artifact digest and collection of associated tags, supporting dev/ops deployment definitions

## Signature

A Notary v2 signature is clear-signed signature of manifest metadata, including but not limited to

- [OCI Image Index](https://github.com/opencontainers/image-spec/blob/master/image-index.md)
- [OCI Image Manifest](https://github.com/opencontainers/image-spec/blob/master/manifest.md)
- [Docker Image Manifest List](https://docs.docker.com/registry/spec/manifest-v2-2/#manifest-list)
- [Docker Image Manifest](https://docs.docker.com/registry/spec/manifest-v2-2/#image-manifest)

### Signing an Artifact Manifest

For any [OCI Artifact][oci-artifacts] submitted to a registry, an [OCI Manifest][oci-manifest] and an optional [OCI Manifest List/Index][oci-manifest-list] is required.

The nv2 prototype signs a combination of:

- Key properties
- The target artifact manifest digest
- *optional:* list of associated tags

#### Generating a self-signed x509 cert

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

An nv2 client would generate the following content to be signed:

``` JSON
{
  "signed": {
      "exp": 1626938793,
      "nbf": 1595402793,
      "iat": 1595402793,
      "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
      "size": 528,
      "references": [
          "registry.example.com/hello-world:latest",
          "registry.example.com/hello-world:v1.0"
      ]
  }
```

The signature of the above would be represented as:

``` JSON
{
  "signature": {
    "typ": "x509",
    "sig": "uFKaCyQ4MtVHemfLVq5gYZyeiClS20tksXzP7hhpeqqjCNK9DiHnoDpkq91sutLqd1o6RCxpfFVuGXy20oqRu1/ZoXXAVC3y7lS6z/wqJ4VDBKSj/H6xyYn7pH3GE8GHR6kjFPqrGsl/OS4yYH2oNXEm9W8Pju2wC381+FCgf4LNf7k6u2Uf4Fb0/Fl40qzvr0m2Fv5pXtRY+wdJctqJb+t408VcXJkNj0U7xoOe0zUr3l1A6xLYqjd0ZY08JBQ8FQul0Vpxrmg0Xdtwd/wEolvia48lxD1x7yphW5bFvJOTd62rOJgd4uI7jYJF3ZLmwjY+geMk5e6Wkp5OyXGjXw==",
    "alg": "RS256",
    "x5c": [
      "MIIDmzCCAoOgAwIBAgIUFSzsIT4/pKtGzywuZWWE7ydiLBIwDQYJKoZIhvcNAQELBQAwXTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEWMBQGA1UEAwwNKi5leGFtcGxlLmNvbTAeFw0yMDA3MjIwMzA2MTBaFw0yMTA3MjIwMzA2MTBaMF0xCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFjAUBgNVBAMMDSouZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDM0MNLy/f1SyRM0ZQu3AtJnCU3O5x8nnOeV1mySmZNr2SCqR8+jENAoKE5FrrSi2ffMnFPP/7DqGnbb9+b1nD9ucFNsI1iW7IrF/GlqOM7jJhUMNnOyatz8mddtQgXr3SZ9bigbc/lxuVGacvi64DewoWzMFr4ZMGq8ik7aDnHryUDwXJFE+KGNbsReO1ePqKmPiLvkLG4sBTqeTuCk+Grrr5t1COujwuFWfhMjmRfq34QGqUZ3SHJYXPzOAxgV3fCmBP9IgHuSv/b1udx5Htf1BV7WlARtXfE21xuA6FM1Gq0pANUhcRF39KJRu4/RBZBmAxg7ces8hrZWTQ4LTo/AgMBAAGjUzBRMB0GA1UdDgQWBBR2pI+c2dexlOZCXLy84Baqu8NR8DAfBgNVHSMEGDAWgBR2pI+c2dexlOZCXLy84Baqu8NR8DAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCH2tChjmvs6/2acw+cJYWkEExdXMEyjdUvqEIcs7W7Ce32My7RcMtJxybtqjV+UVghEVUzq1pNf0Dt5FhFkC6BDHnHv2SIO9jq2TvfDUcJgMMgwSZdSaISmxk+iFD9Cll+RU8KgeoYSnwojOixTksyeBRi5rePdO5smz/n4Bd4ToluKaw42tdWhF4SMgx2Y1nlyHpFlkdUYtJ6D8rOvbVRGQaxo8Td3mWCWPMBYcGvjwO9ESCP1JAK+Z6WXD46JWilsIUd3Y+0NrfvOYKUdhLWuz9LrQ5060qi1pHfYBOTAbyXfnW97EB3TAuMtqBBe6h3VNw00c1p7qrilE1Of9uN"
    ]
  }
}
```

### Signature Persisted within an OCI Artifact Enabled Registry

Both values are persisted in a `signature.json` file. The file would be submitted to a registry as an Artifact with null layers.
The `signature.json` would be persisted wthin the `manifest.config` object

``` SHELL
oras push \
  registry.example.com/hello-world:v1.0 \
  --manifest-config signature.json:application/vnd.cncf.notary.config.v2+json
```

Would push the following manifest:

``` JSON
{
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.cncf.notary.config.v2+json",
    "size": 1906,
    "digest": "sha256:c7848182f2c817415f0de63206f9e4220012cbb0bdb750c2ecf8020350239814"
  },
  "layers": []
}
```

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

- **`signature`** *string*

  This REQUIRED property provides the signature of the signed content. The entire signature file is valid if any signature is valid. The `signature` object is influenced by JSON Web Signature (JWS) at [RFC 7515](https://tools.ietf.org/html/rfc7515).

  - **`typ`** *string*

    This REQUIRED property identifies the signature type. Implementations MUST support at least the following types

    - `x509`: X.509 public key certificates. Implementations MUST verify that the certificate of the signing key has the `digitalSignature` `Key Usage` extension ([RFC 5280 Section 4.2.1.3](https://tools.ietf.org/html/rfc5280#section-4.2.1.3)).

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

### x509 Signature

Example showing a formatted `x509` signature file [examples/x509_x5c.nv2.json](examples/x509_x5c.nv2.json) with certificates provided by `x5c`:

```json
{
  "signed": {
    "exp": 1626938793,
    "nbf": 1595402793,
    "iat": 1595402793,
    "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
    "size": 528,
    "references": [
      "registry.example.com/example:latest",
      "registry.example.com/example:v1.0"
    ]
  },
  "signature": {
    "typ": "x509",
    "sig": "uFKaCyQ4MtVHemfLVq5gYZyeiClS20tksXzP7hhpeqqjCNK9DiHnoDpkq91sutLqd1o6RCxpfFVuGXy20oqRu1/ZoXXAVC3y7lS6z/wqJ4VDBKSj/H6xyYn7pH3GE8GHR6kjFPqrGsl/OS4yYH2oNXEm9W8Pju2wC381+FCgf4LNf7k6u2Uf4Fb0/Fl40qzvr0m2Fv5pXtRY+wdJctqJb+t408VcXJkNj0U7xoOe0zUr3l1A6xLYqjd0ZY08JBQ8FQul0Vpxrmg0Xdtwd/wEolvia48lxD1x7yphW5bFvJOTd62rOJgd4uI7jYJF3ZLmwjY+geMk5e6Wkp5OyXGjXw==",
    "alg": "RS256",
    "x5c": [
      "MIIDmzCCAoOgAwIBAgIUFSzsIT4/pKtGzywuZWWE7ydiLBIwDQYJKoZIhvcNAQELBQAwXTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEWMBQGA1UEAwwNKi5leGFtcGxlLmNvbTAeFw0yMDA3MjIwMzA2MTBaFw0yMTA3MjIwMzA2MTBaMF0xCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxFjAUBgNVBAMMDSouZXhhbXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDM0MNLy/f1SyRM0ZQu3AtJnCU3O5x8nnOeV1mySmZNr2SCqR8+jENAoKE5FrrSi2ffMnFPP/7DqGnbb9+b1nD9ucFNsI1iW7IrF/GlqOM7jJhUMNnOyatz8mddtQgXr3SZ9bigbc/lxuVGacvi64DewoWzMFr4ZMGq8ik7aDnHryUDwXJFE+KGNbsReO1ePqKmPiLvkLG4sBTqeTuCk+Grrr5t1COujwuFWfhMjmRfq34QGqUZ3SHJYXPzOAxgV3fCmBP9IgHuSv/b1udx5Htf1BV7WlARtXfE21xuA6FM1Gq0pANUhcRF39KJRu4/RBZBmAxg7ces8hrZWTQ4LTo/AgMBAAGjUzBRMB0GA1UdDgQWBBR2pI+c2dexlOZCXLy84Baqu8NR8DAfBgNVHSMEGDAWgBR2pI+c2dexlOZCXLy84Baqu8NR8DAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCH2tChjmvs6/2acw+cJYWkEExdXMEyjdUvqEIcs7W7Ce32My7RcMtJxybtqjV+UVghEVUzq1pNf0Dt5FhFkC6BDHnHv2SIO9jq2TvfDUcJgMMgwSZdSaISmxk+iFD9Cll+RU8KgeoYSnwojOixTksyeBRi5rePdO5smz/n4Bd4ToluKaw42tdWhF4SMgx2Y1nlyHpFlkdUYtJ6D8rOvbVRGQaxo8Td3mWCWPMBYcGvjwO9ESCP1JAK+Z6WXD46JWilsIUd3Y+0NrfvOYKUdhLWuz9LrQ5060qi1pHfYBOTAbyXfnW97EB3TAuMtqBBe6h3VNw00c1p7qrilE1Of9uN"
    ]
  }
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
  "signature": {
    "typ": "x509",
    "sig": "JQWZ9/H1oQyuBxyYsPaKE7Xh4+U0uITmPwRpPOBNFOxe0qnIxmkyQD0g/W5eQRt1Jwa+2hn35EamqERmdT6ji2f/6haqfIwcSjjaiDu1q1sXGDQhk+ZVzOCCcqRaFNV0fPRwaVMwxeizTUy9ENe1ksZqAPI1SCyzSr6pAa5xKeoJXFUToPjjMm1VMzwj9qwphGk8sXhSqCAt9P9/PV50pxuWU1Dbe+y6M6ZlnET2YIswBze3EjloROQtKniy87Xb2ZwJp81R0XUbWRk5LqiJVT9jDN8/RMDBvMj8eymrjbcb/F3TugvZ99jkkEVjk6tH+dvXu9HbS9HtGh0KRO1XQw==",
    "alg": "RS256",
    "kid": "SE4Z:F3CT:DZ64:ONJX:6CRE:PTD2:Z755:DG7W:TSUI:I5GZ:RFKR:JCHY"
  }
}
```

[distribution-spec]:    https://github.com/opencontainers/distribution-spec
[oci-artifacts]:        https://github.com/opencontainers/artifacts
[oci-manifest]:         https://github.com/opencontainers/image-spec/blob/master/manifest.md
[oci-manifest-list]:    https://github.com/opencontainers/image-spec/blob/master/image-index.md
