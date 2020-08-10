# OCI Distribution - Notary v2 Signature Support

To support [Notary v2 goals][notaryv2-goals], upload, persistance and discovery of signatures must be supported.

To minimize the complexity of registry operators and projects to adopt Notary v2, a balance between leveraging what already exists and new patterns to support secure discovery are explored.

## Table of Contents

* [Signature Persistance](#signature-persistance)
* [Signature Push](#signature-push)
* [Signature Discovery](#signature-discovery)
* [Signature Pull](#signature-pull)
* [Example Artifacts](#example-artifacts)

## Signature Persistance

Several options for how to persist a signature were explored. We measure these options against the [goals of Notary v2][notaryv2-goals], specifically:

* Maintain the original artifact digest and collection of associated tags, supporting existing dev through deployment workflows
* Multiple signatures per artifact, enabling the originating vendor signature, public registry certification and user/environment signatures
* Native persistance within an OCI Artifact enabled, distribution*spec based registry
* Artifact and signature copying within and across OCI Artifact enabled, distribution*spec based registries
* Support multi-tenant registries enabling cloud providers and enterprises to support managed services at scale
* Support private registries, where public content may be copied to, and new content originated within
* Air-gapped environments, where the originating registry of content is not accessible

To support the above requirements, signatures are stored as separate [OCI Artifacts][oci-artifacts]. They are maintained as any other artifact in a registry, supporting standard operations such as listing, deleting, garbage collection and any other content addressable operations within a registry.

Following the [OCI Artifacts][oci-artifacts] design, signatures are identified with: `config.mediaType: "application/vnd.cncf.notary.config.v2+json"`.
The config object contains the signature and signed content. See [nv2-signature-spec][nv2-signature-spec] for details.

Storing a signature as a separate artifact enables the above goals, most importantly the ability to maintain the existing tag and and digest for a given artifact.

### Persistance as Manifest or Index

Below will discuss the options and tradeoffs for persisting a signature as an [oci-manifest][oci-manifest] or [oci-index][oci-index].
| [OCI manifest](#signature-persistance---option-1-oci-manifest) | [OCI index](#signature-persistance---option-2-oci-index)  |
| - | - |
|![index](../../media/signature-as-manifest.png)| ![manifest](../../media/signature-as-index.png)

### Signature Persistance - Option 1: oci-manifest

[OCI Artifacts][oci-artifacts] currently supports [OCI manifest][oci-manifest], but doesn't yet support [OCI manifest list/index][oci-index]. To work with what's currently supported, the following design is proposed.

1. An artifact (`net-monitor:v1` container image) is pushed to the registry
1. Signature artifacts are pushed using standard [OCI distribution][oci-distribution] apis. For example, using [ORAS][oras].

The challenge with using oci-manifest is how the registry tracks the linkage between the signature and the original artifact.

### Signature Persistance - Option 2: oci-index

This option is similar to using oci-manifest. However, instead of parsing the signature object to determine the linkage between an artifact and signature, the `index.manifests` collection is utilized.

**Pros with this approach:**

* Utilize the existing `index.manifests` collection for linking artifacts.
* Registries that support oci index already have infrastructure for tracking `index.manifests`, including delete operations and garbage collection.
* Existing distribution-spec upload APIs are utilized.
* Unlike the manifest proposal, no additional artifact handler would be required to parse the config object for linking artifacts.
* Based on the artifact type:  `manifest.config.mediaType: "application/vnd.cncf.notary.config.v2+json"`, role check may be done to confirm the identity has a signer role.

**Cons with this approach:**

* OCI index does not yet support the [OCI config descriptor][oci-descriptor]. This would require a schema change to oci-index, with a version bump.
  * This has been a [desired item for OCI Artifacts][oci-artifacts-index] to support other artifact types which would base on Index.
* An additional role check is performed, based on the artifact type.

## Linking signatures to artifacts

### Parse the signature object for the referenced artifact

* The [nv2 signature specification][nv2-signature-spec] identifies the referenced artifact by its digest and optional tags.
* As the registry receives artifacts, the artifact type is parsed, evaluating the `manifest.config.mediaType` of `"application/vnd.cncf.notary.config.v2+json"`
* A role check is performed, confirming the identity of the PUT has **signer** rights
* The registry uses the config objects reference to link the signature with signed digest. This would enable registry tracking for garbage collection

Partial config object, referring to the digest and tag of the `net-monitor:v1` container image:

```json
{
    "signed": {
        "digest": "sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
        "mediaType": "application/vnd.oci.image.manifest.v2+json",
        "size": 528,
        "references": [
            "registry.acme-rockets.com/net-monitor:v1"
        ],
```

**Pros with this approach:**

* Utilize the existing OCI Artifact manifest design, requiring no additional changes to registry implementations.
* Existing distribution-spec upload APIs are utilized
* An artifact handler would be added to registries. Many already parse the config objects to understand which platform and architectures they support.
* Based on the artifact type:  `manifest.config.mediaType: "application/vnd.cncf.notary.config.v2+json"`, role check may be done to confirm the identity has a signer role.

**Cons with this approach:**

* An artifact handler is added that must parse the signature object, creating a new flow for tracking object linkage
* An additional role check is performed, based on the artifact type.

### Signature Linking - distinct API call

Similar to the manifest or index options, the client pushes the artifact and signatures through standard oci-distribution upload apis.
However, no linkage is made between the signature object and the signed artifact. Rather a signatures api is added.

1. Push all artifacts to the registry:  
   * Push `net-monitor:v1` container image: `sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c`
   * Push **acme-rockets** signature artifact `sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b`
1. Link **acme-rockets** signature to the `net-monitor:v1` container image

``` REST
PUT https://localhost:6000/v2/net-monitor/manifests/sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c/signatures/sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
```

**Pros with this approach:**

* Using a unique API enables RBAC to the API, without having to parse the content to determine the role check.

**Cons with this approach:**

* The client must make two calls to achieve a single operation of uploading a signature object, which by definition has the linking information.

### Unique Signature Upload API

In this option the signature artifact (manifest or index) is uploaded through a new signature API.

**Pros with this approach:**

* The signature artifact upload and role check are coupled to a signature API.

**Cons with this approach:**

* Signature upload, which is just another artifact type, is uploaded differently than other artifacts.

## Signature Discovery

### Signature Discovery - Option 1: distribution-spec consistent paging

The [OCI distribution-spec][distribution-spec-paging] identifies paging with `n` and `last` parameters:

Get a list of paginated signatures from the registry. The response will include an opaque URL that can be followed to obtain the next page of results.

#### Sample Request

``` REST
GET http://localhost:5000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/
```

#### Sample Response

```json
{
    "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
    "@nextLink": "{opaqueUrl}",
    "signatures": [
        "sha256:1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
        "sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b"
    ]
}
```

### Option 2: Client-Side Paging

Get a list of paginated signatures from the registry by specifying the last retrieved item and page size.

**Sample Request**

``` REST
GET http://localhost:5000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/?last=sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b&max=10
```

**URI Parameters**

| Parameter | Description                                                  |
| --------- | ------------------------------------------------------------ |
| `last`    | Query parameter for the last item in previous query. Result set will include values lexically after last. |
| `max`     | Query parameter for max number of items.                     |

**Sample Response**

```json
{
    "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
    "signatures": [
        "sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
        "sha256:1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c"
    ]
}
```

### Signature Discovery - Option 1: REST API standard paging

TBD:

## Signature Pull

For the purpose of discussion, each API has different approaches to achieve the results. Below, we explore the two APIs with the pros & cons of each to facilitate the discussion.
Notary v2 requirements state an artifact can have more than one signature. The signatures are pushed as independent artifacts, allowing workflows to provide additional attestation to the state of an artifact, from the point of the entity that provides the signature. While we don't expect endless signatures for a given artifact, we do not want to limit to an arbitrary number as well.

To facilitate retrieving a list of signatures, we introduce two api patterns:

## Storage Layout

Here we illustrate how signature objects are stored in the registry storage backend as different OCI objects are pushed and linked together.

On pushing target manifest `sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042` to repository `hello-world`:

```
<root>
└── v2
    └── repositories
        └── hello-world
            └── _manifests
                └── revisions
                    └── sha256
                        └── 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                            └── link
```

On pushing signature manifest `sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b` to repository `hello-world`:

```
<root>
└── v2
    └── repositories
        └── hello-world
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                        │   └── link
                        └── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                            └── link
```

On pushing signature manifest `sha256:1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c` to repository `hello-world`:

```
<root>
└── v2
    └── repositories
        └── hello-world
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                        │   └── link
                        │── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │   └── link
                        └── 1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            └── link
```

On pushing signature manifest `sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c` to repository `hello-world`:

```
<root>
└── v2
    └── repositories
        └── hello-world
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                        │   └── link
                        │── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │   └── link
                        │── 1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │   └── link
                        └── 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            └── link
```

On linking signature `sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b` to target manifest `90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`:

```
<root>
└── v2
    └── repositories
        └── hello-world
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                        │   ├── link
                        │   └── signatures
                        │       └── sha256
                        │           └── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │               └── link
                        │── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │   └── link
                        │── 1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │   └── link
                        └── 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            └── link
```

On linking signature `sha256:1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c` to target manifest `90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`:

```
<root>
└── v2
    └── repositories
        └── hello-world
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                        │   ├── link
                        │   └── signatures
                        │       └── sha256
                        |           └── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │               └── link
                        │           └── 1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │               └── links
                        │── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │   └── link
                        │── 1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │   └── link
                        └── 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            └── link
```

On linking signature `sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c` to target manifest `90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`:

```
<root>
└── v2
    └── repositories
        └── net-monitor
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │   ├── link
                        │   └── signatures
                        │       └── sha256
                        |           └── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │               └── link
                        │           └── 1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │               └── link
                        │           └── 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │               └── link
                        │── 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                        │   └── link
                        │── 1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                        │   └── link
                        └── 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            └── link
```

## Example Artifacts

The following are references used in the examples below.

These assume:

* The original net-monitors image was sourced from `registry.wabbit-networks.com/net-monitor:v1`.
* Wabbit Networks signed the original image
* ACME Rockets imported the net-monitor image into `registry.acme-rockets.com/net-monitor:v1`
* The Wabbit Networks signature was copied into `registry.acme-rockets.com/net-monitor:v1`
* ACME Rockets added a verification signature.
* Signature objects do NOT have tags. However, they are placed in the same repo as the artifact they reference.
* Per the design options, a signature object may be persisted as an OCI Manifest or OCI Index.

|Artifact                               |`config.mediaType`                          |Tag                                          | Digest                                                                  |
|---------------------------------------|--------------------------------------------|---------------------------------------------|-------------------------------------------------------------------------|
|net-monitor image                      |`application/vnd.oci.image.config.v1+json`  |`registry.acme-rockets.com/net-monitor:v1`   |`sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c`|
|wabbit-networks signature as a manifest|`application/vnd.cncf.notary.config.v2+json`|`registry.acme-rockets.com/net-monitor@sha:*`|`sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`|
|wabbit-networks signature as an index  |`application/vnd.cncf.notary.config.v2+json`|`registry.acme-rockets.com/net-monitor@sha:*`|`sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`|
|acme-rockets signature as a manifest   |`application/vnd.cncf.notary.config.v2+json`|`registry.acme-rockets.com/net-monitor@sha:*`|`sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b`|
|acme-rockets signature as as an index  |`application/vnd.cncf.notary.config.v2+json`|`registry.acme-rockets.com/net-monitor@sha:*`|`sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b`|

### Example manifest for the **container image**: `registry.acme-rockets.com/net-monitor:v1`:

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "digest": "sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
    "size": 1906
  },
  "layers": [
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0",
      "size": 32654
    },
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736",
      "size": 73109
    }
  ]
}
```

### Example **manifest** for a **Notary v2 signature**

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.cncf.notary.config.v2+json",
    "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
    "size": 1906
  },
  "layers": []
}
```

### Example **index** for a **Notary v2 signature**

``` json
{
  "schemaVersion": 2.1,
  "mediaType": "application/vnd.oci.image.index.v2+json",
  "config": {
    "mediaType": "application/vnd.cncf.notary.config.v2+json",
    "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
    "size": 1906
  },
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
      "size": 7023,
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    }
  ]
}
```

### Example **config object** for the **Notary v2 signature artifact**

See [nv2 signature spec][nv2-signature-spec] for more details.

```json
{
    "signed": {
        "digest": "sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
        "size": 528,
        "references": [
            "registry.acme-rockets.com/net-monitor:v1"
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

[cnab]:                     https://cnab.io
[distribution-spec-paging]: https://github.com/opencontainers/distribution-spec/blob/master/spec.md#listing-image-tags
[notaryv2-goals]:           https://github.com/notaryproject/requirements/blob/52c1ba2f5696a98b317aff84288d3564b4041ad5/README.md#goals
[nv2-signature-spec]:       https://github.com/notaryproject/nv2/blob/efe151ddf6a7fd3848fea340cab7553d0a7d295b/docs/signature/README.md
[oci-artifacts]:            https://github.com/opencontainers/artifacts
[oci-artifacts-index]:      https://github.com/opencontainers/artifacts/issues/25
[oci-index]:                https://github.com/opencontainers/image-spec/blob/master/image-index.md
[oci-descriptor]:           https://github.com/opencontainers/image-spec/blob/master/descriptor.md
[oci-distribution]:         https://github.com/opencontainers/distribution-spec
[oci-manifest]:             https://github.com/opencontainers/image-spec/blob/master/manifest.md
[oras]:                     https://github.com/deislabs/oras
