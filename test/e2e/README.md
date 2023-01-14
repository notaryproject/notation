# A Quick Introduction on how Notation end-to-end test works

## Framework
Using [Ginkgo](https://onsi.github.io/ginkgo/) as the e2e framework, which is based on the Golang standard testing library.

Using [Gomega](https://onsi.github.io/gomega/) as the matching library.

## Introduction

### Data：testdata contains data needed by e2e tests, including:
- *config*: notation test key and cert files.
- *registry*: OCI layout files and registry config files.
### For developer
- *Test registry*: a test registry started before running tests.
- *Config isolation*: notation needs a few configuration files in user level directory, which can be isolated by modify `XDG_CONFIG_HOME` environment variable. In Notation E2E test framework, a VirtualHost abstraction is designed for isolating user level configuration.
- *Parallelization*: In order to speed up testing, Ginkgo will launch several processes to run e2e test cases. 
- *Randomization*: By default, Ginkgo will run specs in a suite in random order. Please make sure the test cases can be runned independently. If the test cases depend on the execution order, consider using [Ordered Containers](https://onsi.github.io/ginkgo/#ordered-containers).


## Setting up
### Github Actions
- Please check `Run e2e tests` steps in **workflows/build.yml** for detail.
### Local environment
- Install Golang.
- Install Docker.
- Clone the repository.
- Run `cd ./test/e2e` 
- Run `./run.sh <absolute_path_to_notation_binary>` 