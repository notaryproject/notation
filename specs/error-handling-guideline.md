# Notation CLI Error Handling and Message Guideline

This document aims to provide the guidelines for Notation contributors to improve existing error messages and error handling method as well as the new error output format. It will also provide recommendations and examples for Notation CLI contributors for how to write friendly and standard error messages, avoid generating inconsistent and ambiguous error messages.

## General guiding principles

A clear and actionable error message is very important when raising an error, so make sure your error message describes clearly what the error is and tells users what they need to do if possible.

First and foremost, make the error messages descriptive and informative. Error messages are expected to be helpful to troubleshoot where the user has done something wrong and the program is guiding them in the right direction. A great error message is recommended to contain the following elements:

- HTTP status code: optional, when the logs are generated from the server side, it's recommended to print the HTTP status code in the error description.
- Error description: describe what the error is.
- Suggestion: for those errors that have potential solution, print out the recommended action. Versioned troubleshooting document link is nice to have.

Second, when necessary, it is highly suggested for Notation CLI contributors to provide recommendations for users how to resolve the problems based on the error messages they encountered. Showing descriptive words and straightforward prompt with executable commands as a potential solution is a good practice for error messages.

Third, for unhandled errors you didn't expect the user to run into. For that, have a way to view full traceback information as well as full debug or verbose logs output, and instructions on how to submit a bug.

Fourth, signal-to-noise ratio is crucial. The more irrelevant output you produce, the longer it's going to take the user to figure out what they did wrong. If your program produces multiple errors of the same type, consider grouping them under a single explanatory header instead of printing many similar-looking lines.

Fifth, CLI program termination should follow the standard [Exit Status conventions](https://www.gnu.org/software/libc/manual/html_node/Exit-Status.html) to report execution status information about success or failure. Notation returns `EXIT_FAILURE` if and only if Notation reports one or more errors.

Last, error logs can also be useful for post-mortem debugging and can also be written to a file, truncate them occasionally and make sure they don't contain ansi color codes.

## Error output recommendation

### Dos

- Provide full description if the user input does not match what Notation CLI expected. A full description should include the actual input received from the user and expected input.
- Use the capital letter ahead of each line of any error message.
- Print human readable error message. If the error message is mainly from the server and varies by different servers, tell users that the error response is from server. This implies that users may need to contact server side for troubleshooting.
- Provide specific and actionable prompt message with argument suggestion or show the example usage for reference. (e.g, Instead of showing flag or argument options is missing, please provide available argument options and guide users to `--help` to view more examples).
- If the actionable prompt message is too long to show in the CLI output, consider guide users to Notation user manual or troubleshooting guide with the versioned permanent link.
- If the error message is not enough for troubleshooting, guide users to use `--verbose` to print much more detailed logs.
- If server returns an error without any [message or detail](https://github.com/opencontainers/distribution-spec/blob/v1.1.0/spec.md#error-codes), consider providing customized error logs to make it clearer. The original server logs can be displayed in debug mode.
- As a security tool, `notation` SHOULD prompt users to stop upon verification errors. 

### Don'Ts

- Do not use a formula-like or a programming expression in the error message. (e.g, `json: cannot unmarshal string into Go value of type map[string]map[string]string.`, or `Parameter 'xyz' must conform to the following pattern: '^[-\\w\\._\\(\\)]+$'`).
- Do not use ambiguous expressions which mean nothing to users. (e.g, `Something unexpected happens`, or `Error: accepts 2 arg(s), received 0`).
- Do not print irrelevant error message to make the output noisy. The more irrelevant output you produce, the longer it's going to take the user to figure out what they did wrong.
- As a security tool, `notation` MUST NOT prompt users to update the trust policy to bypass the verification error when verification fails.
- Signing key information is considered sensitive so it's not recommended to print them out in the error logs. 

## How to write friendly error message

### Recommended error message structure

Here is a sample structure of an error message:

```text
{Error|Error response from a remote registry or a server}: {Error description (HTTP status code can be printed out if any)}
[Usage: {Command usage}]
[{Recommended action}]
```

- HTTP status code is an optional information. Printed out the HTTP status code if the error message is generated from the server side. 
- Command usage is also an optional information but it's recommended to be printed out when user input doesn't follow the standard usage or examples.
- Recommended action (if any) should follow the general guiding principles described above.

## Reference

Parts of the content are borrowed from these guidelines:

- [Command Line Interface Guidelines](https://clig.dev/#errors)
- [ORAS CLI Error Handling Guideline](https://github.com/oras-project/oras/blob/v1.2.0/docs/proposals/error-handling-guideline.md)
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)