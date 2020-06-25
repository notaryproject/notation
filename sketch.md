# A Sketch for Notary v2 Prototyping and Experimentation

To enable various SMEs, project owners and customers the ability to provide feedback for the [Notary v2 e2e scenarios][nv2-scenarios], we provide the following sketch for what we intend to build.

By developing a sketch, we enable a lightweight form of discussion of ideas, enabling collaboration across the different entities.

Based on this sketch, various prototypes will be built and iterated upon, instanced in an [experimental environment](.experimental-environment.md).

## e2e Top View

![Notary v2 e2e workflow](media/notary-e2e-scenarios.svg)

To illustrate an e2e workflow, we provide the following example:

- A public registry with public content (`debian` & `node` images)
- A customer, _ACME Rockets_ with a cloud presence and an on-prem, IoT, air-gapped environment
- Public artifacts and their signatures are copied to the private registries of ACME Rockets
- ACME Rockets will additionally certify the public content, adding a signature for the standard/approved base-artifacts
- ACME Rockets builds many apps, including a `web` image that will be built from the `/base-artifact/node` image
- As the `web` image moves from dev through production, additional signatures are added, providing proof the image was verified
- Policy management in the staging environment requires a signature from development, and an SBoM of the content in the image.
- Policy management in the production environment requires an additional signature from the staging environment.
- The web image, along with its signatures are copied to the air-gapped IoT environment for local verification and deployment
- An additional signature is required to run in the secured IoT environments

## Public base artifacts

To represent a set of public base artifacts, that are signed, we will create a mock public registry.

- Linux base image
  - Signed by a fictitious penguin that mimics an entity that would own and sign a public linux image. We'll use debian as it's the base for the node image we'll use
  - Include a mock SBoM to represent the content in the image
- Node base image
  - Signed by a fictitious entity that would represent the node community
  - Include a mock SBoM to represent the content in the image
  - We'll defer the inclusion of the source for all the npm packages. Although, this would be an interesting exercise to see how registries could de-dupe source references to specific npm packages, represented as [oci artifacts][oci-artifacts].

The creation and management of the base artifacts will duplicate much of the validation workflow for the developers app workflows. To minimize duplication of efforts and complexity, we'll optimize the creation and maintenance of these public artifacts.

### Public base artifact build environment

