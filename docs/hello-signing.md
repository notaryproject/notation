# Hello World for Notation

To get started with Notation, the most basic steps involve:

- Acquire the Notation CLI
- Generate a Private Key
- Sign an Artifact
- Verify an Artifact with a Notary Signature

This document outlines a range of scenarios progressing from the first use of Notation to a production deployment.

## Scenarios

To demonstrate how to store and sign a set of supply chain artifacts, we will walk through a set of scenarios:

- Sign a single container image
- Publish the container image across public registries,
- Import public content to a private registry
- Promote from dev through production

![Notary end-to-end scenarios](../media/notary-e2e-scenarios.svg)

To illustrate these scenarios we will introduce two example companies:

- **Wabbit Networks**: A small software company that produces network monitoring software
- **ACME Rockets**: A consumer, who runs network monitoring software

ACME Rockets doesn't know about Wabbit Networks, but finds their [net-monitor software in Docker Hub](https://hub.docker.com/r/wabbitnetworks/net-monitor).
Since they've never heard of Wabbit Networks, they're a little reluctant to run network software without some validations.
ACME Rockets has a policy to only import Docker Hub certified content, and other approved vendors.

Wabbit Networks works with Docker Hub to get certified, to help with their customer confidence.

ACME Rockets will only deploy software that's been scanned and approved by the ACME Rockets security team. They know it's been approved because all approved software has been signed by the ACME Rockets security team.

## Getting Started

- Setup a few environment variables.
  >Note see [Simulating a Registry DNS Name](#simulating-a-registry-dns-name) to use `registry.wabbit-networks.io`

  ```bash
  export PORT=5000
  export REGISTRY=localhost:${PORT}
  export REPO=${REGISTRY}/net-monitor
  export IMAGE=${REPO}:v1
  ```

- Install [Docker Desktop](https://www.docker.com/products/docker-desktop) for local docker operations
- Run a local instance of the [CNCF Distribution Registry][cncf-distribution], with [ORAS Artifacts][artifact-manifest] support.
  ```bash
  docker run -d -p ${PORT}:5000 ghcr.io/oras-project/registry:v0.0.3-alpha
  ```

- Acquire the Notation CLI (Notation releases can be found at: [Notation Releases][notation-releases])

  ```bash
  #LINUX, including WSL
  curl -Lo notation.tar.gz https://github.com/notaryproject/notation/releases/download/v0.7.1-alpha.1/notation_0.7.1-alpha.1_linux_amd64.tar.gz
  tar xvzf notation.tar.gz -C ~/bin notation
  ```

## Building and Pushing

- Build and Push the `net-monitor` software

  ```bash
  docker build -t $IMAGE https://github.com/wabbit-networks/net-monitor.git#main

  docker push $IMAGE
  ```

- List the image, and any associated signatures

  ```bash
  notation list --plain-http $IMAGE
  ```

  At this point, the results are empty, as there are no existing signatures

## Signing a Container Image

To get things started quickly, the Notation cli supports generating self signed certificates. As you automate the signing of content, you will most likely want to create and store the private keys in a key vault. (Detailed production steps will be covered later)

- Generate a self-signed test certificate for signing artifacts
  The following will generate a self-signed X.509 certificate under the `~/config/notation/` directory

  ```bash
  notation cert generate-test --default "wabbit-networks.io"
  ```

- Sign the container image

  ```bash
  notation sign --plain-http $IMAGE
  ```

- List the image, and any associated signatures

  ```bash
  notation list --plain-http $IMAGE
  ```

## Verify a Container Image Using Notation Signatures

To avoid a Trojan Horse attack, and before pulling an artifact into an environment, it is important to verify that the artifact was unmodified after it was created (integrity), and from an trusted entity (authenticity). Notation uses a set of configured public keys that represent trusted entities, to verify the content. The `notation cert generate-test` command created the public key, however it must be implicitly added for verification to succeed.

- Attempt to verify the $IMAGE notation signature

  ```bash
  notation verify --plain-http $IMAGE
  ```

  *The above verification should fail, as you haven't yet configured the keys to trust.*

  ```bash
  2021/09/07 11:40:51 trust certificate not specified
  ```

- To assure users opt-into the public keys they trust, add the key to the trusted store

  ```bash
  notation cert add --name "wabbit-networks.io" ~/.config/notation/certificate/wabbit-networks.io.crt
  ```

- Verify the `net-monitor:v1` notation signature

  ```bash
  notation verify --plain-http $IMAGE
  ```

  This should now succeed because the image is signed with a trusted public key

## Reset

To resetting the environment

- Remove keys, certificates and notation `config.json`
  `rm -r ~/.config/notation/`
- Restart the local registry
  `docker rm -f $(docker ps -q)`

## Simulating a Registry DNS Name

Here are the additional steps to simulate a fully qualified DNS name for wabbit-networks.

- Setup names and variables for `registry.wabbit-networks.io`

  ```bash
  export PORT=80
  export REGISTRY=registry.wabbit-networks.io
  export REPO=${REGISTRY}/net-monitor
  export IMAGE=${REPO}:v1
  ```

- Edit `~/.config/notation/config.json` to support local, insecure registries

    ```json
    {
      "insecureRegistries": [
        "registry.wabbit-networks.io"
      ]
    }
    ```

- Add a `etc/hosts` entry to simulate pushing to registry.wabbit-networks.io
  - If running on windows, _even if using wsl_, add the following entry to: `C:\Windows\System32\drivers\etc\hosts`

    ```hosts
    127.0.0.1 registry.wabbit-networks.io
    ```

- Continue with [Getting Started](#getting-started), but skip the environment variable configurations

[notation-releases]:      https://github.com/notaryproject/notation/releases/tag/v0.7.1-alpha.1
[artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[cncf-distribution]:      https://github.com/oras-project/distribution
