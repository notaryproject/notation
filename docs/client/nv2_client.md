# Notary V2 Client
We illustrate a sample client flow, exploring scenarios described at [github.com/notaryproject/requirements](https://github.com/notaryproject/requirements/blob/master/scenarios.md).

## Sample Artifact
- Tag Reference:
  - `hello-world:v1.0`
- Digest:
  - `sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`

## Sample Client Configuration
For purposes of signing and verification, here's an example client configuration file:
```json
{
  TODO
}
```
------

## Scenario #1: Local Build, Sign, Validate

### Build 
Build image `hello-world:dev`. This can be done using a tool of your choice. For example, using `docker`:
  - `docker build -t hello-world:dev .`

### Sign
Sign image `hello-world:dev`
  - `nv2 sign docker://hello-world:dev`
  - This creates a local verification object `hello-world_dev.nv2`
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
                    "hello-world:dev"
                ]
            }
        ]
    },
    "signatures": [
        {
            "typ": "gpg",
            "iss": "Image Developer",
            "sig": "iQEzBAABCgAdFiEEwJjX8wKzoB/U5VaL5TFYlOcxMTUFAl79dnkACgkQ5TFYlOcxMTVtPAf9HwVwBDnDal6JA+jqUsy1MqLB00grOAyclSfejUcXsdI5on6BGkPgksiTRexCZhPNKumcYw32uhR/+2V5rkBelP55/ER9xGtV4u00QKBBAwlUWkUe8exO6R4VDiWAYl2bCzDMdaATiiYiOXaM5MujK438qL9P0/QlTUUv51ErvRSE6ofoLmaEB+I0vG7DpmYVVq4iVTpWtK08i9CHlwWttlIBz/+72akxUJ/TjX/WgasgpQM89viBSsxwhftfUyQKexRscL7RruAg4IgLvDwH1CXVqO69oT0UoEFtZxa2CYUcZJscf2zsiWl4wn2aUEa7e4EgDFwpGq8F8C9DfDq5BA=="
        }
      ] 
  }
  ```

### Validate
Validate image `hello-world:v1.0`.
  - `nv2 verify docker://hello-world:v1.0`
    - `sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`

Run image `hello-world:v1.0` locally.
  - `docker run hello-world@sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042`

## Scenario #2: Sign, Rename, Push, Validate in Dev

### Sign
Sign image `hello-world:v1.0`
  - `nv2 sign docker://hello-world:v1.0`
  - This creates a local verification object `hello-world_v1-0.nv2`

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
                    "hello-world:v1.0"
                ]
            }
        ]
      },
      "signatures": [
          {
              "typ": "gpg",
              "iss": "Image Developer",
              "sig": "rMEzBAABCgAdFiEEwJjX8wKzoB/U5VaL5TFYlOcxMTUFAl79dnkACgkQ5TFYlOcxMTVtPAf9HwVwBDnDal6JA+jqUsy1MqLB00grOAyclSfejUcXsdI5on6BGkPgksiTRexCZhPNKumcYw32uhR/+2V5rkBelP55/ER9xGtV4u00QKBBAwlUWkUe8exO6R4VDiWAYl2bCzDMdaATiiYiOXaM5MujK438qL9P0/QlTUUv51ErvRSE6ofoLmaEB+I0vG7DpmYVVq4iVTpWtK08i9CHlwWttlIBz/+72akxUJ/TjX/WgasgpQM89viBSsxwhftfUyQKexRscL7RruAg4IgLvDwH1CXVqO69oT0UoEFtZxa2CYUcZJscf2zsiWl4wn2aUEa7e4EgDFwpGq8F8C9DfDq5CC=="
          }
      ]
  }
  ```

### Rename
Rename local artifact to include registry FQDN:
  - `docker tag hello-world:v1.0 localhost:5000/hello-world:v1.0`

### Push
Push target artifact, together with its signature. 
- `nv2 push --signature hello-world_v1-0.nv2 docker://localhost:5000/hello-world:v1.0`

