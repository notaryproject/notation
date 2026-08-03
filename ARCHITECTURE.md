# Notation Architecture

`notation` is the reference CLI of the [Notary Project](https://notaryproject.dev), a CNCF incubating project providing a cross-industry standard for securing software supply chains. It signs and verifies OCI artifacts and arbitrary blobs, storing OCI signatures alongside the artifact they sign so they remain portable across any OCI-compliant registry.

> **The Notary Project specifications are the source of truth.** This document describes how *this CLI* is put together and where it delegates to the shared libraries. It is a map for contributors, not a restatement of the specs. For normative behaviour always refer to [notaryproject/specifications](https://github.com/notaryproject/specifications).

- **Module:** `github.com/notaryproject/notation/v2` (Go `>= 1.24`)
- **Entry point:** [`cmd/notation/main.go`](cmd/notation/main.go) — `run()` assembles the root [Cobra](https://github.com/spf13/cobra) command and registers every subcommand
- **Package detail:** run `go doc ./...` or browse [pkg.go.dev](https://pkg.go.dev/github.com/notaryproject/notation/v2); this document deliberately stays architectural so it does not drift from the code

## Table of Contents

- [Where Notation Sits in the Notary Project](#where-notation-sits-in-the-notary-project)
- [Specification Map](#specification-map)
- [Design Philosophy](#design-philosophy)
- [High-Level Architecture](#high-level-architecture)
- [Core Workflows](#core-workflows)
- [Trust Model](#trust-model)
- [Signature Storage in OCI Registries](#signature-storage-in-oci-registries)
- [Plugin Extensibility](#plugin-extensibility)
- [Developer Workflow](#developer-workflow)
- [User-Facing CLI Workflows](#user-facing-cli-workflows)

## Where Notation Sits in the Notary Project

The CLI is intentionally thin. Signing and verification algorithms, envelope formats, trust evaluation and revocation all live in shared libraries that other tools reuse.

```mermaid
flowchart TD
    SPEC("<b>notaryproject/specifications</b><br/>normative behaviour<br/>signature · trust policy · workflows · plugins")
    CLI("<b>notation</b> (this repo)<br/>Cobra CLI · flags · output rendering")
    NG("<b>notation-go</b><br/>signer · verifier · trust policy<br/>trust store · plugin manager · registry")
    NCG("<b>notation-core-go</b><br/>signature envelopes (JWS · COSE)<br/>x509 · revocation (CRL/OCSP)")
    TSP("<b>tspclient-go</b><br/>RFC 3161 timestamping")
    PFW("<b>notation-plugin-framework-go</b><br/>plugin contract")
    ORAS("<b>oras-go/v2</b><br/>OCI registry & OCI-layout I/O")
    PLUGIN("Third-party plugins<br/>KMS / HSM signing binaries")

    SPEC -.->|"governs"| CLI
    SPEC -.->|"governs"| NG
    CLI -->|"orchestrates"| NG
    NG --> NCG
    NG --> TSP
    NG --> ORAS
    NG -->|"invokes"| PLUGIN
    PFW -.->|"defines contract"| PLUGIN

    classDef spec fill:#fef3c7,stroke:#b45309,stroke-width:1.5px,color:#78350f;
    classDef cli fill:#dbeafe,stroke:#1d4ed8,stroke-width:2px,color:#1e3a8a;
    classDef lib fill:#dcfce7,stroke:#15803d,stroke-width:1.5px,color:#14532d;
    classDef ext fill:#f3e8ff,stroke:#7e22ce,stroke-width:1.5px,color:#581c87;

    class SPEC spec
    class CLI cli
    class NG,NCG,TSP,PFW lib
    class ORAS,PLUGIN ext
```

| Dependency | Version in [`go.mod`](go.mod) | Role |
|---|---|---|
| [`notation-go`](https://github.com/notaryproject/notation-go) | `v1.2.0-beta.1.0.20250512015818-2bc67e7695ef` | Signing/verification engine, trust policy & trust store evaluation, plugin manager, registry glue |
| [`notation-core-go`](https://github.com/notaryproject/notation-core-go) | `v1.3.0` | Signature envelope generation/parsing (JWS, COSE), certificate chain validation, CRL/OCSP revocation |
| [`notation-plugin-framework-go`](https://github.com/notaryproject/notation-plugin-framework-go) | `v1.0.0` | Contract shared with out-of-process plugin binaries |
| [`tspclient-go`](https://github.com/notaryproject/tspclient-go) | `v1.0.1-0.20250306063739-4f55b14d9f01` | RFC 3161 Time-Stamp Protocol client |
| [`oras-go/v2`](https://oras.land) | `v2.6.0` | OCI registry client, registry auth, OCI image layout I/O |

> Two dependencies are pinned to untagged commits (`notation-go`, `tspclient-go`) rather than released tags, so this tree tracks pre-release library work.

## Specification Map

Each Notary Project specification maps onto concrete parts of this CLI:

| Specification | What it governs here |
|---|---|
| [signature-specification](https://github.com/notaryproject/specifications/blob/main/specs/signature-specification.md) | Signature payload, OCI signature manifest shape, blob signature file naming, required algorithms |
| [signature-envelope-jws](https://github.com/notaryproject/specifications/blob/main/specs/signature-envelope-jws.md) · [signature-envelope-cose](https://github.com/notaryproject/specifications/blob/main/specs/signature-envelope-cose.md) | The two envelope formats selectable via `--signature-format` |
| [trust-store-trust-policy](https://github.com/notaryproject/specifications/blob/main/specs/trust-store-trust-policy.md) | `notation cert *` (trust store) and `notation policy` / `notation blob policy` (trust policy), plus revocation evaluation |
| [signing-and-verification-workflow](https://github.com/notaryproject/specifications/blob/main/specs/signing-and-verification-workflow.md) | The four workflows exposed as `sign`, `verify`, `blob sign`, `blob verify` |
| [signing-scheme](https://github.com/notaryproject/specifications/blob/main/specs/signing-scheme.md) | `notary.x509` vs `notary.x509.signingAuthority` — which trust store type is consulted and how authentic signing time is established |
| [plugin-extensibility](https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md) | `notation plugin *`, plugin discovery and the signing/verification plugin contracts |

This repo additionally carries **CLI-level** (non-normative) documentation under [`specs/`](specs/): [`notation-cli.md`](specs/notation-cli.md), [`error-handling-guideline.md`](specs/error-handling-guideline.md), [`registry-auth.md`](specs/registry-auth.md), per-command references in [`specs/cmd/`](specs/cmd/), and design proposals in [`specs/proposals/`](specs/proposals/).

## Design Philosophy

1. **Thin CLI layer.** Command files in `cmd/notation/` parse flags and orchestrate. Cryptographic work lives in `notation-go` / `notation-core-go`, keeping the CLI auditable and the libraries reusable.
2. **Centralized registry plumbing.** All registry, auth and OCI-layout access flows through [`registry.go`](cmd/notation/registry.go) (`getRepository`), so transport, credentials and referrers-fallback logic live in one place.
3. **Two `internal/` trees, two concerns.** `cmd/notation/internal/` holds CLI *presentation* concerns; `internal/` holds reusable lower-level utilities.
4. **Pluggable by design.** Signing keys, verification and revocation are all pluggable via the plugin framework and JSON trust policies.

## High-Level Architecture

```mermaid
flowchart TD
    A("CLI layer<br/><b>cmd/notation/</b><br/>sign · verify · list · inspect · login · key<br/>+ blob/ · cert/ · plugin/ · policy/")
    B("CLI-internal<br/><b>cmd/notation/internal/</b><br/>flag · display · sign · verify<br/>truststore · experimental · errors · plugin")
    C("Shared utilities<br/><b>internal/</b><br/>auth · config · envelope<br/>httputil · revocation · x509")
    D("External Notary libraries<br/>notation-go · notation-core-go · oras-go")

    A -->|"delegates to"| B
    B -->|"built on"| C
    B -->|"calls"| D

    classDef cli fill:#dbeafe,stroke:#1d4ed8,stroke-width:1.5px,color:#1e3a8a;
    classDef internal fill:#fef3c7,stroke:#b45309,stroke-width:1.5px,color:#78350f;
    classDef utils fill:#dcfce7,stroke:#15803d,stroke-width:1.5px,color:#14532d;
    classDef libs fill:#f3e8ff,stroke:#7e22ce,stroke-width:1.5px,color:#581c87;

    class A cli
    class B internal
    class C utils
    class D libs
```

## Core Workflows

### Signing (`runSign`, [`cmd/notation/sign.go`](cmd/notation/sign.go))

```mermaid
flowchart LR
    A("GetSigner<br/>local key or plugin") --> B("getRepository<br/>registry or OCI layout")
    B --> C("prepareSigningOpts<br/>envelope · expiry · TSA · revocation")
    C --> D("resolveReference<br/>tag → digest")
    D --> E("notation.SignOCI<br/>pushes signature as OCI referrer")

    classDef step fill:#dbeafe,stroke:#1d4ed8,stroke-width:1.5px,color:#1e3a8a;
    class A,B,C,D,E step
```

Signing always resolves a tag to an immutable digest first, so a signature is bound to specific content rather than a movable tag.

### Verification (`runVerify`, [`cmd/notation/verify.go`](cmd/notation/verify.go))

```mermaid
flowchart LR
    A("GetVerifier<br/>trust store · trust policy<br/>revocation · plugins") --> B("getRepository +<br/>resolveReference")
    B --> C("notation.Verify")
    C -->|"failure"| D("ComposeVerification<br/>FailurePrintout")
    C --> E("display handler<br/>tree · json · text")
    D --> E

    classDef step fill:#dcfce7,stroke:#15803d,stroke-width:1.5px,color:#14532d;
    classDef fail fill:#fee2e2,stroke:#b91c1c,stroke-width:1.5px,color:#7f1d1d;
    class A,B,C,E step
    class D fail
```

Blob commands mirror both flows using `notation.BlobSigner` / `BlobVerifier` against local files instead of a registry.

## Trust Model

> Trust store layout and trust policy semantics are defined normatively in [trust-store-trust-policy](https://github.com/notaryproject/specifications/blob/main/specs/trust-store-trust-policy.md). The summary below is a practical orientation to the on-disk artifacts this CLI manages.

Verification is governed by three artifacts under the Notation config directory (overridable with `NOTATION_CONFIG`):

```mermaid
flowchart TD
    CFG("Notation config directory")
    TS("<b>Trust store</b><br/>truststore/x509/{ca,signingAuthority,tsa}/&lt;store&gt;/<br/>managed by <code>notation cert</code>")
    TP("<b>Trust policy</b><br/>trustpolicy.oci.json · trustpolicy.blob.json<br/>managed by <code>notation policy</code> / <code>notation blob policy</code>")
    KEY("<b>Signing keys</b><br/>key list in config.json<br/>managed by <code>notation key</code>")

    CFG --> TS
    CFG --> TP
    CFG --> KEY
    TP -->|"selects"| TS

    classDef root fill:#e0e7ff,stroke:#4338ca,stroke-width:2px,color:#312e81;
    classDef leaf fill:#dcfce7,stroke:#15803d,stroke-width:1.5px,color:#14532d;
    class CFG root
    class TS,TP,KEY leaf
```

- **Trust store** holds X.509 roots under three identity types: `ca`, `signingAuthority` and `tsa`. The `tsa` store is what makes RFC 3161 timestamp verification meaningful.
- **Trust policy** maps registry scopes (or blob policy names) to trust stores, verification levels and trusted identities.
- **Signing scheme** (`notary.x509` vs `notary.x509.signingAuthority`) determines which trust store type is consulted and how authentic signing time is established.

Revocation (CRL/OCSP, with an on-disk CRL cache in [`internal/revocation/crl`](internal/revocation/crl)) and RFC 3161 timestamping layer on top of both signing and verification.

## Signature Storage in OCI Registries

Signatures are stored as OCI **referrers** of the artifact they sign — discoverable via the [Referrers API](https://github.com/opencontainers/distribution-spec/blob/v1.1.0/spec.md#listing-referrers), with a [referrers tag schema](https://github.com/opencontainers/distribution-spec/blob/v1.1.0/spec.md#referrers-tag-schema) fallback for registries that do not implement it.

```mermaid
flowchart BT
    ART("Signed artifact<br/>registry.example.com/app:v1<br/>@sha256:abc…")
    SIG1("Signature manifest #1<br/>artifactType:<br/>application/vnd.cncf.notary.signature")
    SIG2("Signature manifest #2<br/>re-signed / rotated key")
    BLOB1("Signature envelope<br/>JWS or COSE")
    BLOB2("Signature envelope<br/>JWS or COSE")

    BLOB1 --> SIG1
    BLOB2 --> SIG2
    SIG1 -->|"subject"| ART
    SIG2 -->|"subject"| ART

    classDef art fill:#dbeafe,stroke:#1d4ed8,stroke-width:2px,color:#1e3a8a;
    classDef sig fill:#fef3c7,stroke:#b45309,stroke-width:1.5px,color:#78350f;
    classDef blob fill:#f3e8ff,stroke:#7e22ce,stroke-width:1.5px,color:#581c87;
    class ART art
    class SIG1,SIG2 sig
    class BLOB1,BLOB2 blob
```

An artifact can carry multiple signatures. `notation list` enumerates them and `notation inspect` decodes each envelope. Because signatures live beside the artifact rather than in a separate database, they travel with the artifact when it is copied between registries.

## Plugin Extensibility

Plugins are standalone executables invoked by `notation-go` over the contract defined in [plugin-extensibility](https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md), letting keys stay in a KMS/HSM and never touch the CLI process.

```mermaid
flowchart LR
    CLI("notation sign --key mykey") --> PM("notation-go<br/>plugin manager")
    PM -->|"describe-key<br/>generate-signature /<br/>generate-envelope"| PB("notation-&lt;name&gt; binary")
    PB --> KMS("KMS · HSM · signing service")

    classDef cli fill:#dbeafe,stroke:#1d4ed8,stroke-width:1.5px,color:#1e3a8a;
    classDef mgr fill:#dcfce7,stroke:#15803d,stroke-width:1.5px,color:#14532d;
    classDef ext fill:#f3e8ff,stroke:#7e22ce,stroke-width:1.5px,color:#581c87;
    class CLI cli
    class PM mgr
    class PB,KMS ext
```

`notation plugin install/list/uninstall` manage the plugin lifecycle; [`cmd/notation/internal/plugin`](cmd/notation/internal/plugin) handles install sources (file vs URL) and download limits.

## Developer Workflow

Build tooling is driven by the [`Makefile`](Makefile):

| Command | Description |
|---------|-------------|
| `make build` | Builds `bin/notation` with version info injected via `-ldflags` |
| `make install` | Builds and copies `notation` to `~/bin/` |
| `make test` | Runs unit tests with race detector and coverage (`coverage.txt`) |
| `make e2e` | Builds the CLI and runs the Ginkgo e2e suite against a `zot` registry |
| `make e2e-covdata` | Runs e2e with binary coverage instrumentation |
| `make download` / `make vendor` | Downloads / vendors Go module dependencies |
| `make check-line-endings` / `fix-line-endings` | Enforces LF line endings on `.go` files |

Version metadata (`Version`, `BuildMetadata`, `GitCommit`) lives in [`internal/version`](internal/version) and is injected at link time — see the `LDFLAGS` block in the Makefile.

### CI/CD Pipeline ([`.github/workflows/`](.github/workflows))

```mermaid
flowchart TD
    subgraph CI["build.yml — on push / pull_request"]
        direction TB
        S1("Check signed commits") --> S2("Setup Go 1.24 + module cache")
        S2 --> S3("make build")
        S3 --> S4("make test<br/>unit · race · coverage")
        S4 --> S5("make e2e-covdata<br/>against zot")
        S5 --> S6("Upload coverage<br/>to codecov.io")
    end

    subgraph REL["release-github.yml — on tag v*"]
        direction TB
        T1("GoReleaser v2") --> T2("GitHub release<br/>cross-platform binaries")
    end

    classDef step fill:#dbeafe,stroke:#1d4ed8,stroke-width:1.5px,color:#1e3a8a;
    classDef rel fill:#dcfce7,stroke:#15803d,stroke-width:1.5px,color:#14532d;
    class S1,S2,S3,S4,S5,S6 step
    class T1,T2 rel
```

Additional workflows: `codeql.yml` (SAST), `scorecard.yml` (OpenSSF), `license-checker.yml`, `stale.yml`, `add-to-project.yml`. Dependencies are kept current by `dependabot.yml`. **All commits must be signed.**

## User-Facing CLI Workflows

### Sign and verify an image in a registry

```sh
# 1. Authenticate to the registry
notation login registry.example.com

# 2. Generate a test key + certificate (production: use a plugin/KMS key)
notation cert generate-test --default "example.com"

# 3. Sign the artifact — the signature is pushed as an OCI referrer
notation sign registry.example.com/app:v1

# 4. Configure a trust policy, then verify
notation policy import ./trustpolicy.json
notation verify registry.example.com/app:v1

# 5. Enumerate and decode signatures
notation list    registry.example.com/app:v1
notation inspect registry.example.com/app:v1
```

### Detached blob signing

```sh
notation blob policy import ./trustpolicy.blob.json
notation blob sign   ./artifact.tar.gz            # emits a detached .sig
notation blob verify --signature artifact.tar.gz.sig ./artifact.tar.gz
```

### Plugin / KMS-backed signing

```sh
notation plugin install --file ./notation-myplugin.tar.gz
notation key add --plugin myplugin --id <key-id> mykey
notation sign --key mykey registry.example.com/app:v1
```

---

*This document reflects the repository state at the time of writing. When command wiring in [`cmd/notation/main.go`](cmd/notation/main.go) or the high-level structure changes, please update the corresponding sections.*
