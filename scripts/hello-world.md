# Hello World for Notation

To get started with Notation, the most basic steps involve:

1. Acquiring the Notation CLI
2. Generating a Private Key
3. Signing an Artifact
4. Pushing the Notation signature

## Getting Started
- Setup a few environment variables
  ```bash
  export PORT=5000
  export REGISTRY=localhost:${PORT}
  export REPO=${REGISTRY}/net-monitor
  export IMAGE=${REPO}:v1
  ```
- Run a Local Registry  
  To work locally, start a local instance of CNCF Distribution, with support of the [Artifact Manifest][artifact-manifest]
  ```bash
  docker run ghcr.io/oras-project/registry:v4.0.1-alpha
  ```
- Build and Push a Container Image
  ```bash
  docker build \
  -t $IMAGE \
  https://github.com/wabbit-networks/net-monitor.git#main

  docker push $IMAGE
  ```
- Acquiring the Notation CLI  
Notation releases can be found at: [Notation Releases][notation-releases]  
Notation can be installed by running the following script:
  ```bash
  curl https://raw.githubusercontent.com/notaryproject/notation/main/scripts/get-notation | sh
  ```
- Generating a Private Key  
  The following will generate an X-509 private key under the ./notation/ directory
  ```bash
  notation cert generate --type x-509 --subject "wabbit-networks"
  ```
- Sign the `net-monitor:v1` container image
  ```bash
  notation sign \
    --key wabbit-networks.key \
    --crt wabbit-networks.crt \
    --push \
    $IMAGE
  ```
- Validate the `net-monitor:v1` notation signature
  ```bash
  notation validate $IMAGE
  ```

[notation-releases]:      https://github.com/notaryproject/notation/releases
[artifact-manifest]:      https://github.com/oras-project/artifacts-spec/blob/main/artifact-manifest.md