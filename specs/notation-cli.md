# Notation CLI

This spec contains reference information on using notation commands. Each command has a reference page along with usages.

## Notation Commands

| Command                                     | Description                                                            |
| ------------------------------------------- | ---------------------------------------------------------------------- |
| [certificate](./commandline/certificate.md) | Manage certificates in trust store                                     |
| [inspect](./commandline/inspect.md)         | Inspect signatures                                                     |
| [key](./commandline/key.md)                 | Manage keys used for signing                                           |
| [list](./commandline/list.md)               | List signatures of the signed artifact                                 |
| [login](./commandline/login.md)             | Login to registries                                                    |
| [logout](./commandline/logout.md)           | Log out from the logged in registries                                  |
| [plugin](./commandline/plugin.md)           | Manage plugins                                                         |
| [policy](./commandline/policy.md)           | [Preview] Manage trust policy configuration for signature verification |
| [sign](./commandline/sign.md)               | Sign artifacts                                                         |
| [verify](./commandline/verify.md)           | Verify artifacts                                                       |
| [version](./commandline/version.md)         | Print the version of notation CLI                                      |

## Notation Outline

```text
Notation - Notary V2 - a tool to sign and verify artifacts

Usage:
  notation [command]

Available Commands:
  certificate Manage certificates in trust store
  inspect     Inspect all signatures associated with the signed artifact
  key         Manage keys used for signing
  list        List signatures of the signed artifact
  login       Login to registry
  logout      Log out from the logged in registries
  plugin      Manage plugins
  policy      [Preview] Manage trust policy configuration for signature verification
  sign        Sign artifacts
  verify      Verify artifacts
  version     Show the notation version information

Flags:
  -h, --help         Help for notation
```
