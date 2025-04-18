# Improve Notation Diagnostic Experience

> [!NOTE]
> The version of this specification is for `notation` v2.0.0 Beta.1. It is subject to change until v2.0.0 is released. 

## Overview

Notation v1.x offers two global options, `--verbose` and `--debug`, which allows users to generate verbose output and debug logs respectively. These features facilitate both users and developers in inspecting `notation`'s performance, interactions with external services and internal systems, and in diagnosing issues by providing a clear picture of the tool's operations.

This proposal document aims to:

1. Identify the usability issues of the `--verbose` and `--debug` options.
2. Clarify the concepts of different types of output and logs for diagnostic purposes.
3. List the guiding principles to write comprehensive, clear, and conducive debug output and debug logs for effective diagnosis.
4. Propose solutions to improve the diagnostic experience for `notation` CLI users and developers.

## Problem Statement

Given the diverse roles and scenarios in which `notation` CLI is utilized, we have received feedback from users and developers to improve the diagnostic experience as described in [GitHub issue #1247](https://github.com/notaryproject/notation/issues/1247). 

Specifically, there are known issues or confusion when diagnosing a problem.

- The user is confused about when to use `--verbose` and `--debug`. 
- Both `--verbose` and `--debug` output the INFO level logs which is duplicated information. 
- Poor readability of debug logs. No separator lines between request and response information.
- Timestamp does not use [Nanoseconds](https://pkg.go.dev/time#Duration.Nanoseconds) precision, which is not accurate to trace historical operation.
- No easy way to get the user environment information. This causes a higher cost to reproduce the issues for `notation` developers.

## Scenarios

Alice is a DevSecOps engineer who uses `notation` CLI to sign artifacts in a CI/CD pipeline. The signing step failed for some reasons with debug logs generated. Alice tries to figure out the root cause by reading the debug logs but she can't locate the problem. In this scenario, she has to submit a GitHub issue to request help from the Notary Project community and provide debug logs. The debug logs will be used by `notation` developers to analyze the operations and locate the root cause. `notation` developers need to reproduce the failed steps for troubleshooting. 

## Concepts

There are differences between output and logs in `notation`:

### Logs

Logs focus on providing technical details for in-depth diagnosing and troubleshooting issues. It is intended for developers or technical users who need to understand the inner workings of the tool. Debug logs are detailed and technical, often including HTTP request and response from interactions between client and server, timestamps, as well as code-specific information. In general, there are different levels of logs. [Logrus](https://github.com/sirupsen/logrus) is the logging framework used by `notation`, which provides seven logging levels: `Trace`, `Debug`, `Info`, `Warning`, `Error`, `Fatal` and `Panic`. Only `Debug`, `Info`, `Warning`, and `Error` are used by `notation` debug logs. 

- **Purpose**: Debug logs are specifically aimed to facilitate `notation` developers to diagnose `notation` tool itself. They contain detailed technical information that is useful for troubleshooting problems.
- **Target users**: Primarily intended for developers or technical users who are trying to understand the inner workings of the code and identify the root cause of a possible issue with the tool itself.
- **Content**: Debug logs focus on providing context needed to troubleshoot issues, like variable values, execution paths, error stack traces, and internal states of the application.
- **Level of Detail**: Extremely detailed, providing insights into the application's internal workings and logic, often including low-level details that are essential for debugging.

Currently, the verbose output of `notation` prints INFO level of logs, which is overlapped with debug logs. This is duplicated information for users.

### Output

There are four types of output in `notation` CLI:

- **Status output**: such as operation progress information and command execution result.
- **Metadata output**: showing what has been executed in specified format, such as JSON, text.
- **Content output**: it is to output the raw data obtained from the remote registry server or file system, such as the generated signature file.
- **Error output**: error messages are expected to be helpful to troubleshoot where the user has done something wrong and the program is guiding them in the right direction.

The target users of these types of output are general users. 

## Proposals

### Common Conventions

By defining the common conventions, it helps `notation` print out clear and analyzable debug logs.

- Timestamp Each Log Entry with precise timing: Ensure each log entry has a precise timestamp to trace the sequence of events accurately. `notation` SHOULD use the [Nanoseconds](https://pkg.go.dev/time#Duration.Nanoseconds) precision to print the timestamp in the first field of each line. Example: `[2024-08-02 23:56:02.6738192Z] `
- Avoid logging sensitive information for privacy and security requirement: Abstain from logging sensitive information such as passwords, personal data, or authentication tokens. Example: `[2024-08-02 23:56:02.7338192Z] Attempting to authenticate user [UserID: usr123]` (`notation` SHOULD exclude authentication token and password information).

### Enhancements

Here are the proposals for `notation` diagnostic experience enhancements: 

- Deprecate the `--verbose` flag but keep `--debug` flag to avoid ambiguity and duplicated INFO level logs in two outputs. It is reasonable to continue using `--debug` to enable logs with different levels as it is in `notation`.
- Add two empty lines as the separator between each request and response for readability.
- Use the [Nanoseconds](https://pkg.go.dev/time#Duration.Nanoseconds) precision to print the timestamp for each request and response at the beginning.
- Debug log level SHOULD be colored-coded on terminal for better readability 
- Show running environment details of `notation` such as `OS/Arch` in the output of `notation version`. It would be helpful to help the notation developers locate and reproduce the issue easier. 

These proposals are applicable for all `notation` commands. This document uses the debug log of `notation sign` as an example below.

### Example

Current debug logs of `notation sign` command:

```bash
$ notation sign ghcr.io/notaryproject/hello-world:v1 --debug
```

```console
DEBU[2025-04-17T21:44:58-07:00] Request #0
> Request: "HEAD" "https://ghcr.io/notaryproject/v2/hello-world/manifests/v1"
> Request headers:
   "Accept": "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json, application/vnd.oci.artifact.manifest.v1+json"
   "User-Agent": "notation/2.0.0-alpha.1" 
DEBU[2025-04-17T21:44:58-07:00] Response #0
< Response status: "401 Unauthorized"
< Response headers:
   "Date": "Fri, 18 Apr 2025 04:44:58 GMT"
   "Content-Type": "application/json; charset=utf-8"
   "Access-Control-Expose-Headers": "Docker-Content-Digest, WWW-Authenticate, Link, X-Ms-Correlation-Request-Id"
   "X-Content-Type-Options": "nosniff"
   "Strict-Transport-Security": "max-age=31536000; includeSubDomains, max-age=31536000; includeSubDomains"
   "Www-Authenticate": "Bearer realm=\"https://ghcr.io/notaryproject/oauth2/token\",service=\"ghcr.io/notaryproject\",scope=\"repository:hello-world:pull\""
   "X-Ms-Correlation-Request-Id": "0ce5e379-fa9b-4f6c-b064-65664e7931d3"
   "Content-Length": "205"
   "Connection": "keep-alive"
   "Docker-Distribution-Api-Version": "registry/2.0"
   "Server": "ghcr.io" 
DEBU[2025-04-17T21:44:58-07:00] started executing credential helper program docker-credential-desktop with action get 
DEBU[2025-04-17T21:44:59-07:00] successfully finished executing credential helper program docker-credential-desktop with action get 
DEBU[2025-04-17T21:44:59-07:00] Request #1
> Request: "GET" "https://ghcr.io/notaryproject/oauth2/token?scope=repository%3Ahello-world%3Apull&service=ghcr.io/notaryproject"
> Request headers:
   "Authorization": "*****"
   "User-Agent": "notation/2.0.0-alpha.1" 
DEBU[2025-04-17T21:44:59-07:00] Response #1
< Response status: "200 OK"
< Response headers:
   "X-Ms-Correlation-Request-Id": "7ba65cb0-a9a7-436c-a0c8-863adc26e2ea"
   "X-Ms-Ratelimit-Remaining-Calls-Per-Second": "166.65"
   "Strict-Transport-Security": "max-age=31536000; includeSubDomains"
   "Server": "ghcr.io"
   "Date": "Fri, 18 Apr 2025 04:44:59 GMT"
   "Content-Type": "application/json; charset=utf-8"
   "Connection": "keep-alive" 
DEBU[2025-04-17T21:44:59-07:00] Request #2
> Request: "HEAD" "https://ghcr.io/notaryproject/v2/hello-world/manifests/v1"
> Request headers:
   "Accept": "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json, application/vnd.oci.artifact.manifest.v1+json"
   "Authorization": "*****"
   "User-Agent": "notation/2.0.0-alpha.1" 
DEBU[2025-04-17T21:44:59-07:00] Response #2
< Response status: "200 OK"
< Response headers:
   "Etag": "\"sha256:3141b50a3ab08741923d746a2a7cd4c96b6043d72fb37a0f739b842d08469f62\""
   "X-Ms-Correlation-Request-Id": "61fbcac5-2710-469c-abe4-b97f86c62cb3"
   "Content-Type": "application/vnd.oci.image.manifest.v1+json"
   "Connection": "keep-alive"
   "Access-Control-Expose-Headers": "Docker-Content-Digest, WWW-Authenticate, Link, X-Ms-Correlation-Request-Id"
   "Docker-Distribution-Api-Version": "registry/2.0"
   "Strict-Transport-Security": "max-age=31536000; includeSubDomains, max-age=31536000; includeSubDomains"
   "X-Content-Type-Options": "nosniff"
   "Date": "Fri, 18 Apr 2025 04:44:59 GMT"
   "X-Ms-Client-Request-Id": ""
   "X-Ms-Request-Id": "f1e00810-8456-419b-89bb-d499a9c09674"
   "Server": "ghcr.io"
   "Content-Length": "535"
   "Docker-Content-Digest": "sha256:3141b50a3ab08741923d746a2a7cd4c96b6043d72fb37a0f739b842d08469f62" 
INFO[2025-04-17T21:44:59-07:00] Reference v1 resolved to manifest descriptor: {MediaType:application/vnd.oci.image.manifest.v1+json Digest:sha256:3141b50a3ab08741923d746a2a7cd4c96b6043d72fb37a0f739b842d08469f62 Size:535 URLs:[] Annotations:map[] Data:[] Platform:<nil> ArtifactType:} 
Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:v1) because tags are mutable and a tag reference can point to a different artifact than the one signed.
```

Suggested debug logs of `notation sign`:

```
[2025-04-01 23:56:02.6738192Z] DEBUG --> Request #0
> Request: "HEAD" "https://ghcr.io/notaryproject/v2/hello-world/manifests/v1"
> Request headers:
   "Accept": "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json, application/vnd.oci.artifact.manifest.v1+json"
   "User-Agent": "notation/2.0.0-alpha.1" 


[2025-04-01 23:55:04.6738192Z] DEBUG <-- Response #0
< Response status: "401 Unauthorized"
< Response headers:
   "Connection": "keep-alive"
   "X-Content-Type-Options": "nosniff"
   "X-Ms-Correlation-Request-Id": "cbb9bab0-3e2f-4a6c-bfa4-9c55cf57ace3"
   "Server": "ghcr.io"
   "Content-Type": "application/json; charset=utf-8"
   "Content-Length": "205"
   "Docker-Distribution-Api-Version": "registry/2.0"
   "Www-Authenticate": "Bearer realm=\"https://ghcr.io/notaryproject/oauth2/token\",service=\"ghcr.io/notaryproject\",scope=\"repository:hello-world:pull\""
   "Access-Control-Expose-Headers": "Docker-Content-Digest, WWW-Authenticate, Link, X-Ms-Correlation-Request-Id"
   "Strict-Transport-Security": "max-age=31536000; includeSubDomains, max-age=31536000; includeSubDomains"
   "Date": "Fri, 18 Apr 2025 02:55:53 GMT" 
[2025-04-01 23:56:04.2138192Z] DEBUG started executing credential helper program docker-credential-desktop with action get 
[2025-04-01 23:56:04.2238192Z] DEBUG successfully finished executing credential helper program docker-credential-desktop with action get 


[2025-04-01 23:56:04.2738192Z] DEBUG --> Request #1
> Request: "GET" "https://ghcr.io/notaryproject/oauth2/token?scope=repository%3Ahello-world%3Apull&service=ghcr.io/notaryproject"
> Request headers:
   "User-Agent": "notation/2.0.0-alpha.1"
   "Authorization": "*****" 


[2025-04-01 23:56:04.3738192Z] DEBUG <-- Response #1
< Response status: "200 OK"
< Response headers:
   "X-Ms-Correlation-Request-Id": "adbfe683-9f71-4914-aa48-38d030750234"
   "X-Ms-Ratelimit-Remaining-Calls-Per-Second": "166.65"
   "Strict-Transport-Security": "max-age=31536000; includeSubDomains"
   "Server": "ghcr.io"
   "Date": "Fri, 18 Apr 2025 02:55:54 GMT"
   "Content-Type": "application/json; charset=utf-8"
   "Connection": "keep-alive" 


[2025-04-01 23:56:04.4738192Z] DEBUG --> Request #2
> Request: "HEAD" "https://ghcr.io/notaryproject/v2/hello-world/manifests/v1"
> Request headers:
   "Authorization": "*****"
   "User-Agent": "notation/2.0.0-alpha.1"
   "Accept": "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.oci.image.index.v1+json, application/vnd.oci.artifact.manifest.v1+json" 


[2025-04-01 23:56:04.5738192Z] DEBUG <-- Response #2
< Response status: "200 OK"
< Response headers:
   "Etag": "\"sha256:3141b50a3ab08741923d746a2a7cd4c96b6043d72fb37a0f739b842d08469f62\""
   "X-Ms-Client-Request-Id": ""
   "X-Ms-Correlation-Request-Id": "bcc62e7f-48d0-4431-9fbc-d46766358a97"
   "Strict-Transport-Security": "max-age=31536000; includeSubDomains, max-age=31536000; includeSubDomains"
   "Server": "ghcr.io"
   "Content-Type": "application/vnd.oci.image.manifest.v1+json"
   "Content-Length": "535"
   "X-Ms-Request-Id": "486bf410-f68e-4f08-a93d-342ba87b8668"
   "Date": "Fri, 18 Apr 2025 02:55:54 GMT"
   "Docker-Distribution-Api-Version": "registry/2.0"
   "X-Content-Type-Options": "nosniff"
   "Connection": "keep-alive"
   "Access-Control-Expose-Headers": "Docker-Content-Digest, WWW-Authenticate, Link, X-Ms-Correlation-Request-Id"
   "Docker-Content-Digest": "sha256:3141b50a3ab08741923d746a2a7cd4c96b6043d72fb37a0f739b842d08469f62" 


[2025-04-01 23:56:04.6738192Z] INFO Reference v1 resolved to manifest descriptor: {MediaType:application/vnd.oci.image.manifest.v1+json Digest:sha256:3141b50a3ab08741923d746a2a7cd4c96b6043d72fb37a0f739b842d08469f62 Size:535 URLs:[] Annotations:map[] Data:[] Platform:<nil> ArtifactType:} 


Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:v1) because tags are mutable and a tag reference can point to a different artifact than the one signed.

Error: certificate-chain is invalid, certificate with subject "CN=blog-sign,O=Notary,L=Seattle,ST=WA,C=US" was invalid at signing time of 2025-04-01 23:56:04.6738192Z UTC. Certificate is valid from [2025-04-01 23:56:04.6738192Z UTC] to [2025-04-10 23:56:04.6738192Z UTC]
```

### Show user's environment details

Output the system environment of `notation` user could help the `notation` developers reproduce the issue easier. 

For example, the operating system and architecture are supposed to be outputted in `notation version`: 

```bash
$ notation version

Notation - a tool to sign and verify artifacts.

notation Version:  2.0.0-alpha.1
Go version:      go1.24.1
OS/Arch:         linux/amd64
Git commit:      xxxxxxxxxxxx
```