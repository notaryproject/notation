# Monthly Notation patch releases

The canonical release controller lives in `notaryproject/notation`. It assesses
Core, Go, and CLI together, reserves independent stable patch versions, and
dispatches workers in dependency order. This draft is disabled by default.
The [Proposal: automated monthly Notation patch releases](MONTHLY_PATCH_UPSTREAM_PROPOSAL.md)
describes the shared scope and maintainer responsibilities.

## Authorization and approval

The read-only assessment agent produces evidence-bound release, skip, or defer
decisions. It cannot authorize a release, merge a PR, change repository settings,
or waive a finding. The controller authenticates the successful default-main
agent run, revalidates its artifact, and checks the maintainer-owned
`.github/monthly-patch-policy.yml`.

The implemented routine policy permits stable, nondecreasing Go dependency
patches on the same major/minor, without changing module identities, dependency
membership, or Go requirements. Workflow dependency changes and broader updates
require a manual controller dispatch by `MONTHLY_PATCH_ACTOR` with
`approve=true`. The policy does not authorize feature merges, source bug fixes,
or CVE backports. Reachable source and binary `govulncheck` findings block
canonical releases at every severity; fork advisory dispositions cannot waive
them.

Routine cycle authorization does not replace release voting. Go and CLI workers
first open dependency-only backport PRs to the stable release branch. After
ordinary successful CI and required branch reviews, the worker merges the exact
backport with a normal merge commit. The worker then opens a separate signed
single-commit release-voting PR. Go changes only `signer/signer.go`'s
`signingAgent`; CLI changes only `Version` and `BuildMetadata` in
`internal/version/version.go`.

The voting PR must receive current-head approval from a strict majority of the
active repo-level maintainers in `MAINTAINERS` at the pinned assessed `main`.
Dismissed or stale approvals do not count. Changes-requested reviews block the
vote. The worker requires ordinary CI and branch reviews and uses a normal
merge commit, never an administrative merge. Core has no inferred version edit
or mandatory majority vote; its backport still follows ordinary branch rules.

Go and CLI qualify and tag the recorded signed voting commit. The release
branch remains at the verified normal merge tip. The worker checks the voted
tree, merge parents, signatures, majority, and immutable receipt again before
tagging and publication.

## Repository setup

Maintainers must install the shared worker, scripts, tests, and validation
workflow in all three default `main` branches. The CLI additionally installs the
agent, compiled lock, controller, policy, and CLI packaging configuration.
No product dependency or toolchain upgrade is required by this installation.

