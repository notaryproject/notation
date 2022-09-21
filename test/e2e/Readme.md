# A Quick Introduction on how Notation end-to-end test works

## Framework
Using [Ginkgo](https://onsi.github.io/ginkgo/) as the e2e framework, which is based on the Golang standard testing library.

Using [Gomega](https://onsi.github.io/gomega/) as the matching library.

## Introduction
1. **Data**： **testdata** contains data needed by e2e tests, including:
    * **images**: a test artifact to be signed.
    * **config**: username and password used by the test registry.

2. **Initialization**： **test/e2e/internal/utils/init.go** contains init functions, including setting up notation-binary path, registry password and cert path. These variables can be set through environment variables.

3. **TestScenario** will run all tests in the package. Pay attention to *test/e2e/scenario/sign.go*. This is where I wrote the spec(test cases).

4. **For developer**
    * **TestRegistry**: a test registry started before running tests.
    * Some tests may use user config dir. To create a separate environment, developers can call `utils.SetUpUserDir()` to get a clean dir for user config.
    * Some tests may need a new Image because there maybe some signatures pointing to the same artifact. Developers can use `utils.TestRegistry.PushRandomImage()` to create a new, clean Image.
    * **Executing notation commands**: Calling `utils.ExecCommandGroup` will execute a batch of commands in the order provided by arguments. 
    The most common scenario can be: 
        ```bash
        notation sign ...
        notation verify ...
        ```
        or 
        ```bash
        notation policy add
        notation policy show
        ```
    * **Executing notation commands in docker**: Calling `utils.ExecCommandGroupInContainer` will launch a docker container and run all commands in it. This is used to test system-level config.
    * **Parallelization**: In order to speed up testing, Ginkgo will launch several processes to run e2e test cases. 
    * **Randomization**: By default, Ginkgo will run specs in a suite in random order. Please make sure the test cases can be runned independently. 
        If the test cases depend on the execution order, consider using [Ordered Containers](https://onsi.github.io/ginkgo/#ordered-containers).


## Setting up
1. Github Actions

    * Please check `Run e2e tests` steps in **workflows/build.yml** for detail.
2. Local environment

    * Install Golang.
    * Install Docker.
    * install Ginkgo CLI.

        `go install github.com/onsi/ginkgo/v2/ginkgo`
    * Clone the repository.
    * Run `./test/e2e/local.sh absolute_path_to_your_notation_binary` in the repository's root directory. It is a simple shell script which only sets some environment variables, builds the notation image and starts the test registry.
    
