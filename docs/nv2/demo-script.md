# nv2 Demo Script

The following is a summary when presenting a demo of Notary v2.

## Demo Setup

Perform the following steps prior to the demo:

- Install [Docker Desktop](https://www.docker.com/products/docker-desktop) for local docker operations
- [Install and Build the nv2 Prerequisites](./README.md#prerequisites)
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
  docker run -d -p 80:5000 --restart always --name registry notaryv2/registry:nv2-prototype-1
  ```
- Add a `etc/hosts` entry to simulate pushing to registry.wabbit-networks.io
  - If running on windows, _even if using wsl_, add the following entry to: `C:\Windows\System32\drivers\etc\hosts`
    ```hosts
    127.0.0.1 registry.wabbit-networks.io
    ```

## Demo Reset

If iterating through the demo, these are the steps required to reset to a clean state:

- Remove docker alias:
  ```bash
  unalias docker
  ```
- Reset the local registry:
  ```bash
  docker rm -f $(docker ps -a -q)
  docker run -d -p 80:5000 --restart always --name registry notaryv2/registry:nv2-prototype-1
  ```
- Remove the `net-monitor:v1` image:
  ```bash
  docker rmi -f registry.wabbit-networks.io/net-monitor:v1
  ```
- Remove `wabbit-networks.crt` from `"verificationCerts"`:
  ```bash
  code ~/.docker/nv2.json
  ```

## Demo Steps

### Explain the end to end experience being presented

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

To see the extension capabilities, review the `Management Commands:`:

```bash
docker --help

Management Commands:
  ...
  generate*   Generate artifacts (github.com/shizhMSFT, 0.1.0)
  nv2*        Notary V2 Signature extension (Sajay Antony, Shiwei Zhang, 0.1.0)

```

To see the sub commands of `nv2` and `nv2 notary`:

```bash
docker nv2 --help
docker nv2 notary --help
```

To avoid having to type `docker nv2` each time, we'll create an alias to mask over this:

```bash
alias docker="docker nv2"
```

To avoid having to type the fully qualified registry name, we'll create an environment variable:

```bash
export image=registry.wabbit-networks.io/net-monitor:v1
```

## Wabbit Networks Build, Sign, Promote Process

Let's walk through the sequence of operations Wabbit Networks takes to build, sign and promote their software.

Within the automation of Wabbit Networks, the following steps are completed:

### Build the `net-monitor` image

```bash
docker build \
    -t $image \
    https://github.com/wabbit-networks/net-monitor.git#main
```

### Acquire the private key

- As a best practice, we'll always build on an ephemeral client, with no previous state.
- The ephemeral client will retrieve the private signing key from the companies secured key vault provider.

These specific steps are product/cloud specific, so we'll assume these steps have been completed.

### Sign the image

Using the private key, we'll sign the net-monitor image. Note, we're signing the image with a registry name that we haven't yet pushed to. This enables offline signing scenarios. This is important as the image will eventually be published on `registry.wabbit-networks.io/`, however their internal staging and promotion process may publish to internal registries before promotion to the public registry.

- Generate an [nv2 signature][nv2-signature], persisted locally as `net-monitor_v1.signature.config.jwt`

  ```shell
  docker notary --enabled

  docker notary sign \
    --key ./wabbit-networks.key \
    --cert ./wabbit-networks.crt \
    $image
  ```
- view the signature referenced from docker notary sign
  ```bash
  cat <output reference>
  ```

- View the manifest the signature is based upon:
  ```bash
  docker generate manifest $image
  ```

### Push the image & signature to the registry

Push the image, and its signature in one user gesture. Note the push links the signature to the image for later retrevial by a `:tag` or `digest`.

```shell
docker push $image
```

### Clear the local image

- To simulate another client, we'll clear out the `net-monitor:v1` image
  ```bash
  docker rmi -f registry.wabbit-networks.io/net-monitor:v1
  rm  ~/.docker/nv2/sha256/*.*
  ```

### Validate the image

To validate an image, `docker pull` with `docker notary --enabled` will attempt to validate the image, based on the local keys.

- Attempt to pull the `net-monitor:v1` image:
  ```bash
  docker pull $image
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
      "/home/stevelas/nv2-demo/wabbit-networks.crt"
    ]
  }
  ```

- Pull the `net-monitor:v1` image, using the public key for verification:
  ```bash
  docker pull $image
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

This shows the target experience we're shooting for, within various build and container runtime tooling.

[nv2-signature]:    ../signature/README.md
[docker-generate]:  https://github.com/shizhMSFT/docker-generate