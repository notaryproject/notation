# Demo Steps

The following demonstrates the capabilities of [Notary v2 - Prototype-2][nv2-prototype-2].

> At this point, this is a target experience, that is still being developed.

## Demo Setup

Perform the following steps prior to the demo:

- Install [Docker Desktop](https://www.docker.com/products/docker-desktop) for local docker operations
- [Install and Build the nv2 Prerequisites](./README.md#prerequisites)
- [Install and Build the ORAS Prototype-2 branch](https://github.com/deislabs/oras/blob/prototype-2/docs/artifact-manifest.md)
- Edit the `~/.docker/nv2.json` file to support local, insecured registries
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
- Create an empty working directory:
  ```bash
  mkdir nv2-demo
  cd nv2-demo/
  ```
- Generate the Wabbit Networks Public and Private Keys:
  ```bash
  openssl req \
    -x509 \
    -sha256 \
    -nodes \
    -newkey rsa:2048 \
    -days 365 \
    -subj "/CN=registry.wabbit-networks.io/O=wabbit-networks inc/C=US/ST=Washington/L=Seattle" \
    -addext "subjectAltName=DNS:registry.wabbit-networks.io" \
    -keyout ./wabbit-networks.key \
    -out ./wabbit-networks.crt
  ```
- Start a local registry instance:
  ```bash
  docker run -d -p 80:5000 --name oci-artifact-registry notaryv2/registry:nv2-prototype-2
  ```
- Add a `etc/hosts` entry to simulate pushing to registry.wabbit-networks.io
  - If running on windows, _even if using wsl_, add the following entry to: `C:\Windows\System32\drivers\etc\hosts`
    ```hosts
    127.0.0.1 registry.wabbit-networks.io
    ```

## Demo The End to End Experience

![](../../media/notary-e2e-scenarios.svg)

- Wabbit Networks is a small software company that produces network monitoring software.
- ACME Rockets wishes to acquire network monitoring software.
- ACME Rockets doesn't know about Wabbit Networks, but finds their [net-monitor software in Docker Hub](https://hub.docker.com/r/wabbitnetworks/net-monitor)
- Since they've never heard of Wabbit Networks, they're a little reluctant to run network software without some validations.
- ACME Rockets has a policy to only import Docker Hub certified software, or approved vendors.
- Wabbit Networks works with Docker Hub to get certified, to help with their customer confidence.
- ACME Rockets will only deploy software that's been scanned and approved by the ACME Rockets security team. They know it's been approved because all approved software has been signed by the ACME Rockets security team.

### Show Notary Extension

To demonstrate a clean end to end experience, using the docker cli, we're using the [docker-generate][docker-generate] extension.

- To see the extension capabilities, review the `Management Commands:`. They include some familiar ones, like buildx, and some new ones for the Notary v2 prototype:
  ```bash
  docker --help

  Management Commands:
    ...
    buildx*     Build with BuildKit (Docker Inc., v0.5.1-docker)
    generate*   Generate artifacts (github.com/shizhMSFT, 0.1.0)
    nv2*        Notary V2 Signature extension (Sajay Antony, Shiwei Zhang, 0.1.0)
  ```
- To see the sub commands of `nv2` and `nv2 notary`:
  ```bash
  docker nv2 --help
  docker nv2 notary --help
  ```
- To avoid having to type the fully qualified registry name, we'll create an environment variable:
  ```bash
  export REPO=registry.wabbit-networks.io/net-monitor
  export IMAGE=${REPO}:v1
  ```
  ```

### Intro `nv2` Commands

To achieve an proposed experience, an extension is added to implement Notary v2 signing and OCI Registry persistence and discovery of artifacts.

- View Docker Extensions
  ```bash
  docker --help
  ```
- Note the Management Commands with *. Also note the `nv2` extension
  ```bash
  Management Commands:
    app*        Docker App (Docker Inc., v0.9.1-beta3)
    buildx*     Build with BuildKit (Docker Inc., v0.5.1-docker)
    generate*   Generate artifacts (github.com/shizhMSFT, 0.1.0)
    nv2*        Notary V2 Signature extension (Sajay Antony, Shiwei Zhang, 0.2.3)
    scan*       Docker Scan (Docker Inc., v0.5.0)
  ```
  This command wraps `docker push`, _pushing both the image and its signature_, and `docker pull` to _verify signatures_ before pulling the image.
- View nv2 commands
  ```bash
  docker nv2 --help
  ```
- To avoid having to type `docker nv2` each time, we'll create an alias to mask over this:
  ```bash
  alias docker="docker nv2"
  ```

## Wabbit Networks Build, Sign, Promote Process

Let's walk through the sequence of operations Wabbit Networks takes to build, sign and promote their software.

Within the automation of Wabbit Networks, the following steps are completed:

- Building the `net-monitor` image
- Sign the `net-monitor` image
- Sign and Push the Image and Signature
- Create and Push an SBoM

### Build the `net-monitor` image

- Build the image, directly from GitHub to simplify the sequence.
  ```bash
  docker build \
    -t $IMAGE \
    https://github.com/wabbit-networks/net-monitor.git#main
  ```
  _To represent an ephemeral client in an air-gapped environment, git clone, then build with `.` as the context_

### Acquire the private key

- As a best practice, we'll always build on an ephemeral client, with no previous state.
- The ephemeral client will retrieve the private signing key from the companies secured key vault provider.

These specific steps are product/cloud specific, so we'll assume these steps have been completed.

### Sign and Push the Image and Signature

Using the private key, we'll sign the net-monitor image. Note, we're signing the image with a registry name that we haven't yet pushed to. This enables offline signing scenarios. This is important as the image will eventually be published on `registry.wabbit-networks.io/`, however their internal staging and promotion process may publish to internal registries before promotion to the public registry.

- Generate an [nv2 signature][nv2-signature], persisted locally as `net-monitor_v1.signature.config.jwt`

- Enable notary, for the nv2 extension to account for signing and verification steps
  ```shell
  docker notary --enabled
  ```
- Generate an [nv2 signature][nv2-signature], persisted within the `/.docker/nv2/sha256/` directory:
  ```shell
  docker notary sign \
    --key ./wabbit-networks.key \
    --cert ./wabbit-networks.crt \
    $IMAGE
  ```
- Push the container image
  ```bash
  docker push $IMAGE
  ```
- View the output, that includes pushing the signature as a reference:
  ```bash
  The push refers to repository [registry.wabbit-networks.io/net-monitor]
  8ea3b23f387b: Preparing
  8ea3b23f387b: Pushed
  v1: digest: sha256:31c6d76b9a0af8d2c0a7fc16b43b7d8d9b324aa5ac3ef8c84dc48ab5ba5c0f49 size: 527
  Pushing signature
  signature manifest: digest: sha256:8eb7394c8f287ebd0e84a4659f37a2688c6e07e39906933ccb83d9011fb29034 size: 2534
  refers to manifests: digest: sha256:31c6d76b9a0af8d2c0a7fc16b43b7d8d9b324aa5ac3ef8c84dc48ab5ba5c0f49 size: 527
  ```
- Discover the references using the [oras prototype-2 branch](https://github.com/deislabs/oras/tree/prototype-2).
  ```bash
  oras discover \
    --plain-http \
    $IMAGE
  ```

### Validate the image

To validate an image, `docker pull` with `docker notary --enabled` will attempt to validate the image, based on the local keys.

- Attempt to pull the `net-monitor:v1` image:
  ```bash
  docker pull $IMAGE
  ```

- The above command will fail, as we haven't configured the `nv2` client access to the public keys.
  ```bash
  Looking up for signatures
  Found 1 signatures
  2021/03/02 18:34:47 none of the signatures are valid: verification failure: x509: certificate signed by unknown authority
  ```

- Open the `nv2.json` configuration file:
  ```bash
  code ~/.docker/nv2.json
  ```

- Add the wabbit networks public key:
  ```json
  {
    "enabled": true,
    "verificationCerts": [
      "/home/[USER]/nv2-demo/wabbit-networks.crt"
    ]
  }
  ```

- Pull the `net-monitor:v1` image, using the public key for verification:
  ```bash
  docker pull $IMAGE
  ```
- The validated pull can be seen:
  ```bash
  v1 digest: sha256:48575dfb9ef2ebb9d67c6ed3cfbd784d635fcfae8ec820235ffa24968b3474dc size: 527
  Looking up for signatures
  Found 1 signatures
  Found valid signature: sha256:282f5475ac4788f5c0ce3c0c44995726192385c2cae85d0f04da12595707a73f
  The image is originated from:
  - registry.wabbit-networks.io/net-monitor:v1
  registry.wabbit-networks.io/net-monitor@sha256:48575dfb9ef2ebb9d67c6ed3cfbd784d635fcfae8ec820235ffa24968b3474dc: Pulling from net-monitor
  Digest: sha256:48575dfb9ef2ebb9d67c6ed3cfbd784d635fcfae8ec820235ffa24968b3474dc
  Status: Downloaded newer image for registry.wabbit-networks.io/net-monitor@sha256:48575dfb9ef2ebb9d67c6ed3cfbd784d635fcfae8ec820235ffa24968b3474dc
  registry.wabbit-networks.io/net-monitor@sha256:48575dfb9ef2ebb9d67c6ed3cfbd784d635fcfae8ec820235ffa24968b3474dc  
  ```

### Create and Push an SBoM

Push the image, and its signature in one user gesture. Note the push links the signature to the image for later retrieval by a `:tag` or `digest`.

- Create an overly simplistic SBoM
  ```bash
  echo '{"version": "0.0.0.0", "artifact": "registry.wabbit-networks.io/net-monitor:v1", "contents": "good"}' > sbom.json
  ```
- Push the SBoM with ORAS, saving the manifest for signing
  ```bash
  oras push $REPO \
    --artifact-type application/x.example.sbom.v0 \
    --artifact-reference $IMAGE \
    --export-manifest sbom-manifest.json \
    --plain-http \
    ./sbom.json:application/tar
  ```
- View the references with `oras discover`:
  ```bash
  oras discover \
    --plain-http \
    $IMAGE
  ```

### Sign the SBoM

In the above case, the SBoM has already been pushed to the registry. To sign it before pushing, we could have used `oras push` with the `--dry-run` and `--export-manifest` options. 

- For non-container images, we'll use the `nv2` cli to sign and  the `oras` cli to push to a registry. We'll use the `oras discover` cli to find the sbom digest the signature will reference.
  ```bash
  nv2 sign \
    -m x509 \
    -k wabbit-networks.key \
    -c wabbit-networks.crt \
    --plain-http \
    --push \
    --push-reference oci://${REPO}@$(oras discover \
      --artifact-type application/x.example.sbom.v0 \
      --output-json \
      --plain-http \
      $IMAGE | jq -r .references[0].digest) \
    file:sbom-manifest.json
  ```
- View the referenced artifacts, starting with the `net-monitor:v1` image
  ```bash
  oras discover \
    --plain-http \
    $IMAGE

  Discovered 2 artifacts referencing registry.wabbit-networks.io/net-monitor:v1
  Digest: sha256:0da7b8db631b5faeff09f6217de7ac47bdcd53e0e7a15cec559a8140ac164f5c

  Artifact Type                    Digest
  application/x.example.sbom.v0    sha256:13d93a1ba883976b5b59491fe76c2dc94863db3820dc09b63b033ba8194cd96d
  application/vnd.cncf.notary.v2   sha256:75a4b865eda581ec35ac51d3ac8283a37bf7507550d60ce94ee208a9d3edd167
  ```
  The above shows the Notary v2 signature of the `net-monitor:v1` image, and the SBoM.
- View the SBoM signature, using `oras discover` and the digest from above
  ```bash
  # set the digest from the output above, referencing application/x.example.sbom.v0
  SBOM_DIGEST=sha256:13d93a1ba883976b5b59491fe76c2dc94863db3820dc09b63b033ba8194cd96d
  oras discover \
    --plain-http \
    ${REPO}@${SBOM_DIGEST}
  ```
- Dynamically get the SBoM digest
  ```bash
  oras discover \
    --plain-http \
    ${REPO}@$(oras discover \
      --artifact-type application/x.example.sbom.v0 \
      --output-json \
      --plain-http \
      $IMAGE | jq -r .references[0].digest) \
    --output-json | jq
  ```

  ```json
  {
    "digest": "sha256:13d93a1ba883976b5b59491fe76c2dc94863db3820dc09b63b033ba8194cd96d",
    "references": [
      {
        "digest": "sha256:c3b2c533e1ae852a99c1f52f749d26e4d4f3f1979e1955bf77726ef4e81d886c",
        "manifest": {
          "schemaVersion": 2,
          "mediaType": "application/vnd.oci.artifact.manifest.v1+json",
          "artifactType": "application/vnd.cncf.notary.v2",
          "blobs": [
            {
              "mediaType": "application/vnd.cncf.notary.signature.v2+jwt",
              "digest": "sha256:bf13c043317b22b99b317cfd7f6a70bc546477b93480d25c6b97ac0017849f19",
              "size": 2454
            }
          ],
          "manifests": [
            {
              "mediaType": "application/vnd.oci.artifact.manifest.v1+json",
              "digest": "sha256:13d93a1ba883976b5b59491fe76c2dc94863db3820dc09b63b033ba8194cd96d",
              "size": 500
            }
          ]
        }
      }
    ]
  }
  ```


This shows the target experience we're shooting for, within various build and container runtime tooling.

## Demo Reset

If iterating through the demo, these are the steps required to reset to a clean state:

- Remove docker alias:
  ```bash
  unalias docker
  ```
- Reset the local registry:
  ```bash
  docker rm -f $(docker ps -a -q)
  docker run -d -p 80:5000 --name oci-artifact-registry notaryv2/registry:nv2-prototype-2
  ```
- Remove the `net-monitor:v1` image:
  ```bash
  docker rmi -f registry.wabbit-networks.io/net-monitor:v1
  ```
- Remove `wabbit-networks.crt` from `"verificationCerts"` in the `nv2.json` configuration file:
  ```bash
  code ~/.docker/nv2.json
  ```

[docker-generate]:        https://github.com/shizhMSFT/docker-generate
[nv2-signature]:          ../signature/README.md
[oci-image-manifest]:     https://github.com/opencontainers/image-spec/blob/master/manifest.md
[oci-image-index]:        https://github.com/opencontainers/image-spec/blob/master/image-index.md
[oci-artifact-manifest]:  https://github.com/SteveLasker/artifacts/blob/oci-artifact-manifest/artifact-manifest.md
[oras]:                   https://github.com/deislabs/oras
[nv2-prototype-2]:        https://github.com/notaryproject/notaryproject/issues/53
