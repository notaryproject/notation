# Notation

[![codecov](https://codecov.io/gh/notaryproject/notation/branch/main/graph/badge.svg)](https://codecov.io/gh/notaryproject/notation)

Notation is a CLI project to add signatures as standard items in the registry ecosystem, and to build a set of simple tooling for signing and verifying these signatures. This should be viewed as similar security to checking git commit signatures, although the signatures are generic and can be used for additional purposes. Notation is an implementation of the [Notary V2 specifications][notaryv2-specs].

## Table of Contents

- [Notation Quick Start](#notation-quick-start)
- [Contributing](#contributing)
- [Core Documents](#core-documents)
- [Community](#community)
- [Release Management](#release-management)
- [Support](#support)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

## Notation Quick Start

- Install the Notation CLI from [Notation Releases][notation-releases]
    ```bash
    curl -Lo notation.tar.gz https://github.com/notaryproject/notation/releases/download/v0.10.0-alpha.3/notation_0.10.0-alpha.3_linux_amd64.tar.gz
    tar xvzf notation.tar.gz -C ~/bin notation
    ```
- Run a local instance of the [CNCF Distribution Registry][cncf-distribution], with [ORAS Artifacts][artifact-manifest] support.
  ```bash
  docker run -d -p 5000:5000 ghcr.io/oras-project/registry:v1.0.0-rc
  ```

- Build, Push, Sign, Verify the `net-monitor` software

  ```bash
  export IMAGE=localhost:5000/net-monitor:v1
  docker build -t $IMAGE https://github.com/wabbit-networks/net-monitor.git#main
  docker push $IMAGE
  notation cert generate-test --default --trust "wabbit-networks-dev"
  notation sign --plain-http $IMAGE
  notation list --plain-http $IMAGE
  notation verify --plain-http $IMAGE
  ```

Signatures are persisted as [ORAS Artifacts manifests][artifact-manifest].


## Documents

- [Hello World for Notation: Local signing and verification](docs/hello-signing.md)
- [Build, sign, and verify container images using Notary and Azure Key Vault](https://docs.microsoft.com/azure/container-registry/container-registry-tutorial-sign-build-push)


## Community

### Development and Contributing

- [Build Notation from source code](/building.md)
- [Governance for Notation](https://github.com/notaryproject/notary/blob/master/GOVERNANCE.md)
- [Maintainers and reviewers list](https://github.com/notaryproject/notary/blob/master/MAINTAINERS)
- Regular conversations for Notation occur on the [Cloud Native Computing Slack](https://slack.cncf.io/) **notary-v2** channel.

### Notary v2 Community Meeting

- Mondays 5-6pm pacific time, 8-9pm US Eastern, 8-9am Shanghai
- Thursdays 9-10am pacific time, 12pm US Eastern, 5pm UK

Join us at [Zoom Dial-in link](https://zoom.us/my/cncfnotaryproject) / Passcode: 77777. Please see the [CNCF Calendar](https://www.cncf.io/calendar/) for community meeting details. Meeting notes are captured on [hackmd.io](https://hackmd.io/_vrqBGAOSUC_VWvFzWruZw).

## Release Management

The Notation release process is defined in [RELEASE_MANAGEMENT.md](RELEASE_MANAGEMENT.md#supported-releases).

## Support

Support for the Notation project is defined in [supported releases](RELEASE_MANAGEMENT.md#supported-releases).

## Code of Conduct

This project has adopted the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for further details.

## License

This project is covered under the Apache 2.0 license. You can read the license [here](LICENSE).

[Notation Releases]:      https://github.com/notaryproject/notation/releases
[Notary v2 Specs]:         https://github.com/notaryproject/notaryproject
[ORAS artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[CNCF Distribution]:      https://github.com/oras-project/distribution
