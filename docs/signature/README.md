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
        "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
        "size": 528,
        "references": [
            "registry.example.com/example:latest",
            "registry.example.com/example:v1.0"
        ],
        "exp": 1627555319,
        "nbf": 1596019319,
        "iat": 1596019319
    }
}
```

The signature of the above would be represented as:

``` JSON
{
    "signature": {
        "typ": "x509",
        "sig": "UFqN24K2fLj7/h2slM68PLTfF9CDhrEVGuMQ8m3kkQJ4SKusj9fNxYV78tTiedqB+E8SqVH66mZbdlTrVQFJAd7aL2c3NZFfo92pE9SaHnqEDqnnGWXGRVjtBRM13YyRDm2wD8aRyuL5jEDUkTw7jBLY0+LfKHMDuYCsOOzvedof7aiaFc3qA+qKiW53jn2uEGCFfAs0LmsNafGfAtVmdGSO4zX4fdnQFAGT8sbUmL71uXl9W1B6tGeLfx5nBoQUvtplQipHly/yMQvWw7qMXsaAsf/BbGDmivN06CRahSb7VOwNq6K7Py4zYeiW40hEFVz9L7/5xT5XI1unKPZDuw==",
        "alg": "RS256",
        "x5c": [
            "MIIDszCCApugAwIBAgIUL1anEU/yJy67VJTbHkNX0bBNAnEwDQYJKoZIhvcNAQELBQAwaTEdMBsGA1UEAwwUcmVnaXN0cnkuZXhhbXBsZS5jb20xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTAeFw0yMDA3MjcxNDQzNDZaFw0yMTA3MjcxNDQzNDZaMGkxHTAbBgNVBAMMFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDkKwAcV44psjN8nno1eZ3zv1ZKUhJAoxwBOIGfIxIe+iHtpXLvFFVwk5Jbxu+Pkig2N4B3Ilrj/Vryi0hxp4mag02M733bXLRENSOFONRkslpO8zHUN5pYdnhTSwYTLap1+1bgcFSuUXLWieqZB6qc7kiv3bj3SPaf42+s48V49t/OpXxLtgiWL9XkuDTZctpJJA4vHHk6Ou0bcg7iGm+L1xwIfb8Ml4oWvT0SF35fgW08bbLXZ2v1XCLRsrWUgbq4U+KxtEpG3XIYcYhKx1rIrUhfEJkuHzgPglM11gG5W+Cyfg+wfOJig5q6axIKWzIf6C8m8lmy6bM+N5EsD9SvAgMBAAGjUzBRMB0GA1UdDgQWBBTf1hM6/ibGF+u/SVAK88FUMjzRoTAfBgNVHSMEGDAWgBTf1hM6/ibGF+u/SVAK88FUMjzRoTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBgvVau5+2wAuCsmOyyG28h1zyC4IPmMmpRZTDOp/pLdwXeHjJr8kEC3l92qJEvc+WAboJ1RoucHycUe7RWh2C6ZF/WPCBLyWGwnlyqGyRM9/j86UJ1OgiuZl7kl9zxwWoaxPBCmHa0RHowdQB7AVlpqg1c7FhKjhUCBmGT4Ve8tV0hdZtrZoQV+6xHPbUd37KV1B1Bmfo3o4ekoJKhUu99Eo03OpE3JLtM13A1HxABEuQGHTI0tycDBBdRn3b03HoIhU0VnqjvpV1KPvsrgYi/0VStLNezZPgGe0fG3Xgy8yekdB9NMUn+zZLATI4+z8j4QH5Wj5ZPaUkyoAD2oUJO"
        ]
    }
}
```

### Signature Persisted within an OCI Artifact Enabled Registry

Both values are persisted in a `signature.json` file. The file would be submitted to a registry as an Artifact with null layers.
The `signature.json` would be persisted within the `manifest.config` object

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
        "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
        "size": 528,
        "references": [
            "registry.example.com/example:latest",
            "registry.example.com/example:v1.0"
        ],
        "exp": 1627555319,
        "nbf": 1596019319,
        "iat": 1596019319
    },
    "signature": {
        "typ": "x509",
        "sig": "UFqN24K2fLj7/h2slM68PLTfF9CDhrEVGuMQ8m3kkQJ4SKusj9fNxYV78tTiedqB+E8SqVH66mZbdlTrVQFJAd7aL2c3NZFfo92pE9SaHnqEDqnnGWXGRVjtBRM13YyRDm2wD8aRyuL5jEDUkTw7jBLY0+LfKHMDuYCsOOzvedof7aiaFc3qA+qKiW53jn2uEGCFfAs0LmsNafGfAtVmdGSO4zX4fdnQFAGT8sbUmL71uXl9W1B6tGeLfx5nBoQUvtplQipHly/yMQvWw7qMXsaAsf/BbGDmivN06CRahSb7VOwNq6K7Py4zYeiW40hEFVz9L7/5xT5XI1unKPZDuw==",
        "alg": "RS256",
        "x5c": [
            "MIIDszCCApugAwIBAgIUL1anEU/yJy67VJTbHkNX0bBNAnEwDQYJKoZIhvcNAQELBQAwaTEdMBsGA1UEAwwUcmVnaXN0cnkuZXhhbXBsZS5jb20xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTAeFw0yMDA3MjcxNDQzNDZaFw0yMTA3MjcxNDQzNDZaMGkxHTAbBgNVBAMMFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDkKwAcV44psjN8nno1eZ3zv1ZKUhJAoxwBOIGfIxIe+iHtpXLvFFVwk5Jbxu+Pkig2N4B3Ilrj/Vryi0hxp4mag02M733bXLRENSOFONRkslpO8zHUN5pYdnhTSwYTLap1+1bgcFSuUXLWieqZB6qc7kiv3bj3SPaf42+s48V49t/OpXxLtgiWL9XkuDTZctpJJA4vHHk6Ou0bcg7iGm+L1xwIfb8Ml4oWvT0SF35fgW08bbLXZ2v1XCLRsrWUgbq4U+KxtEpG3XIYcYhKx1rIrUhfEJkuHzgPglM11gG5W+Cyfg+wfOJig5q6axIKWzIf6C8m8lmy6bM+N5EsD9SvAgMBAAGjUzBRMB0GA1UdDgQWBBTf1hM6/ibGF+u/SVAK88FUMjzRoTAfBgNVHSMEGDAWgBTf1hM6/ibGF+u/SVAK88FUMjzRoTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBgvVau5+2wAuCsmOyyG28h1zyC4IPmMmpRZTDOp/pLdwXeHjJr8kEC3l92qJEvc+WAboJ1RoucHycUe7RWh2C6ZF/WPCBLyWGwnlyqGyRM9/j86UJ1OgiuZl7kl9zxwWoaxPBCmHa0RHowdQB7AVlpqg1c7FhKjhUCBmGT4Ve8tV0hdZtrZoQV+6xHPbUd37KV1B1Bmfo3o4ekoJKhUu99Eo03OpE3JLtM13A1HxABEuQGHTI0tycDBBdRn3b03HoIhU0VnqjvpV1KPvsrgYi/0VStLNezZPgGe0fG3Xgy8yekdB9NMUn+zZLATI4+z8j4QH5Wj5ZPaUkyoAD2oUJO"
        ]
    }
}
```

