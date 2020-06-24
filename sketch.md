# A Sketch for Notary v2 Prototyping and Experimentation

To enable various SMEs, project owners and customers the ability to provide feedback for the [Notary v2 e2e scenarios][nv2-scenarios], we provide the following sketch for what we intend to build.

By developing a sketch, we enable a lightweight form of discussion of ideas, enabling collaboration across the different entities.

Based on this sketch, various prototypes will be built and iterated upon, instanced in an [experimental environment](.experimental-environment.md).

## e2e Top View

![Notary v2 e2e workflow](media/notary-e2e-scenarios.svg)

> TODO: Complete description, based on the [Notary v2 e2e scenarios][nv2-scenarios]

To illustrate an e2e workflow, we provide the following example:

- A public registry with public content (debian & node images)
- A customer (ACME Rockets) with a cloud presence and an on-prem, IoT, air-gapped environment
- Artifacts are signed, when pushed to the public registry
- Public artifacts and their signatures are copied to the private registries of ACME Rockets
- ACME Rockets will additionally certify the public content, adding a signature for the standard/approved base-artifacts
- ACME Rockets builds many apps, including a `web` image that will be built from the `/base-artifact/node` image
- As the `web` image moves from dev through production, additional signatures are added, providing proof the image was verified
- Policy management in the staging environment requires a signature from development, and an SBoM of the content in the image.
- Policy management in the production environment requires an additional signature from the staging environment.
- The web image, along with its signatures are copied to the air-gapped environment for local verification and deployment
- An additional signature is required to run in the secured IoT environments

## Base Artifacts

To represent a set of public base artifacts, create a set of public artifacts signed by their representative entities.

- Linux base image
  - Signed by a fictitious penguin that mimics an entity that would own and sign a public linux image. We'll use debian as it's the base for the node image we'll use
  - Include a mock SBoM to represent the content in the image
- Node base image
  - Signed by a fictitious entity that would represent the node community
  - Include a mock SBoM to represent the content in the image
  - We'll defer the inclusion of the source for all the npm packages. Although, this would be an interesting exercise to see how registries could de-dupe source references to specific npm packages, represented as [oci artifacts][oci-artifacts].

To represent these base artifacts, which would be signed, we'll need to create this [fictitious environment](./experimental-environment.md#mock-public-content).

The creation and management of the base artifacts will duplicate much of the validation workflow for the developers app workflows As a result, we'll optimize the creation and maintenance of these artifacts.

### Base artifact build environment

A single build environment will be created to manage the building and signing of the linux and node images.

- For the purpose of isolation, we will create separate keys for the debian and node images.
- For simplicity of managing multiple key vault instances, the node and debian keys will be stored in the same key vault instance.
- With a focus on signing, we can optimize by simply importing a selected linux and node image.

The imported content will be:

**node image**:

- Built `FROM docker.io/node`
  - tagged `node:[version]-[os]-[version]`
  - Signed by the fictitious node community key
  - Include an SBoM `node:[version]-sbom`
  - Have an oci-index that includes the image and it's SBoM `node:[version]`

**debian image**:

- Built `FROM docker.io/debian`
  - tagged `debian:[version]
  - Signed by the fictitious penguin distro key
  - Include an SBoM `debian:[version]-sbom`
  - Have an oci-index that includes the image an it's SBoM `debian:[version]`

The build environment will be triggered by a git commit to the backing git repo to ease rebuilding.

## ACME Rockets flow

### ACME Rockets keys

The ACME Rockets organization provides a set of keys that will be used for signing their corporate standard artifacts and their custom applications. The following keys will be created:

- Corporate base artifacts key `acme-rockets-base-artifact`
  - The company will maintain a set of corporate standard artifacts, including a linux base image and various runtime images. (node)
  - All corporate standard artifacts will be imported from the public registry (`registry.notaryv2.io`), tested and signed with the corporate key
- Development key `acme-rockets-ateam-dev`
  - As artifacts are built in development, they are unit tested and scanned before being promoted to a staging environment. Only artifacts that pass unit tests and pass scanning are signed with the `acme-rockets-ateam-dev` key. Only artifacts that are signed by a set of known development teams will be permitted into the staging environment.
- Production validation `acme-rockets-ateam-prod`
  - As artifacts are validated in the staging environment, they will be signed with an additional production key. Only artifacts signed with production keys will be permitted to be run in the production environments.

These keys will be stored in the companies key-vault solution.

### ACME Rockets build environment

In a build environment, the following occurs:

- An [oci-image][oci-image] is created, representing a runnable container image `hello-world:a1b2c3`
  - A node.js, hello word web app, referencing a small set of npm packages.
  - The app source is stored in a git repository
  - **note:** this is the equivalent of a `docker build`, with the minor but important difference that an [oci-image manifest][oci-image-manifest] is what's generated.
- An SBoM is generated
  - The SBoM contains:
    - the list of npm packages and versions referenced
    - the node version
    - ... any other minimal information to enable policy management decisions
- The `./src` of the project is added as an additional OCI Artifact, supporting gpl type license requirements
- An OCI-index that groups the above elements together as a single tag
- All 4 artifacts (image, SBoM, src, index) are signed

### ACME Rockets staging/validation

### ACME Rockets production environment

### ACME Rockets IoT environment

An air-gapped environment that must account for secured and signed content.

[nv2-scenarios]:        https://github.com/notaryproject/requirements/blob/master/scenarios.md
[oci-artifacts]:        https://github.com/opencontainers/artifacts/
[oci-image]:            https://github.com/opencontainers/image-spec/
[oci-image-manifest]:   https://github.com/opencontainers/image-spec/blob/master/manifest.md