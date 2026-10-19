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

The controller creates one release-scope issue per repository after assessment.
Each issue shows its own versions and Dependabot source revisions. Go links the
Core issue; CLI links Core and Go and shows overall status, without copying
their change lists. Source details and the central technical state are collapsed.
Metadata voting commits remain separate from Dependabot source commits.
Creation starts with CLI, then Core and Go, then wires exact issue links.
An interrupted creation resumes from authenticated placeholders. Duplicate,
missing, or mismatched issue records fail closed before execution.

An active CLI repo-level maintainer must post the exact, unedited command
displayed in the CLI issue: `/approve`. The controller polls every
six hours or on manual resume. It checks the current `MAINTAINERS` roster,
the issue's automation author/editor and exact rendered body, comment author,
comment identity/body/edit timestamps, policy, source heads, and immutable
worker pins before execution. It re-fetches the same approval on later writes.
Editing/deleting the comment revokes remaining execution; retired maintainers,
closing issues, checklists, unknown actors, and AI verdicts cannot approve.
Reassessment produces a new internal plan ID and review revision requiring a
fresh comment. The receipt binds the comment ID, plan ID, and review revision;
an older `/approve` cannot authorize a later plan.

The implemented routine policy permits stable, nondecreasing Go dependency
patches on the same major/minor, without changing module identities, dependency
membership, or Go requirements. Workflow dependency changes and broader updates
require a manual controller dispatch by `MONTHLY_PATCH_ACTOR` with
`approve=true` to request exceptional plan review; this input never bypasses
the maintainer comment gate. The policy does not authorize feature merges, source bug fixes,
or CVE backports. Reachable source and binary `govulncheck` findings block
canonical releases at every severity. Advisory dispositions cannot waive them.

The read-only collector also compares `main` against the selected release line
and supplies source/PR evidence for proposed bug-fix backports. The assessment
must explain applicability, known cherry-pick provenance or reverts, and missing
evidence. Titles or labels alone do not establish a bug fix. Candidates appear
only in their repository's issue, separate from the included Dependabot scope.
Proposal metadata cannot authorize a backport. Including one requires separately
reviewed execution support and fresh scope approval; new `main` commits never
silently join a frozen plan.

Plan approval does not replace release voting. Go and CLI workers
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

## Qualified draft review and manual publication

The isolated stager pushes the exact signed annotated tag with the existing
branch/tag leases, then creates a real GitHub draft release (`draft=true`,
`prerelease=false`) at that stable tag. The tag is already public. A draft does
not hide the tag or prevent module consumers from observing it.
For upstream use, a manual approval before the final stable tag remains required
but unimplemented. The existing draft-publication gate does not provide that
earlier boundary; this issue-topology change does not redesign publication.

The stager records the release ID and exact qualified asset names, IDs, sizes,
and SHA-256 hashes. Source, packaging, and native tests receive no release-write
credential. Native runners fetch the recorded draft assets with only target
`GITHUB_TOKEN` Contents read, validate bytes, and run the same package gates
before any qualified-review request. An asset API response may stream bytes
or redirect to signed GitHub storage; credentials are never forwarded to the
storage redirect. Denied read-only access blocks the gate without a write-token
fallback. Maintainers must validate live draft access before activation.
Interrupted uploads retain the reserved release ID and original hashes.
Retries may upload only missing identical bytes; a different rebuild is blocked,
never used to replace existing assets. Orphaned unrecorded drafts require
explicit reconciliation rather than adoption.

After every draft/native gate passes, the worker comments in the existing
repository release-scope issue. It lists actual
Dependabot source receipts, the backport candidate, signed voted HEAD and merge
tip, release ID, asset IDs/hashes/sizes, and source/scan/native evidence. Failed
drafts stay blocked and never receive a qualified-review request.

The maintainer inspects that issue and the exact GitHub draft, chooses **Edit**
in Releases, and chooses **Publish release** without changing its tag or assets.
An approval comment or issue close cannot publish it. This implementation never
turns a draft into a public release; future automatic publication is not
implemented and needs separate review and opt-in. The draft author's identity
is not used to infer who later published it.

On resume, the worker requires the same release ID, stable tag, signed/voted
commit, asset IDs, sizes, and hashes. A deleted/replaced release, changed assets,
or publication before qualification blocks the cycle. It then downloads public
assets anonymously and repeats native package verification. Only those passing
checks mark the producer published and unblock exact-version consumers.

## Failure notifications

Each repository's release issue is its human notification thread. When a worker
fails, the controller comments only in that issue and mentions that repository's
active repo-level maintainers. It includes the affected repo, failed stage when
available, workflow-run/attempt link, retained evidence location, and blocked
downstream releases. Go and CLI receive only concise links when blocked by a
dependency, not repeated dependency error details. Required check/review blockers
and qualified draft-ready requests use the same thread. Routine progress does
not send comments. Source tests, scans, packaging, draft checks, and public
verification remain blocking gates.

Comments are deduplicated by plan, repo, run, run attempt, and stage. Reconciliation
does not repeat an alert, but a newly failed retry gets a new comment.
Notifications do not rewrite the approved scope or count as approval.
Controller validation/execution failures also post to the existing release
issue. Notification errors fail visibly; they are not treated as delivery.

