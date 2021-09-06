# Building Notation

The notation repo contains the following:

- `notation` - A CLI for signing and verifying with Notation
- `docker-generate` - Extends docker with `docker generate` to create locally persisted manifest for signing, without having to push to a registry.
- `docker-notation` - Extends docker with `docker notation` to enable, sign and verify Notation signatures.

Building above binaries requires [golang](https://golang.org/dl/) with version `>= 1.17`.

## Windows with WSL

- Build the binaries, installing them to:
  - `~/bin/notation`
  - `~/.docker/cli-plugins/docker-generate`
  - `~/.docker/cli-plugins/docker-notation`
  ```sh
  git clone https://github.com/notaryproject/notation.git
  cd notation
  make install
  ```
- Verify binaries are installed
  ```sh
  docker --help
  # look for 
  Management Commands:
    generate*   Generate artifacts (github.com/shizhMSFT, 0.1.0)
    notation*   Manage signatures on Docker images (Sajay Antony, Shiwei Zhang, 0.5.1)
  
  which notation
  # output
  /home/<user>]/bin/notation
  ```
