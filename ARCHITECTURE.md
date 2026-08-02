# Notation Architecture

Notation is the [Notary Project](https://notaryproject.dev) CLI for signing and
verifying artifacts in the OCI registry ecosystem. It stores signatures as OCI
**referrers** alongside the artifact they sign — conceptually similar to git
commit signing, but generic and usable for arbitrary security purposes.

- **Module:** `github.com/notaryproject/notation/v2` (Go `>= 1.24`)
- **Entry point:** [`cmd/notation/main.go`](cmd/notation/main.go) → `run()` assembles the root Cobra command and registers all subcommands
- **Core frameworks:** [Cobra](https://github.com/spf13/cobra) (CLI), [`oras-go/v2`](https://oras.land) (OCI registry & OCI-layout I/O)
- **Notary libraries** (where the real work happens):
  - `notation-go` — sign / verify / list algorithms, trust policy, plugin manager
  - `notation-core-go` — signature envelopes, X.509, revocation (CRL/OCSP)
  - `notation-plugin-framework-go` — plugin contract
  - `tspclient-go` — RFC 3161 timestamping

## Table of Contents

- [Design Philosophy](#design-philosophy)
- [High-Level Architecture](#high-level-architecture)
- [Package Layout](#package-layout)
- [Core Workflows](#core-workflows)
- [Trust Model](#trust-model)
- [Developer Workflow](#developer-workflow)
- [User-Facing CLI Workflows](#user-facing-cli-workflows)
- [Improvement Opportunities](#improvement-opportunities)

## Design Philosophy

1. **Thin CLI layer.** Command files in `cmd/notation/` only parse flags and
   orchestrate. All cryptographic algorithms live in `notation-go` /
   `notation-core-go`. This keeps the CLI auditable and lets the libraries be
   reused by other tools.
2. **Centralized registry plumbing.** All OCI registry, auth, and OCI-layout
   access flows through [`registry.go`](cmd/notation/registry.go)
   (`getRepository`), so transport, credentials, and referrers-fallback logic
   live in one place.
3. **Two `internal/` trees, two concerns.**
   - `cmd/notation/internal/` — CLI *presentation* concerns (flags, output
     rendering, signer/verifier construction, experimental gating).
   - `internal/` — reusable *lower-level* utilities (auth store, config cache,
     HTTP client, revocation, x509, version).
4. **Pluggable everything.** Signing keys, verification, and revocation are all
   pluggable via the plugin framework and JSON trust policies.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  CLI Layer  (cmd/notation/*.go — Cobra commands)             │
│  sign · verify · list · inspect · login/logout · key         │
│  subcommands: blob/ · cert/ · plugin/ · policy/              │
└───────────────┬─────────────────────────────────────────────┘
                │ delegates to
┌───────────────▼─────────────────────────────────────────────┐
│  CLI-internal (cmd/notation/internal/)                       │
│  flag · display · sign(GetSigner) · verify(GetVerifier)      │
│  truststore · experimental · errors · plugin                 │
└───────────────┬─────────────────────────────────────────────┘
                │ uses
┌───────────────▼──────────────┐   ┌──────────────────────────┐
│  Shared utils (internal/)    │   │  External Notary libs     │
│  auth · config · envelope    │   │  notation-go   (sign/vfy) │
│  httputil · revocation · x509│   │  notation-core-go (crypto)│
│  osutil · trace · version    │   │  oras-go       (registry) │
└──────────────────────────────┘   └──────────────────────────┘
```

## Package Layout

### `cmd/notation/` — Top-Level Commands

| File | Purpose |
|------|---------|
| `sign.go` | Signs an OCI artifact. Resolves a signer, builds `SignOptions`, calls `notation.SignOCI`, pushes signature as a referrer. |
| `verify.go` | Verifies signatures on an OCI artifact via `notation.Verify` and renders results. |
| `list.go` | Lists all signatures for an artifact (`ListSignatures`), rendered as a tree. |
| `inspect.go` | Decodes each signature envelope and renders as tree or JSON. |
| `login.go` / `logout.go` | Registry credential management via oras `credentials`. |
| `key.go` | Manages the signing-key list in `config.json` (`key add/update/delete/list`). |
| `registry.go` | **Not a command** — OCI registry plumbing (`getRepository`, auth client, referrers fallback). |
| `manifest.go` | **Not a command** — reference resolution (tag→digest), mutable-tag warnings. |
| `version.go` | Prints version, Go version, OS/arch, git commit. |

### `cmd/notation/` — Subcommand Packages

- **`blob/`** — detached-signature workflow for arbitrary files (`sign`, `verify`, `inspect`) plus `blob/policy/` for blob trust-policy management.
- **`cert/`** — trust-store certificate management (`add`, `list`, `show`, `delete`, `generate-test`, `cleanup-test`).
- **`plugin/`** — plugin lifecycle (`list`, `install` from file/URL, `uninstall`).
- **`policy/`** — OCI trust-policy management (`import`, `show`).

### `cmd/notation/internal/` — CLI Presentation

| Package | Purpose |
|---------|---------|
| `flag/` | Shared flag option structs (signer, logging, secure/credentials, output format). |
| `display/` | Output-handler factory + `metadata/` renderers (json/text/tree) and `output/` printer. |
| `sign/` | `GetSigner` — resolves a local X.509 key pair or a plugin signer. |
| `verify/` | `GetVerifier`/`GetBlobVerifier` — builds a verifier with trust policy, trust store, revocation, plugins; composes failure messages. |
| `truststore/` | Trust-store filesystem operations and validation. |
| `experimental/` | Gates experimental features behind `NOTATION_EXPERIMENTAL=1`. |
| `errors/` | Typed CLI errors (referrers-API unsupported, missing reference, max signatures). |
| `plugin/` | Plugin source handling (file vs URL), download-size limits, media types. |

### `internal/` — Shared Utilities

| Package | Purpose |
|---------|---------|
| `auth/` | Builds an oras credentials store (Notation → Docker → OS-native). |
| `config/` | Cached read of `config.json`; `IsRegistryInsecure` helper. |
| `envelope/` | Signature envelope constants (COSE/JWS) and payload wrapping. |
| `httputil/` | oras `auth.Client`/`http.Client` with retry, debug logging, user agent. |
| `revocation/` | CRL/OCSP revocation validators with on-disk CRL cache (`crl/`). |
| `x509/` | X.509 helpers (root cert detection, cert pools for TSA roots). |
| `osutil/` | File utilities (SHA-256, size-limited writes/downloads). |
| `slices/` | Generic `Contains` helper. |
| `trace/` | HTTP debug-logging transport and logger-level context. |
| `version/` | Version constants (`Version`, `BuildMetadata`, `GitCommit`) injected at build time. |

### Supporting Directories

- **`specs/`** — design docs and per-command specifications, plus `proposals/` (blob signing, diagnostic experience, dm-verity).
- **`test/e2e/`** — a separate Go module of Ginkgo end-to-end tests, plus a mock plugin module.

## Core Workflows

### Signing (`runSign`, `cmd/notation/sign.go`)

```
sign.GetSigner ──► notation-go signer (local key pair OR plugin signer)
      │
getRepository ──► oras OCI repository (remote registry or OCI layout)
      │
prepareSigningOpts ──► notation.SignOptions
      │   (envelope media type, expiry, plugin config,
      │    user metadata, optional TSA timestamper + revocation)
      │
resolveReference ──► tag resolved to immutable digest
      │
notation.SignOCI(ctx, signer, sigRepo, opts)
      └──► notation-go signs and PUSHES the signature to the registry
           as an OCI referrer of the target artifact.
```

### Verification (`runVerify`, `cmd/notation/verify.go`)

```
verify.GetVerifier ──► verifier with:
      │   • X.509 trust store
      │   • OCI trust policy (trustpolicy.LoadOCIDocument)
      │   • revocation validators (CRL/OCSP)
      │   • plugin manager
      │
getRepository + resolveReference
      │
notation.Verify(ctx, sigVerifier, sigRepo, opts)
      │
ComposeVerificationFailurePrintout (on failure)
      │
display handler renders result (tree / json / text)
```

Blob commands mirror these flows using `notation.BlobSigner` / `BlobVerifier`.

## Trust Model

Verification is governed by three on-disk artifacts, all under the Notation
config directory (overridable with `NOTATION_CONFIG`):

1. **Trust store** — X.509 CA/certificate roots under
   `truststore/x509/<type>/<store>/`, managed by `notation cert`.
2. **Trust policy** — a JSON document (`trustpolicy.oci.json` /
   `trustpolicy.blob.json`) mapping registry scopes to trust stores, signature
   verification levels, and trusted identities; managed by `notation policy` /
   `notation blob policy`.
3. **Signing keys** — the key list in `config.json`, managed by `notation key`.

Revocation (CRL/OCSP) and RFC 3161 timestamping are layered on top during both
signing and verification.

## Developer Workflow

Build tooling is driven by the [`Makefile`](Makefile). Key targets:

| Command | Description |
|---------|-------------|
| `make build` | Builds `bin/notation` with version info injected via `-ldflags`. |
| `make install` | Builds and copies `notation` to `~/bin/`. |
| `make test` | Runs unit tests with race detector and coverage (`coverage.txt`). |
| `make lint` | Runs `golangci-lint` using `.golangci.yml` (adopted incrementally via `new-from-rev`). |
| `make e2e` | Builds the CLI and runs the Ginkgo e2e suite against a `zot` registry. |
| `make e2e-covdata` | Runs e2e with binary coverage instrumentation. |
| `make download` | Downloads Go module dependencies. |
| `make vendor` | Vendors Go modules. |
| `make check-line-endings` / `fix-line-endings` | Enforces LF line endings on `.go` files. |

Version metadata (`GitCommit`, `BuildMetadata`) is injected at link time from
git state — see the `LDFLAGS` block in the Makefile.

### CI/CD Pipeline (`.github/workflows/`)

```
 push / pull_request
        │
   build.yml (Continuous Integration)
        ├── check signed commits (PRs must be signed)
        ├── setup Go 1.24 + module cache
        ├── make download → make build
        ├── golangci-lint (new-issue gating)
        ├── make test           (unit, race, coverage)
        ├── make e2e-covdata     (e2e against zot)
        └── upload coverage → codecov.io

 push tag v*
        │
   release-github.yml
        └── GoReleaser (v2) → GitHub release with cross-platform binaries
```

Additional workflows: `codeql.yml` (SAST), `scorecard.yml` (OpenSSF),
`license-checker.yml`, `stale.yml`, `add-to-project.yml`. Dependencies are kept
current by `dependabot.yml`. All commits must be **signed**.

## User-Facing CLI Workflows

### Sign & verify an image (registry)

```sh
# 1. Authenticate to the registry
notation login registry.example.com

# 2. Register a signing key (or use a plugin/KMS key)
notation cert generate-test --default "example.com"

# 3. Sign the artifact (signature stored as an OCI referrer)
notation sign registry.example.com/app:v1

# 4. Configure a trust policy and verify
notation policy import ./trustpolicy.json
notation verify registry.example.com/app:v1

# 5. Inspect / list signatures
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

## Improvement Opportunities

The following are suggestions, not existing commitments. They are ordered by
approximate value-to-effort.

### Documentation & onboarding
- **This document** — keep `ARCHITECTURE.md` linked from `README.md` and
  `building.md` so new contributors have a map of the codebase.
- Add a short **"first contribution" walkthrough** (build → run a unit test →
  run one e2e case) to `building.md`; currently it only covers `make install`.

### Developer workflow / CI
- **Split CI jobs** — unit and e2e run in one job; separating them (with e2e as
  a matrix over registry backends) improves signal and parallelism.
- **Cache e2e registry images** to reduce flakiness and runtime.
- **Burn down the lint backlog.** `make lint` is adopted incrementally
  (`new-from-rev` in `.golangci.yml`) so only new issues fail CI. Roughly ~76
  pre-existing findings remain; clearing them lets `new-from-rev` be removed to
  enforce linting across the whole tree.
- **Pin and document the Go toolchain** via a `go.work`/toolchain directive so
  local and CI builds are reproducible.

### Codebase / architecture
- **Consolidate reference-resolution logic.** `registry.go` and `manifest.go`
  both handle registry-vs-OCI-layout branching; a single `target resolver`
  abstraction would reduce duplication across `sign`/`verify`/`list`/`inspect`.
- **Unify OCI and blob flows.** Signing/verification exist in near-parallel OCI
  and blob variants; extracting a shared orchestration interface (signer +
  target + options) would cut duplicated flag/plumbing code.
- **Structured error taxonomy.** Extend `cmd/notation/internal/errors` into a
  consistent, exit-code-mapped error hierarchy so scripts can reliably branch on
  failure classes (network vs. policy vs. crypto).
- **Machine-readable output everywhere.** `--output json` is available on some
  commands (inspect/verify); making it uniform across `list`, `sign`, and `key`
  would improve scriptability and CI integration.

### User experience
- **Progress + diagnostics.** A consistent `--debug`/verbosity story (already
  partly present via `trace/`) surfaced uniformly across commands aids
  troubleshooting of registry/auth failures.
- **`notation doctor`** — a diagnostic command that validates config directory,
  trust policy syntax, trust-store contents, and registry connectivity in one
  shot would shorten support loops.

---

*This document reflects the repository state at the time of writing. When
command wiring in `cmd/notation/main.go` or the package layout changes, update
the corresponding sections here.*
