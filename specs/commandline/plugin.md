# notation plugin

## Description

Use `notation plugin` to manage plugins. See notation [plugin documentation](https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details. The `notation plugin` command by itself performs no action. In order to operate on a plugin, one of the subcommands must be used.

## Outline

### notation plugin

```text
Manage plugins

Usage:
  notation plugin [command]

Available Commands:
  list        List installed plugins

Flags:
  -h, --help          help for plugin
```

### notation plugin list

```text
List installed plugins

Usage:
  notation plugin list [flags]

Flags:
  -h, --help          help for list

Aliases:
  list, ls
```

## Usage

### Install a plugin

Currently there is no subcommand available for plugin installation. Plugin publisher should provide instructions to download and install the plugin.

### Uninstall a plugin

Currently there is no subcommand available for plugin un-installation. Plugin publisher should provide instructions to uninstall the plugin.

### List installed plugins

```shell
notation plugin list
```

Upon successful execution, a list of plugins are printed out with information of name, description, version, capabilities and errors. The capabilities show what the plugin is capable of, for example, the plugin can generate signatures or verify signatures. The errors column indicates whether the plugin was installed properly or not. `<nil>` of Error indicates that the plugin installed successfully.

An example of output from `notation plugin list`:

```text
NAME       DESCRIPTION                                   VERSION             CAPABILITIES                ERROR
azure-kv   Sign artifacts with keys in Azure Key Vault   v0.3.1-alpha.1      [SIGNATURE_GENERATOR.RAW]   <nil>
```
