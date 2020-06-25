# Notary v2 Experimental Environment

To enable various SMEs and project owners the ability to provide feedback for the [Notary v2 e2e scenarios][nv2-scenarios], we provide the following reference environment for what we intend to build.

Based on the [Notary v2 sketch](sketch.md), a set of prototypes will be built and instanced in this experimental environment. Following a set of [milestones](milestones.md) the environment will be updated to reflect the goals of the milestone.

A experimental reference environment will be provided for the maintainers of the Notary v2 efforts in [Azure](http://www.azure.com/).

## Self instancing the reference environment

For each milestone, we will provide instructions and scripts, to create the reference environment for users to experiment and iterate themselves.

## Milestone 0 environment

### Mock public artifacts

To represent a mock public registry, we'll need to build and host a mock docker hub environment. This includes the building and hosting of two base images: (linux and node)

To represent ths public content, we'll create:

- Public registry
  - An instance of [notary/distribution][notary-distribution] that represents a public registry like: (registry.notaryv2.io)
  - We may create a simple markdown page for the list of images, but we're not attempting to duplicate a public registry user interface.
- Public content build environment
  - The Linux and Node base images will need to be signed and hosted for this experimental environment
  - To optimize, we'll create one public-build environment for both the Linux and Node images.
- Public content key vault
  - To build the public images, we'll need to store the private keys used to build this content.
  - This key vault is not publicly accessible. It's the equivalent of the linux distros internal key vault store. And, the node communities key vault store.
  - To optimize, we'll use one key vault solution to host both the linux and node images, but we will store separate keys.
- Public content git repository
  - A git repo to build the linux image https://github.com/notaryproject/mock-linux
  - A git repo to build the node image https://github.com/notaryproject/mock-node

### ACME Rockets content

To represent a customer environment, we'll build-out the ACME Rockets fictitious company.

- Private registry
  - An instance of [notary/distribution][notary-distribution] that represents a private registry: (registry.acmerockets.io)
- Corporate git repo
  - https://github.com/acme-rockets/*
- Corporate standard artifacts build environment
  - linux image
    - https://github.com/acme-rockets/base-linux
  - node image
    - https://github.com/acme-rockets/base-node
  - hello-world image
    - https://github.com/acme-rockets/hello-world
- Key vault
  - An instance of a [cncf][cncf-projects] key vault

## Milestone 1 environment

[cncf-projects]:        https://www.cncf.io/projects/
[notary-distribution]:  https://github.com/notaryproject/distribution
[nv2-scenarios]:        https://github.com/notaryproject/requirements/blob/master/scenarios.md
