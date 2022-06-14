# Notation Directory Structure

The `notation` CLI requires local file systems support for the following components across various platforms.

- `notation` binary
- Plugins
- Configurations
- Trust stores
- Trust policies
- Signature caches
- Signing key store

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

On Unix, `$XDG_CONFIG_HOME` is default to `~/.config` and `$XDG_CACHE_HOME` is default to `~/.cache` if XDG environment variables are empty.

There is no default `BIN` path at user level since the `notation` binary can be put anywhere as long as it in the `PATH` environment variable. Common directories on Unix/Darwin are `~/bin` and `~/.local/bin` where manual `PATH` update by users may be required.

## Structure

The overall directory structure for `notation` is summarized as follows.

```
{BIN}
└── notation
{CACHE}
└── notation
    └── signatures
        └── {manifest-digest-algorithm}
            └── {manifest-digest}
                └── {signature-blob-digest-algorithm}
                    └── {signature-digest}.sig
{CONFIG}
└── notation
    ├── config.json
    ├── localkeys
    │   ├── {key-name}.crt
    │   └── {key-name}.pem
    ├── signingkeys.json
    ├── trustpolicy.json
    └── truststore
        └── {trust-store-type}
            └── {named-store}
                └── {cert-file}
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

On Windows, the `.exe` extension is required for executables.

```
{BIN}/notation.exe
```

### Plugin

[Plugins][Plugin] are binaries not meant to be executed directly by users' shell or scripts. The path of a plugin follows the pattern below.

```
{LIBEXEC}/notation/plugins/{plugin-name}/notation-{plugin-name}
```

On Windows, the `.exe` extension is required for executables.

```
{LIBEXEC}/notation/plugins/{plugin-name}/notation-{plugin-name}.exe
```

### General Configuration

The path of the general configuration file of the `notation` CLI is

```
{CONFIG}/notation/config.json
```

### Trust Store

The path of a certificate file in a [Trust Store][TS] follows the pattern of

```
{CONFIG}/notation/truststore/{trust-store-type}/{named-store}/{cert-file}
```

### Trust Policy

The path of the [Trust Policy][TP] file is

```
{CONFIG}/notation/trustpolicy.json
```

### Signature Caches

The signatures are cached to optimize the network traffic. The path of cached signatures for a certain target manifest (e.g. an image manifest) follows the pattern below.

```
{CACHE}/notation/signatures/{manifest-digest-algorithm}/{manifest-digest}/{signature-blob-digest-algorithm}/{signature-digest}.sig
```

or in a hierarchical view

```
{CACHE}
└── notation
    └── signatures
        └── {manifest-digest-algorithm}
            └── {manifest-digest}
                └── {signature-blob-digest-algorithm}
                    └── {signature-digest}.sig
```

### Signing Key Store

Developers sign artifacts using local private keys with associated certificate chain. The signing key information is tracked in a JSON file at

```
{CONFIG}/notation/signingkeys.json
```

Since the signing key store is user-specific, the system level `{CONFIG}` is not recommended. Developers SHOULD consider safe places to store the passphrase-protected key and certificate pairs, or opt to remote signing.

For testing purpose, the following directory structure is suggested.

```
{CONFIG}/notation/localkeys/{key-name}.crt
{CONFIG}/notation/localkeys/{key-name}.pem
```

Since `signingkeys.json` takes references in absolute paths, it is not required to copy the private keys and certificates used for signing to the above directory structure.

## Examples

Examples are shown on various platforms where the user `exampleuser` overrides the `notation` config and the trust policy.

### Unix

```
/
├── etc
│   └── notation
│       ├── config.json
│       ├── trustpolicy.json
│       └── truststore
│           └── x509
│               ├── ca
│               │   ├── acme-rockets
│               │   │   ├── cert1.pem
│               │   │   └── cert2.pem
│               │   └── wabbit-networks
│               │       └── cert3.pem
│               └── tsa
│                   └── publicly-trusted-tsa
│                       └── tsa-cert1.pem
├── home
│   └── exampleuser
│       ├── .cache
│       │   └── notation
│       │       └── signatures
│       │           └── sha256
│       │               └── 05b3abf2579a5eb66403cd78be557fd860633a1fe2103c7642030defe32c657f
│       │                   └── sha256
│       │                       ├── 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.sig
│       │                       ├── b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9.sig
│       │                       └── fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9.sig
│       └── .config
│           └── notation
│               ├── config.json
│               ├── localkeys
│               │   ├── dev.crt
│               │   ├── dev.pem
│               │   ├── test.crt
│               │   └── test.pem
│               ├── plugins
│               │   └── com.example.bar
│               │       └── notation-com.example.bar
│               ├── signingkeys.json
│               ├── trustpolicy.json
│               └── truststore
│                   └── x509
│                       ├── ca
│                       │   └── acme-rockets
│                       │       └── cert4.pem
│                       └── tsa
│                           └── publicly-trusted-tsa
│                               └── tsa-cert2.pem
└── usr
    ├── bin
    │   └── notation
    └── libexec
        └── notation
            └── plugins
                └── com.example.foo
                    └── notation-com.example.foo
