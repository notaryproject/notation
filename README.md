# Debug 12312
# Notation

[![Go Report Card](https://goreportcard.com/badge/github.com/notaryproject/notation)](https://goreportcard.com/report/github.com/notaryproject/notation)
[![codecov](https://codecov.io/gh/notaryproject/notation/branch/main/graph/badge.svg)](https://codecov.io/gh/notaryproject/notation)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/notaryproject/notation/badge)](https://api.securityscorecards.dev/projects/github.com/notaryproject/notation)

Notation is a CLI project to add signatures as standard items in the registry ecosystem, and to build a set of simple tooling for signing and verifying these signatures. This should be viewed as similar security to checking git commit signatures, although the signatures are generic and can be used for additional purposes. Notation is an implementation of the [Notary project specifications][notaryv2-specs].

> **Note** The documentation for using `notation` is available [here](https://notaryproject.dev/docs/).

## Table of Contents

  - [Documents](#documents)
  - [Community](#community)
    - [Development and Contributing](#development-and-contributing)
    - [Notary project Community Meeting](#notary-project-community-meeting)
  - [Release Management](#release-management)
  - [Support](#support)
  - [Code of Conduct](#code-of-conduct)
  - [License](#license)

## Documents

- [Quick start: Sign and validate a container image](https://notaryproject.dev/docs/quickstart/)
- Build, sign, and verify container images using Notation with [Azure Key Vault](https://docs.microsoft.com/azure/container-registry/container-registry-tutorial-sign-build-push?wt.mc_id=azurelearn_inproduct_oss_notaryproject) or [AWS Signer](https://docs.aws.amazon.com/signer/latest/developerguide/container-workflow.html)
## Community

### Development and Contributing

- [Build Notation from source code](/building.md)
- [Governance for Notation](https://github.com/notaryproject/.github/blob/master/GOVERNANCE.md)
- [Maintainers and reviewers list](https://github.com/notaryproject/notation/blob/main/CODEOWNERS)
- Regular conversations for Notation occur on the [Cloud Native Computing Slack](https://slack.cncf.io/) **notary-project** channel.

### Notary project Community Meeting

- Mondays 5-6 PM PDT, 4-5 PM PST, 8-9 PM EDT, 7-8 PM EST, 8-9 AM Shanghai
- Thursdays 9-10 AM PDT, 8-9 AM PST, 12 PM EDT, 11 AM EST, 5 PM UK

Join us at [Zoom Dial-in link](https://zoom.us/my/cncfnotaryproject) / Passcode: 77777. Please see the [CNCF Calendar](https://www.cncf.io/calendar/) for community meeting details. Meeting notes are captured on [hackmd.io](https://hackmd.io/_vrqBGAOSUC_VWvFzWruZw).

## Release Management

The Notation release process is defined in [RELEASE_MANAGEMENT.md](RELEASE_MANAGEMENT.md#supported-releases).

## Support

Support for the Notation project is defined in [supported releases](RELEASE_MANAGEMENT.md#supported-releases).

## Code of Conduct

This project has adopted the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for further details.

## License

This project is covered under the Apache 2.0 license. You can read the license [here](LICENSE).

[notation-releases]:      https://github.com/notaryproject/notation/releases
[notaryv2-specs]:         https://github.com/notaryproject/notaryproject
[artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[cncf-distribution]:      https://github.com/oras-project/distribution
