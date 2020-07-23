# Notary V2 Artifact

[Notary v2 signatures](../signature/README.md) can be stored as [OCI artifacts](https://github.com/opencontainers/artifacts). Precisely, it is a [OCI manifest](https://github.com/opencontainers/image-spec/blob/master/manifest.md) with a config of type

- `application/vnd.cncf.notary.config.v2+json`

and no layers.

## Example Artifact

Example showing the manifest ([examples/manifest.json](examples/manifest.json)) of an artifact.

```json
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
