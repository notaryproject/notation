# Release Checklist

## Overview

This document describes the checklist to publish a release for Notation CLI via GitHub workflow.

## Release Process

- Check if there are any security vulnerabilities fixed and security advisories published before a release. Security advisories should be linked on the release notes.
- Determine a [SemVer2](https://semver.org/)-valid version prefixed with the letter `v` for release. For example, `version="v1.0.0-alpha.1"`.
- If there is new release in [notation-go](https://github.com/notaryproject/notation-go) or [notation-core-go](https://github.com/notaryproject/notation-core-go) library that are required to be upgraded in Notation CLI, update the dependency versions in the follow `go.mod` and `go.sum` files of Notation CLI:
  - [go.mod](go.mod), [go.sum](go.sum)
  - [test/e2e/go.mod](test/e2e/go.mod), [test/e2e/go.sum](test/e2e/go.sum)
  - [test/e2e/plugin/go.mod](test/e2e/plugin/go.mod) and [test/e2e/plugin/go.sum](test/e2e/plugin/go.sum)
- Open a PR submit the changes in the previous step to the notation repository. Please make sure this PR is merged with all E2E test cases passed before starting the next step. See [PR #754](https://github.com/notaryproject/notation/pull/754) as an example.
- Create another PR to update the Notation CLI version with a single commit when PRs in above steps are merged. The commit message MUST follow the [conventional commit](https://www.conventionalcommits.org/en/v1.0.0/) and could be `bump: tag and release $version`. Record the digest of that commit as `<commit_digest>`. This PR is also used for voting purpose of the new release. Add the link of change logs and repo-level maintainer list in the PR's description. The PR title could be `bump: tag and release $version`. Make sure to reach a majority of approvals from the [repo-level maintainers](MAINTAINERS) before releasing it. This PR should be merged using [Create a merge commit](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/about-merge-methods-on-github) method in GitHub. See [PR #748](https://github.com/notaryproject/notation/pull/748) as an example. 
- After the voting PR is merged in the [Notation](https://github.com/notaryproject/notation.git) repository, execute `git clone git@github.com:notaryproject/notation.git` to clone the repository to your local file system.
- Enter the cloned repository and execute `git checkout <commit_digest>` to switch to the specified branch based on the voting result.
- Create a tag by running `git tag -am $version $version -s`.
- Run `git tag` and ensure the desired tag name in the list looks correct, then push the new tag directly to the repository by running `git push origin $version`.
- Wait for the completion of the GitHub action [release-github](https://github.com/notaryproject/notation/actions/workflows/release-github.yml).
- Check the new draft release, revise the release description, and publish the release.
- Update the necessary documentation in the [notaryproject.dev](https://github.com/notaryproject/notaryproject.dev) repository to reflect the changes of the release on the Notary Project website, includes but not limited to [installation guide](https://github.com/notaryproject/notaryproject.dev/blob/main/content/en/docs/installation/cli.md), [user guide](https://github.com/notaryproject/notaryproject.dev/tree/main/content/en/docs/how-to), [banner](https://github.com/notaryproject/notaryproject.dev/blob/main/layouts/partials/banner.html), [release blog](https://github.com/notaryproject/notaryproject.dev/tree/main/content/en/blog).
- Announce the new release in the Notary Project community.
