# Release Checklist

## Overview

This document describes the checklist to publish a release via GitHub workflow.

## Release Process

1. Determine a [SemVer2](https://semver.org/)-valid version prefixed with the letter `v` for release. For example, `version="v1.0.0-alpha.1"`.
2. Bump up the `Version` in [internal/version/version.go](internal/version/version.go#L5) and open a PR for the changes.
3. Wait for the PR merge.
4. Make a fresh clone of the repository, check the `git log`, and create a tag by `git tag $version`.
5. After double checking the digest of the tag, push the tag directly to the repository by `git push origin $version`.
6. Wait for the completion of the GitHub action `release-github`.
7. Check the new draft release, revise the release description, and publish the release.
8. Announce the release in the community.
