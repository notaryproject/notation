# Building Notation

The notation repo contains the following:

- `notation` - A CLI for signing and verifying artifacts with Notation

Building above binaries require [golang](https://golang.org/dl/) with version `>= 1.20`.

## Windows with WSL or Linux

- Build the binaries, installing them to:
  - `~/bin/notation`
  ```sh
  git clone https://github.com/notaryproject/notation.git
  cd notation
  make install
  ```
- Verify binaries are installed
  ```sh
  which notation
  # expected output
  /home/<user>/bin/notation
  ```

  If you confront `notation not found`, please add `~/bin/` to your $PATH:
  ```sh
  export PATH="$HOME/bin:$PATH"
  ```
  If you would like to add the path permanently, add the command to your Bash `~/.profile` file:
  ```sh
  echo 'export PATH="$HOME/bin:$PATH"' >> ~/.profile
  source ~/.profile
  ```
  For other shells, please add the command to the corresponding `profile` file.
  - Bash: ~/.bash_profile or ~/.profile
  - Zsh: ~/.zprofile
  - Ksh: ~/.profile