This single command does three operations:
1. Push docker image `localhost:5000/hello-world:v1.0`
2. Push signature artifact `hello-world_v1-0.nv2`
3. Link signature `hello-world_v1-0.nv2` to target artifact `localhost:5000/hello-world:v1.0`.

Pushing the signature and linking it to its target artifact are separate operations for the following reasons:
- A signature artifact is like any other OCI artifact and has no special handling in the registry.
- Separate `push` and `link` operations allow for fine-grained RBAC.

### Validate
A consumer of the target artifact, such as an orchestrator deploying an image, can verify the signatures on it.
- `nv2 verify docker://localhost:5000/hello-world:v1.0`

This command would fetch the signatures on this artifact form the registry and verify each one of them (or a configured few.) As in our example, if gpg signatures are used, the consumer needs to have the verification keys configured in their local keyring.

## Scenario #6: Multiple Signatures
A client may download an artifact from an origin registry, verify its signatures, add its own signature, and then push everything to a private destination registry.

### Pull artifact from origin
As an example, we use `docker`:
- `docker pull localhost:5000/hello-world:v1.0`

### Verify signatures
- `nv2 verify docker://localhost:5000/hello-world:v1.0`

This command will pull the signature artifacts from the registry and verify each one of them. Verification would ensure that the signature includes tag reference metadata the target artifact is referenced by tag. As a part of this verification, the signature `localhost-5000_hello-world_v1-0.nv2` is downloaded.

### Rename
Rename local artifact to include private registry FQDN:
  - `docker tag localhost:5000/hello-world:v1.0 localhost:6000/hello-world:v1.0-test`

### Add signature
Sign image `localhost:6000/hello-world:v1.0-test`. 
  - `nv2 sign docker://localhost:6000/hello-world:v1.0-test`
  - This creates a local verification object `localhost-6000_hello-world_v1-0-test.nv2`

  ```json
  {
    "signed": {
        "exp": 1693660592,
        "nbf": 1693659992,
        "iat": 1693659992,
        "manifests": [
            {
                "digest": "sha256:90659bf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cbc042",
                "size": 525,
                "references": [
                    "localhost:6000/hello-world:v1.0-test"
                ]
            }
        ]
      },
      "signatures": [
          {
              "typ": "gpg",
              "iss": "Image Tester",
              "sig": "tYEzBAABCgAdFiEEwJjX8wKzoB/U5VaL5TFYlOcxMTUFAl79dnkACgkQ5TFYlOcxMTVtPAf9HwVwBDnDal6JA+jqUsy1MqLB00grOAyclSfejUcXsdI5on6BGkPgksiTRexCZhPNKumcYw32uhR/+2V5rkBelP55/ER9xGtV4u00QKBBAwlUWkUe8exO6R4VDiWAYl2bCzDMdaATiiYiOXaM5MujK438qL9P0/QlTUUv51ErvRSE6ofoLmaEB+I0vG7DpmYVVq4iVTpWtK08i9CHlwWttlIBz/+72akxUJ/TjX/WgasgpQM89viBSsxwhftfUyQKexRscL7RruAg4IgLvDwH1CXVqO69oT0UoEFtZxa2CYUcZJscf2zsiWl4wn2aUEa7e4EgDFwpGq8F8C9DfDq5ER=="
          }
      ]
  }
  ```

### Push
- `nv2 push --signatures localhost-5000_hello-world_v1-0.nv2,localhost-6000_hello-world_v1-0-test.nv2  docker://localhost:6000/hello-world:v1.0-test`

This single command does five operations:
1. Push docker image `localhost:6000/hello-world:v1.0-test`
2. Push signature artifact `localhost-5000_hello-world_v1-0.nv2`
3. Link signature `localhost-5000_hello-world_v1-0.nv2` to target artifact `localhost:6000/hello-world:v1.0-test`.
4. Push signature artifact `localhost-6000_hello-world_v1-0-test.nv2`
5. Link signature `localhost-6000_hello-world_v1-0-test.nv2` to target artifact `localhost:6000/hello-world:v1.0-test`.