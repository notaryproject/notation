# notation plugin

## Description

Use `notation plugin` to manage plugins. See notation [plugin documentation](https://github.com/notaryproject/notaryproject/blob/v1.0.0/specs/plugin-extensibility.md) for more details. The `notation plugin` command by itself performs no action. In order to manage notation plugins, one of the subcommands must be used.

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
  notation plugin install [flags] <--file|--url> <plugin_source>

Flags:
  -d, --debug              debug mode
      --file               install plugin from a file on file system
      --force              force the installation of the plugin
  -h, --help               help for install
      --sha256sum string   must match SHA256 of the plugin source, required when "--url" flag is set
      --url                install plugin from an HTTPS URL. The plugin download timeout is 10m0s
  -v, --verbose            verbose mode

Aliases:
  install, add
```

### notation plugin uninstall

```text
Uninstall a plugin

Usage:
  notation plugin uninstall [flags] <plugin_name>

Flags:
  -d, --debug         debug mode
  -h, --help          help for remove
  -v, --verbose       verbose mode
  -y, --yes           do not prompt for confirmation
Aliases:
  uninstall, remove, rm
```

## Usage

## Install a plugin 

### Install a plugin from file system

Install a Notation plugin from the host file system. `.zip`, `.tar.gz`, and `single plugin executable file` formats are supported. In this scenario, SHA-256 checksum validation is optional.

```shell
$ notation plugin install --file <file_path>
```

Upon successful execution, the plugin is copied to Notation's plugin directory. If the plugin directory does not exist, it will be created. The name and version of the installed plugin are displayed as follows. 

```console
Successfully installed plugin <plugin name>, version <x.y.z>
```

If the entered plugin checksum digest doesn't match the published checksum, Notation will return an error message and will not start installation.

```console
Error: plugin installation failed: plugin sha256sum does not match user input. Expecting <sha256sum>
```

If the plugin version is higher than the existing plugin, Notation will start installation and overwrite the existing plugin.

```console
Successfully updated plugin <plugin name> from version <x.y.z> to <a.b.c>
```

If the plugin version is equal to the existing plugin, Notation will not start installation and return the following message. Users can use a flag `--force` to skip plugin version check and force the installation.

```console
Error: plugin installation failed: plugin <plugin-name> with version <x.y.z> already exists.
```

If the plugin version is lower than the existing plugin, Notation will return an error message and will not start installation. Users can use a flag `--force` to skip plugin version check and force the installation.

```console
Error: plugin installation failed: failed to install plugin <plugin-name>. The installing plugin version <a.b.c> is lower than the existing plugin version <x.y.z>.
It is not recommended to install an older version. To force the installation, use the "--force" option.
```
### Install a plugin from URL

Install a Notation plugin from a URL. Notation only supports HTTPS URL, which means that the URL must start with "https://". The URL MUST point to a resource in `.zip`, `.tar.gz`, or `single plugin executable file` format.  In this scenario, the SHA-256 checksum of the resource MUST be provided.

```shell
$ notation plugin install --sha256sum <digest> --url <HTTPS_URL>
```

### Uninstall a plugin

```shell
notation plugin uninstall <plugin_name>
```

Upon successful execution, the plugin is uninstalled from the plugin directory. 

```shell
Are you sure you want to uninstall plugin "<plugin name>"? [y/n] y
Successfully uninstalled plugin <plugin_name> 
```

Uninstall the plugin without prompt for confirmation.

```shell
notation plugin uninstall <plugin_name> --yes
```

If the plugin is not found, an error is returned showing the syntax for the plugin list command to show the installed plugins.

```shell
Error: unable to find plugin <plugin_name>. 
To view a list of installed plugins, use `notation plugin list`
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
