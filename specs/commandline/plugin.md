# notation plugin

## Description

Use `notation plugin` to manage the lifecycle of plugins. See notation [plugin documentation](https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details. The `notation plugin` command by itself performs no action. In order to operate on a plugin, one of the subcommands must be used. The current available subcommand is `notation plugin list`.

## Outline

### notation plugin

```text
Manage the lifecycle of plugins

Usage:
  notation plugin <command>

Available Commands:
  list        List registered plugins

Flags:
  -h, --help   help for plugin
```

### notation plugin list

```text
List installed plugins

Usage:
  notation plugin list [flags]

Aliases:
  list, ls

Flags:
  -h, --help   help for list

Global Flags:
      --plain-http   Registry access via plain HTTP
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

Upon successful execution, a list of plugins are printed out with information of name, description, version, capabilities and errors. The capabilities show what the plugin is capable of, for example, the plugin can generate signatures or verify signatures. The information of errors indicates whether the plugin installed properly. `<nil>` of Error indicates the plugin installed successfully.
