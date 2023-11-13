# notation plugin

## Description

Use `notation plugin` to manage plugins. See notation [plugin documentation](https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md) for more details. The `notation plugin` command by itself performs no action. In order to manage notation plugins, one of the subcommands must be used.

## Outline

### notation plugin

```text
Manage plugins

Usage:
  notation plugin [command]

Available Commands:
  install     Install a plugin
  list        List installed plugins
  uninstall   Uninstall a plugin
  upgrade     Upgrade a plugin

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
Install a plugin

Usage:
  notation plugin install [flags] <plugin_source>

Flags:
  -h, --help                    help for install
  -f, --force                   force the installation of a plugin
      --sha256sum string         must match SHA256 of the plugin source               

Aliases:
  install, add
```

### notation plugin upgrade

```text
Upgrade a plugin

Usage:
  notation plugin upgrade [flags] <plugin_source>

Flags:
  -h, --help                     help for upgrade
      --plugin-name string       plugin name 
      --plugin-version           plugin version           
      --sha256sum string         must match SHA256 of the plugin source                 
```

### notation plugin uninstall

```text
Uninstall a plugin

Usage:
  notation plugin uninstall [flags] <plugin_name>

Flags:
  -h, --help          help for remove
  -y, --yes           do not prompt for confirmation
Aliases:
  uninstall, remove, rm
```

## Usage

## Install a plugin 

### Install a plugin from file system

Install a Notation plugin from file system. The checksum validation is optional for this case.

```shell
$ notation plugin install <file_path>
```

Upon successful execution, the plugin is copied to Notation's plugin directory. The name and version of the installed plugin is displayed as follows. 

```console
Successfully installed plugin <plugin name>, version <x.y.z>
```

If the plugin directory does not exist, it will be created. When an existing plugin is detected, it fails to install and returns the error as follows. Users can use a flag `--force` to skip existence check and force the installation.

```console
Error: failed to install the plugin, <plugin_name> already installed. 
To view a list of installed plugins, use "notation plugin list".
To force the installation, use a flag `--force`.
```

If the entered plugin checksum digest doesn't match the published checksum, Notation will return an error message and will not start installation.

```console
Error: failed to install the plugin, input checksum does not match the published checksum, expected <digest>
```

### Install a plugin from URL

Install a Notation plugin from a remote shared address and verify the plugin checksum. Notation only supports installing plugins from an HTTPS URL.

```shell
$ notation plugin install --sha256sum <digest> <URL>
```

### Install a plugin as an OCI artifact from a registry (for future iteration)

Install a Notation plugin from a registry. Users can verify the plugin's signature with `notation verify` before the plugin installation.

```shell
$ notation plugin install <registry>/<repository>@<digest>
```

### Upgrade a plugin to a higher version from file system 

Upgrade a Notation plugin to a higher version from file system and verify the plugin checksum.

```shell
$ notation plugin upgrade <file_path>
```

Upon successful execution, the plugin is copied to Notation's plugin directory. The name and version of the installed plugin is displayed as follows. 

```console
Successfully upgraded plugin <plugin name> to version <x.y.z>
```

If the upgrade version is equal to or lower than an existing plugin, Notation will return an error message and will not start upgrade.

```console
Error: failed to upgrade the plugin, <plugin name> version should be higher than <x.y.z>
```

If the plugin does not exist, Notation will return an error message and will not start upgrade.

```console
Error: failed to upgrade the plugin, <plugin name> does not exist.
To install a plugin, use "notation plugin install".
```

### Upgrade a plugin from URL

When upgrading a Notation plugin from GitHub release page, Notation upgrades the plugin to the latest version by default.  

```
$ notation plugin upgrade --plugin-name <plugin-name>
```

When upgrading a Notation plugin from GitHub release page, users can also specify a plugin version to upgrade.

```
$ notation plugin upgrade --plugin-version <plugin-version>
```

Upgrade a Notation plugin from a remote shared address (e.g, object storage) and verify the plugin checksum. Notation only supports upgrade a plugin from an HTTPS URL.

```shell
$ notation plugin upgrade --sha256sum <digest> <URL>
```

### Upgrade a plugin as an OCI artifact from a registry (for future iteration)

Upgrade a Notation plugin from a registry. Users can verify the plugin's signature with `notation verify` before the plugin installation.

```shell
$ notation plugin upgrade --plugin-name <plugin-name>
```

### Uninstall a plugin

```shell
notation plugin uninstall <plugin_name>
```

Upon successful execution, the plugin is uninstalled from the plugin directory. 

```shell
Are you sure you want to uninstall plugin "<plugin name>"? [y/N] y
Successfully uninstalled <plugin_name> 
```

Uninstall the plugin without prompt for confirmation.

```shell
notation plugin uninstall <plugin_name> --yes
```

If the plugin is not found, an error is returned showing the syntax for the plugin list command to show the installed plugins.

```shell
Error: unable to find plugin <plugin_name>. 
To view a list of installed plugins, use "notation plugin list".
```

### List installed plugins

```shell
notation plugin list
```

Upon successful execution, a list of plugins are printed out with information of name, description, version, capabilities and errors. The capabilities show what the plugin is capable of, for example, the plugin can generate signatures or verify signatures. The errors column indicates whether the plugin was installed properly or not. `<nil>` of Error indicates that the plugin installed successfully.

An example of output from `notation plugin list`:

```text
NAME                                   DESCRIPTION                                   VERSION       CAPABILITIES                                                                                            ERROR
azure-kv                               Sign artifacts with keys in Azure Key Vault   v1.0.0        Signature generation                                                                                <nil>
com.amazonaws.signer.notation.plugin   AWS Signer plugin for Notation                1.0.290       Signature envelope generation, Trusted Identity validation, Certificate chain revocation check   <nil>
```
