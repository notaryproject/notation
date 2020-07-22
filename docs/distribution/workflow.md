# Notary V2 Workflow

We illustrate a sample client flow, exploring scenarios described at [github.com/notaryproject/requirements](https://github.com/notaryproject/requirements/blob/master/scenarios.md).

The detailed documentation of the `nv2` command is [available](../nv2/README.md).

## Scenario #1: Local Build, Sign, Validate

### Build 

Build image `hello-world:dev`. This can be done using a tool of your choice. For example, using `docker`:

```shell
docker build -t hello-world:dev .
```

### Sign

Sign image `hello-world:dev`.

```shell
docker generate manifest hello-world:dev | nv2 sign -m gpg -i demo -o hello-world.nv2
```

This creates a signature `hello-world.nv2`.

```json
{
  "signed": {
    "iat": 1595418878,
    "manifests": [
      {
        "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
        "size": 528
      }
    ]
  },
  "signatures": [
    {
      "typ": "gpg",
      "sig": "wsDcBAABCAAQBQJfGCj+CRDvXc1GQtqlQgAA+0QMABcsQ0wU2oY78SgHkm7MsYyHdsAkrWBpLG1hRT02InRj18LUmnwGrvpl6sZm7h5pYbfAg1tST9ta+KQCCXNzP4axGS6cNwilPh7V8kUCgXSPaYyzHIptpBbr5HIaOGBCNPIJTFmnvKGYum1AZng+YudRY2UalS1K4vYWMFEsS5xUJNwoHk06nr+DY68QEBUpBGf689iSH7eIGE9XN4+1mtpnOHhI33FbjCFf3ksh+caE91gch/H4H4CQ5RRfjuvnD0xEBVDCVA/0XygBR1IGT9upoVFUA8XNbuhtATej1MHpOd3mIfeg1rBb2sP0j5tZrbyjBBBB4EbI2GfRYfczlaqfRvmAug4AI9Ya7/RFaZTX15A9X+zTpLH0I34BWwh6BKF9TwoFybFPJODYdZ0+rOmE9Renlc4GwPn0LnXX/PVQ3h6rlWznpdaVUFSPYhPg4bbQnW3XL9nCM8zPu2oVoQGVVNqhIVZpq1es7zc0BkrTT+n3eJyBG/WiLpxwGJneNw==",
      "iss": "Demo User <demo@example.com>"
    }
  ]
}
  ```

### Validate

Validate image `hello-world:dev`.

```
$ docker generate manifest hello-world:dev | nv2 verify -f hello-world.nv2
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

## Scenario #2: Sign, Rename, Push, Validate in Dev

### Sign

Sign image `hello-world:dev`.

```shell
docker generate manifest hello-world:dev | nv2 sign -m gpg -i demo -o hello-world.nv2
```

This creates a signature `hello-world.nv2`.

### Rename

Rename local artifact to include registry FQDN:

```shell
docker tag hello-world:dev localhost:5000/hello-world:v1.0
```

### Push

Push target artifact, together with its signature. 

1. Push docker image `localhost:5000/hello-world:v1.0`
2. Push signature artifact `hello-world.nv2`
3. Link signature `hello-world.nv2` to target artifact `localhost:5000/hello-world:v1.0`.

Pushing the signature and linking it to its target artifact are separate operations for the following reasons:

- A signature artifact is like any other OCI artifact and has no special handling in the registry.
- Separate `push` and `link` operations allow for fine-grained RBAC.

### Validate

A consumer of the target artifact, such as an orchestrator deploying an image, can verify the signatures on it.

Fetch the signatures on this artifact form the registry and verify each one of them (or a configured few). After fetching the signature artifact `hello-world.nv2`, the system will do the following equivalent 

```
$ nv2 verify -f hello-world.nv2 --insecure docker://localhost:5000/hello-world:v1.0
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

 As in our example, if `gpg` signatures are used, the consumer needs to have the verification keys configured in their local keyring.