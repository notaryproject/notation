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
  -addext "subjectAltName=DNS:registry.example.com" \
  -keyout example.key \
  -out example.crt
```

An nv2 client would generate the following header and claims to be signed.

The header would be a base64 URL encoded string without paddings:

```
eyJhbGciOiJSUzI1NiIsInR5cCI6Ing1MDkiLCJ4NWMiOlsiTUlJRDFEQ0NBcnlnQXdJQkFnSVVNRVdDTW1vSHU3VGFZb2pBVXJDdGhIKy9nUFl3RFFZSktvWklodmNOQVFFTEJRQXdhVEVkTUJzR0ExVUVBd3dVY21WbmFYTjBjbmt1WlhoaGJYQnNaUzVqYjIweEZEQVNCZ05WQkFvTUMyVjRZVzF3YkdVZ2FXNWpNUXN3Q1FZRFZRUUdFd0pWVXpFVE1CRUdBMVVFQ0F3S1YyRnphR2x1WjNSdmJqRVFNQTRHQTFVRUJ3d0hVMlZoZEhSc1pUQWVGdzB5TURBNE1qQXdNalF4TUROYUZ3MHlNVEE0TWpBd01qUXhNRE5hTUdreEhUQWJCZ05WQkFNTUZISmxaMmx6ZEhKNUxtVjRZVzF3YkdVdVkyOXRNUlF3RWdZRFZRUUtEQXRsZUdGdGNHeGxJR2x1WXpFTE1Ba0dBMVVFQmhNQ1ZWTXhFekFSQmdOVkJBZ01DbGRoYzJocGJtZDBiMjR4RURBT0JnTlZCQWNNQjFObFlYUjBiR1V3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRQ3pvbHN0YXFMUGt5cmR1dmRCNktoK2o4Vi9XckZRM1dtYnc3VW1pbWlFOVdLRFJHK3d6TndqL3hSN0YyVzZoNUduekdkSSs1eUlzbmx3UzUrZm9NYndOdWJwZlAxK0c0Qk05UjFtTUd0dWxDYUFvbHZRWHBqTkwwbytvbC84RDBLdGFvMlFCbEIyamFZVjlSMlhGWDFvOFo1OWkwbXA1Q1ZnUmdGM2lsQm9ycjFHUXN3aWFXb2RPUEJXZWJXSGhVcGVPWWRRZWY4bXNOTTdKQjJkM3VMU0sxc2FrUGtYVThHZCthQXZXSUtmY3did00yTHZxRmtKY2NQYlNoT3ErdE1PSnJ2MGdhTW4zbVNCSGlXQXRuNGJ0ZEJVZjFUTGtMQmpTTDk4dncwaVVNUUJVbk1aWWtvT2pCSmlpSkNESE1mOW9Ba3hVcVdtRjVqeEdBZlJudk14QWdNQkFBR2pkREJ5TUIwR0ExVWREZ1FXQkJSVG5pb1hJMUpIYkp5blJoRzhyMHF2c2craUJqQWZCZ05WSFNNRUdEQVdnQlJUbmlvWEkxSkhiSnluUmhHOHIwcXZzZytpQmpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUI4R0ExVWRFUVFZTUJhQ0ZISmxaMmx6ZEhKNUxtVjRZVzF3YkdVdVkyOXRNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUNPNy94YmJXTzRXUzMxQ1V0bGx2N2xEN05VbFpHM0t6VGZSWWhIZTFtN0xiakRpY3o0MVgzb2Q3Q0J3REJiYjhFNkxSdGVlZUlyK3k1Y2FRQm9FUHlWamFab2xwd0xpdy9LMytQYm5rSjBVTm5wdnRJMlpaNy9OS1daSTdqK3pTTkh0N24wTWpOcXZENFFlY04wWWQxOFFTVTBUd3FuVnphLzhUK0lRbnRMMDRZNU1CaFA2ZldNR1p5ZURob2lEYlhJRm1HRzd4N3pmN3FYcFphU0FLbHFHTDkvdzJETUY1OTh0V2YxT3NTN1BDdE9xUGZPZ2JnR3FKMSs4S2FoWTlxN0pBQUZ6bDY4Y3Jmdkg2MzNGcXdtNVErT2lyZUtuMXA2d20yZzhBK0ZsOWtHWEdQVVdqd05wdUZCL0JSSXZ0TEhYRjJ6R3Flam9ET1dzekJ1TmhmOCJdfQ
```

The parsed and formatted header would be:

```json
{
    "alg": "RS256",
    "typ": "x509",
    "x5c": [
        "MIID1DCCArygAwIBAgIUMEWCMmoHu7TaYojAUrCthH+/gPYwDQYJKoZIhvcNAQELBQAwaTEdMBsGA1UEAwwUcmVnaXN0cnkuZXhhbXBsZS5jb20xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTAeFw0yMDA4MjAwMjQxMDNaFw0yMTA4MjAwMjQxMDNaMGkxHTAbBgNVBAMMFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCzolstaqLPkyrduvdB6Kh+j8V/WrFQ3Wmbw7UmimiE9WKDRG+wzNwj/xR7F2W6h5GnzGdI+5yIsnlwS5+foMbwNubpfP1+G4BM9R1mMGtulCaAolvQXpjNL0o+ol/8D0Ktao2QBlB2jaYV9R2XFX1o8Z59i0mp5CVgRgF3ilBorr1GQswiaWodOPBWebWHhUpeOYdQef8msNM7JB2d3uLSK1sakPkXU8Gd+aAvWIKfcwbwM2LvqFkJccPbShOq+tMOJrv0gaMn3mSBHiWAtn4btdBUf1TLkLBjSL98vw0iUMQBUnMZYkoOjBJiiJCDHMf9oAkxUqWmF5jxGAfRnvMxAgMBAAGjdDByMB0GA1UdDgQWBBRTnioXI1JHbJynRhG8r0qvsg+iBjAfBgNVHSMEGDAWgBRTnioXI1JHbJynRhG8r0qvsg+iBjAPBgNVHRMBAf8EBTADAQH/MB8GA1UdEQQYMBaCFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQCO7/xbbWO4WS31CUtllv7lD7NUlZG3KzTfRYhHe1m7LbjDicz41X3od7CBwDBbb8E6LRteeeIr+y5caQBoEPyVjaZolpwLiw/K3+PbnkJ0UNnpvtI2ZZ7/NKWZI7j+zSNHt7n0MjNqvD4QecN0Yd18QSU0TwqnVza/8T+IQntL04Y5MBhP6fWMGZyeDhoiDbXIFmGG7x7zf7qXpZaSAKlqGL9/w2DMF598tWf1OsS7PCtOqPfOgbgGqJ1+8KahY9q7JAAFzl68crfvH633Fqwm5Q+OireKn1p6wm2g8A+Fl9kGXGPUWjwNpuFB/BRIvtLHXF2zGqejoDOWszBuNhf8"
    ]
}
```

The claims would be a base64 URL encoded string without paddings:

```
eyJkaWdlc3QiOiJzaGEyNTY6YzQ1MTZiOGEzMTFlODVmMWYyYTYwNTczYWJmNGM2Yjc0MGNhM2FkZTQxMjdlMjliMDU2MTY4NDhkZTQ4N2QzNCIsImV4cCI6MTYyOTEwNTk5NCwiaWF0IjoxNTk3ODkzOTk0LCJtZWRpYVR5cGUiOiJhcHBsaWNhdGlvbi92bmQuZG9ja2VyLmRpc3RyaWJ1dGlvbi5tYW5pZmVzdC52Mitqc29uIiwibmJmIjoxNTk3ODkzOTk0LCJyZWZlcmVuY2VzIjpbInJlZ2lzdHJ5LmV4YW1wbGUuY29tL2V4YW1wbGU6bGF0ZXN0IiwicmVnaXN0cnkuZXhhbXBsZS5jb20vZXhhbXBsZTp2MS4wIl0sInNpemUiOjUyOH0
```

The parsed and formatted claims would be:

``` JSON
{
    "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
    "exp": 1629105994,
    "iat": 1597893994,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "nbf": 1597893994,
    "references": [
        "registry.example.com/example:latest",
        "registry.example.com/example:v1.0"
    ],
    "size": 528
}
```

The signature of the above would be represented as a base64 URL encoded string without paddings:

``` 
Zxlv4mVVFKUFwGmyBC85pEiRD-qzwu4myUT89frvfsZ8PRNDV1_B4eoHQVJFpAn5HsTJ_MtHSP1dHd1OeoLRYlOpeDFy_rQlio--zXsyJDMvxIK4SIwnwhohf41hOoF-N_tAUv93fUunEdW5tt55pWrb3jOewYAWbpTRkBoEVHgHmiJriT4eeJqhi8C_gXn9B6KejH4RGUq3GQtclAKTNpZju0g0l1gp_zfHMhm4kn-5NlXlg96MA3JVkib9TglliCbMvPqaMdAZi74JQwd6KyH-2QMspw3tFZKyy9gQnqC9LFaMN2gZgAX1XX0bMB2XV5hA2-_S4uCgVTG7lpgehA
```

Putting everything together:

```
eyJhbGciOiJSUzI1NiIsInR5cCI6Ing1MDkiLCJ4NWMiOlsiTUlJRDFEQ0NBcnlnQXdJQkFnSVVNRVdDTW1vSHU3VGFZb2pBVXJDdGhIKy9nUFl3RFFZSktvWklodmNOQVFFTEJRQXdhVEVkTUJzR0ExVUVBd3dVY21WbmFYTjBjbmt1WlhoaGJYQnNaUzVqYjIweEZEQVNCZ05WQkFvTUMyVjRZVzF3YkdVZ2FXNWpNUXN3Q1FZRFZRUUdFd0pWVXpFVE1CRUdBMVVFQ0F3S1YyRnphR2x1WjNSdmJqRVFNQTRHQTFVRUJ3d0hVMlZoZEhSc1pUQWVGdzB5TURBNE1qQXdNalF4TUROYUZ3MHlNVEE0TWpBd01qUXhNRE5hTUdreEhUQWJCZ05WQkFNTUZISmxaMmx6ZEhKNUxtVjRZVzF3YkdVdVkyOXRNUlF3RWdZRFZRUUtEQXRsZUdGdGNHeGxJR2x1WXpFTE1Ba0dBMVVFQmhNQ1ZWTXhFekFSQmdOVkJBZ01DbGRoYzJocGJtZDBiMjR4RURBT0JnTlZCQWNNQjFObFlYUjBiR1V3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRQ3pvbHN0YXFMUGt5cmR1dmRCNktoK2o4Vi9XckZRM1dtYnc3VW1pbWlFOVdLRFJHK3d6TndqL3hSN0YyVzZoNUduekdkSSs1eUlzbmx3UzUrZm9NYndOdWJwZlAxK0c0Qk05UjFtTUd0dWxDYUFvbHZRWHBqTkwwbytvbC84RDBLdGFvMlFCbEIyamFZVjlSMlhGWDFvOFo1OWkwbXA1Q1ZnUmdGM2lsQm9ycjFHUXN3aWFXb2RPUEJXZWJXSGhVcGVPWWRRZWY4bXNOTTdKQjJkM3VMU0sxc2FrUGtYVThHZCthQXZXSUtmY3did00yTHZxRmtKY2NQYlNoT3ErdE1PSnJ2MGdhTW4zbVNCSGlXQXRuNGJ0ZEJVZjFUTGtMQmpTTDk4dncwaVVNUUJVbk1aWWtvT2pCSmlpSkNESE1mOW9Ba3hVcVdtRjVqeEdBZlJudk14QWdNQkFBR2pkREJ5TUIwR0ExVWREZ1FXQkJSVG5pb1hJMUpIYkp5blJoRzhyMHF2c2craUJqQWZCZ05WSFNNRUdEQVdnQlJUbmlvWEkxSkhiSnluUmhHOHIwcXZzZytpQmpBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUI4R0ExVWRFUVFZTUJhQ0ZISmxaMmx6ZEhKNUxtVjRZVzF3YkdVdVkyOXRNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUNPNy94YmJXTzRXUzMxQ1V0bGx2N2xEN05VbFpHM0t6VGZSWWhIZTFtN0xiakRpY3o0MVgzb2Q3Q0J3REJiYjhFNkxSdGVlZUlyK3k1Y2FRQm9FUHlWamFab2xwd0xpdy9LMytQYm5rSjBVTm5wdnRJMlpaNy9OS1daSTdqK3pTTkh0N24wTWpOcXZENFFlY04wWWQxOFFTVTBUd3FuVnphLzhUK0lRbnRMMDRZNU1CaFA2ZldNR1p5ZURob2lEYlhJRm1HRzd4N3pmN3FYcFphU0FLbHFHTDkvdzJETUY1OTh0V2YxT3NTN1BDdE9xUGZPZ2JnR3FKMSs4S2FoWTlxN0pBQUZ6bDY4Y3Jmdkg2MzNGcXdtNVErT2lyZUtuMXA2d20yZzhBK0ZsOWtHWEdQVVdqd05wdUZCL0JSSXZ0TEhYRjJ6R3Flam9ET1dzekJ1TmhmOCJdfQ.eyJkaWdlc3QiOiJzaGEyNTY6YzQ1MTZiOGEzMTFlODVmMWYyYTYwNTczYWJmNGM2Yjc0MGNhM2FkZTQxMjdlMjliMDU2MTY4NDhkZTQ4N2QzNCIsImV4cCI6MTYyOTEwNTk5NCwiaWF0IjoxNTk3ODkzOTk0LCJtZWRpYVR5cGUiOiJhcHBsaWNhdGlvbi92bmQuZG9ja2VyLmRpc3RyaWJ1dGlvbi5tYW5pZmVzdC52Mitqc29uIiwibmJmIjoxNTk3ODkzOTk0LCJyZWZlcmVuY2VzIjpbInJlZ2lzdHJ5LmV4YW1wbGUuY29tL2V4YW1wbGU6bGF0ZXN0IiwicmVnaXN0cnkuZXhhbXBsZS5jb20vZXhhbXBsZTp2MS4wIl0sInNpemUiOjUyOH0.Zxlv4mVVFKUFwGmyBC85pEiRD-qzwu4myUT89frvfsZ8PRNDV1_B4eoHQVJFpAn5HsTJ_MtHSP1dHd1OeoLRYlOpeDFy_rQlio--zXsyJDMvxIK4SIwnwhohf41hOoF-N_tAUv93fUunEdW5tt55pWrb3jOewYAWbpTRkBoEVHgHmiJriT4eeJqhi8C_gXn9B6KejH4RGUq3GQtclAKTNpZju0g0l1gp_zfHMhm4kn-5NlXlg96MA3JVkib9TglliCbMvPqaMdAZi74JQwd6KyH-2QMspw3tFZKyy9gQnqC9LFaMN2gZgAX1XX0bMB2XV5hA2-_S4uCgVTG7lpgehA
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
    "alg": "RS256",
    "typ": "x509",
    "x5c": [
        "MIID1DCCArygAwIBAgIUMEWCMmoHu7TaYojAUrCthH+/gPYwDQYJKoZIhvcNAQELBQAwaTEdMBsGA1UEAwwUcmVnaXN0cnkuZXhhbXBsZS5jb20xFDASBgNVBAoMC2V4YW1wbGUgaW5jMQswCQYDVQQGEwJVUzETMBEGA1UECAwKV2FzaGluZ3RvbjEQMA4GA1UEBwwHU2VhdHRsZTAeFw0yMDA4MjAwMjQxMDNaFw0yMTA4MjAwMjQxMDNaMGkxHTAbBgNVBAMMFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMRQwEgYDVQQKDAtleGFtcGxlIGluYzELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0b24xEDAOBgNVBAcMB1NlYXR0bGUwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCzolstaqLPkyrduvdB6Kh+j8V/WrFQ3Wmbw7UmimiE9WKDRG+wzNwj/xR7F2W6h5GnzGdI+5yIsnlwS5+foMbwNubpfP1+G4BM9R1mMGtulCaAolvQXpjNL0o+ol/8D0Ktao2QBlB2jaYV9R2XFX1o8Z59i0mp5CVgRgF3ilBorr1GQswiaWodOPBWebWHhUpeOYdQef8msNM7JB2d3uLSK1sakPkXU8Gd+aAvWIKfcwbwM2LvqFkJccPbShOq+tMOJrv0gaMn3mSBHiWAtn4btdBUf1TLkLBjSL98vw0iUMQBUnMZYkoOjBJiiJCDHMf9oAkxUqWmF5jxGAfRnvMxAgMBAAGjdDByMB0GA1UdDgQWBBRTnioXI1JHbJynRhG8r0qvsg+iBjAfBgNVHSMEGDAWgBRTnioXI1JHbJynRhG8r0qvsg+iBjAPBgNVHRMBAf8EBTADAQH/MB8GA1UdEQQYMBaCFHJlZ2lzdHJ5LmV4YW1wbGUuY29tMA0GCSqGSIb3DQEBCwUAA4IBAQCO7/xbbWO4WS31CUtllv7lD7NUlZG3KzTfRYhHe1m7LbjDicz41X3od7CBwDBbb8E6LRteeeIr+y5caQBoEPyVjaZolpwLiw/K3+PbnkJ0UNnpvtI2ZZ7/NKWZI7j+zSNHt7n0MjNqvD4QecN0Yd18QSU0TwqnVza/8T+IQntL04Y5MBhP6fWMGZyeDhoiDbXIFmGG7x7zf7qXpZaSAKlqGL9/w2DMF598tWf1OsS7PCtOqPfOgbgGqJ1+8KahY9q7JAAFzl68crfvH633Fqwm5Q+OireKn1p6wm2g8A+Fl9kGXGPUWjwNpuFB/BRIvtLHXF2zGqejoDOWszBuNhf8"
    ]
}.{
    "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
    "exp": 1629105994,
    "iat": 1597893994,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "nbf": 1597893994,
    "references": [
        "registry.example.com/example:latest",
        "registry.example.com/example:v1.0"
    ],
    "size": 528
}.[Signature]
```

Example showing a formatted `x509` signature file [examples/x509_kid.nv2.jwt](examples/x509_kid.nv2.jwt) with certificates referenced by `kid`:

```json
{
    "alg": "RS256",
    "kid": "GLCY:N6YH:YD7T:7TKW:B3L3:MXER:AS63:EAYF:PJL7:DS4R:ESJN:4MZQ",
    "typ": "x509"
}.{
    "digest": "sha256:c4516b8a311e85f1f2a60573abf4c6b740ca3ade4127e29b05616848de487d34",
    "exp": 1629106005,
    "iat": 1597894005,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "nbf": 1597894005,
    "references": [
        "registry.example.com/example:latest",
        "registry.example.com/example:v1.0"
    ],
    "size": 528
}.[Signature]
```

[distribution-spec]:    https://github.com/opencontainers/distribution-spec
[oci-artifacts]:        https://github.com/opencontainers/artifacts
[oci-manifest]:         https://github.com/opencontainers/image-spec/blob/master/manifest.md
[oci-manifest-list]:    https://github.com/opencontainers/image-spec/blob/master/image-index.md
