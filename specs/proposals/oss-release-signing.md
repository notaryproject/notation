## Proposal for Signing and Verifying OSS Release Assets

### Overview

This document proposes enhancing Notation's capabilities for signing and verifying Open Source Software (OSS) community released assets. It outlines the design for a new vendor-neutral `notation-signer` plugin and introduces a simplified, stateless verification mode for the `notation blob verify` command.

### Problem Statement & Motivation

Notation v2.0.0-alpha.1 introduced blob signing, enabling the signing of assets like GitHub releases. However, several issues remain:
1.  A key challenge is the lack of a vendor-neutral Notation plugin; existing plugins are cloud-provider specific (AKV, AWS), while the OSS community often prefers neutral solutions.
2.  Furthermore, the standard verification process is overly complex for typical community users whose primary need is simple verification of a downloaded file and its signature. This complexity, requiring multiple manual steps (adding certificates to a trust store, configuring trust policies) for each publisher, hinders adoption for simple asset verification use cases.

Addressing these challenges is crucial to significantly broaden Notation's use case and improve its adoption within the OSS community for asset signing.

**Goals:**

* Provide a vendor-neutral signing solution for OSS publishers.
* Simplify the verification experience for OSS users needing to verify downloaded assets and their signatures.
* Significantly broaden the use case and improve adoption of Notation within the OSS community for asset signing.

**Non-Goals:**

* Overhaul the general Notation trust store and trust policy management system for complex or persistent verification scenarios.
* Address signing or verification of container images (as blob signing specifically targets arbitrary files).

### Background & Context

Notation's existing blob signing feature provides the foundational capability to sign arbitrary files. However, the initial implementation focused on cloud-specific key management solutions like AKV and AWS, leaving a gap for users seeking a more universally deployable or self-hosted signing method. Simultaneously, the standard `notation verify` command workflow, designed for robust and policy-driven verification, presents a barrier to entry for users with straightforward, one-off verification needs.

### Scenarios

**OSS Publishers:** A publisher needs a reliable, automated way to sign release binaries or tarballs before uploading them to platforms like GitHub Releases. They require a signing solution that is easily integrated into CI/CD and doesn't depend on specific cloud providers, ideally supporting keys stored securely (e.g., encrypted locally or in secrets) and certificates publicly available (e.g., in the code repository).

**OSS Users:** A user downloads an asset file and its detached signature from an OSS project release page. They need a simple command-line tool to quickly verify the file's authenticity and integrity using the signature file and publicly available certificate information (like a URL, fingerprint, and trusted identity).

### Impact to Users and Ecosystem

Implementing a vendor-neutral signer plugin would empower a wider range of OSS projects to adopt Notation for signing their releases without relying on specific cloud infrastructure, fostering a more diverse signing ecosystem. The simplified verification workflow will dramatically lower the barrier to entry for users wanting to verify OSS downloads, increasing confidence in the provenance of community-released software and potentially improving overall supply chain security practices within the OSS space.

### Existing Solutions or Expectations

Currently, verifying a detached blob signature with Notation requires users to manually download the root certificate, add it to their local trust store, define a trust policy referencing the trust store, and then execute the `notation verify` command. This multi-step process, involving at least 4-5 commands, is cumbersome for one-time verifications. Users often expect a simpler mechanism, perhaps a single command, to check the signature of a downloaded file against a publicly available certificate and identity claim.

### Proposal

This proposal suggests two key enhancements:

