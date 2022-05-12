# Notation Directory Structure

The `notation` CLI requires local file systems support for the following components across various platforms.

- `notation` binary
- Plugins
- Configurations
- Trust stores
- Trust policies
- Signature caches
- Local key store

This documentation specifies the recommended directory structure for those components.

## Category

The directories for various components are classified into the following catagories.

| Alias     | Description                                                                         |
| --------- | ----------------------------------------------------------------------------------- |
| `BIN`     | Directory for executable binaries                                                   |
| `LIBEXEC` | Directory for binaries not meant to be executed directly by users' shell or scripts |
| `CONFIG`  | Directory for configurations                                                        |
| `CACHE`   | Directory for user-specific cache                                                   |

Although it is recommended to install `notation` with its plugins and default configurations at the system level, it is possible to install at the user level.

On Unix systems, `notation` follows [Filesystem Hierarchy Standard][FHS] for system level directories and [XDG Base Directory Specification][XDG] for user level directories. On Windows, [Known Folders][KF] and [App Settings][AS] are followed equivalently. On Darwin, [macOS File System][macOS_FS] with [System Integrity Protection][SIP] is followed equivalently. If a file with the same name exists at the system level and the user level, the file at the user level takes over the priority.

### System Level

Default directory paths for various operating systems at system level are specified as below.

| Directory | Unix           | Windows                       | Darwin                         |
| --------- | -------------- | ----------------------------- | ------------------------------ |
| `BIN`     | `/usr/bin`     | `%ProgramFiles%/notation/bin` | `/usr/local/bin`               |
| `LIBEXEC` | `/usr/libexec` | `%ProgramFiles%`              | `/usr/local/lib`               |
| `CONFIG`  | `/etc`         | `%ProgramData%`               | `/Library/Application Support` |

`CACHE` is omitted since it is user specific.

### User Level

Default directory paths for various operating systems at user level are specified as below.

| Directory | Unix               | Windows          | Darwin                          |
| --------- | ------------------ | ---------------- | ------------------------------- |
| `LIBEXEC` | `$XDG_CONFIG_HOME` | `%AppData%`      | `~/Library/Application Support` |
| `CONFIG`  | `$XDG_CONFIG_HOME` | `%AppData%`      | `~/Library/Application Support` |
| `CACHE`   | `$XDG_CACHE_HOME`  | `%LocalAppData%` | `~/Library/Caches`              |

There is no default `BIN` path at user level since the `notation` binary can be put anywhere as long as it in the `PATH` environment variable. Common directories on Unix/Darwin are `~/bin` and `~/.local/bin` where manual `PATH` update by users may be required.

## Structure

The overall directory structure for `notation` is summarized as follows.

```
{BIN}
└── notation
{CONFIG}
└── notation
    ├── config.json
    ├── private
    │   ├── {key-name}.crt
    │   └── {key-name}.key
    └── trust
        ├── policy.json
        └── store
            └── {trust-store-type}
                └── {named-store}
                    └── {cert-file}
{CACHE}
└── notation
    └── signature
        └── {manifest-digest-algorithm}
            └── {manifest-digest}
                └── {signature-digest-algorithm}
                    └── {signature-digest}.sig
{LIBEXEC}
└── notation
    └── plugins
        └── {plugin-name}
            └── notation-{plugin-name}
```

### Notation Binary

The path for the `notation` binary is

```
{BIN}/notation
```

### Plugin

Plugin are binaries not meant to be executed directly by users' shell or scripts. The path of a plugin follows the pattern below.

```
{LIBEXEC}/notation/plugins/{plugin-name}/notation-{plugin-name}
```

### General Configuration

The path of the general configuration file of the `notation` CLI is

```
{CONFIG}/notation/config.json
```

### Trust Store

The path of a certificate file in a [Trust Store][TS] follows the pattern of

```
{CONFIG}/notation/trust/store/{trust-store-type}/{named-store}/{cert-file}
```

### Trust Policy

The path of the [Trust Policy][TP] file is

```
{CONFIG}/notation/trust/policy.json
```

### Signature Caches

The signatures are cached to optimize the network traffic. The path of a cached signature for a certain manifest follows the pattern below.

```
{CACHE}/notation/signature/{manifest-digest-algorithm}/{manifest-digest}/{signature-digest-algorithm}/{signature-digest}.sig
```

or in a hierarchical view

```
{CACHE}
└── notation
    └── signature
        └── {manifest-digest-algorithm}
            └── {manifest-digest}
                └── {signature-digest-algorithm}
                    └── {signature-digest}.sig
```

### Local key store

Developers sign artifacts using local private keys with associated certificate chain. The default directory structure for testing purpose is suggested as follows.

```
{CONFIG}/notation/private/{key-name}.crt
{CONFIG}/notation/private/{key-name}.key
```

Developers SHOULD consider safer places to store the passphrase-protected key and certificate pair, or opt to remote signing.

[References]::

[FHS]: https://refspecs.linuxfoundation.org/fhs.shtml "Filesystem Hierarchy Standard"
[XDG]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html "XDG Base Directory Specification"
[KF]: https://docs.microsoft.com/windows/win32/shell/knownfolderid "Known Folders"
[AS]: https://docs.microsoft.com/windows/apps/design/app-settings/store-and-retrieve-app-data "App Settings"
[macOS_FS]: https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/FileSystemOverview/FileSystemOverview.html#//apple_ref/doc/uid/TP40010672-CH2-SW14 "macOS File System"
[SIP]: https://support.apple.com/HT204899 "System Integrity Protection"
[TS]: https://github.com/notaryproject/notaryproject/blob/main/trust-store-trust-policy-specification.md#trust-store "Trust Store"
[TP]: https://github.com/notaryproject/notaryproject/blob/main/trust-store-trust-policy-specification.md#trust-policy "Trust Policy"