| Location | Required configuration |
|---|---|
| All repos, **Settings > General** | Enable issues for controller and worker cycle records. Enable **Allow merge commits** for backport and voting PRs. |
| All repos, **Settings > Actions > General** | Enable the installed pinned Actions and reusable tooling. The workflows use read-only `GITHUB_TOKEN` permissions. |
| All repos, **Settings > Rules > Rulesets**, or **Settings > Branches** | Protect `main` and the latest stable `release-MAJOR.MINOR` branch with the normal required checks and reviews. Do not give the release actor bypass rights. |
| All repos, **Settings > Advanced Security** (or the repository's security settings) | Enable Dependabot alerts and security updates. Review `.github/dependabot.yml` for Go modules and GitHub Actions on `main`; CLI includes root, E2E, and plugin modules. |
| All repos, **Settings > Secrets and variables > Actions > Secrets** | Add the same `MONTHLY_PATCH_TOKEN`, a dedicated authenticated-user token scoped to the three repositories with Contents, Pull requests, Issues, and Actions read/write. Explicitly approved workflow updates also need Workflows write permission. The token must resolve to `MONTHLY_PATCH_ACTOR` through `GET /user`; GitHub App installation tokens are not supported by this draft's credential check. |
| All repos, **Settings > Secrets and variables > Actions > Secrets** | Add `MONTHLY_PATCH_SIGNING_KEY`, a dedicated unencrypted SSH private signing key. Register its public key under the signer account's **Settings > SSH and GPG keys > New SSH key > Signing key**. |
| All repos, **Settings > Secrets and variables > Actions > Variables** | Set `MONTHLY_PATCH_ACTOR`, `MONTHLY_PATCH_SIGNER_LOGIN`, `MONTHLY_PATCH_SIGNER_EMAIL`, and optionally `MONTHLY_PATCH_SIGNER_NAME`. Use the same release actor in every repo. |
| CLI only, **Settings > Secrets and variables > Actions > Secrets** | Add separate `COPILOT_GITHUB_TOKEN` with **Copilot Requests: Read** and access to the Copilot engine. The agent receives no release token or signing key. |
| CLI only, reviewed PRs on `main` and the stable release branch | Install the `MONTHLY_PATCH_ACTOR` exclusion in `release-github.yml` on both branches. The worker inspects the release branch guard before preparing a publication, preventing a second tag-triggered publisher. |

The stable baseline must be a verified signed annotated tag. The release branch
must already exist; the worker does not create or repair stable release lines.

## Activation and recovery

Every switch is false when unset. Maintainers can manually dispatch the agent
on `main` for read-only assessment, then dispatch the controller with
`assessment_run=<successful run ID>,mode=dry-run` to inspect the validated plan.
Dry runs receive no release secrets and create no cycle issues or release PRs.

Only after setup and policy approval, maintainers set
`MONTHLY_PATCH_ENABLED=true` in all three repositories. In CLI they separately
set `NOTATION_RELEASE_AGENT_ENABLED=true`,
`NOTATION_RELEASE_COORDINATOR_ENABLED=true`, and
`NOTATION_RELEASE_AUTO_APPROVAL_ENABLED=true`. The first switch enables the
monthly assessment at `0 9 1 * *`; the coordinator resumes at `0 */6 * * *`.
The automatic-approval switch authorizes new routine cycles only; existing
authorized cycles retain their approvals. Clearing `MONTHLY_PATCH_ENABLED`
stops writing worker stages.

The controller records exact PR heads, stable baselines, source checkpoints,
automation tree fingerprints, workflow identities, and dispatch attempts.
On canonical `main`, only recorded worker dependency merges can advance the
source checkpoint without invalidating the assessment. Changed automation,
governance, policy, or unrecorded source changes require reassessment. A changed
maintainer roster requires a new release vote, not reuse of retired members'
approvals. Consumers wait for the exact planned, signed, publicly verified producer version and a
qualifying Dependabot update; absence is recorded as waiting, not success.

For a failed attempt, the configured actor manually dispatches the controller
with `retry_failed=true`. A changed assessment requires
`assessment_run=<run>,approve=true,replan=true`; replanning retains reserved
versions and refuses in-flight workers or replacement public candidates.
Verification-only recovery additionally requires
`recover_public_verification=true` and preserves the original tag, candidate,
and asset hashes. Ambiguous dispatches, conflicting PR branches, or duplicate
records require explicit reconciliation and are never blindly repeated.

Source tests, packaging, the agent, and downloaded-package tests run without
release secrets. A separate trusted-default-main publisher signs and publishes
only after qualification. It uses exact branch/tag leases, never moves existing
tags, and verifies the signed tag and public asset hashes. Successful native
downloaded-package checks are required before producer completion.

## Local validation

The shared regression suite preserves fork authorization behavior separately
from canonical tests:

```sh
GITHUB_REPOSITORY=yizha1/notation python3 -m unittest discover \
  -s .github/scripts -p 'test_*release*.py'
GITHUB_REPOSITORY=notaryproject/notation python3 -m unittest discover \
  -s .github/scripts -p 'test_upstream_authorization.py'
```

The validation workflow also runs pinned `actionlint v1.7.12`. Compile the
assessment lock with existing pinned `gh-aw v0.89.21`; do not edit it by hand:

```sh
gh aw compile notation-release-agent --strict \
  --gh-aw-ref c35393777e5604a63721d09512263b1383301d4f --no-check-update
```

Local fixture tests do not prove repository credential permissions, live
Dependabot propagation, production release votes, or platform-specific public
package checks. Maintainers must validate those settings and gates before
production activation. This draft has not dispatched or published an upstream
release.