1.  **Notation Signer Plugin (`notation-signer`)**: A new plugin implementation allowing users to sign blobs or artifacts using a locally stored key and certificate bundle. The key would ideally be encrypted, with the password manageable via environment variables or secure inputs (e.g., GitHub Actions secrets). The certificate could be stored alongside the signed assets or in the code repository for easy public access. This provides a vendor-neutral signing option. 

    **Desired User Experience / CLI Operation:**
    1. The publisher obtains a private key `notation.key` with a certificate bundle `notation.pem`, either self-signed or CA-issued.
    2. Download the `notation-signer` plugin locally, and call `encrypt` command with password to get an encrypted key `notation.key.enc`.
        ```sh
        # on local machine
        notation-signer encrypt notation.key
        ```
    3. The publisher adds the content of the encrypted key `notation.key.enc` and the password to the CI workflow secret (GitHub Action Secret), then exports them as environment variables `NOTATION_SIGNER_KEY` and `NOTATION_SIGNER_KEY_PASSWORD` later for the `notation` binary.
    4. Configure the release workflow to set up `notation` with the `notation-signer` plugin.
    5. Export the secrets as environment variables for notation and call the `notation blob sign` command to sign the release asset:
        ```sh
        # on CI workflow
        notation blob sign --id signer --plugin signer \
          --plugin-config certificate_bundle_path=./notation.pem \
          <release-asset-path>
        ```
    6. The publisher attaches the assets, along with their signatures, to the release page.

2.  **Simplified Notation Blob Verify Command (`notation blob verify`)**: This proposes adding flags to the existing `notation blob verify` command to enable a simplified, stateless verification mode for blob signatures. These flags are:
    * `--certificate-url <url>`: A URL to the root certificate.
    * `--certificate-sha256-fingerprint <fingerprint>`: A SHA256 fingerprint of the root certificate to verify the downloaded file.
    * `--certificate-path <path>`: *Alternatively*, provide a local root certificate file path instead of a remote URL and fingerprint.
    * `--trusted-identity <identity>`: A trusted identity claim for the signing certificate (e.g., `x509.subject:...`).

    When these flags are provided, the command will operate in a **stateless mode**:
    * If `--certificate-url` is used, download the root certificate and verify its SHA256 fingerprint against the specified value. If `--certificate-path` is used, load the local certificate.
    * Create an in-memory trust store containing the verified root certificate.
    * Create an in-memory, temporary trust policy based on the provided trusted identity and the in-memory trust store.
    * Perform the signature verification of the target file using the signature file against the in-memory trust store and policy.
    * Report success or failure.

    This approach eliminates the need for users to manage persistent trust stores or policy files for simple verification tasks.

    **Desired User Experience / CLI Operation:**

    ```bash
    SIGNATURE=<signature-path>
    TARGET_FILE=<signed-target-file-path>

    # Using remote certificate
    notation blob verify \
      --certificate-url "https://raw.githubusercontent.com/JeyJeyGao/notation-local-signer/refs/tags/v0.1.0/notation.crt" \
      --certificate-sha256-fingerprint "F3:5E:B5:3F:6A:BF:55:89:BA:51:EB:39:7B:1A:BA:3A:0A:30:77:14:2C:12:BD:86:EF:5F:CD:54:C5:BE:8B:C4" \
      --trusted-identity "x509.subject: C = US, ST = Redmond, L = Redmond, O = notation, CN = notation-local-signer" \
      --signature $SIGNATURE \
      $TARGET_FILE

    # Or using a local certificate file
     notation blob verify \
      --certificate-path "./notation.crt" \
      --trusted-identity "x509.subject: C = US, ST = Redmond, L = Redmond, O = notation, CN = notation-local-signer" \
      --signature $SIGNATURE \
      $TARGET_FILE
    ```

    **Security Considerations for Stateless Verification:**

    The security of the stateless verification process heavily relies on the user providing correct and trustworthy values for the certificate source (URL/path and fingerprint) and the trusted identity. The certificate fingerprint check (when using a URL) is a critical step to mitigate risks associated with downloading the certificate from a potentially compromised location. Publishers must ensure these values are accurate. Providing these verification details from a source separate from where the signed asset and signature are hosted (e.g., the project's official website rather than just the GitHub release page) can enhance trust by avoiding reliance on a single point of compromise.