A single [public artifact build environment](./experimental-environment.md#mock-public-artifacts) will be created to manage the building and signing of the linux and node images.

- For the purpose of isolation, we will create separate keys for the debian and node images.
- For simplicity of managing multiple key vault instances, the node and debian keys will be stored in the same key vault instance.
- With a focus on signing, we can optimize by simply importing a selected linux and node image.

The imported content will be:

**node image**:

- Built `FROM docker.io/node`
  - tagged `node:[version]-[os]-[version]`
  - Signed by the fictitious node community key
  - Include an [SBoM][sbom] `node:[version]-sbom`
  - Have an [v2 oci-index][oci-index-v2] that includes the image and its SBoM `node:[version]`

**debian image**:

- Built `FROM docker.io/debian`
  - tagged `debian:[version]
  - Signed by the fictitious penguin distro key
  - Include an [SBoM][sbom] `debian:[version]-sbom`
  - Have an [v2 oci-index][oci-index-v2] that includes the image an its SBoM `debian:[version]`

The build environment will be triggered by a git commit to the backing git repo.

## ACME Rockets e2e flow

ACME Rockets follows [best practices][registry-best-practices] for securing the content they depend upon (base images and other artifacts like helm charts). As part of this process, they copy their content from public registries to their private owned registries.

With a set of in-house base images, the company builds many custom apps which they validate, sign and move across several private registries.

For air-gapped, IoT environments, ACME Rockets will copy the signed artifacts to additional air-gapped registries.

### ACME Rockets keys

The ACME Rockets organization provides a set of keys that will be used for signing their corporate standard artifacts and their custom applications. The following keys will be created:

- Corporate base artifacts key `acme-rockets-base-artifact`
  - The company will maintain a set of corporate standard artifacts, including a debian and node base images
  - All corporate standard artifacts will be imported from the public registry (`registry.notaryv2.io`), tested and signed with the corporate key
- Development key `acme-rockets-dev-team-a`
  - As artifacts are built in development, they are unit tested and scanned before being promoted to a staging environment. Only artifacts that pass unit tests and pass scanning are signed with the `acme-rockets-dev-team-a` key. Only artifacts that are signed by a set of known development teams will be permitted into the staging environment.
- Staging environment key `acme-rockets-prod-team-a`
  - As artifacts are validated in the staging environment, they will be signed with an additional production key. Only artifacts signed with production keys will be permitted to be run in the production environments. Each team is provided a production key, uniquely assigned to each team for traceability.
- IoT environment key `acme-rockets-prod-iot-team-a`
  - As artifacts are moved to the IoT air-gapped environment, they will be signed with an additional IoT production key. Only artifacts signed with production keys will be permitted to be run in the IoT environments.

These keys will be stored in the ACME Rockets key-vault.

### Base artifact maintenance

ACME Rockets brings into their environment any artifacts they depend upon, validating them and assigning an `acme-rockets` signature.

- How these base-artifacts are maintained is outside the scope of the Notary v2 effort. Different vendors have differentiated offerings supporting a customers ability to maintain a buffered set of dependent artifacts.
- For the purposes of this exercise, a build server will handle importing public images, testing and signing them with an acme-rockets-base-artifacts key.

### Custom app flow

ACME Rockets maintains a set of custom apps they develop and deploy within their organization.

### ACME Rockets `web image` build environment

- An [oci-image][oci-image] is created, representing a runnable container image with a unique tag: `web:[build-id]`
  - A node.js, hello word web app, referencing a small set of npm packages.
  - The app source is stored in a git repository: [github.com/acme-rockets/web](https://github.com/acme-rockets/web)
- Base images that aren't signed with the `acme-rockets-base-artifacts` key fail the build
- An [SBoM][sbom] is generated, saved as an [OCI Artifact][oci-artifacts] with a `manifest.config.mediaType` = `application/vnd.oci.prototype.sbom.config.v01`
- The `./src` of the project is added as an additional [OCI Artifact][oci-artifacts], supporting gpl type license requirements. The `manifest.config.mediaType` = `application/vnd.oci.prototype.src.config.v01`
- An OCI-index, that includes: image, SBoM, src with a tag of `web:[build-id]-package`
- All 4 artifacts (image, SBoM, src, index) are signed with the `acme-rockets-dev-team-a` key
- All 4 artifacts are pushed to the acme-rockets private registry, using the `/team-a/` repository

### ACME Rockets `web image` staging validation

- Validation requires:
  - an `acme-rockets-dev-team-a` signature
  - an SBoM that meets ACME Rockets production compliance requirements
  - a functional test that checks the background color of the home page
    - if the `back-color = red`, the functional test fails
- If the image passes the functional testing, the image, sbom and src artifacts are signed with an additional `acme-rockets-prod-team-a` key
  - The collection of artifacts are pushed to the `/prod/team-a/web` repository
  - The additional signatures are uniquely added to the `/prod/team-a/web` repository

### ACME Rockets `web image` deployment

- A helm 3 chart is used to deploy the new image
  - The `build-id` is used to pass into the helm chart for deployment
- The policy manager enforces policies that includes:
  - The Helm chart is signed by the `acme-rockets-prod-team-a` key
  - Images referenced, in the merged helm chart, are signed by the `acme-rockets-prod-team-a` key
  - The SBoM is also signed by the `acme-rockets-prod-team-a` key and doesn't reference blocked packages in the SBoM content
- Images that pass policy management are deployed to the cloud production cluster

### ACME Rockets IoT deployment

- Images promoted to the prod repository are also copied to an on-prem/IoT environment (_note: the on-prem environment will be mocked in the cloud experimental environment_)
  - The contents of the index, along with their signatures, are copied to the air-gapped environment for local verification and deployment
- Within the air-gapped environment, an additional `acme-rockets-prod-iot-team-a` signature,  attesting to approved content is added
- Deployments are initiated to several mocked IoT devices. These mocked devices, (micro-vms with a containerd host), will validate the content is signed with the `acme-rockets-prod-team-a` and `acme-rockets-prod-iot-team-a` keys. Deployments that aren't properly signed fail the deployment, logging the notaryv2 failure in the host logs.

[nv2-scenarios]:           https://github.com/notaryproject/requirements/blob/master/scenarios.md
[oci-artifacts]:           https://github.com/opencontainers/artifacts/
[oci-image]:               https://github.com/opencontainers/image-spec/
[oci-image-manifest]:      https://github.com/opencontainers/image-spec/blob/master/manifest.md
[registry-best-practices]: https://stevelasker.blog/2018/11/14/choosing-a-docker-container-registry/
[sbom]:                    ./mock-sbom.md
[oci-index-v2]:            ./oci-index-v2.md
