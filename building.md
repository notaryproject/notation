# Building NV2

The nv2 repo contains the following:

- `nv2` - A CLI for signing and verifying with Notary v2
- `docker-generate` - Extends docker with `docker generate` to create locally persisted manifest for signing, without having to push to a registry.
- `docker-nv2` - Extends  docker with `docker nv2 notary` to enable, sign and verify Notary v2 signatures on `docker pull`

This plugin requires [golang](https://golang.org/dl/) with version `>= 1.16`.

## Windows with WSL

- Build the binaries, installing them to:
  - `~/bin/nv2`
  - `~/.docker/cli-plugins/docker-generate`
  - `~/.docker/cli-plugins/docker-nv2`
  ```shell
  git clone https://github.com/notaryproject/nv2.git
  cd nv2
  git checkout prototype-2
  make install
  ```
- Verify binaries are installed
  ```bash
  docker --help
  # look for 
  Management Commands:
    generate*   Generate artifacts (github.com/shizhMSFT, 0.1.0)
    nv2*        Notary V2 Signature extension (Sajay Antony, Shiwei Zhang, 0.2.3)
  
  which nv2
  # output
  /home/<user>]/bin/nv2
  ```
