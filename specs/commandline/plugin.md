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
  install     Installs a plugin
  remove      Removes a plugin

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
  notation plugin install [flags] <plugin path>

Flags:
  -h, --help                    help for install
  -f, --force                   force the installation of a plugin

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
  remove, rm, uninstall, delete
```

## Usage

### Install a plugin

```shell
notation plugin install <plugin path>
```

Upon successful execution, the plugin is copied to plugins directory and name+version of plugin is displayed. If the plugin directory does not exist, it will be created. When an existing plugin is detected, the versions are compared and if the existing plugin is a lower version then it is replaced by the newer version.

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
NAME                                   DESCRIPTION                                   VERSION       CAPABILITIES                                                                                            ERROR
azure-kv                               Sign artifacts with keys in Azure Key Vault   v0.5.0-rc.1   [SIGNATURE_GENERATOR.RAW]                                                                                <nil>
com.amazonaws.signer.notation.plugin   AWS Signer plugin for Notation                1.0.290       [SIGNATURE_GENERATOR.ENVELOPE SIGNATURE_VERIFIER.TRUSTED_IDENTITY SIGNATURE_VERIFIER.REVOCATION_CHECK]   <nil>
```