```

### Windows

```
C:.
├── Program Files
│   └── notation
│       ├── bin
│       │   └── notation.exe
│       └── plugins
│           └── com.example.foo
│               └── notation-com.example.foo.exe
├── ProgramData
│   └── notation
│       ├── config.json
│       ├── trustpolicy.json
│       └── truststore
│           └── x509
│               ├── ca
│               │   ├── acme-rockets
│               │   │   ├── cert1.pem
│               │   │   └── cert2.pem
│               │   └── wabbit-networks
│               │       └── cert3.pem
│               └── tsa
│                   └── publicly-trusted-tsa
│                       └── tsa-cert1.pem
└── Users
    └── exampleuser
        └── AppData
            ├── Local
            │   └── notation
            │       └── signatures
            │           └── sha256
            │               └── 05b3abf2579a5eb66403cd78be557fd860633a1fe2103c7642030defe32c657f
            │                   └── sha256
            │                       ├── 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.sig
            │                       ├── b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9.sig
            │                       └── fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9.sig
            └── Roaming
                └── notation
                    ├── config.json
                    ├── localkeys
                    │   ├── dev.crt
                    │   ├── dev.pem
                    │   ├── test.crt
                    │   └── test.pem
                    ├── plugins
                    │   └── com.example.bar
                    │       └── notation-com.example.bar.exe
                    ├── signingkeys.json
                    ├── trustpolicy.json
                    └── truststore
                        └── x509
                            ├── ca
                            │   └── acme-rockets
                            │       └── cert4.pem
                            └── tsa
                                └── publicly-trusted-tsa
                                    └── tsa-cert2.pem
```

### Darwin

```
/
├── Library
│   └── Application Support
│       └── notation
│           ├── config.json
│           ├── trustpolicy.json
│           └── truststore
│               └── x509
│                   ├── ca
│                   │   ├── acme-rockets
│                   │   │   ├── cert1.pem
│                   │   │   └── cert2.pem
│                   │   └── wabbit-networks
│                   │       └── cert3.pem
│                   └── tsa
│                       └── publicly-trusted-tsa
│                           └── tsa-cert1.pem
├── Users
│   └── exampleuser
│       └── Library
│           ├── Application Support
│           │   └── notation
│           │       ├── config.json
│           │       ├── localkeys
│           │       │   ├── dev.crt
│           │       │   ├── dev.pem
│           │       │   ├── test.crt
│           │       │   └── test.pem
│           │       ├── plugins
│           │       │   └── com.example.bar
│           │       │       └── notation-com.example.bar
│           │       ├── signingkeys.json
│           │       ├── trustpolicy.json
│           │       └── truststore
│           │           └── x509
│           │               ├── ca
│           │               │   └── acme-rockets
│           │               │       └── cert4.pem
│           │               └── tsa
│           │                   └── publicly-trusted-tsa
│           │                       └── tsa-cert2.pem
│           └── Caches
│               └── notation
│                   └── signatures
│                       └── sha256
│                           └── 05b3abf2579a5eb66403cd78be557fd860633a1fe2103c7642030defe32c657f
│                               └── sha256
│                                   ├── 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae.sig
│                                   ├── b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9.sig
│                                   └── fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9.sig
└── usr
    └── local
        ├── bin
        │   └── notation
        └── lib
            └── notation
                └── plugins
                    └── com.example.foo
                        └── notation-com.example.foo
```

[References]::

[FHS]: https://refspecs.linuxfoundation.org/fhs.shtml "Filesystem Hierarchy Standard"
[XDG]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html "XDG Base Directory Specification"
[KF]: https://docs.microsoft.com/windows/win32/shell/knownfolderid "Known Folders"
[AS]: https://docs.microsoft.com/windows/apps/design/app-settings/store-and-retrieve-app-data "App Settings"
[macOS_FS]: https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/FileSystemOverview/FileSystemOverview.html#//apple_ref/doc/uid/TP40010672-CH2-SW14 "macOS File System"
[SIP]: https://support.apple.com/HT204899 "System Integrity Protection"
[Plugin]: https://github.com/notaryproject/notaryproject/blob/main/specs/plugin-extensibility.md "Notation Extensibility for Signing and Verification"
[TS]: https://github.com/notaryproject/notaryproject/blob/main/trust-store-trust-policy-specification.md#trust-store "Trust Store"
[TP]: https://github.com/notaryproject/notaryproject/blob/main/trust-store-trust-policy-specification.md#trust-policy "Trust Policy"
