# Notation CLI

This spec contains reference information on using notation commands. Each command has a reference page along with usages.

## Notation Commands

| Command                                     | Description                                                                   |
| ------------------------------------------- | ----------------------------------------------------------------------------- |
| [blob](./cmd/blob.md)               | Sign, verify and inspect signatures associated with blobs                     |
| [certificate](./cmd/certificate.md) | Manage certificates in trust store                                            |
| [inspect](./cmd/inspect.md)         | Inspect OCI signatures                                                        |
| [key](./cmd/key.md)                 | Manage keys used for signing                                                  |
| [list](./cmd/list.md)               | List signatures of a signed OCI artifact                                      |
| [login](./cmd/login.md)             | Log into OCI registries                                                       |
| [logout](./cmd/logout.md)           | Log out from the logged in OCI registries                                     |
| [plugin](./cmd/plugin.md)           | Manage plugins                                                                |
| [policy](./cmd/policy.md)           | Manage OCI trust policy configuration for OCI artifact signature verification |
| [sign](./cmd/sign.md)               | Sign OCI artifacts                                                            |
| [verify](./cmd/verify.md)           | Verify OCI artifacts                                                          |
| [version](./cmd/version.md)         | Print the version of notation CLI                                             |

## Notation Outline

```text
Notation - a tool to sign and verify artifacts

Usage:
  notation [command]

Available Commands:
  blob        Sign, verify and inspect signatures associated with blobs
  certificate Manage certificates in trust store
  inspect     Inspect all signatures associated with a signed OCI artifact
  key         Manage keys used for signing
  list        List signatures of a signed OCI artifact
  login       Log into OCI registries
  logout      Log out from the logged in OCI registries
  plugin      Manage plugins
  policy      Manage trust policy configuration for OCI signature verification
  sign        Sign OCI artifacts
  verify      Verify OCI artifacts
  version     Show the notation version information

Flags:
  -h, --help         Help for notation
```
