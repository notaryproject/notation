# Building Notation

The notation repo contains the following:

- `notation` - A CLI for signing and verifying artifacts with Notation

Building above binaries require [golang](https://golang.org/dl/) with version `>= 1.17`.

## Windows with WSL

- Build the binaries, installing them to:
  - `~/bin/notation`
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
    generate*   Generate artifacts (CNCF Notary Project, 0.1.0)
    notation*   Manage signatures on Docker images (CNCF Notary Project, 0.5.3-alpha)
  
  which notation
  # output
  /home/<user>/bin/notation
  ```
