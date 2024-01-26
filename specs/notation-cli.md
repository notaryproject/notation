# Notation CLI

This spec contains reference information on using notation commands. Each command has a reference page along with usages.

## Notation Commands

| Command                                     | Description                                                            |
| ------------------------------------------- | ---------------------------------------------------------------------- |
| [blob](./commandline/blob.md)               | Sign, verify and inspect singatures associated with blobs              |
| [certificate](./commandline/certificate.md) | Manage certificates in trust store                                     |
| [inspect](./commandline/inspect.md)         | Inspect OCI signatures                                                 |
| [key](./commandline/key.md)                 | Manage keys used for signing                                           |
| [list](./commandline/list.md)               | List signatures of a signed OCI artifact                               |
| [login](./commandline/login.md)             | Log into OCI registries                                                |
| [logout](./commandline/logout.md)           | Log out from the logged in OCI registries                              |
| [plugin](./commandline/plugin.md)           | Manage plugins                                                         |
| [policy](./commandline/policy.md)           | Manage trust policy configuration for OCI signature verification       |
| [sign](./commandline/sign.md)               | Sign OCI artifacts                                                     |
| [verify](./commandline/verify.md)           | Verify OCI artifacts                                                   |
| [version](./commandline/version.md)         | Print the version of notation CLI                                      |

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
