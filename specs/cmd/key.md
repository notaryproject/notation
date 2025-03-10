# notation key

## Description

Use ```notation key``` command to manage keys used for signing. User can add/update/list/remove key to/from Notation signing key list. Please be noted this command doesn't manage the lifecycle of signing key itself, it manages the Notation signing key list only.

## Outline

### notation key command

```text
Manage keys used for signing

Usage:
  notation key [command]

Available Commands:
  add         Add key to Notation signing key list
  delete      Remove key from Notation signing key list
  list        List keys used for signing
  update      Update key in Notation signing key list

Flags:
  -h, --help  help for key
```

### notation key add

```text
Add key to Notation signing key list

Usage:
  notation key add --plugin <plugin_name> [flags] <key_name>

Flags:
  -d, --debug                       debug mode
      --default                     mark as default
  -h, --help                        help for add
      --id string                   key id (required if --plugin is set)
      --plugin string               signing plugin name
      --plugin-config stringArray   {key}={value} pairs that are passed as it is to a plugin, refer plugin's documentation to set appropriate values
  -v, --verbose                     verbose mode
```

### notation key delete

```text
Remove key from Notation signing key list

Usage:
  notation key delete [flags] <key_name>...

Flags:
  -d, --debug     debug mode
  -h, --help      help for delete
  -v, --verbose   verbose mode
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

### notation key update

```text
Update key in Notation signing key list

Usage:
  notation key update [flags] <key_name>

Aliases:
  update, set

Flags:
  -d, --debug     debug mode
      --default   mark as default
  -h, --help      help for update
  -v, --verbose   verbose mode
```

## Usage

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

### Remove a specified key from Notation signing key list

```shell
notation key delete <key_name>
```

- Upon successful removal of a local testing key created by notation, the output message is printed out as follows:

```text
Removed <key_name> from Notation signing key list. The source key still exists.
```
- Upon successful removal of a key associated with a KMS, the output message is printed out as follows:

```text
Removed <key_name> from Notation signing key list. The source key still exists.
```

- Upon successful removal of the default key, the output message is printed out as follows:

```text
Removed default key <key_name> from Notation signing key list. The source key still exists.
```

### Remove two keys from Notation signing key list

```shell
notation key delete <key_name_1> <key_name_2>
```

Upon successful execution, the output message is printed out as follows. Please be noted if default signing key is removed, Notation will not automatically assign a new default signing key. User needs to update the default signing key explicitly.

```text
Removed the following keys from Notation signing key list. The source keys still exist.
<key_name_1>
<key_name_2> (default)
```