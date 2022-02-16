# Notation

Notation is a project to add signatures as standard items in the registry ecosystem, and to build a set of simple tooling for signing and verifying these signatures. This should be viewed as similar security to checking git commit signatures, although the signatures are generic and can be used for additional purposes. Notation is an implementation of the [Notary V2 specifications][notaryv2-specs].

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
    curl -Lo notation.tar.gz https://github.com/notaryproject/notation/releases/download/v0.7.1-alpha.1/notation_0.7.1-alpha.1_linux_amd64.tar.gz
    tar xvzf notation.tar.gz -C ~/bin notation
    ```
- Run a local instance of the [CNCF Distribution Registry][cncf-distribution], with [ORAS Artifacts][artifact-manifest] support.
  ```bash
  docker run -d -p 5000:5000 ghcr.io/oras-project/registry:v0.0.3-alpha
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

For more detailed samples, see [hello-signing](docs/hello-signing.md)

## Core Documents

- [Governance for Notation](https://github.com/notaryproject/notary/blob/master/GOVERNANCE.md)
- [Maintainers and reviewers list](https://github.com/notaryproject/notary/blob/master/MAINTAINERS)

## Community

- Regular conversations for Notation occur on the [Cloud Native Computing Slack](https://app.slack.com/client/T08PSQ7BQ/CQUH8U287?) channel.
- Please see the [CNCF Calendar](https://www.cncf.io/calendar/) for community meeting details.
- Meeting notes are captured on [hackmd.io](https://hackmd.io/_vrqBGAOSUC_VWvFzWruZw).

## Release Management

The Notation release process is defined in [RELEASE_MANAGEMENT.md](RELEASE_MANAGEMENT.md#supported-releases).

## Support

Support for the Notation project is defined in [supported releases](RELEASE_MANAGEMENT.md#supported-releases).

## Code of Conduct

This project has adopted the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for further details.

## License

This project is covered under the Apache 2.0 license. You can read the license [here](LICENSE).

[notation-releases]:      https://github.com/notaryproject/notation/releases/tag/v0.7.1-alpha.1
[notaryv2-specs]:         https://github.com/notaryproject/notaryproject
[artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[cncf-distribution]:      https://github.com/oras-project/distribution
