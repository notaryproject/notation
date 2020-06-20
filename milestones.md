# Milestones for the Notary v2 Prototype

To enable various SMEs, project owners and users the ability to provide feedback for the [Notary v2 e2e scenarios][nv2-scenarios], we provide the following milestones by which we plan to execute.

Based on the milestones, various prototypes will be built and iterated upon, instanced in an [experimental environment](.experimental-environment.md).

The milestones will be used to focus specific efforts, while understanding when additional efforts may be covered.

## Table of contents

- [Milestone 0](#milestone-0)
- [Milestone 1](#milestone-1)
- [Milestone 2](#milestone-2)
- [Milestone ___](#milestone-___)
- [Constraints](#constraints)

## Milestone 0

Enables the initial baseline for the prototyping phase. We will build-out a baseline, required to start design.

Including:

- Iterate and incorporate feedback to the [prototype sketch](.sketch.md)
- nv2 client for pushing and pulling signatures to a registry.
  - As a baseline, we'll have a golang based cli style project enabling parameters and help documentation: `nv2 --help`
- A build environment to build images and the associated artifacts, such as a mock SBoM
  - A breakout group will decide what standard infrastructure shall be used.
- Three instances of the [notary/distribution][notary-distribution] reference implementation
  - Equivalent of a public registry with anonymous pull of artifacts
  - Equivalent of a private registry, with secured push/pull of artifacts
  - Equivalent of a private, air-gapped registry where content must be moved to, and all resources in that air-gapped environment are limited from making external calls
- Three instances of k8s
  - Equivalent of a development environment
  - Equivalent of a staging environment, in network that has public access
  - Equivalent of a air-gapped environment, with no public network access
- An instance of an OSS project for key management from [cncf][cncf-projects]
- An instance of a policy management solution from [cncf][cncf-projects]

## Milestone 1

Begins iterative design changes required to satisfy the [sketch](.sketch.md), supporting the [Notary v2 Scenarios][nv2-scenarios]

- Registry lookup of signatures using the [signature verification lookup design #22](https://github.com/notaryproject/requirements/pull/22)
  - Add `index.config.mediaType` to oci-index, enabling the identification of indexes that are of type signature
  - Design how signatures of an artifact will be stored in a registry - can they be stored _in_ the `index.config`? See: [verification-config][config-signatures]
- Mock SBoM document
  - Contain the most basic information to enable policy management decisions
  - Provide an additional artifact to include in an oci-index of a collection of objects related to an artifact
- Key management solution
  - any unique changes required to store offline keys
- Policy management changes to read the mock SBoM

## Milestone 2

Unplanned, but as we complete milestone 0 and start milestone 1, we will start to scope the current milestone +1

## Milestone ___

As we manage the backlog of issues, we will start to frame out future milestones for teams to do appropriate design planning.

## Constraints

### Cloud neutrality / oss implementations

To assure cloud neutrality, the prototypes and experimental environment will prioritize cloud neutral projects for the various components. The purpose here is to assure we're designing with cloud neutrality in mind. The cloud neutrality will have a balanced approach, utilize the clouds core infrastructure that doesn't have an impact on the overall design.

For example: an instance of [notaryproject/distribution][notary-distribution] will be used in the experimental environment. Key and policy management will use one of the [cncf projects][cncf-projects]. However, we won't attempt to implement a different authentication or storage service as we don't see these as fundamental to  Notary v2 design decisions.

### Cloud/vendor implementations

An effort like this is driven by the ability to instance the Notary v2 e2e workflows for each cloud and vendors customers.

While the Notary v2 working group will focus on the cloud neutral implementations, we desire and expect cloud vendors to work on their unique implementations. It is through these vendor specific incubations we expect to get extensibility requirements, pull requests and customer scenarios to assure the projects success.

[cncf-projects]:        https://www.cncf.io/projects/
[notary-distribution]:  https://github.com/notaryproject/distribution
[nv2-scenarios]:        https://github.com/notaryproject/requirements/blob/master/scenarios.md
[config-signatures]:    https://github.com/notaryproject/requirements/blob/e7743d8e700f591a1f2b0ffb256909783c6a0881/verification-by-reference.md#weba2b2-staging-verification-config
