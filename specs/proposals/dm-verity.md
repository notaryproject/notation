# Proposal: Code Integrity for OCI Containers – Per-Layer DM-Verity Signing with Notation

## Overview

This proposal extends the Notation CLI to support signing and verifying OCI container image layers with dm-verity Merkle tree root hashes. While Notation currently supports signing container images at the manifest level, this proposal enables container image per-layer integrity protection that can be continuously enforced at runtime using dm-verity and [Integrity Policy Enforcement](https://docs.kernel.org/next/admin-guide/LSM/ipe.html) (IPE). Container layer DM-verity signing is critical for extending the trust from kernel code-integrity guarantees to workloads within trusted containers, while blocking the execution of untrusted containers and binaries created in mutable container state.

## Problem Statement & Motivation

Modern Linux container hosts may achieve a high level of security by running an immutable host OS, preventing tampering with system binaries. However, OCI container images themselves have traditionally not been held to the same standard – integrity is only verified at image pull time, with no continuous enforcement at runtime. This leaves a gap where if an attacker injects or executes a malicious binary inside a container, the host has no built-in mechanism to prevent it from running.

In developing [Azure Linux with OS Guard](https://techcommunity.microsoft.com/blog/linuxandopensourceblog/azure-linux-with-os-guard-immutable-container-host-with-code-integrity-and-open-/4437473) we aim to to extend code integrity protections into OCI containers using dm-verity and IPE. Each container image layer is backed by a read-only dm-verity block device whose integrity is ensured by a Merkle tree root hash. The root hash is signed by a key that the Linux kernel trusts. At container start, the kernel verifies each root hash signature before allowing the layer to mount. IPE policies then allow execution only from layers with verified hashes.

Signing container images at the manifest level alone is not sufficient to ensure continuous runtime integrity. While manifest signatures protect against image tampering during distribution (image pull), they do not enable enforcement at runtime. Container layers must also have kernel-verifiable signatures to ensure their integrity. With the increasing adoption of immutable infrastructure and zero-trust security models, securing these artifacts with continuous kernel enforcement is critical. By adding per-layer container image signing, Notation can extend its capabilities to enable kernel-enforced container integrity.

The challenge is that OCI registries and container image tools currently support distributing signatures but there is no good tooling for creating per-layer signatures. Existing Notation signatures use JWS or COSE formats, which cannot be verified by the Linux kernel. Kernel-level dm-verity enforcement requires each layer's dm-verity root hash to be accompanied by a PKCS#7 signature that the kernel can check at mount time.

## Scenarios

The following describes how per-layer dm-verity signing can enhance container security across different attack scenarios:

### Scenario 1: Runtime Layer Tampering (Current Implementation)

Sarah, a DevOps engineer, deploys a containerized application to a production Kubernetes cluster. The container images are signed using standard Notation manifest signatures. An attacker gains access to a worker node and modifies one of the container layers on the host filesystem by injecting a malicious binary. When the container restarts, the modified layer is mounted without detection because:
- Notation's manifest signature only verifies the image at pull time
- The digest in the manifest still matches the original layer blob in the registry
- The layer is mounted without verification of root hashes after being tampered with offline (offline attack allowed)
- The malicious binary cannot be detected at runtime since IPE is not present to prevent binaries executing from unsigned dm-verity volumes (runtime attack allowed)


The malicious binary executes successfully, compromising the application. Current Notation signatures cannot prevent this attack because they don't provide continuous runtime enforcement at the image layer level.

### Scenario 2: Kernel-Enforced Layer Integrity (Proposed Solution)

With the proposed per-layer dm-verity signing:

David, a platform engineer, uses Notation with the proposed dm-verity signing changes. The following happens:
1. Each container image layer is processed to generate an EROFS filesystem with dm-verity metadata
2. The dm-verity Merkle tree root hash for each layer is computed deterministically
3. Each root hash is signed using PKCS#7 format with the company's signing key
4. Image layer signatures and metadata are injected into a referrer artifact attached to the image manifest in the registry

When containers are deployed:
1. The EROFS containerd snapshotter fetches the OCI image and its attached referrer artifact that contains the  layer signatures from the registry
2. For each layer, the snapshotter creates a dm-verity block device, passing the root hash and PKCS#7 signature to the kernel
3. The Linux kernel verifies the PKCS#7 signed root hash against trusted keys before mounting the layer (offline attack blocked)
4. IPE policies enforce that only correctly signed dm-verity volumes can execute code at runtime (runtime attack blocked)

If an attacker attempts to modify a layer like in the first scenario, the root hash verification fails immediately and the kernel refuses to mount the tampered layer. If an attacker drops an unsigned binary into a running container and tries to execute it, IPE blocks execution because it was not loaded from a signed dm-verity volume.

## Proposal

This section outlines the proposed solution for signing and verifying OCI container image layers with dm-verity root hashes using Notation CLI commands. The following topics are outside the scope of this document:

- Detailed command usage, which will be covered in individual command specifications
- [EROFS](https://erofs.docs.kernel.org/en/latest/) filesystem implementation details
- dm-verity kernel subsystem internals
- IPE policy configuration 

**Requirements:**
1. Support for PKCS#7 signature format (in addition to existing JWS/COSE)
2. Per-layer signing capability
3. Deterministic EROFS image and Merkle tree generation
4. OCI registry distribution for a new artifact containing the signed layer root hashes via ORAS Attached Artifacts

### Extended Notation CLI

Extend `notation sign` with a new `--dm-verity` flag to enable automated per-layer signing. While the command below assumes the container image exists in a remote registry, this argument should also work when signing [OCI image layouts](https://github.com/notaryproject/notation/blob/main/specs/cmd/sign.md#experimental-sign-container-images-stored-in-oci-layout-directory) with argument `--oci-layout` for local signing.

The manifest from the default sign behavior is signed with the expected JWS/COSE formats that can be verified in userspace while the layer hashes are signed with the PKCS#7 format by default until other formats are supported. 

This command will not recursively sign multi-arch container images. In this case, the command should be run for each individual image for the requested architecture.

**Sample command:**
```bash
notation sign --dm-verity \
  --id myKeyId \
  myregistry.azurecr.io/myapp@sha256:def456...
```
**Sample output:**
```
Successfully signed myregistry.azurecr.io/myapp@sha256:def456...
Pushed the dm-verity signatures to myregistry.azurecr.io/myapp@sha256:439dd2...
```

**Steps:**

1. **Pull the image manifest** from the registry.
2. **Iterate through all layers** in the manifest. For each layer:
    - Pull the layer blob
    - Generate an EROFS image by decompressing the tar layer and using `mkfs.erofs` (deterministic, read-only filesystem image)
    - Compute the dm-verity Merkle tree root hash from the EROFS image
    - Sign the root hash using PKCS#7
3. **Create a signature envelope** for each layer containing:
    - The signed root hash of the EROFS image.
    - Signer cert embedded inside the PKCS#7 signature blob
    - Digest info of the original layer digest blob
    - Digest and size of the PKCS#7 signature file
4. **Create a signature manifest** containing:
    - All per-layer signature envelopes
5. **Attach the signature manifest** to the image manifest as an attached artifact (referrer) in the registry.

**Example signature manifest structure:**

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "artifactType": "application/vnd.cncf.notary.signature.dm-verity",
  "config": {
    "mediaType": "application/vnd.oci.empty.v1+json",
    "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
    "size": 2
  },
  "layers": [
    {
      "mediaType": "application/pkcs7-signature",
      "digest": "sha256:abc123...",
      "size": 1479,
      "annotations": {
        "io.cncf.notary.layer.digest": "sha256:layer0digest...",
        "io.cncf.notary.dm-verity.root-hash": "0dcd29977f675344645e8c907b5a86b490335e7a2657a2ba45d00e7944701eed",
        "org.opencontainers.image.title": "layer-0.pkcs7.sig"
      }
    }
  ],
  "subject": {
    "mediaType": "application/vnd.oci.image.manifest.v1+json",
    "digest": "sha256:imagemanifestdigest...",
    "size": 1234
  },
  "annotations": {
    "io.cncf.notary.dm-verity.signature": "true"
  }
}
```

The layers section now contains metadata for each image layer signature. Each entry describes:
- mediaType: The format of the signature (PKCS#7).
- digest: The SHA-256 hash of the signature file.
- size: The signature file size in bytes.
- annotations: Extra metadata, including:
  - The digest of the signed layer.
  - The dm-verity root hash for integrity verification.
  - A human-readable title for the signature file.

The new entries are described below:
- io.cncf.notary.layer.digest: The digest of the original image layer
- io.cncf.notary.dm-verity.root-hash: The root hash value of the dm-verity block device
- io.cncf.notary.dm-verity.signature=true: This is a flag that notifies Notation that dm-verity signatures and root hashes exist in the artifact

**Performance Metrics:**
  - Registry overhead: ~4 KB per layer
    - PKCS#7 signature blob: ~2 KB
    - Manifest entry: ~2 KB
  - Signing time: ~4-5 seconds per layer
  - Timeout: 5 minutes for EROFS conversion with no hardcoded maximum layer size


**Verification command:**

```bash
notation verify myregistry.azurecr.io/myapp@sha256:def456...
```

The command interface will not change. This command will not check layer signature compatibility with any keys. It will only output additional information for dm-verity. If the image has dm-verity signatures attached, Notation should:
1. Detect the `io.cncf.notary.dm-verity.signature=true` annotation
2. Inform the user that kernel-level verification is required

**Sample output:**

```
Successfully verified signature for myregistry.azurecr.io/myapp@sha256:def456...
Note: This image includes dm-verity layer signatures for kernel-enforced integrity.
```

### PKCS#7 Signature Format Support

To enable kernel verification, Notation must support PKCS#7 signature envelopes:

**Requirements:**
- PKCS#7 envelope generation for signing
- X.509 certificate chain embedding
- Compatibility with Linux kernel key rings
- User-mode verification support (for build-time validation)


### Runtime Verification Workflow (implemented in containerd)

This work is ongoing in the containerd project and is described in milestone 1 of the [RFC for Code Integrity](https://github.com/containerd/containerd/issues/12081). The EROFS containerd snapshotter implements the following:

1. **Container Start**: When a container is scheduled, fetch the image manifest
2. **Signature Discovery**: Fetch the signature referrer artifact if one is attached to the image manifest
3. **Layer Processing**: For each layer:
   - Fetch the PKCS#7 signature and metadata from the signature manifest
   - Generate the EROFS image from the decompressed tar file
   - Compute the dm-verity Merkle tree root hash from the layer content
   - Create a dm-verity block device, passing the root hash and PKCS#7 signature to the kernel
4. **Kernel Verification**: The kernel verifies the PKCS#7 signature against the trusted keyring
5. **Mount**: If verification succeeds, the kernel mounts the dm-verity protected layer
6. **IPE Enforcement**: IPE policies allow code execution only from verified dm-verity volumes

This provides continuous integrity protection. Any tampering with layer content causes dm-verity verification to fail, preventing mount and execution.
