# Release Checklist

## Overview

This document describes the checklist to publish a release via GitHub workflow.

## Release Process

1. Check if there are any security vulnerabilities fixed and security advisories published before a release. Security advisories should be linked on the release notes
1. Determine a [SemVer2](https://semver.org/)-valid version prefixed with the letter `v` for release. For example, `version="v1.0.0-alpha.1"`.
1. Create an issue to vote for the new release. Add the link of change logs and determine a specific git commit to tag on the issue description. List all repo-level maintainers and make sure have majority of approvals from the maintainers before releasing it. 
1. If there are code changes in notation-go or notation-core-go library, follow the three steps above and cut the release for the library. Then submit a PR to update the dependency versions in notation [go.mod](go.mod) and [go.sum](go.sum). Run `go mod tidy` to ensure the `go.sum` file is also updated with any potential changes.
1. Bump up the `Version` in [internal/version/version.go](internal/version/version.go#L5) and open a PR for the changes in notation repository. 
1. After the version and dependencies are updated, be on the `main` branch of the notation repository (not a fork) and execute `git pull`. 
1. Run `git log -1` to show the latest commit on the `main` branch and make sure you are on the up-to-date commit.
1. Create a tag by running `git tag -am $version $version`.
1. Run `git tag` and ensure the name in the list added looks correct, then push the tag directly to the repository by running `git push --follow-tags`.
1. Wait for the completion of the GitHub action [release-github](https://github.com/notaryproject/notation/actions/workflows/release-github.yml).
1. Check the new draft release, revise the release description, and publish the release.
1. Update the necessary documentation in the [notaryproject.dev](https://github.com/notaryproject/notaryproject.dev) repository to reflect the changes of the release on the Notary website, includes but not limited to installation guide, user guide, banner, release blog.
1. Announce the release in the community.
