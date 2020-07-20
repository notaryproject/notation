# Notary v2 Artifact
Notary v2 signing / verification objects are stored as [OCI artifacts](https://github.com/opencontainers/artifacts). Precisely, it is a [OCI manifest](https://github.com/opencontainers/image-spec/blob/master/manifest.md) ([example](examples/nv2_manifest.json)) with a config of type

- `application/vnd.cncf.notary.config.v2+json`

and no layers.

```json
{
    "schemaVersion": 2,
    "config": {
        "mediaType": "application/vnd.cncf.notary.config.v2+json",
        "size": 970,
        "digest": "sha256:3050007db1743bfb40df955fa99bfef7ab451a51"
    },
    "layers": []
}
```

Note: All JSON files should be compact with no whitespaces in storage.

## Notary V2 Config
- See [schema](../../schemas/signature-config-schema.json).

The config JSON file consists two parts `signed` and `signatures`:

```json
{
    "signed": {
        "exp": 1593660592,
        "nbf": 1593659992,
        "iat": 1593659992,
        "manifests": [
            {
                "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
                "size": 525,
                "references": [
                    "docker.io/library/hello-world:sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042"
                    "docker.io/library/hello-world:latest",
                    "docker.io/library/hello-world:linux"
                ]
            }
        ]
    },
    "signatures": []
}
```

The `signed` part contains the image metadata for notary.

- The claims `exp`, `nbf`, `iat` follows [RFC7519](https://tools.ietf.org/html/rfc7519#section-10.1.2) and are optional.
- The private claim `manifests` contains the metadata of the manifests / images for notary, and is **required**. A single signing object can notarize multiple manifests for storage and bandwidth efficiency.
  - `digest` is the digest of the manifest
  - `size` is the size of the manifest
  - `references` are the original references of this manifest, and is optional. If `references` is present, signatures are valid only if signatures are signed by a `x509` cert with `Key Usage` extension of `digitalSignature`  and its common name `CN` in the `Subject` field, or with explicit trust configured at the client-side.

The `signatures` part contains signatures for the `signed` part. The signatures have OR logic.