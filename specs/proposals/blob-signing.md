# Proposal for Supporting Blob Signing and Verification in Notation

## Overview 

A **blob**, short for [Binary Large Object](https://wikipedia.org/wiki/Binary_large_object) is a collection of binary data stored as a single entity. In this proposal, a blob refers to any binary data or file, including SBOM (Software Bill of Materials) files, release assets, AI model files (such as model files on [Hugging Face](https://huggingface.co/)), WebAssembly files or other forms of unstructured data. While Notation currently supports signing and verifying OCI (Open Container Initiative) artifacts such as container images, this document outlines a proposal to extend the Notation tool's capabilities to include blob signing and verification.

## Problem Statement & Motivation 

Signing container images alone is not enough to secure the software supply chain for containers. In addtion to images, SBOMs, configuration files, release binaries or other artifacts must also be signed and verified to ensure their integrity and authenticity. With the increasing use of AI model files and WebAssembly files, securing these artifacts is equally critical. By adding blob signing, Notation can extend its capabilities to handle a broader range of artifacts across various scenarios, enhancing its versatility while leveraging its existing key management and trust policy framework.

## Scenarios 

The following describes how blob signing and verification can be used for various scenarios:

### Scenario 1: Blob signing and verification with file-based distribution

Sarah, a software engineer, creates binaries for applications that are distributed via her company's website. In CI/CD pipelines, she creates tasks that generates an SBOM for each binary and signs the SBOM to ensure the authenticity and integrity of the SBOM. The pipeline automatically generates a digital signature for each SBOM, which is stored alongside the SBOM on the filesystem. With SBOM and signatures created, the pipeline tasks automatically upload the binaries, SBOMs, and SBOM signatures to the company website as separate downloadable artifacts. The compliance team downloads the SBOM and signature to check whether the corresponding binary is compliant with both company and governance rules. Before analyzing the SBOM, they verify its integrity against the corresponding signature, ensuring the SBOM hasn't been tampered with since being signed in the pipeline and that it originates from the Sarah's company.

David, a release manager, has set up signing task in CI/CD pipelines to sign release artifacts, such as binaries, deployment scripts, and configuration files before they are publicly distributed to ensure that they haven't been tampered with during distribution. After the artifacts pass verification, they are signed and published on the release website. Users download both release artifacts and corresponding signatures, and verify the signatures to ensure that the artifacts are untampered and trustworthy before using them.

### Scenario 2: Blob signing and verification with registry-based distribution

The platform team at Sarah’s company leverages their existing OCI-compliant registry to distribute SBOMs while preserving the current generation process. The CI/CD pipeline continues to build binaries, generate SBOMs, and sign them as usual. However, instead of uploading these artifacts to the company website, the pipeline pushes the SBOM and its signature as separate OCI artifacts. To facilitate easy discovery, each signature references its corresponding SBOM in the registry. The compliance team then downloads both SBOMs and their associated signatures to the filesystem, verifies the signatures, and proceeds with SBOM analysis.

## Proposal 

This section outlines the proposed solution for signing and verifying blobs using Notation CLI commands. The following topics are outside the scope of this document:  

- **Detailed command usage**, which is covered in the individual command specifications at [Notation CLI specs](https://github.com/notaryproject/notation/tree/main/specs/commandline).  
- **Distribution of certificates** for verification in different scenarios.  
- **Blob signature definitions**, as outlined in the [Notary Project specification](https://github.com/notaryproject/specifications/blob/main/specs/signature-specification.md#blob-payload).  
- **CI/CD integration**, as this document specifically focuses on using Notation CLI commands.  

In general, all blob-related commands are grouped under the `notation blob` command per the [blob command spec](https://github.com/notaryproject/notation/blob/main/specs/commandline/blob.md).

### Blob Sign and verify with file-based distribution

For file-based distribution, such as SBOMs or release artifacts shared via a website or file transfer system, the process of signing and verifying blobs follows a similar flow. The steps below outline the user experience for signing and verifying a file, using an SBOM file named `sbom.json` as an example.

**Prerequisites:**
- SBOMs, release artifacts or other files are created and ready for publishing

**Steps:**

1. Sign a file on filesystem with a specified signature envelope:

    ```shell
    notation blob sign --id myKeyId --signature-format "cose" --media-type "application/spdx+json" sbom.json
    ```
    This command generates a signature file named `sbom.json.cose.sig` in the current working directory. The file name follows the [Notary Project specification](https://github.com/notaryproject/specifications/blob/main/specs/signature-specification.md#blob-signatures). The signature format is **COSE**, as specified by the `--signature-format` flag. The default format is **JWS**. The `--media-type` flag specifies the **media type** of the blob. In this example, the content of `sbom.json` is in the format `application/spdx+json`. This flag is **optional**. If omitted, the default media type **`application/octet-stream`** is used.  

2. Publish both the file and corresponding signature using any file transfering mechanism. Both the file and its signature can be packaged into one file (e.g., a tarball) for transferring.

3. Download or fetch both the file and signature. If they are packaged into one file, for example, a tarball, unpackage the file to get separate file and signature.

4. Verify the file against the signature:

   - Add the root CA certificate and set up the trust store (The same experience for verifying container images):

        ```shell
        notation cert add --type ca --store myCACerts root.crt
        ```

     Confirm the certificate is added.

        ```shell
        notation cert ls
        ```

        > [!NOTE]
        > Learn more options using `notation cert ls --help`.

    - Set up trust policy for blobs by adding a new command `notation blob policy init`. This command streamlines the process, eliminating the need for users to consult documentation for the correct trust policy format and preventing the accidental use of policies intended for other verification purposes.

        ```shell
        notation blob policy init --name "myBlobPolicy" --trust-store "ca:myCACerts" --trust-identity "x509.subject:C=US,ST=WA,O=wabbit-network.io"
        ```

        Show the policies configured for verifying blobs:

        ```shell
        notation blob policy show
        ```

        > [!NOTE]
        > See the [section](#manage-blob-policies) for more commands for `notation blob policy`.

    - Verify the signature:

        ```shell
        notation blob verify --policy-name myBlobPolicy --signature sbom.json.cose.sig sbom.json
        ```

        If a [global policy](https://github.com/notaryproject/specifications/blob/main/specs/trust-store-trust-policy.md#blob-trust-policy) is set, you can skip the `--policy-name` flag.

### Blob Sign and verify with registry-based distribution

For registry-based distribution, such as using an OCI-compliant container registry, the process is similar but includes additional steps for pushing blobs and sigantures to the registry.

**Prerequisites:**
- SBOMs, release artifacts or other files are created and ready for publishing.

**Steps:**

1. Sign a file on filesystem with a specified signature envelope:

    ```shell
    notation blob sign sbom.json --signature-format cose --id myKeyId --plugin myKMSPluginName
    ```

    This generates a signature named `sbom.json.cose.sig` in the same directory as `sbom.json`.

2. Publish the file and signature to the OCI-compliant registry:

    ```shell
    notation blob push --artifact-type "application/spdx+json" sbom.json:"application/json" --signature sbom.json.cose.sig --reference myregistry/mypath/mysboms:v1
    ```

    This command stores `sbom.json` as an OCI image referenced by `myregistry/mypath/mysboms:v1` in the registry, which has the artifact type `application/spdx+json` and `sbom.json` data stored as a layer with the media type `application/json`. Additionally, this command stores the signature file `sbom.json.cose.sig` as another OCI image with the artifact type `application/vnd.cncf.notary.signature` and the signature data as a layer with media type `application/cose`. The artifact type should be `application/vnd.cncf.notary.signature` as it is Notary Project signature. The layer media type can be detected according to the file name, which is either `application/cose` or `application/jws`. The signature manifest's `subject` property refers to the `myregistry/mypath/mysboms:v1`. To distinguish these signatures from those generated during OCI artifact signing, the specific annotation `io.cncf.notary.blob.signature=true` is added for the signature manifest.

    Below is an example of the manifest for the reference `myregistry/mypath/mysboms:v1`:

    ```json  
    {
        "schemaVersion": 2,
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "artifactType": "application/spdx+json",
        "config": {
            "mediaType": "application/vnd.oci.empty.v1+json",
            "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
            "size": 2,
            "data": "e30="
        },
        "layers": [
            {
            "mediaType": "application/json",
            "digest": "sha256:3154705eb5fc92bc3f91deec6008e6dd4a02b3e53ac18bdb7ea3ff03e62971a8",
            "size": 845,
            "annotations": {
                "org.opencontainers.image.title": "sbom.json"
            }
            }
        ],
        "annotations": {
            "org.opencontainers.image.created": "2025-02-18T06:00:00Z"
        }
    }
    ```

    Below is an example of the manifest for the blob signature in the registry:

    ```json
    {
        "schemaVersion": 2,
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "artifactType": "application/vnd.cncf.notary.signature",
        "config": {
            "mediaType": "application/vnd.oci.empty.v1+json",
            "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
            "size": 2,
            "data": "e30="
        },
        "layers": [
            {
            "mediaType": "application/cose",
            "digest": "sha256:6dc2a336b124d625f87deb3f539fa1465414ddad300a898adf4682e530d6f144",
            "size": 1479,
            "annotations": {
                "org.opencontainers.image.title": "sbom.json.cose.sig"
            }
            }
        ],
        "subject": {
            "mediaType": "application/vnd.oci.image.manifest.v1+json",
            "digest": "sha256:57eed31a09556933bddd4e164e1e085a220cf00489477eb98d89feac902eb42b",
            "size": 553
        },
        "annotations": {
            "io.cncf.notary.blob.signature": "true",
            "io.cncf.notary.x509chain.thumbprint#S256": "[\"aaaa\", \"bbbb\"]",
            "org.opencontainers.image.created": "2025-02-18T06:08:45Z"
        }
    }
    ```

    Alternatively, you can also use ORAS tool to push the blob file first, and then use Notation CLI command to push only the signature:

    ```shell
    oras push myregistry/mypath/mysboms:v1 sbom.json:application/octet-stream 
    notation blob push --signature sbom.json.cose.sig --reference myregistry/mypath/mysboms:v1
    ```

    You can also create an oci image layout and then use ORAS tool to copy it to the target registry. 

    ```shell
    notation blob push --artifact-type "application/spdx+json" sbom.json:"application/json" --signature sbom.json.cose.sig --oci-layout sbom.json:v1
    oras copy -r --from-oci-layout sbom.json:v1 myregistry/mypath/mysboms:v1
    ```

    > NOTE:
    >
    > Support for OCI image layout in Notation is currently an experimental feature.

3. Verify the signature from the registry:

    - Add the root CA certificate and set up the trust store:

        ```shell
        notation cert add --type ca --store myCACerts root.crt
        ```

        Confirm that the cert is added.

        ```shell
        notation cert ls
        ```

    - Set up the trust policy:

        ```shell
        notation blob policy init --name "myBlobPolicy" --trust-store "ca:myCACerts" --trust-identity "x509.subject:C=US,ST=WA,O=wabbit-network.io"
        ```

    - Verify the signature:

        ```shell
        notation blob verify --reference myregistry/mypath/mysboms:v1 --policy-name myBlobPolicy
        ```

### Manage blob policies

The following commands are available for managing blob policies:

- Initialize blob policies:

    ```shell
    notation blob policy init --name "myBlobPolicy" --trust-store "ca:myCACerts" --trust-identity "x509.subject:C=US,ST=WA,O=wabbit-network.io"
    ```

- Overwrite an existing policy with a prompt:

    ```shell
    notation blob policy init --trust-store "ca:myCACerts" --trust-identity "x509.subject:C=US,ST=WA,O=wabbit-network.io"
    ```

- Overwrite an existing policy with a prompt using the flag `--force`:

    ```shell
    notation blob policy init --force --trust-store "ca:myCACerts" --trust-identity "x509.subject:C=US,ST=WA,O=wabbit-network.io"
    ```

- Show the blob policy:

    ```shell
    notation blob policy show
    ```

- Modify the blob policy from a JSON file:

    - Export the existing blob policy to a JSON file:

        ```shell
        notation blob policy show > myBlobPolicy.json
        ```

    - Update and save the JSON file `myBlobPolicy.json`

    - Import the updated policies:

        ```shell
        notation blob policy import myBlobPolicy.json
        ```

- Set the global policy

    ```shell
    notation blob policy update --name myBlobPolicy --global
    ```

### Inspect blob signatures

The `notation blob inspect` command allows users to inspect blob signatures, providing output similar to the `notation inspect` command.

```shell
notation blob inspect sbom.json.cose.sig
```