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

### notation plugin install

```text
Installs a plugin

Usage:
  notation plugin install --name <plugin name> --plugin-package <plugin package> [flags]

Flags:
  -n, --name           string   name of the plugin
  -p, --plugin-package string   path to the plugin package
  -h, --help                    help for install

Aliases:
  install, add
```

### notation plugin remove 

```text
Removes a plugin

Usage:
  notation plugin remove [flags] <plugin name>

Flags:
  -h, --help          help for remove

Aliases:
  remove, rm
```

## Usage

### Install a plugin

```shell
notation plugin install --name <plugin name> --plugin-package <plugin package>
```

Unpon successful execution, the plugins directory is created under the default plugin directory if it does not exist, and the plugin package is extracted to the plugins directory. The plugin installation is then verified. If the verification fails, the plugin is removed from the plugins directory and an error is returned. 

### Uninstall a plugin

```shell
notation plugin remove <plugin name>
```

Upon successful execution, the plugin is removed from the plugins directory. If the plugin is not found, an error is returned showing the syntax for the plugin list command to show the installed plugins.

### List installed plugins

```shell
notation plugin list
```

Upon successful execution, a list of plugins are printed out with information of name, description, version, capabilities and errors. The capabilities show what the plugin is capable of, for example, the plugin can generate signatures or verify signatures. The errors column indicates whether the plugin was installed properly or not. `<nil>` of Error indicates that the plugin installed successfully.

An example of output from `notation plugin list`:

```text
NAME       DESCRIPTION                                   VERSION             CAPABILITIES                ERROR
azure-kv   Sign artifacts with keys in Azure Key Vault   v0.5.0-rc.1     [SIGNATURE_GENERATOR.RAW]   <nil>
```
