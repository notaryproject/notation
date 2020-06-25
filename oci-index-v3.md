# OCI Index v3

An updated of the OCI Index to support `index.config`, enabling an Index to be uniquely identified as a `vnd.cncf.notary.v2` mediaType.

## Additions

- `config` - provides a means to uniquely identify the type of content the Index contains. The config value is optional.
  - `config.mediaType` - Used consistently with the [OCI Artifact usage of `manifest.config.mediaType`][oci-artifact-unique-artifact]

- `manifests[].config.mediaType` - provides a means to identify the type of artifact in the collection.

### Standard OCI Index example

Represents a collection of manifests. The collection is just a typical OCI Index of various manifests. The major difference is the collection may include additional artifact types, such as an SBoM or a gpl style source artifact.

Example: [oci-index-v2-verification.json](./oci-index-v3.json)

```json
{
  "schemaVersion": 3,
  "config": {
    "mediaType": "application/vnd.oci.index.config.v1+json",
    "size": 7023,
    "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  },
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "config.mediaType": "application/vnd.oci.image.v1+json",
      "size": 7143,
      "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "config.mediaType": "application/vnd.oci.prototype.sbom.v1",
      "size": 362,
      "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "config.mediaType": "application/vnd.oci.prototype.src.v1",
      "size": 420,
      "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a433aa23a3a"
    }
  ]
}
```

### Notary verification index example

These are new index types that provide signatures to manifests or other indexes. By posting a new signature as an index, the referenced artifact doesn't need to change it's tag or it's digest. Deployment documents that reference an image by `image:tag123` need not change when an additional signature is added. Using this model, any number of signatures may be added.

Note: the `config.mediaType` of `application/vnd.cncf.notary.config.v2+json`

Example: [oci-index-v2-verification.json](./oci-index-v3-verification.json)
```json
{
  "schemaVersion": 3,
  "config": {
    "mediaType": "application/vnd.cncf.notary.config.v2+json",
    "size": 7023,
    "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
  },
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "config.mediaType": "application/vnd.oci.image.v1+json",
      "size": 7143,
      "digest": "sha256:e692418e4cbaf90ca69d05a66403747baa33ee08806650b51fab815ad7fc331f"
    }
  ]
}
```
[oci-artifact-unique-artifact]:     https://github.com/opencontainers/artifacts/blob/master/artifact-authors.md#defining-a-unique-artifact-type