# Notation

Notation is a project to add signatures as standard items in the registry ecosystem, and to build a set of simple tooling for signing and verifying these signatures. This should be viewed as similar in security to checking git commit signatures, although the signatures are generic and can be used for additional purposes.

## Table of Contents

- [Notation Quick Start](#notation-quick-start)
- [Branch](#branch)
- [Core Documents](#core-documents)
- [Community](#community)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

## Notation Quick Start

- Install the Notation CLI from [Notation Releases][notation-releases]  
  ```bash
  curl -Lo notation.tar.gz https://github.com/shizhMSFT/notation/releases/download/v0.5.2/notation_0.5.2_linux_amd64.tar.gz
  tar xvzf notation.tar.gz -C ~/bin notation
  ```
- Build, Push, Sign, Verify the `net-monitor` software
  ```bash
  export IMAGE=localhost:5000/net-monitor:v1
  docker build -t $IMAGE https://github.com/wabbit-networks/net-monitor.git#main
  docker push $IMAGE
  notation cert generate-test --default --trust "wabbit-networks-dev"
  notation sign $IMAGE
  notation list $IMAGE 
  notation verify $IMAGE
  ```

## Branch

[Prototype 2][prototype-2] - signing and verifying OCI artifacts, using signatures persisted [ORAS Artifacts manifests][artifact-manifest]

[artifact-manifest]:  https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[prototype-2]:      https://github.com/notaryproject/notation/tree/prototype-2

## Core Documents

* [Governance for Notation](https://github.com/notaryproject/notary/blob/master/GOVERNANCE.md)
* [Maintainers and reviewers list](https://github.com/notaryproject/notary/blob/master/MAINTAINERS)

## Community

* Regular conversations for Notation occur on the [Cloud Native Computing Slack](https://app.slack.com/client/T08PSQ7BQ/CQUH8U287?) channel.

* Please see the [CNCF Calendar](https://www.cncf.io/calendar/) for community meeting details.

* Meeting notes are captured on [hackmd.io](https://hackmd.io/_vrqBGAOSUC_VWvFzWruZw).

## Code of Conduct

This project has adopted the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for further details.

## License

This project is covered under the Apache 2.0 license. You can read the license [here](LICENSE).

[notation-releases]:      https://github.com/shizhMSFT/notation/releases/tag/v0.5.0
[artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[cncf-distribution]:      https://github.com/oras-project/distribution