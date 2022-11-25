# Release Checklist

## Overview

This document describes the checklist to publish a release via GitHub workflow.

NOTE: Make sure the dependencies in `go.mod` file are expected by the release. For example, if there are dependencies on certain version of notation library (notation-go or notation-core-go) or ORAS library (oras-go), make sure that version of library is released first, and the version number is updated accordingly in `go.mod` file.  After updating go.mod file, run `go mod tidy` to ensure the go.sum file is also updated with any potential changes.

## Release Process

1. Determine a [SemVer2](https://semver.org/)-valid version prefixed with the letter `v` for release. For example, `version="v1.0.0-alpha.1"`.
1. Bump up the `Version` in [internal/version/version.go](internal/version/version.go#L5) and open a PR for the changes.
1. Wait for the PR merge.
1. Be on the main branch connected to the actual repository (not a fork) and `git pull`.  Ensure `git log -1` shows the latest commit on the main branch.
1. Create a tag `git tag -am $version $version`
1. `git tag` and ensure the name in the list added looks correct, then push the tag directly to the repository by `git push --follow-tags`.
1. Wait for the completion of the GitHub action [release-github](https://github.com/notaryproject/notation/actions/workflows/release-github.yml).
1. Check the new draft release, revise the release description, and publish the release.
1. Announce the release in the community.
