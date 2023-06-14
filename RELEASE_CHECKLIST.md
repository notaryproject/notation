# Release Checklist

## Overview

This document describes the checklist to publish a release via GitHub workflow.

## Release Process

1. Check if there are any security vulnerabilities fixed and security advisories published before a release. Security advisories should be linked on the release notes
1. Determine a [SemVer2](https://semver.org/)-valid version prefixed with the letter `v` for release. For example, `version="v1.0.0-alpha.1"`.
1. If there are code changes in notation-go or notation-core-go library, follow the two steps above and cut the release for the library. Then submit a PR to update the dependency versions in notation [go.mod](go.mod) and [go.sum](go.sum). Run `go mod tidy` to ensure the `go.sum` file is also updated with any potential changes.
1. Bump up the `Version` in [internal/version/version.go](internal/version/version.go#L5) and open a PR for the changes in notation repository. 
1. Create an issue to vote for the new release. Add the link of change logs and determine a specific git commit to tag on the issue description. List all repo-level maintainers and make sure have majority of approvals from the maintainers before releasing it. 
1. After the version and dependencies are updated, be on the `main` branch of the notation repository (not a fork) and execute `git clone https://github.com/notaryproject/notation.git`. 
1. Enter the cloned repository and execute `git checkout <commit_digest>` to switch to the specified branch based on the voting result.
1. Create a tag by running `git tag -s $version`.
1. Run `git tag` and ensure the desired tag name in the list looks correct, then push the new tag directly to the repository by running `git push $version`.
1. Wait for the completion of the GitHub action [release-github](https://github.com/notaryproject/notation/actions/workflows/release-github.yml).
1. Check the new draft release, revise the release description, and publish the release.
1. Update the necessary documentation in the [notaryproject.dev](https://github.com/notaryproject/notaryproject.dev) repository to reflect the changes of the release on the Notary website, includes but not limited to [installation guide](https://github.com/notaryproject/notaryproject.dev/blob/main/content/en/docs/installation/cli.md), [user guide](https://github.com/notaryproject/notaryproject.dev/tree/main/content/en/docs/how-to), [banner](https://github.com/notaryproject/notaryproject.dev/blob/main/layouts/partials/banner.html), [release blog](https://github.com/notaryproject/notaryproject.dev/tree/main/content/en/blog).
1. Announce the release in the community.
