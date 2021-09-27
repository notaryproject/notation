# Validating Artifact Promotion with Signatures

When promoting artifacts across environments, signatures are one means to validate the content _should_ be promoted.

This demo will build upon the [Hello Signing quick-start](./hello-signing.md) demonstrating:
 - How artifacts may be verified before promoting
 - Promoting an artifact
 - Adding a signature, attesting to how the artifact was promoted into the target environment.

## Getting Started

Please [follow the common steps](./common-steps.md) to get started.
- Acquire the Notation CLI
- Generate a Private Key

## Scenario

In this scenario, the user will promote content from a simulated public registry to the ACME Rockets private registry. The same workflow may be applied to internal promotion from development to staging to production environments.

To demonstrate how to store and sign a graph of supply chain artifacts, the following steps will be completed:

1. Push and sign the `net-monitor:v1` container image to the **Wabbit Networks** public registry
2. Validate the `net-monitor:v1` image meets the acceptance criteria of ACME Rockets
3. Promote the `net-monitor:v1` image to the ACME Rockets registry
4. Sign the `net-monitor:v1` image with the **ACME Rockets library** key, indicating its valid for internal consumption

![](../media/notary-e2e-scenarios.svg)

## Getting Started
- Setup a few environment variables.  
  >Note see [Simulating a Registry DNS Name](#simulating-a-registry-dns-name) to use `registry.wabbit-networks.io`
  ```bash
  export PUBLIC_PORT=5000
  export PUBLIC_REGISTRY=localhost:${PUBLIC_PORT}
  export PUBLIC_REPO=${PUBLIC_REGISTRY}/net-monitor
  export PUBLIC_IMAGE=${PUBLIC_REPO}:v1

  export PRIVATE_PORT=5050
  export PRIVATE_REGISTRY=localhost:${PRIVATE_PORT}
  export PRIVATE_REPO=${PRIVATE_REGISTRY}/net-monitor
  export PRIVATE_IMAGE=${PRIVATE_PORT}:v1
  ```
- Install [Docker Desktop](https://www.docker.com/products/docker-desktop) for local docker operations
- Run a local registry representing the Wabbit Networks public registry
  ```bash
  docker run -d -p ${PUBLIC_PORT}:5000 ghcr.io/oras-project/registry:latest
  ```
- Run a local registry representing the ACME Rockets private registry
  ```bash
  docker run -d -p ${PRIVATE_PORT}:5000 ghcr.io/oras-project/registry:latest
  ```
## Building and Pushing the Public Image

- Build and push the `net-monitor` software
  ```bash
  docker build -t $PUBLIC_IMAGE https://github.com/wabbit-networks/net-monitor.git#main

  docker push $PUBLIC_IMAGE
  ```
- Generate a self-signed test certificate for signing artifacts
  The following will generate a self-signed X.509 certificate under the `~/config/notation/` directory
  ```bash
  notation cert generate-test "wabbit-networks.io"
  ```
- Sign the container image
  ```bash
  notation sign --plain-http $PUBLIC_IMAGE
  ```
- List the image, and any associated signatures
  ```bash
  notation list --plain-http $PUBLIC_IMAGE
  ```

## Import the Public Image

- Validate the image is signed with a key that fits within the ACME Rockets policy
 ```bash
 notation validate $PUBLIC_IMAGE
 ```
- The above command should fail, as the Wabbit Networks public key has not yet been configured
- Configure the Wabbit Networks key for validation, and re-validate
  ```bash
  notation cert add -n "wabbit-networks.io" /home/stevelas/.config/notation/certificate/wabbit-networks.io.crt
  notation verify $PUBLIC_IMAGE
  ``` 
> TODO: Promote with ORAS Copy

## Reset
To resetting the environment

- Remove keys, certificates and notation `config.json`  
  `rm -r ~/.config/notation/`
- Restart the local registry  
  `docker rm -f $(docker ps -q)`


[notation-releases]:      https://github.com/shizhMSFT/notation/releases/tag/v0.5.0
[artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md
[cncf-distribution]:      https://github.com/oras-project/distribution