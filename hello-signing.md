# Hello World for Notation

To get started with Notation, the most basic steps involve:

1. Acquire the Notation CLI
2. Generate a Private Key
3. Sign an Artifact
4. Verify an Artifact with a Notary Signature

This document outlines progressive scenarios enabling a range of scenarios from the first use to a production deployment.

## Scenarios

To demonstrate how to store and sign a set of supply chain artifacts, we will walk through a set of scenarios:
- Sign a single container image
- Publish the container image across public registries, 
- Import public content to a private registry
- Promote from dev through production

![](../media/notary-e2e-scenarios.svg)

- Wabbit Networks is a small software company that produces network monitoring software.
- ACME Rockets wishes to acquire network monitoring software.
- ACME Rockets doesn't know about Wabbit Networks, but finds their [net-monitor software in Docker Hub](https://hub.docker.com/r/wabbitnetworks/net-monitor)
- Since they've never heard of Wabbit Networks, they're a little reluctant to run network software without some validations.
- ACME Rockets has a policy to only import Docker Hub certified software, or other approved vendors.
- Wabbit Networks works with Docker Hub to get certified, to help with their customer confidence.
- ACME Rockets will only deploy software that's been scanned and approved by the ACME Rockets security team. They know it's been approved because all approved software has been signed by the ACME Rockets security team.

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
- Run a local instance of the [CNCF Distribution Registry][cncf-distribution]
  ```bash
  docker run -d -p ${PORT}:5000 ghcr.io/oras-project/registry:latest
  ```
- Acquire the Notation CLI  
Notation releases can be found at: [Notation Releases][notation-releases]  
  ```bash
  curl -O ~/bin/notation https://github.com/notaryproject/notation/releases/download/v0.0.1-alpha/notation-v0.0.1-alpha.1-linux-amd64
  curl -O ~/.docker/cli-plugins/docker-notation https://github.com/notaryproject/notation/releases/download/v0.0.1-alpha/docker-notation-v0.0.1-alpha.1-linux-amd64
  ```

## Building and Pushing
- Build and Push the `net-monitor` software
  ```bash
  docker build -t $IMAGE https://github.com/wabbit-networks/net-monitor.git#main

  docker push $IMAGE
  ```
- List the image, and any associated signatures
  ```bash
  notation list $IMAGE
  ```
  At this point, the results are empty, as there are no existing signatures

## Signing a Container Image

To get things started quickly, the notation cli supports generating self signed key-pairs. As you automate the signing of content, you will most likely want to create and store the private keys in a key vault. (Detailed production steps will be covered later)

- Generating a Private Key  
  The following will generate an X-509 private key under the `./notary/` directory
  ```bash
  notation cert generate "wabbit-networks.io"
  ```
- Sign the `net-monitor:v1` container image
  ```bash
  notation sign $IMAGE
  ```
- List the image, and any associated signatures
  ```bash
  notation list $IMAGE
  ```
## Verify a Container Image Using Notation Signatures

To avoid a Trojan Horse attack, before pulling an artifact into an environment, it is important to verify the integrity of the artifact based on the identity of an entity you trust. Notation uses a set of configured public keys to verify the content. The `notation cert generate` command automatically added the public key.
- Attempt to verify the $IMAGE notation signature
  ```bash
  notation verify $IMAGE
  ```
  *The above verification should fail, as you haven't yet configured the keys to trust.*
- To assure users opt-into the public keys they trust, add the key to the trusted store
  ```bash
  notation cert add "wabbit-networks.io" ~/.notary/wabbit-networks.io.cert
  ```
- Verify the `net-monitor:v1` notation signature
  ```bash
  notation verify $IMAGE
  ```
  The above now succeeds as the image is signed with a trusted public key

## Simulating a Registry DNS Name

Here are the additional steps to simulate a fully qualified dns name for wabbit-networks.

- Setup names and variables for `registry.wabbit-networks.io`
  ```bash
  export PORT=80
  export REGISTRY=registry.wabbit-networks.io
  export REPO=${REGISTRY}/net-monitor
  export IMAGE=${REPO}:v1
  ```
- Edit `~/.notary/notation.json` to support local, insecure registries
  ```json
  {
    "enabled": true,
    "verificationCerts": [
    ],
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

[notation-releases]:      https://github.com/notaryproject/notation/releases
[artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[cncf-distribution]:      https://github.com/oras-project/distribution