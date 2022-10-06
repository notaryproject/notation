# notation key

## Description

Use ```notation key``` command to manage keys used for signing. User can add/update/list/remove key to/from signing key list. Please be noted this command doesn't manage the lifecycle of signing key itself, it manages the signing key list only.

## Outline

### notation key command

```text
Manage keys used for signing

Usage:
  notation key [command]

Available Commands:
  add         Add key to signing key list
  list        List keys used for signing
  remove      Remove key from signing key list
  update      Update key in signing key list

Flags:
  -h, --help   help for key
```

### notation key add

```text
Add key to signing key list

Usage:
  notation key add <key_name> [key_path cert_path] [flags]

Flags:
  -d, --default               mark as default
  -h, --help                  help for add
      --id string             key id (required if --plugin is set)
  -p, --plugin string         signing plugin name
      --plugin-config string  list of comma-separated {key}={value} pairs that are passed as is to the plugin, refer plugin documentation to set appropriate values
```

### notation key list

```text
List keys used for signing

Usage:
  notation key list [flags]

Aliases:
  list, ls

Flags:
  -h, --help   help for list

```

### notation key remove

```text
Remove key from signing key list

Usage:
  notation key remove <key_name>... [flags]

Aliases:
  remove, rm

Flags:
  -h, --help   help for remove
```

### notation key update

```text
Update key in signing key list

Usage:
  notation key update <key_name> [flags]

Aliases:
  update, set

Flags:
  -d, --default   mark as default
  -h, --help      help for update

```

## Usage

### Add a local key to signing key list and make it as default signing key

```shell
notation key add --default <key_name> <key_file_path> <cert_file_path> 
```

Upon successful adding, a key name is printed out for added signing key with additional info "marked as default".

### Add a default signing key referencing the key identifier for the remote key, and the plugin associated with it

```shell
notation key add --default --plugin <plugin_name> --id <remote_key_id> <key_name>
```

Upon successful adding, a key name is printed out for added signing key with additional info "marked as default".

### Update the default signing key

```shell
notation key update --default <key_name>
```

Upon successful update, the supplied key name is printed out with additional info "marked as default".

### List signing keys

```text
notation key list
```

Upon successful execution, a list of keys is printed out with information of name, key path, certificate path, key id and plugin name. The default signing key name is preceded by an asterisk. The key id and plugin name are used together to provide the information of the key identifier for the remote key and the plugin associated with it.

### Remove two keys from signing key list

```shell
notation key remove <key_name_1> <key_name_2>
```

Upon successful execution, the names of removed signing keys are printed out. Please be noted if default signing key is removed, Notation will not automatically assign a new default signing key. User needs to update the default signing key explicitly.
