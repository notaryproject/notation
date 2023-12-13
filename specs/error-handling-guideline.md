# Notation CLI Error Handling and Message Guideline

This document aims to provide the guidelines for Notation contributors to improve existing error messages and error handling method as well as the new error output format. It will also provide recommendations and examples for Notation CLI contributors for how to write friendly and standard error messages, avoid generating inconsistent and ambiguous error messages.

## General guiding principles

A clear and actionable error message is very important when raising an error, so make sure your error message describes clearly what the error is and tells users what they need to do if possible.

First and foremost, make the error messages descriptive and informative. Error messages are expected to be helpful to troubleshoot where the user has done something wrong and the program is guiding them in the right direction. A great error message is recommended to contain the following elements:

- Error code: optional, when the logs are generated from the server side
- Error description: describe what the error is
- Suggestion: for those errors that have potential solution, print out the recommended solution. Versioned troubleshooting document link is nice to have.

Second, when necessary, it is highly suggested for Notation CLI contributors to provide recommendations for users how to resolve the problems based on the error messages they encountered. Showing descriptive words and straightforward prompt with executable commands as a potential solution is a good practice for error messages.

Third, for unhandled errors you didn't expect the user to run into. For that, have a way to view full traceback information as well as full debug or verbose logs output, and instructions on how to submit a bug.

Fourth, signal-to-noise ratio is crucial. The more irrelevant output you produce, the longer it's going to take the user to figure out what they did wrong. If your program produces multiple errors of the same type, consider grouping them under a single explanatory header instead of printing many similar-looking lines.

