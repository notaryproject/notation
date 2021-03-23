# How key management is handled in TUF

## Overview

Key management is a vital part of secure distribution because users need to know who is trusted to sign an image and which public keys they control in order to know which signatures to verify. This document describes how TUF addresses key distribution and management to minimize the need for out-of-band communication and make it clear which entities should be trusted to sign each image. This document describes the aspects of key management for which each entity is responsible. In summary:
* The registry or organization that controls the root role is responsible for maintaining root keys, determining how to delegate to developers, and revoking compromised keys.
* Developers manage their own keys and upload them to targets metadata
* Users only need root keys and use delegations to find trusted signatures for any image.


## Registries or organizations

The entity that controls the root metadata, whether this is a registry
or an organization, is responsible for the following aspects of key management:
* *Root key generation*: Root keys should be generated using a secure, offline procedure (for example: [PyPI](https://github.com/psf/psf-tuf-runbook)), used to sign root metadata, then stored in a secure location.
* *Root public key distribution*: Public keys from the root key generation must be securely distributed to users, for example using SPIFFE/SPIRE.
* *Delegate to developers*: Root metadata will delegate to a top-level targets metadata file. This metadata file will then delegate to developers or teams who will sign metadata about the actual images. The registry or organization must:
    * Perform any developer/project vetting. There should be some procedure to determine which developers may be added to the delegation so that the delegation has meaning.
    * Set delegation scope. The delegation can specify an image name or repository that a delegatee is trusted to sign.
    * Set reasonable thresholds for the number of signatures required for each image, and the number of roles that must agree on image contents. These thresholds may provide additional protection, or ensure that different teams (for example dev and security) have vetted the image.
* *Set up automated management of snapshot and timestamp*: These roles are delegated from root, but only need manual interaction from an administrator when compromised.

## Developers

Developers are responsible for their keys, and so must:
* *Generate keys*: This may be done with a hardware token for additional security. Once generated, private keys must be securely stored.
* *Upload keys to registry*: The developer should communicate their public key to the registry so that the registry may delegate to it using targets metadata.
* *Sign images*: The developer should use their private key to sign targets metadata about images, and upload this to the registry.

## Users
A user starts with root key(s) for an organization or registry.

When the user initiates an image download using Notary v2, the following automated
process takes place, described in more detail in the [TUF Specification](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#detailed-client-workflow--detailed-client-workflow):
* The user downloads the root metadata, verifies its signatures, and obtains the latest timestamp, snapshot, and targets keys.
* The user downloads the timestamp metadata, verifies its signatures and timeliness, and establishes the expected hash of snapshot metadata
* The user downloads the snapshot metadata, verifies its signatures and hash, and checks for a rollback attack
* The user downloads the top-level targets metadata, verifies it's signature, and uses any delegations in this file to perform a pre-order depth-first search for the desired image signatures.

The metadata file in the fourth step may be replaced with a client-specified top-level targets file (as explained in the [Design Overview](https://github.com/notaryproject/nv2/blob/prototype-tuf/tuf-design.md#tap-13-client-side-selection-of-the-top-level-target-files-through-mapping-metadata)). In this case, the user will use the delegations and keys from the provided file.
