# Proposal: automated monthly Notation patch releases

We propose monthly patch releases across **`notation-core-go`, `notation-go`,
and `notation` (CLI)**, automated under a maintainer-approved policy.

This draft implements canonical execution and policy-based cycle authorization.
All schedules and publishing remain disabled until maintainers configure and
explicitly activate them. Local tests do not establish production readiness.

## 1. Release scope

Each patch targets the repo's **latest stable release line**, with applicable
dependency changes backported from `main`. Versions remain independent.

The following scope applies to all three repos:

| Change | Monthly patch scope |
|---|---|
| Dependabot PRs | Included |
| Additional CVE fixes | Optional, outside automatic authority in this draft |
| Bug-fix backports | Optional, outside automatic authority in this draft |
| Features and breaking changes | Excluded |
| Older release lines | Excluded |

The following table covers only cross-repository prerequisites.
Each repo must also satisfy its own release checks.

| Repository | Cross-repository release prerequisite |
|---|---|
| Core | None among these three repos |
| Go | Wait for Core if Core is releasing in the same cycle |
| CLI | Wait for Core and Go if they are releasing in the same cycle |

Each repo can release its own dependency updates independently. Repos without updates or a planned library version to consume are skipped.

The implemented routine policy permits nondecreasing Go dependency patches on
the same stable major/minor, with unchanged Go and module requirements.
GitHub Actions updates, broader dependency upgrades, and replans require an
explicit manual dispatch by the configured release actor. Required producer
updates must consume the exact planned canonical versions.

## 2. Release solution

One monthly trigger starts a shared release plan.

| Component | Uses AI? | Responsibility |
|---|---|---|
| Assessment agent | Yes | Assesses all three repos and proposes release, skip, or defer |
| Release controller | No | Authorizes the plan under policy, enforces order, and tracks progress |
| Release workers | No | Merge qualified updates, run checks, sign, publish, and verify |

1. The agent assesses Dependabot PRs, stable baselines, manifests, and checks.
2. The controller validates and authorizes the plan, recording exact inputs
   and versions in a tracking issue.
3. Workers release producers first. Consumers adopt the exact verified library
   versions through checked Dependabot PRs before releasing. Workers open
   dependency-only backport PRs to the stable release branch, then separate
   release-voting PRs where the repository checklist requires them.
4. The controller closes the cycle after public-package verification.

Producer failures block consumers. Callbacks and reconciliation resume the
same cycle without duplicate releases. Later updates enter a new cycle.

## 3. Upstream configuration and maintainer approval

The central agent and controller belong in **`notaryproject/notation`**.

| Repository | Changes required |
|---|---|
| `notation-core-go` | Thin dispatch caller pinned to the central worker's full commit SHA and local setup documentation |
| `notation-go` | Thin pinned caller and local setup documentation; central gates verify adoption of the planned Core version |
| `notation` | Shared reusable worker, all helpers/tests, agent, controller, policy, packaging, and downloaded-CLI tests |

The following setup maps to the code in this draft. The central implementation
must merge into canonical CLI `main` before maintainers finalize and merge the
library callers. CLI uses a relative reusable-workflow call at the same commit.
Core and Go use the same full 40-character SHA in `uses` and `worker_sha`; they
do not copy helpers. Provisional draft pins must be replaced with the final
upstream implementation SHA before library merge or activation. The controller
checks canonical default-main ancestry, caller/tree identity, and the actual
run's called workflow path/SHA. A resolvable unmerged fork-network commit does
not qualify, and there is no fallback to `main`.

