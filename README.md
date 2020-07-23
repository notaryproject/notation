# Notary V2 (nv2) - Prototype

nv2 is an incubation and prototype for the [Notary v2][notary-v2] efforts, securing artifacts stored in [distribution-spec][distribution-spec] based registries.

![nv2-components](media/nv2-client-components.png)

- The nv2 client (1) will sign any OCI artifact type (2) including a Docker Image, Helm Chart, OPA, SBoM or any OCI Artifact type, generating a Notary v2 signature (3)
- The [ORAS][oras] client (4) can then push the artifact (2) and the Notary v2 signature (3) to an OCI Artifacts supported registry (5)
- In a subsequent prototype, signatures may be retrieved from the OCI Artifacts supported registry (5)

## Table of Contents

1. [nv2 signing and verification docs](docs/nv2/README.md)
2. [Notary v2 signature specification](docs/signature/README.md)
3. [OCI Artifact schema for storing signatures](docs/artifact/README.md)
4. [nv2 prototype scope](#prototype-scope)

## Prototype Scope

The `nv2` prototype covers the scenarios outlined in [notaryproject/requirements](https://github.com/notaryproject/requirements/blob/master/scenarios.md#scenarios).

- Client
  - CLI experience
    - Signing
    - Verification
  - Binaries plug-in
    - Actual pull / push should be done by external binaries
- Server
  - Access control
  - HTTP API changes
  - Registry storage changes

Key management is offloaded to the underlying signing tools.

[distribution-spec]:    https://github.com/opencontainers/distribution-spec
[notary-v2]:            http://github.com/notaryproject/
[oras]:                 https://github.com/deislabs/oras