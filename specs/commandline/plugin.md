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
  list        List installed plugins
  install     Installs a plugin
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
  notation plugin install [flags] <plugin_source>

Flags:
  -h, --help                    help for install
  -f, --force                   force the installation of a plugin
      --checksum string         must match SHA256 of the plugin source
      --source string           the location of plugin installation file, options: "file", "url","registry" (default "file")                  
Aliases:
  install, add
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
  remove, rm, uninstall
```

## Usage

## Install a plugin 

### Install a plugin from file system

```shell
$ notation plugin install --file <file_path> --checksum <digest>
```

Upon successful execution, the plugin is copied to Notation's plugin directory. The name and version of the installed plugin is displayed as follows. 

```console
Successfully installed plugin <plugin name>, version <x.y.z>
```

If the plugin directory does not exist, it will be created. When an existing plugin is detected and the version is the same as the installing plugin, it fails to install and returns the error as follows. Users can use a flag `--force` to skip existence check and force the installation with a specified version.

```console
Error: failed to install the plugin, <plugin name> already installed
```

### Install a plugin from URL

```shell
$ notation plugin install --source <URL> --checksum <digest>
```

### Install a plugin as an OCI artifact from a registry (for future iteration)

```shell
$ notation plugin install --source registry <artifact_reference>
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
Error: <plugin_name> does not exist
```

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
