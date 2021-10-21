# Release Checklist

## Overview

This document describes the checklist to publish a release via GitHub workflow.

## Release Process

1. Determine a [SemVer2](https://semver.org/)-valid version prefixed with the letter `v` for release. For example, `version="v1.0.0-alpha.1"`.
2. Bump up the `Version` in [internal/version/version.go](internal/version/version.go#L5) and open a PR for the changes.
3. Wait for the PR merge.
4. Generate a GitHub [personal access token (PAT)](https://github.com/settings/tokens/new) with the `repo:public_repo` permission.
5. In the repository `Settings -> Secrets`, create or update the repository secret `RELEASE_GITHUB_USER_TOKEN` with the PAT generated above.
6. Make a fresh clone of the repository, check the `git log`, and create a tag by `git tag $version`.
7. After double checking the digest of the tag, push the tag directly to the repository by `git push origin $version`.
8. Wait for the completion of the GitHub action `release-github`.
9. Revoke the PAT generated previously.
10. Delete or update the repository secret `RELEASE_GITHUB_USER_TOKEN` with a dummy value.
11. Check the new release, and revise the release description.
12. Announce the release in the community.
