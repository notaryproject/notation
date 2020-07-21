# Notary V2 (nv2) - Prototype

nv2 is an incubation and prototype for designing the [Notary v2](http://github.com/notaryproject/) efforts, securing artifacts stored in [oci-distribution](https://github.com/opencontainers/distribution-spec) based registries.

Public repository is available at GitHub [notaryproject/nv2](https://github.com/notaryproject/nv2).

## Table of Contents
1. [Problem statement](#problem-statement)
2. [Prototype scope](#prototype-scope)
3. [New registry server APIs](docs/distribution/nv2_distribution.md)
4. [New schema for storing signatures](docs/artifacts/nv2_artifact.md)
5. [Client description](docs/client/nv2_client.md)

## Problem Statement

In the world of software, the consumers care about the [authenticity](https://en.wikipedia.org/wiki/Message_authentication) of the software they are using. In terms of OCI registries, the software is referred as images or artifacts where their origin and integrity are the major concerns.

As the current solution, [Notary v1](https://github.com/theupdateframework/notary) implements [the update framework](https://theupdateframework.io/) (NYU papers [available](https://ssl.engineering.nyu.edu/publications)) to ensure the security not only in the aspect of the authenticity but also in the aspect of timeliness. However, Notary v1 is hard to use in the real world as it is 

- Requires maintaining states at the client
- Signatures (trust collections) are stored in an alternative server
- Confusing tagging system
  - The same tag from the notary server may not match the tag in the registry
  - The current implementation does not check if the tag in registry has been updated.
    - This is reflected by CSS cases
    - If an image is pushed to the same tag without signing, the client does not know it. The expected result is that the client should be aware of the new tag and fail the verification.
- Lack of simple signing options
  - Some users do not want the complex features but simple signatures
  - Some users want to do offline signing / verification
- The root trust of the embedded Public Key Infrastructure (PKI) is uncertain
  - Trust on first usage (TOFU)
  - Trust pinning
    - Only available with the notary-cli.
    - Not possible with the docker-cli
- Does not have good support for non-docker products like `ctr`, `oras`, `singularity`, etc.
  - Notary is built-in with docker. However, docker is too heavy for IoT scenarios.
- Signatures (trust collections) are bound with the Globally Unique Name (GUN) so that the signatures cannot be moved / copied to other registries.
- Access control is confusing since registry and notary are two standalone services. 
- TUF targets generic usages not the images in container registries.
  - It does not make sense to make a snapshot of the entire registry, or the entire repository.
  - The users should be able to make a snapshot of a collection of selected images or tags.
- Lack of concurrency
  - The concurrent trust collection publish process is not safe.
  - The notary v1 system is not designed for a distributed system.

More scenarios are presented in [notaryproject/requirements](https://github.com/notaryproject/requirements/blob/master/scenarios.md#scenarios).

## Prototype Scope

The `nv2` prototype covers

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