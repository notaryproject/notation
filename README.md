# nv2 - Prototype

nv2 is an incubation and prototype for designing the [Notary v2][notaryv2] efforts, securing artifacts stored in [OCI distribution-spec][oci-distribution] based registries.

## Table of contents

- [Prototyping components](#prototyping-components)
- [Prototyping approach](#prototyping-approach)
- [Sketch, prototype, experiment, iterate](#sketch-prototype-experiment-iterate)
  - [Prototype Sketch](.sketch.md)
  - [Milestones](./milestones.md)
  - [Experimental Environment](./experimental-environment.md)

## Prototyping components

As the below _end to end_ (e2e) workflow visualizes, there are many components we must account for in this e2e experience. Not accounting for the e2e experience could leave the community with a new set of blocking issues found with the Docker Content Trust implementation of Notary v1.

Components to consider:

- Build environments
- Key management - including the ability to integrate with a vendors native key management solution
- Software Bill of Materials (SBoM)
- Source packaging
- Public and private registries, including air-gapped and [purdue network][purdue-network] isloated registries
- Vulnerability and security scanning products
- Policy Management - to leverage an SBoM and signatures to determine _if_ an artifact is trusted and should be deployed
- Container host environments, like kubernetes

![Notary v2 e2e workflow](media/notary-e2e-scenarios.svg)

## Prototyping approach

There are many approaches to building prototypes. Some approaches cater to simple projects, while others are better at supporting complex projects.

Notary v2 is goaled at securing a complex e2e secure supply chain workflow. This will involve many subject matter experts (SMEs) and various projects to engage. Since no one person or group has a concrete blueprint for what and how we would build this e2e solution, we can be stalled with gaps of communication and differing views.

### The value of different perspectives

We all have experiences and biases that guide us. These can be an asset to forming a diverse set of views, but can also be difficult to overcome when a shared interaction must be made between any two parts. To build out the Notary v2 experience we must incorporate interaction between various projects, with a shared understanding, and individuals within each project must have a shared understanding. However, the shared understanding is assumed to evolve as we all learn together and from each others differing views.

- [“No two persons ever read the same book.”: Edmund Wilson](https://www.goodreads.com/quotes/23977-no-two-persons-ever-read-the-same-book).
- [“How two can see the same thing and interpret it differently”](https://jenalynalbia.wordpress.com/2017/01/11/explain-how-two-can-see-the-same-thing-and-interpret-it-differently/)
- [Mars probe lost due to simple math error](https://www.latimes.com/archives/la-xpm-1999-oct-01-mn-17288-story.html)

### Building complex software

Building a complex solution is not unique to Notary v2. We will bring SMEs from various areas, each with their own views, and we will continue to evolve the design until we're ready to execute. 

In software, there are many models, including waterfall and iterative. However, within the iterative, there are at least two additional approaches:

1. Build and iterate with constant changes, churn and frustration to those dependent on the outcome
    - Consumers of the effort can get lost with complaints of instability
1. Build a prototype, learn, toss, build the real thing, with a reasonable amount of iterations
    - Consumers clearly see this as a prototype, monitor, provide feedback and await the outcome while the SMEs work out all the details

### Prototyping complex projects

In construction, we must bring together various designers, architects and trades:

- Designers provide sketches to quickly iterate ideas, narrowing in a common goal
- Architects provide detailed blueprints, with layered designs from various trades, incorporating their expertise
  - Grading contractors - sculpting the ground by which the property will reside
  - Foundation contractors - providing a solid foundation for the structure, including environmental impact and risk (earthquakes, floods, ...)
  - Framing contractors - accounting for the various contractors that must fit all internals that make a house a home
  - HVAC contractors - have large spaces to heat and cool, requiring the framers to account for the plenums and returns
  - Plumbing contractors - that may provide detailed design for that fancy glassless shower and constant hot water
  - Electrical contractors - needing to place the switches and outlet in all the right areas you blindly reach for
  - ...

Each trade may not know the details of the other trades, but they know they need to work together. The plumbers and electrician must work around the HVAC systems, the grading contractors must provide a solid footing, with water runoff for the foundation to be stable.

While auto-cad and 3D programs allow users to visualize the design, we still often start out with a sketch for where to start the detailed design. For complex designs, modeling is often used to see _how_ the design will actually work. Can you really extend the patio that far out without it bouncing? Or how long and how much water will it take to get hot water to the shower? As productive as auto-cad and 3D programs are, it's still complex and expensive to design a building from scratch. Which is why so many buildings are based on existing proven templates. To build something new, depending on the complexity of the problem, we may need to sketch and model a design before proceeding to detailed blueprints.

### The work of Antoni Gaudí

[Antoni Gaudí](https://en.wikipedia.org/wiki/Antoni_Gaud%C3%AD) is famously known for his amazingly creative works in Barcelona. The [Sagrada Famila](https://simple.wikipedia.org/wiki/Sagrada_Fam%C3%ADlia) was a departure from the massive piers and buttress designs. Gaudí wanted a more natural look, which had no existing templates to work from. Gaudí sketched and modeled many times to work out the intricate details for the various trades to work together. To design natural arches and vaults, Gaudí created an inverted model using small bags of birdshot and string. It's through this sketch, model, design, execute approach that Gaudí was able to enlist the creative skills of various trades to _eventually_ complete the Sagrada Famila.

![Antoni Gaudí](https://upload.wikimedia.org/wikipedia/commons/thumb/7/72/Antoni_Gaudi_1878.jpg/176px-Antoni_Gaudi_1878.jpg)
<img src=https://upload.wikimedia.org/wikipedia/commons/thumb/f/fa/Maqueta_funicular.jpg/800px-Maqueta_funicular.jpg width=200>)
<img src=https://upload.wikimedia.org/wikipedia/commons/a/ab/Gaud%C3%AD-_Martorell-_Catedral_BCN_%281887%29.jpg width=200>

## Sketch, prototype, experiment, iterate

The different views and interaction of the different trades is equivalent to the different views and interaction we need between the different SMEs and project owners for Notary v2.

- Key management folks need to figure out where they should engage, providing input on how keys should be managed
- Key vault solutions must understand where they plug in their key vault provider for each registry
- Policy management folks need to understand what content they can pull from a registry, and how they should trust it to make policy decisions
- The update framework folks must understand where they can plug in their metadata to assure the content is secured and trusted
- The folks working on the secure software supply chain efforts must understand the registry workflows and what they must account for
- The registry vendors must understand the implications for the changes they must make to support Notary v2
- Just as the public provides feedback on public works, customers need something to view for providing feedback

To facilitate the e2e workflows, we'll:

- [Sketch](./sketch.md) an e2e workflow, supporting the [Notary v2 scenarios][nv2-scenarios]
- Prototype various components of the e2e workflow including
  - The nv2 client for signing artifacts
  - A registry that implements any APIs required to store and serve signatures and verification objects
  - A key vault solution for storing signing keys
  - A SBoM document, used for making policy decisions
  - A Policy Manager, used to make policy decisions
  - A container hosting solution to deploy secured containers
- Create an [environment](./experimental-environment.md) to experiment with the e2e workflows
- Iterate, with a set of [milestones](./milestones.md) each team will work towards

As we get to a point where we feel comfortable with the e2e design, that accounts for the [Notary v2 scenarios][nv2-scenarios], balancing the security and usability goals, we can move to a spec (blueprint) for building out the final versions of each component.

[notaryv2]:             http://github.com/notaryproject/
[oci-distribution]:     https://github.com/opencontainers/distribution-spec
[oci-image]:            https://github.com/opencontainers/image-spec
[purdue-network]:       https://en.wikipedia.org/wiki/Purdue_Enterprise_Reference_Architecture
[nv2-scenarios]:        https://github.com/notaryproject/requirements/blob/master/scenarios.md