Example showing a formatted `x509` signature file [examples/x509_kid.nv2.json](examples/x509_kid.nv2.json) with certificates referenced by `kid`:

```json
{
    "signed": {
        "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
        "size": 528,
        "references": [
            "registry.example.com/example:latest",
            "registry.example.com/example:v1.0"
        ],
        "exp": 1627554920,
        "nbf": 1596018920,
        "iat": 1596018920
    },
    "signature": {
        "typ": "x509",
        "sig": "emzP9ygJD3y2ZWMYGO/wyqOhaSxrhd4ZdmjC9Zd+Ba7gGmGzBylsY1CskyZw389Hz2Z0xA6AQLhaNBbbqyxuAxVXtataMRsqCl/cgyNbyYU1URB2aTUZY/3V4iJzH1O/QfwSkpQa3aN1OCL8uMBNCtM6Rde9+SX8Q8XNMByDbuXtyPDvnKunZxpofEn2ibLe2Cm3o+MTK4pgxacEWeld85gTb06NicARf7mcVj7bflLyUIgel4qvmdqT6896Gtd2ES1KawvyjoEyskdlVlneSTdEKGRYxfchwIUK4E7p3EtTnmj+FuD9MpCtP0M4CQiOr19j0NtQe2bHuTo4bwtjuw==",
        "alg": "RS256",
        "kid": "XP5O:Y7W2:PRB6:O355:56CC:P3A6:CBDV:EDMN:QZCK:W5PO:QMV3:T2LX"
    }
}
```

[distribution-spec]:    https://github.com/opencontainers/distribution-spec
[oci-artifacts]:        https://github.com/opencontainers/artifacts
[oci-manifest]:         https://github.com/opencontainers/image-spec/blob/master/manifest.md
[oci-manifest-list]:    https://github.com/opencontainers/image-spec/blob/master/image-index.md

