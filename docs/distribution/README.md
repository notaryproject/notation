# OCI Distribution

We introduce new REST APIs in the registry to support storing signature objects together with the target artifacts, and retrieving them for verification.

Here, we propose a few candidates APIs using example artifacts.

## Table of Contents
1. [Example Artifacts](#example-artifacts)
2. [Retrieving Signatures](#retrieving-signatures)
3. [Storing Signatures](#storing-signatures)
4. [Storage Layout](#storage-layout)

## Example Artifacts

|Artifact|Config Media Type                           |Digest                                                                   |Description     |
|--------|--------------------------------------------|-------------------------------------------------------------------------|----------------|
|Manifest|`application/vnd.cncf.notary.config.v2+json`|`sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c`|Signature object|
|Manifest|`application/vnd.cncf.notary.config.v2+json`|`sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b`|Signature object|
|Manifest|`application/vnd.cncf.notary.config.v2+json`|`sha256:1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c`|Signature object|
|Manifest|`*`                                         |`sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`|Target object   |

## Retrieving Signatures

The list of signatures for an artifact can be retrieved from the registry.

### Option 1: Server-Side Paging

Get a list of paginated signatures from the registry. The response will include an opaque URL that can be followed to obtain the next page of results.

#### Sample Request
```
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

#### Sample Request
```
GET http://localhost:5000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/?last=sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b&max=10
```

#### URI Parameters

| Parameter | Description                                                  |
| --------- | ------------------------------------------------------------ |
| `last`    | Query parameter for the last item in previous query. Result set will include values lexically after last. |
| `max`     | Query parameter for max number of items.                     |

#### Sample Response

```json
{
    "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
    "signatures": [
        "sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
        "sha256:1135d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c"
    ]
}
```

## Storing Signatures

Signatures can be stored in the registry as OCI artifacts and linked to target artifacts that they sign.

### Option 1: Client-Driven Signature Linking

The client pushes all artifacts to the registry independently and then links signatures to their targets.

1. Push all artifacts to the registry (for example, using [oras](https://github.com/deislabs/oras)).
    - Push target artifact `sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`
    - Push signature artifact `sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c`

2. Link signature artifact to the target it signs.

    #### Sample Request
    ```
    PUT https://localhost:6000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
    ```

### Option 2: Server-Side Signature Linking

The server detects a signature push and implicitly links the signature to the target.
#### TODO

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