Maintainers should subscribe to the release issue and configure GitHub
notification/email preferences. Comments and mentions do not guarantee email
delivery, particularly when the automation uses the recipient's own account.
The issue and linked Actions run remain the audit trail.

If no trusted cycle issue exists yet, or checkout/authentication prevents the
notifier from running, GitHub's failed-workflow notification and logs are the
fallback. Bootstrap failures before trusted worker setup are reported when the
controller next reconciles. Neither notification nor retry bypasses a failed
gate or publishes a draft.

## Repository setup

The CLI owns `.github/workflows/monthly-patch-worker.yml`, all release scripts
and tests, the validation workflow, agent, compiled lock, controller, policy,
and packaging configuration. Each repository owns only its dispatch caller
`.github/workflows/monthly-patch-release.yml` and repository-specific setup
documentation. Core and Go do not copy release scripts or tests.
No product dependency or toolchain upgrade is required by this installation.

The CLI caller uses `./.github/workflows/monthly-patch-worker.yml` and passes
`worker_sha: ${{ github.sha }}`, selecting the same commit without a circular
self-SHA literal. Each library caller must use
`notaryproject/notation/.github/workflows/monthly-patch-worker.yml@FULL40SHA`
and pass that identical literal SHA as `worker_sha`. Moving tags, `main`,
different helper pins, arbitrary repositories, and missing pins fail closed.
There is no fallback to the latest shared code.

Maintainers first commit the central implementation and read its exact commit
SHA, then create separate library caller commits with that SHA. A draft can
contain this provisional, unmerged SHA, but library PRs must not merge until the
central CLI PR merges. Maintainers must replace provisional pins with the final
upstream merged implementation SHA before merging either library PR or activating
the system. The controller requires the revision to be reachable from canonical
CLI default-main history and requires unchanged assessed shared automation.
Resolving an unmerged commit through GitHub's fork network is not authorization.

| Location | Required configuration |
|---|---|
| All repos, **Settings > General** | Enable issues for controller and worker cycle records. Enable **Allow merge commits** for backport and voting PRs. |
| All repos, **Settings > Actions > General > Actions permissions** | Permit the pinned Actions and the public reusable workflow from `notaryproject/notation`. If an allow-list is configured, allow that repository's reusable workflow at the reviewed full SHA. Caller permissions include Contents, Actions, Issues, and Pull requests read; the called worker cannot elevate them. |
| All repos, **Settings > Rules > Rulesets**, or **Settings > Branches** | Protect `main` and the latest stable `release-MAJOR.MINOR` branch with the normal required checks and reviews. Ensure normal CI also targets stable-branch backport PRs. Missing checks block execution; do not give the release actor bypass rights. |
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
and `NOTATION_RELEASE_COORDINATOR_ENABLED=true`. The first switch enables the
monthly assessment at `0 9 1 * *`; the coordinator resumes at `0 */6 * * *`.
The legacy `NOTATION_RELEASE_AUTO_APPROVAL_ENABLED` setting is obsolete and has
no approval authority, even when true. Every new/replacement plan needs the
maintainer issue comment. Clearing `MONTHLY_PATCH_ENABLED`
stops writing worker stages.

The controller records exact PR heads, stable baselines, source checkpoints,
caller automation commits/tree fingerprints, the central worker repository,
path, commit/tree fingerprint, workflow identities, and dispatch attempts.
GitHub's reusable context identifies the caller repository, ref, and SHA. The
worker checks the actual run's `referenced_workflows` entry against the approved
central path and SHA before checking out any helper or using release credentials.
Every automation checkout explicitly selects `notaryproject/notation` at that
validated revision; target source and candidate checkouts remain separate.
Progress callbacks and producer completion revalidate the same issued receipt
and actual called revision. A worker-input SHA alone cannot authorize execution.
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
release ID, asset IDs, and hashes. It can reverify an unchanged draft or public
candidate, never rebuild, retag, or republish it. The replacement plan still
requires a fresh maintainer comment. Ambiguous dispatches, conflicting PR branches, or duplicate
records require explicit reconciliation and are never blindly repeated.

Source tests, packaging, the agent, and downloaded-package tests run without
release secrets. A separate trusted-default-main stager signs and creates drafts
only after qualification. It never moves tags or publishes official releases.

The read-only format preview uses the same scope renderer without enabling fork
execution or creating an approval record:

```sh
python3 .github/scripts/notation_release_controller.py preview \
  --output /tmp/notation-plan-preview
```

It reads live canonical Dependabot metadata, not qualification results. An
optional `--inventory <captured-preview.json>` uses a previously captured
read-only preview inventory. A fork issue containing the resulting
`plan-preview.md` cannot authorize canonical execution.

## Local validation

The central regression suite covers canonical authorization, signing, source
and public-package gates, release votes, caller pins, and reusable-run provenance:

```sh
python3 -m unittest discover \
  -s .github/scripts -p 'test_*release*.py'
python3 -m unittest discover \
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
