# Notation CLI

This spec contains reference information on using notation commands. Each command has a reference page along with usages.

## Notation Commands

| Command                                     | Description                            |
| ------------------------------------------- | -------------------------------------- |
| [certificate](./commandline/certificate.md) | Manage certificates in trust store     |
| [key](./commandline/key.md)                 | Manage keys used for signing           |
| [list](./commandline/list.md)               | List signatures of the signed artifact |
| [login](./commandline/login.md)             | Login to registries                    |
| [logout](./commandline/logout.md)           | Log out from the logged in registries  |
| [plugin](./commandline/plugin.md)           | Manage plugins                         |
| [sign](./commandline/sign.md)               | Sign artifacts                         |
| [verify](./commandline/verify.md)           | Verify artifacts                       |
| [version](./commandline/version.md)         | Print the version of notation CLI      |

## Notation Outline

```text
Notation - Notary V2 - a tool to sign and verify artifacts

Usage:
  notation [command]

Available Commands:
  certificate Manage certificates in trust store
  key         Manage keys used for signing
  list        List signatures of the signed artifact
  login       Login to registries
  logout      Log out from the logged in registries
  plugin      Manage plugins
  sign        Sign artifacts
  verify      Verify artifacts
  version     Print the version of notation CLI

Flags:
  -h, --help         Help for notation
```
