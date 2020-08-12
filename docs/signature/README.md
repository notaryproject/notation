# Notary V2 Signature Specification

This section defines the signature file, which is a [JWT](https://tools.ietf.org/html/rfc7519) variant.

## Signature Goals

- Offline signature creation
- Persistence within an [OCI Artifact][oci-artifacts] enabled, [distribution-spec][distribution-spec] based registry
- Artifact and signature copying within and across [OCI Artifact][oci-artifacts] enabled, [distribution-spec][distribution-spec] based registries
- Support public registry acquisition of content - where the public registry may host certified content as well as public, non-certified content
- Support private registries, where public content may be copied to, and new content originated within
- Air-gapped environments, where the originating registry of content is not accessible
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

An nv2 client would generate the following header and claims to be signed.

The header would be a base64 URL encoded string without paddings:

```
eyJ0eXAiOiJ4NTA5IiwiYWxnIjoiUlMyNTYiLCJ4NWMiOlsiTUlJRHN6Q0NBcHVnQXdJQkFnSVVMMWFuRVUveUp5NjdWSlRiSGtOWDBiQk5BbkV3RFFZSktvWklodmNOQVFFTEJRQXdhVEVkTUJzR0ExVUVBd3dVY21WbmFYTjBjbmt1WlhoaGJYQnNaUzVqYjIweEZEQVNCZ05WQkFvTUMyVjRZVzF3YkdVZ2FXNWpNUXN3Q1FZRFZRUUdFd0pWVXpFVE1CRUdBMVVFQ0F3S1YyRnphR2x1WjNSdmJqRVFNQTRHQTFVRUJ3d0hVMlZoZEhSc1pUQWVGdzB5TURBM01qY3hORFF6TkRaYUZ3MHlNVEEzTWpjeE5EUXpORFphTUdreEhUQWJCZ05WQkFNTUZISmxaMmx6ZEhKNUxtVjRZVzF3YkdVdVkyOXRNUlF3RWdZRFZRUUtEQXRsZUdGdGNHeGxJR2x1WXpFTE1Ba0dBMVVFQmhNQ1ZWTXhFekFSQmdOVkJBZ01DbGRoYzJocGJtZDBiMjR4RURBT0JnTlZCQWNNQjFObFlYUjBiR1V3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRRGtLd0FjVjQ0cHNqTjhubm8xZVozenYxWktVaEpBb3h3Qk9JR2ZJeEllK2lIdHBYTHZGRlZ3azVKYnh1K1BraWcyTjRCM0lscmovVnJ5aTBoeHA0bWFnMDJNNzMzYlhMUkVOU09GT05Sa3NscE84ekhVTjVwWWRuaFRTd1lUTGFwMSsxYmdjRlN1VVhMV2llcVpCNnFjN2tpdjNiajNTUGFmNDIrczQ4VjQ5dC9PcFh4THRnaVdMOVhrdURUWmN0cEpKQTR2SEhrNk91MGJjZzdpR20rTDF4d0lmYjhNbDRvV3ZUMFNGMzVmZ1cwOGJiTFhaMnYxWENMUnNyV1VnYnE0VStLeHRFcEczWElZY1loS3gxcklyVWhmRUprdUh6Z1BnbE0xMWdHNVcrQ3lmZyt3Zk9KaWc1cTZheElLV3pJZjZDOG04bG15NmJNK041RXNEOVN2QWdNQkFBR2pVekJSTUIwR0ExVWREZ1FXQkJUZjFoTTYvaWJHRit1L1NWQUs4OEZVTWp6Um9UQWZCZ05WSFNNRUdEQVdnQlRmMWhNNi9pYkdGK3UvU1ZBSzg4RlVNanpSb1RBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCZ3ZWYXU1KzJ3QXVDc21PeXlHMjhoMXp5QzRJUG1NbXBSWlRET3AvcExkd1hlSGpKcjhrRUMzbDkycUpFdmMrV0Fib0oxUm91Y0h5Y1VlN1JXaDJDNlpGL1dQQ0JMeVdHd25seXFHeVJNOS9qODZVSjFPZ2l1Wmw3a2w5enh3V29heFBCQ21IYTBSSG93ZFFCN0FWbHBxZzFjN0ZoS2poVUNCbUdUNFZlOHRWMGhkWnRyWm9RVis2eEhQYlVkMzdLVjFCMUJtZm8zbzRla29KS2hVdTk5RW8wM09wRTNKTHRNMTNBMUh4QUJFdVFHSFRJMHR5Y0RCQmRSbjNiMDNIb0loVTBWbnFqdnBWMUtQdnNyZ1lpLzBWU3RMTmV6WlBnR2UwZkczWGd5OHlla2RCOU5NVW4relpMQVRJNCt6OGo0UUg1V2o1WlBhVWt5b0FEMm9VSk8iXX0
```

The parsed and formatted header would be:

```json
{
    "typ": "x509",
    "alg": "RS256",
    "x5c": [
        "MIIDszCCApugAwIBAgIUL1anEU/yJy67VJTbHkNX0bBNAnEwDQYJKoZIhvcNAQELBQAwaTEdMBsGA1UEAwwUcmVnaXN0cnkuZXhhbXBsZS5jb20xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTAeFw0yMDA3MjcxNDQzNDZaFw0yMTA3MjcxNDQzNDZaMGkxHTAbBgNVBAMMFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDkKwAcV44psjN8nno1eZ3zv1ZKUhJAoxwBOIGfIxIe+iHtpXLvFFVwk5Jbxu+Pkig2N4B3Ilrj/Vryi0hxp4mag02M733bXLRENSOFONRkslpO8zHUN5pYdnhTSwYTLap1+1bgcFSuUXLWieqZB6qc7kiv3bj3SPaf42+s48V49t/OpXxLtgiWL9XkuDTZctpJJA4vHHk6Ou0bcg7iGm+L1xwIfb8Ml4oWvT0SF35fgW08bbLXZ2v1XCLRsrWUgbq4U+KxtEpG3XIYcYhKx1rIrUhfEJkuHzgPglM11gG5W+Cyfg+wfOJig5q6axIKWzIf6C8m8lmy6bM+N5EsD9SvAgMBAAGjUzBRMB0GA1UdDgQWBBTf1hM6/ibGF+u/SVAK88FUMjzRoTAfBgNVHSMEGDAWgBTf1hM6/ibGF+u/SVAK88FUMjzRoTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBgvVau5+2wAuCsmOyyG28h1zyC4IPmMmpRZTDOp/pLdwXeHjJr8kEC3l92qJEvc+WAboJ1RoucHycUe7RWh2C6ZF/WPCBLyWGwnlyqGyRM9/j86UJ1OgiuZl7kl9zxwWoaxPBCmHa0RHowdQB7AVlpqg1c7FhKjhUCBmGT4Ve8tV0hdZtrZoQV+6xHPbUd37KV1B1Bmfo3o4ekoJKhUu99Eo03OpE3JLtM13A1HxABEuQGHTI0tycDBBdRn3b03HoIhU0VnqjvpV1KPvsrgYi/0VStLNezZPgGe0fG3Xgy8yekdB9NMUn+zZLATI4+z8j4QH5Wj5ZPaUkyoAD2oUJO"
    ]
}
```

The claims would be a base64 URL encoded string without paddings:

```
eyJtZWRpYVR5cGUiOiJhcHBsaWNhdGlvbi92bmQuZG9ja2VyLmRpc3RyaWJ1dGlvbi5tYW5pZmVzdC52Mitqc29uIiwiZGlnZXN0Ijoic2hhMjU2OmM0NTE2YjhhMzExZTg1ZjFmMmE2MDU3M2FiZjRjNmI3NDBjYTNhZGU0MTI3ZTI5YjA1NjE2ODQ4ZGU0ODdkMzQiLCJzaXplIjo1MjgsInJlZmVyZW5jZXMiOlsicmVnaXN0cnkuZXhhbXBsZS5jb20vZXhhbXBsZTpsYXRlc3QiLCJyZWdpc3RyeS5leGFtcGxlLmNvbS9leGFtcGxlOnYxLjAiXSwiZXhwIjoxNjI4NTg3MTE5LCJpYXQiOjE1OTcwNTExMTksIm5iZiI6MTU5NzA1MTExOX0
```

The parsed and formatted claims would be:

``` JSON
{
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
    "size": 528,
    "references": [
        "registry.example.com/example:latest",
        "registry.example.com/example:v1.0"
    ],
    "exp": 1628587119,
    "iat": 1597051119,
    "nbf": 1597051119
}
```

The signature of the above would be represented as a base64 URL encoded string without paddings:

``` 
MtQBOL2FERM2fMSikruHOMQdHuEXAE1wf6J6TfDY2W_7PfQQllBKbJJE0HqJ5ENAbuqNYHNZeIeKUCYFrNx2XgtrKuTe7WCa1ZZKDtp5bmANp484ekdl6lW23YB8r_SRtseJuibqjI3HuiMyELj9uYV1CdRYaD2BIZ_qxraYH1fMpjDWjehU4RYLI37hsSuDQ90o09BwaNfzbQXYPsGmkSUSmej7rOFPDnuwhNy4WcUed3kRKYEW8eIjO9OUBGQq3PWvhDjxZi3QF4QFDoiKBOXL70AjaiVIveQRkJI9-xHZSYwje9OFEMioeNWB5ceZR-r4L7VzDcU-Fxqjxn79Fw
```

Putting everything together:

```
eyJ0eXAiOiJ4NTA5IiwiYWxnIjoiUlMyNTYiLCJ4NWMiOlsiTUlJRHN6Q0NBcHVnQXdJQkFnSVVMMWFuRVUveUp5NjdWSlRiSGtOWDBiQk5BbkV3RFFZSktvWklodmNOQVFFTEJRQXdhVEVkTUJzR0ExVUVBd3dVY21WbmFYTjBjbmt1WlhoaGJYQnNaUzVqYjIweEZEQVNCZ05WQkFvTUMyVjRZVzF3YkdVZ2FXNWpNUXN3Q1FZRFZRUUdFd0pWVXpFVE1CRUdBMVVFQ0F3S1YyRnphR2x1WjNSdmJqRVFNQTRHQTFVRUJ3d0hVMlZoZEhSc1pUQWVGdzB5TURBM01qY3hORFF6TkRaYUZ3MHlNVEEzTWpjeE5EUXpORFphTUdreEhUQWJCZ05WQkFNTUZISmxaMmx6ZEhKNUxtVjRZVzF3YkdVdVkyOXRNUlF3RWdZRFZRUUtEQXRsZUdGdGNHeGxJR2x1WXpFTE1Ba0dBMVVFQmhNQ1ZWTXhFekFSQmdOVkJBZ01DbGRoYzJocGJtZDBiMjR4RURBT0JnTlZCQWNNQjFObFlYUjBiR1V3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRRGtLd0FjVjQ0cHNqTjhubm8xZVozenYxWktVaEpBb3h3Qk9JR2ZJeEllK2lIdHBYTHZGRlZ3azVKYnh1K1BraWcyTjRCM0lscmovVnJ5aTBoeHA0bWFnMDJNNzMzYlhMUkVOU09GT05Sa3NscE84ekhVTjVwWWRuaFRTd1lUTGFwMSsxYmdjRlN1VVhMV2llcVpCNnFjN2tpdjNiajNTUGFmNDIrczQ4VjQ5dC9PcFh4THRnaVdMOVhrdURUWmN0cEpKQTR2SEhrNk91MGJjZzdpR20rTDF4d0lmYjhNbDRvV3ZUMFNGMzVmZ1cwOGJiTFhaMnYxWENMUnNyV1VnYnE0VStLeHRFcEczWElZY1loS3gxcklyVWhmRUprdUh6Z1BnbE0xMWdHNVcrQ3lmZyt3Zk9KaWc1cTZheElLV3pJZjZDOG04bG15NmJNK041RXNEOVN2QWdNQkFBR2pVekJSTUIwR0ExVWREZ1FXQkJUZjFoTTYvaWJHRit1L1NWQUs4OEZVTWp6Um9UQWZCZ05WSFNNRUdEQVdnQlRmMWhNNi9pYkdGK3UvU1ZBSzg4RlVNanpSb1RBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCZ3ZWYXU1KzJ3QXVDc21PeXlHMjhoMXp5QzRJUG1NbXBSWlRET3AvcExkd1hlSGpKcjhrRUMzbDkycUpFdmMrV0Fib0oxUm91Y0h5Y1VlN1JXaDJDNlpGL1dQQ0JMeVdHd25seXFHeVJNOS9qODZVSjFPZ2l1Wmw3a2w5enh3V29heFBCQ21IYTBSSG93ZFFCN0FWbHBxZzFjN0ZoS2poVUNCbUdUNFZlOHRWMGhkWnRyWm9RVis2eEhQYlVkMzdLVjFCMUJtZm8zbzRla29KS2hVdTk5RW8wM09wRTNKTHRNMTNBMUh4QUJFdVFHSFRJMHR5Y0RCQmRSbjNiMDNIb0loVTBWbnFqdnBWMUtQdnNyZ1lpLzBWU3RMTmV6WlBnR2UwZkczWGd5OHlla2RCOU5NVW4relpMQVRJNCt6OGo0UUg1V2o1WlBhVWt5b0FEMm9VSk8iXX0.eyJtZWRpYVR5cGUiOiJhcHBsaWNhdGlvbi92bmQuZG9ja2VyLmRpc3RyaWJ1dGlvbi5tYW5pZmVzdC52Mitqc29uIiwiZGlnZXN0Ijoic2hhMjU2OmM0NTE2YjhhMzExZTg1ZjFmMmE2MDU3M2FiZjRjNmI3NDBjYTNhZGU0MTI3ZTI5YjA1NjE2ODQ4ZGU0ODdkMzQiLCJzaXplIjo1MjgsInJlZmVyZW5jZXMiOlsicmVnaXN0cnkuZXhhbXBsZS5jb20vZXhhbXBsZTpsYXRlc3QiLCJyZWdpc3RyeS5leGFtcGxlLmNvbS9leGFtcGxlOnYxLjAiXSwiZXhwIjoxNjI4NTg3MTE5LCJpYXQiOjE1OTcwNTExMTksIm5iZiI6MTU5NzA1MTExOX0.MtQBOL2FERM2fMSikruHOMQdHuEXAE1wf6J6TfDY2W_7PfQQllBKbJJE0HqJ5ENAbuqNYHNZeIeKUCYFrNx2XgtrKuTe7WCa1ZZKDtp5bmANp484ekdl6lW23YB8r_SRtseJuibqjI3HuiMyELj9uYV1CdRYaD2BIZ_qxraYH1fMpjDWjehU4RYLI37hsSuDQ90o09BwaNfzbQXYPsGmkSUSmej7rOFPDnuwhNy4WcUed3kRKYEW8eIjO9OUBGQq3PWvhDjxZi3QF4QFDoiKBOXL70AjaiVIveQRkJI9-xHZSYwje9OFEMioeNWB5ceZR-r4L7VzDcU-Fxqjxn79Fw
```

### Signature Persisted within an OCI Artifact Enabled Registry

All values are persisted in a `signature.jwt` file. The file would be submitted to a registry as an Artifact with null layers.
The `signature.jwt` would be persisted within the `manifest.config` object

``` SHELL
oras push \
  registry.example.com/hello-world:v1.0 \
  --manifest-config signature.json:application/vnd.cncf.notary.config.v2+jwt
```

Would push the following manifest:

``` JSON
{
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.cncf.notary.config.v2+jwt",
    "size": 1906,
    "digest": "sha256:c7848182f2c817415f0de63206f9e4220012cbb0bdb750c2ecf8020350239814"
  },
  "layers": []
}
```

## *Signature* Property Descriptions

### Header

- **`typ`** *string*

  This REQUIRED property identifies the signature type. Implementations MUST support at least the following types

  - `x509`: X.509 public key certificates. Implementations MUST verify that the certificate of the signing key has the `digitalSignature` `Key Usage` extension ([RFC 5280 Section 4.2.1.3](https://tools.ietf.org/html/rfc5280#section-4.2.1.3)).

  Implementations MAY support the following types

  - `tuf`: [The update framework](https://theupdateframework.io/).

  Although the signature file is a JWT, type `JWT` is not used as it is not an authentication or authorization token.

- **`alg`** *string*

  This REQUIRED property for the `x509` type identifies the cryptographic algorithm used to sign the content. This field is based on [RFC 7515 Section 4.1.1](https://tools.ietf.org/html/rfc7515#section-4.1.1).

- **`x5c`** *array of strings*

  This OPTIONAL property for the `x509` type contains the X.509 public key certificate or certificate chain corresponding to the key used to digitally sign the content. The certificates are encoded in base64. This field is based on [RFC 7515 Section 4.1.6](https://tools.ietf.org/html/rfc7515#section-4.1.6).

- **`kid`** *string*

  This OPTIONAL property for the `x509` type is a hint (key ID) indicating which key was used to sign the content. This field is based on [RFC 7515 Section 4.1.4](https://tools.ietf.org/html/rfc7515#section-4.1.4).

### Claims

- **`iat`** *integer*

  This OPTIONAL property identities the time at which the manifests were presented to the notary. This field is based on [RFC 7519 Section 4.1.6](https://tools.ietf.org/html/rfc7519#section-4.1.6). When used, it does not imply the issue time of any signature in the `signatures` property.

- **`nbf`** *integer*

  This OPTIONAL property identifies the time before which the  signed content MUST NOT be accepted for processing. This field is based on [RFC 7519 Section 4.1.5](https://tools.ietf.org/html/rfc7519#section-4.1.5).

- **`exp`** *integer*

  This OPTIONAL property identifies the expiration time on or after which the signed content MUST NOT be accepted for processing. This field is based on [RFC 7519 Section 4.1.4](https://tools.ietf.org/html/rfc7519#section-4.1.4).

- **`mediaType`** *string*

  This REQUIRED property contains the media type of the referenced content. Values MUST comply with [RFC 6838][rfc6838], including the [naming requirements in its section 4.2][rfc6838-s4.2].

- **`digest`** *string*

  This REQUIRED property is the *digest* of the target manifest, conforming to the requirements outlined in [Digests](https://github.com/opencontainers/image-spec/blob/master/descriptor.md#digests). If the actual content is fetched according to the *digest*, implementations MUST verify the content against the *digest*.

- **`size`** *integer*

  This REQUIRED property is the *size* of the target manifest. If the actual content is fetched according the *digest*, implementations MUST verify the content against the *size*.

- **`references`** *array of strings*

  This OPTIONAL property claims the manifest references of its origin. The format of the value MUST matches the [*reference* grammar](https://github.com/docker/distribution/blob/master/reference/reference.go). With used, the `x509` signatures are valid only if the domain names of all references match the Common Name (`CN`) in the `Subject` field of the certificate.

## Example Signatures

### x509 Signature

Example showing a formatted `x509` signature file [examples/x509_x5c.nv2.jwt](examples/x509_x5c.nv2.jwt) with certificates provided by `x5c`:

```json
{
    "typ": "x509",
    "alg": "RS256",
    "x5c": [
        "MIIDszCCApugAwIBAgIUL1anEU/yJy67VJTbHkNX0bBNAnEwDQYJKoZIhvcNAQELBQAwaTEdMBsGA1UEAwwUcmVnaXN0cnkuZXhhbXBsZS5jb20xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTAeFw0yMDA3MjcxNDQzNDZaFw0yMTA3MjcxNDQzNDZaMGkxHTAbBgNVBAMMFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDkKwAcV44psjN8nno1eZ3zv1ZKUhJAoxwBOIGfIxIe+iHtpXLvFFVwk5Jbxu+Pkig2N4B3Ilrj/Vryi0hxp4mag02M733bXLRENSOFONRkslpO8zHUN5pYdnhTSwYTLap1+1bgcFSuUXLWieqZB6qc7kiv3bj3SPaf42+s48V49t/OpXxLtgiWL9XkuDTZctpJJA4vHHk6Ou0bcg7iGm+L1xwIfb8Ml4oWvT0SF35fgW08bbLXZ2v1XCLRsrWUgbq4U+KxtEpG3XIYcYhKx1rIrUhfEJkuHzgPglM11gG5W+Cyfg+wfOJig5q6axIKWzIf6C8m8lmy6bM+N5EsD9SvAgMBAAGjUzBRMB0GA1UdDgQWBBTf1hM6/ibGF+u/SVAK88FUMjzRoTAfBgNVHSMEGDAWgBTf1hM6/ibGF+u/SVAK88FUMjzRoTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBgvVau5+2wAuCsmOyyG28h1zyC4IPmMmpRZTDOp/pLdwXeHjJr8kEC3l92qJEvc+WAboJ1RoucHycUe7RWh2C6ZF/WPCBLyWGwnlyqGyRM9/j86UJ1OgiuZl7kl9zxwWoaxPBCmHa0RHowdQB7AVlpqg1c7FhKjhUCBmGT4Ve8tV0hdZtrZoQV+6xHPbUd37KV1B1Bmfo3o4ekoJKhUu99Eo03OpE3JLtM13A1HxABEuQGHTI0tycDBBdRn3b03HoIhU0VnqjvpV1KPvsrgYi/0VStLNezZPgGe0fG3Xgy8yekdB9NMUn+zZLATI4+z8j4QH5Wj5ZPaUkyoAD2oUJO"
    ]
}.{
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
    "size": 528,
    "references": [
        "registry.example.com/example:latest",
        "registry.example.com/example:v1.0"
    ],
    "exp": 1628587119,
    "iat": 1597051119,
    "nbf": 1597051119
}.[Signature]
```

Example showing a formatted `x509` signature file [examples/x509_kid.nv2.jwt](examples/x509_kid.nv2.jwt) with certificates referenced by `kid`:

```json
{
    "typ": "x509",
    "alg": "RS256",
    "kid": "XP5O:Y7W2:PRB6:O355:56CC:P3A6:CBDV:EDMN:QZCK:W5PO:QMV3:T2LX"
}.{
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
    "size": 528,
    "references": [
        "registry.example.com/example:latest",
        "registry.example.com/example:v1.0"
    ],
    "exp": 1628587341,
    "iat": 1597051341,
    "nbf": 1597051341
}.[Signature]
```

[distribution-spec]:    https://github.com/opencontainers/distribution-spec
[oci-artifacts]:        https://github.com/opencontainers/artifacts
[oci-manifest]:         https://github.com/opencontainers/image-spec/blob/master/manifest.md
[oci-manifest-list]:    https://github.com/opencontainers/image-spec/blob/master/image-index.md

