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
docker generate manifest hello-world:dev | nv2 sign -m x509 -k key.pem -o hello-world.nv2
```

This creates a signature `hello-world.nv2`.

```json
{
    "signed": {
        "digest": "sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55",
        "size": 528,
        "iat": 1595562742
    },
    "signatures": [
        {
            "typ": "x509",
            "sig": "UhKwzvYsYN/IBWT7N/19HuG6x/Jo15OrWlwB6X/AySMbnB76ZvzID6zHZxD1l9bDRugAL+HV1KGQV1Vv6P9b7NosT2Z0mbVeTdMltdZndRTZkv5ozZtUXgknuGg9EkvvNLP3THsfK6Tm5dU3uk+rdk//cJ+T9/sYizt0zzAXC0gR/MJ3SxXwaGyQ6TqqQr94QyzPgEpn5ActZwNJ4WRPRpTutic95Na99cxjAYLKyhusPUYbXu1BICUv2EkUviSISrtyM+yHe4tX1m5Q4Qc0+labgsD3K82ezCGhRYQb2jCPSlDw0r2x1s3KbK2dlGXpSgz9DrhM+x4L2UEyp0cnsg==",
            "alg": "RS256",
            "kid": "DD2E:DW7J:OVJK:KCZR:2PGY:SYC5:WFJF:RMMV:FH6W:VGYM:2WW4:7ZGC"
        }
    ]
}
```

### Validate

Validate image `hello-world:dev`.

```
$ docker generate manifest hello-world:dev | nv2 verify -f hello-world.nv2 -c cert.pem
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```

## Scenario #2: Sign, Rename, Push, Validate in Dev

### Sign

Sign image `hello-world:dev`.

```shell
docker generate manifest hello-world:dev | nv2 sign -m x509 -k key.pem -o hello-world.nv2
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
$ nv2 verify -f hello-world.nv2 -c cert.pem --insecure docker://localhost:5000/hello-world:v1.0
sha256:3351c53952446db17d21b86cfe5829ae70f823aff5d410fbf09dff820a39ab55
```
