# OCI Distribution
We introduce new REST APIs in the registry to support storing signature objects together with the target artifacts, and retrieving them for verification. 

- [Registry OpenAPI Spec](../../specs/distribution/signatures.yml)

Here, we illustrate a few sample requests for the new APIs.

## GET list of signatures
The list signatures for an artifact can be retrieved from the registry.
### Requests
- Get all signatures:
  - `GET http://localhost:5000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/`
- Get signatures with optional parameters: 
  - Paginated request:
    - `GET http://localhost:5000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/?last={last}&max={max}`
  - Query by signer:
    - `GET http://localhost:5000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/?iss={iss}`

### URI Parameters
|Parameter|Description|
|---|---|
|`last`|Query parameter for the last item in previous query. Result set will include values lexically after last.|
|`max`|Query parameter for max number of items.|
|`iss`|Query parameter for issuer, example: `Open Image Scanner`|

### Sample Response
```json
{
    "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
    "signatures": [
        "sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c",
        "sha256:3335d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc88d"
    ]
}
```

## Link signatures to target artifacts.
The signature objects `sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c` and `sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b` can be linked to a target artifact `sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042` as follows:
  - `PUT https://localhost:6000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c`
  - `PUT https://localhost:6000/v2/hello-world/manifests/sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042/signatures/sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b`

## Registry storage layout
Here we illustrate how signature objects are stored in the registry storage backend as different OCI objects are pushed and linked together.

On pushing target manifest `sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042` to repository `hello-world`:

```
<root>
|__ v2
    |__ repositories
        |__ hello-world
            |__ _manifests
                |__ revisions
                    |__ sha256
                        |__ 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                            |__ link
```

On pushing signature manifest `sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c` to repository `hello-world`:

```
<root>
|__ v2
    |__ repositories
        |__ hello-world
            |__ _manifests
                |__ revisions
                    |__ sha256
                        |__ 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                            |__ link
                        |__ 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            |__ link

```

On pushing signature manifest `sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b` to repository `hello-world`:

```
<root>
|__ v2
    |__ repositories
        |__ hello-world
            |__ _manifests
                |__ revisions
                    |__ sha256
                        |__ 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                            |__ link
                        |__ 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            |__ link
                        |__ 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                            |__ link 
```

On linking signature `sha256:2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c` to target manifest `90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`: 

```
<root>
|__ v2
    |__ repositories
        |__ hello-world
            |__ _manifests
                |__ revisions
                    |__ sha256
                        |__ 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                            |__ link
                            |__ signatures
                                |__ sha256
                                    |__ 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                                        |__ link
                        |__ 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            |__ link
                        |__ 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                            |__ link
```

On linking signature `sha256:007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b` to target manifest `90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`: 

```
<root>
|__ v2
    |__ repositories
        |__ hello-world
            |__ _manifests
                |__ revisions
                    |__ sha256
                        |__ 90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042
                            |__ link
                            |__ signatures
                                |__ sha256
                                    |__ 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                                        |__ link
                                    |__ 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                                        |__ link
                        |__ 2235d2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49cc77c
                            |__ link
                        |__ 007170c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75a153b
                            |__ link
```