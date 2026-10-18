# Proposal: automated monthly Notation patch releases

We propose monthly patch releases across **`notation-core-go`, `notation-go`,
and `notation` (CLI)**, automated under a maintainer-approved policy.

This draft requires maintainer approval of each plan and maintainer publication
of each qualified draft release. Automatic approval and publication are not implemented.
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
| Release controller | No | Opens the plan review issue, verifies maintainer approval, enforces order, and tracks progress |
| Release workers | No | Merge reviewed updates, run checks, sign tags, stage draft assets, and verify publication |

1. The agent assesses Dependabot PRs, stable baselines, manifests, and checks.
2. The controller opens a CLI plan issue showing each repository's baseline,
   proposed version, Dependabot PR/title, exact proposed HEAD and commits or
   already-merged source commit, and producer prerequisites. An active CLI
   repo-level maintainer reviews it and posts `/approve-plan <plan_id>`.
3. Workers prepare producers first. Consumers adopt the exact verified library
   versions through checked Dependabot PRs before releasing. Workers open
   dependency-only backport PRs to the stable release branch, then separate
   release-voting PRs where the repository checklist requires them.
4. The worker stages a true draft release and verifies its downloaded assets
   using read-only access. Only successful checks create a repository-local
   draft review issue with the exact release ID, commits, asset IDs/hashes, and
   evidence. A maintainer inspects it and uses GitHub's **Publish release**.
   The worker then verifies the same release and bytes anonymously before
   consumers proceed or the cycle closes.

| Gate | Maintainer action | What remains blocked |
|---|---|---|
| Shared plan issue | Post the exact approval command displayed in the issue | All worker execution until approval |
| Qualified draft review issue | Inspect evidence and publish the exact draft through GitHub Releases | Official release and consumers until manual publication and public verification |

An issue close, checklist, AI verdict, or unknown/comment-edited approval does
not pass either gate. Replans require new exact plan approval. Tags are signed
and public before draft review; draft visibility does not hide tags.

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
| 6. Review the manual-first policy | CLI repo | Review `.github/monthly-patch-policy.yml`, dependency compatibility, blocking reachable-vulnerability gates, both manual gates, and repository release votes. The legacy `NOTATION_RELEASE_AUTO_APPROVAL_ENABLED` variable is obsolete and cannot bypass plan approval, even if set. A changed or unsupported policy fails closed. | Required for the policy and implementation PR |
| 7. Activate staging and review | All three repos, plus CLI settings | After approving setup, set Actions variable `MONTHLY_PATCH_ENABLED=true` in all three repos. In CLI, set `NOTATION_RELEASE_AGENT_ENABLED=true` and `NOTATION_RELEASE_COORDINATOR_ENABLED=true`. The agent schedule is `0 9 1 * *`; reconciliation is `0 */6 * * *` and can also be dispatched manually. Maintainers still approve each plan and publish each qualified draft. | Required for production activation |

During operation:

| Step | Workflow action | Maintainer approval |
|---|---|---|
| Assess and review a cycle | Validate the shared plan and open its readable CLI issue | An active CLI repo-level maintainer must post the exact unedited `/approve-plan <plan_id>` command before execution |
| Merge dependency or producer-update PRs | Require successful checks and satisfied branch rules | Required wherever normal PR reviews are required |
| Merge release backport PRs | Require successful checks, satisfied branch rules, and exact candidate scope | Required wherever normal PR reviews are required |
| Vote on Go and CLI releases | Update only `signingAgent` or CLI version metadata in a separate, signed, single-commit PR, require current-head approval from a strict majority of the repo-level maintainers pinned at assessed `main`, and use a normal merge commit | Required by each repository's existing release checklist |
| Include optional CVE fixes or bug-fix backports | Keep these changes outside this draft's automatic authority | Separate reviewed implementation and scope approval are required |
| Handle a conflict, failed check, or blocking security finding | Stop the affected chain and record the blocker | Maintainer resolution and, where refs changed, explicit actor-approved reassessment before resuming; canonical advisory exceptions are not implemented |
| Review qualified draft releases | Record successful draft/native checks and open each repository's review issue | A maintainer inspects assets/evidence and uses GitHub Publish release; automation never publishes |
| Verify public patches | Compare the recorded release ID, tag, asset IDs, hashes, and sizes, then run anonymous public-package checks | Consumers stay blocked until checks pass |

The agent is read-only. Source tests and the agent receive no release secrets.
Required reviews and release gates are not bypassed. Advisory exceptions are
not implemented.

Core has no repository-local release-voting requirement in the inspected
baseline. Its dependency backport PR still requires the repository's ordinary
branch checks and reviews. Go and CLI tag the signed voting PR commit recorded
by their checklists; the release branch remains at the verified normal merge
tip. Qualification tests and public-package metadata identify that exact voted
commit.

Read-only draft asset access must succeed in the target repository with
`GITHUB_TOKEN` Contents read. Denied access blocks qualification; workers never
pass the release-write token to source or native tests. Future automatic
publication requires separate reviewed implementation and maintainer opt-in.