Fifth, CLI program termination should follow the standard [Exit Status conventions](https://www.gnu.org/software/libc/manual/html_node/Exit-Status.html) to report execution status information about success or failure. 

Last, error logs can also be useful for post-mortem debugging, truncate them occasionally so they don't eat up space on disk, and make sure they don't contain ansi color codes. Thereby, error logs can be written to a file.

## Error output recommendation

### Dos

- Provide full description if the user input does not match what Notation CLI expected. A full description should include the actual input received from the user and expected input
- Use the capital letter ahead of each line of any error message
- Print human readable error message. If the error message is mainly from the server and varies by different servers, tell users that the error response is from server. This implies that users may need to contact server side for troubleshooting.
- Provide specific and actionable prompt message with argument suggestion or show the example usage for reference. (e.g, Instead of showing flag or argument options is missing, please provide available argument options and guide users to "--help" to view more examples)
- If the actionable prompt message is too long to show in the CLI output, consider guide users to Notation user guide or troubleshooting guide with the permanent link.
- If the error message is not enough for troubleshooting, guide users to use "--verbose" to print much more detailed logs

### Don'Ts

- Do not use a formula-like or a programming expression in the error message. (e.g, `json: cannot unmarshal string into Go value of type map[string]map[string]string.`, or `Parameter 'xyz' must conform to the following pattern: '^[-\\w\\._\\(\\)]+$'`)
- Do not use ambiguous expressions which mean nothing to users. (e.g, `Something unexpected happens`, or `Error: accepts 2 arg(s), received 0`)
- Do not print irrelevant error message to make the output noisy. The more irrelevant output you produce, the longer it's going to take the user to figure out what they did wrong.

## How to write friendly error message

### Recommended error message structure

Here is a sample structure of an error message:

```text
Error: [Error code]  [Error description] 
Usage: [Command usage]
[Recommended solution]
```

- Error code is an optional information. If the error message is generated from the server side, it may include error code or [warn code](https://www.rfc-editor.org/rfc/rfc7234#section-5.5). It could be printed out alongside the error description.
- Command usage is also an optional information but it's recommended to be printed out when user input doesn't follow the standard usage or examples. 
- Recommended solution is required and should follow the general guiding principles described above.

### Examples

#### Sign an artifact without artifact reference or signing key

Current behavior and output:

```
$ notation sign --signature-format cose
Error: missing reference
```

Suggested error message:

```
$ notation sign --signature-format cose
Error: missing artifact reference required for signing. 
Please specify an artifact reference. Run "notation sign --help" for more options and examples.
```

#### Sign an artifact with an non-existing signing key in a key vault

Current behavior and output:

```
$ notation sign localhost:5000/test-repo:v1  --signature-format cose --plugin wabbitnetworks-kv --id https://feynman-kv.vault.wabbit.net/keys/feynmankv-networks-io/6670ffa5cb694c49b1e0a6bb6bdefaaa
Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:v1) because tags are mutable and a tag reference can point to a different artifact than the one signed.
Error: describe-key command failed: ERROR: A certificate with (name/id) feynmankv-networks-io/versions/6670ffa5cb694c49b1e0a6bb6bdefaaa was not found in this key vault. If you recently deleted this certificate you may be able to recover it using the correct recovery command. For help resolving this issue, please see https://go.wabbit.net/fwlink/?linkid=2125182
Status: 404 (Not Found)
ErrorCode: CertificateNotFound

Content:
{"error":{"code":"CertificateNotFound","message":"A certificate with (name/id) feynmankv-networks-io/versions/6670ffa5cb694c49b1e0a6bb6bdefaaa was not found in this key vault. If you recently deleted this certificate you may be able to recover it using the correct recovery command. For help resolving this issue, please see https://go.wabbit.net/fwlink/?linkid=2125182"}}

Headers:
Cache-Control: no-cache
Pragma: no-cache
x-ms-keyvault-region: eastus
x-ms-client-request-id: a2923244-ed47-461b-9dc1-d0b9f4202788
x-ms-request-id: 96103d99-c372-449f-adba-8d24b7d5da7e
x-ms-keyvault-service-version: 1.9.1116.1
x-ms-keyvault-network-info: conn_type=Ipv4;addr=20.65.162.175;act_addr_fam=InterNetwork;
X-Content-Type-Options: REDACTED
Strict-Transport-Security: REDACTED
Date: Wed, 13 Dec 2023 07:27:33 GMT
Content-Length: 376
Content-Type: application/json; charset=utf-8
Expires: -1
```

Suggested error message:

```
$ notation sign localhost:5000/test-repo:v1  --signature-format cose --plugin wabbitnetworks-kv --id https://feynman-kv.vault.wabbit.net/keys/feynmankv-networks-io/6670ffa5cb694c49b1e0a6bb6bdefaaa
Warning: Always sign the artifact using digest(@sha256:...) rather than a tag(:v1) because tags are mutable and a tag reference can point to a different artifact than the one signed.
ERROR: A certificate with (name/id) feynmankv-networks-io/versions/6670ffa5cb694c49b1e0a6bb6bdefaaa was not found in this key vault. 
Please make sure the certificate is available in the key vault. Use "--verbose" to see detailed logs.
```

#### Sign an artifact with an error signature format parameter

Current behavior and output:

```
$ notation sign localhost:5000/test-repo:v1  --signature-format cosee
Error: signature format "cosee" not supported
```

Suggested error message:

```
$ notation sign localhost:5000/test-repo:v1  --signature-format dsse
Error: signature format "dsse" not supported
Please use the supported signature envelope format "jws" or "cose"
```

#### When the plugin name doesn't not follow the plugin spec

Current behavior and output:

```
$ notation plugin ls 
NAME       DESCRIPTION   VERSION   CAPABILITIES   ERROR

azure-kv                           []             stat /home/azureuser/.config/notation/plugins/azure-kv/notation-azure-kv: no such file or directory
```

Suggested error message:

```
$ notation plugin ls 
NAME       DESCRIPTION   VERSION   CAPABILITIES   ERROR

azure-kv                           []             Plugin file should follow the naming convention "notation-{plugin-name}"
```

## Reference

Parts of the content are borrowed from these guidelines.

- [Command Line Interface Guidelines](https://clig.dev/#errors)
- [ORAS CLI Error Handling Guideline](https://github.com/oras-project/oras/pull/1163/files)
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)