| Step | Where | Configuration | Maintainer approval |
|---|---|---|---|
| 1. Merge automation | CLI first, then libraries | Merge the central worker/helpers into CLI `main`, initially disabled. Replace library draft pins with its final upstream SHA, then merge only the thin callers/local setup docs. Use canonical module requirements, not trial `replace` directives. In CLI, guard `release-github.yml` on `main` and the selected release branch so it does not publish monthly-worker tags. | Required for the PRs, final pin review, and release-branch changes |
| 2. Configure repositories | All three repos | Enable Actions in **Settings > Actions > General**, allow the public `notaryproject/notation` reusable workflow at its reviewed SHA under **Actions permissions**, and enable issues in **Settings > General**. Maintain the latest stable `release-MAJOR.MINOR` branch and its required checks/reviews. Configure `.github/dependabot.yml` for Go modules and GitHub Actions on `main`; include CLI's root, E2E, and plugin modules. Enable Dependabot alerts and security updates in repository security settings. | Required for settings and configuration PRs |
| 3. Provision release credentials | All three repos | In **Settings > Secrets and variables > Actions > Secrets**, add the same `MONTHLY_PATCH_TOKEN`, an authenticated user token for one common release actor. Scope it to these three repos with contents, PRs, issues, and Actions write permissions; include workflow-write permission for explicitly approved workflow dependency updates. The controller verifies the token with `GET /user`; installation-token rotation is not implemented. Do not grant protection bypass. | Required for credential provisioning |
| 4. Configure signing and identity | All three repos | Add secret `MONTHLY_PATCH_SIGNING_KEY`. Register its public key as a GitHub signing key on the signer account. In the **Variables** tab, set `MONTHLY_PATCH_ACTOR`, `MONTHLY_PATCH_SIGNER_LOGIN`, and `MONTHLY_PATCH_SIGNER_EMAIL`; optionally set `MONTHLY_PATCH_SIGNER_NAME`. Use the same actor across all three repos. | Required for signer and actor setup |
| 5. Configure AI access | CLI repo only | Add Actions secret `COPILOT_GITHUB_TOKEN` with **Copilot Requests: Read** permission and access to the selected Copilot engine. Keep it separate from release credentials. | Required for account access and secret provisioning |
| 6. Approve automatic-release policy | CLI repo | Review `.github/monthly-patch-policy.yml`, including the implemented dependency compatibility rules, unchanged routine Go requirements, blocking reachable-vulnerability gates, and repository-specific release votes. `NOTATION_RELEASE_AUTO_APPROVAL_ENABLED` defaults to false when unset. A changed or unsupported policy fails closed. | Required for the policy and implementation PR |
| 7. Activate the cycle | All three repos, plus CLI settings | After approving the port and setup, set Actions variable `MONTHLY_PATCH_ENABLED=true` in all three repos. In the CLI, set `NOTATION_RELEASE_AGENT_ENABLED=true`, `NOTATION_RELEASE_COORDINATOR_ENABLED=true`, and `NOTATION_RELEASE_AUTO_APPROVAL_ENABLED=true`. The agent schedule is `0 9 1 * *` (monthly, 09:00 UTC); reconciliation is `0 */6 * * *`. Regenerate the agent lock when changing its source schedule. | Required for production activation |

During operation:

| Step | Workflow action | Maintainer approval |
|---|---|---|
| Assess and authorize a routine cycle | Validate the shared plan against the approved policy | No separate cycle approval |
| Merge dependency or producer-update PRs | Require successful checks and satisfied branch rules | Required wherever normal PR reviews are required |
| Merge release backport PRs | Require successful checks, satisfied branch rules, and exact candidate scope | Required wherever normal PR reviews are required |
| Vote on Go and CLI releases | Update only `signingAgent` or CLI version metadata in a separate, signed, single-commit PR, require current-head approval from a strict majority of the repo-level maintainers pinned at assessed `main`, and use a normal merge commit | Required by each repository's existing release checklist |
| Include optional CVE fixes or bug-fix backports | Keep these changes outside this draft's automatic authority | Separate reviewed implementation and scope approval are required |
| Handle a conflict, failed check, or blocking security finding | Stop the affected chain and record the blocker | Maintainer resolution and, where refs changed, explicit actor-approved reassessment before resuming; canonical advisory exceptions are not implemented |
| Publish and verify eligible patches | Execute the authorized plan in dependency order | No additional approval unless release-environment rules require it |

The agent is read-only. Source tests and the agent receive no release secrets.
Required reviews and release gates are not bypassed. Advisory exceptions are
not implemented.

Core has no repository-local release-voting requirement in the inspected
baseline. Its dependency backport PR still requires the repository's ordinary
branch checks and reviews. Go and CLI tag the signed voting PR commit recorded
by their checklists; the release branch remains at the verified normal merge
tip. Qualification tests and public-package metadata identify that exact voted
commit.
