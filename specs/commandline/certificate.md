# Use CLI to manage trust store

## Description

Use ```notation certificate``` command to add/list/delete certificates in a trust store. Update an existing certificate is not allowed since the thumbprint will be inconsistent, which results in a new certificate.

The trust store is in the format of a directory in the filesystem as`x509/<type>/<name>/*.crt|*.cer|*.pem`. Currently two types of trust store are supported:

* `Certificate Authority`: The directory name is `ca`.
* `Signing Authority`: The directory name is `signingAuthority`

There could be more trust store types introduced in the future.

Here is an example of trust store directory structure:

```text
$XDG_CONFIG_HOME/notation/trust-store
    /x509
        /ca
            /acme-rockets
                cert1.pem
                cert2.pem
                  /sub-dir       # sub directory is ignored
                    cert-3.pem   # certs under sub directory is ignored

        /signingAuthority
            /wabbit-networks
                cert3.crt
```

In this example, there are two certificates stored in trust store named `acme-rockets` of type `ca`. There is one certificate stored in trust store named `wabbit-networks` of type `signingAuthority`.

Please be noted there wil be user level trust store and system level trust store. See [Directory spec](https://github.com/notaryproject/notation/blob/main/specs/directory.md) for more details. The commands `notation certificate add` and `notation certificate delete` are performed only on user level trust store.

## Outline

### notation certificate

```text
Manage certificates in trust store

Usage:
  notation certificate [command]

Aliases:
  certificate, cert

Available Commands:
  add           Add certificates to the trust store. This command only operates on User level.
  delete        Delete certificates from the trust store. This command only operates on User level.
  generate-test Generate a test RSA key and a corresponding self-generated certificate
  list          List certificates used for verification. This command operates on User level and System level.
  show          Show certificate details given trust store type, named store, and certificate file name. If certificate file contains certificate chain, all certificates in the chain are displayed starting from the leaf. Certificate file on User level is displayed prior to  System level.

Flags:
  -h, --help   help for certificate

Global Flags:
      --plain-http   Registry access via plain HTTP
```

### notation certificate add

```text
Add certificates to the trust store. This command only operates on User level.

Usage:
  notation certificate add --type <type> --store <name> <filepath...> [flags]

Flags:
  -h, --help   Help for certificate
  -s, --store string   Specify named store
  -t, --type string    Specify trust store type, options: ca, signingAuthority

Global Flags:
      --plain-http   Registry access via plain HTTP
```

### notation certificate list

```text
List certificates in the trust store. This command operates on both User level and System level.

Usage:
  notation certificate list [flags]

Aliases:
  list, ls

Flags:
  -h, --help           help for list
  -s, --store string   specify named store
  -t, --type string    specify trust store type, options: ca, signingAuthority

Global Flags:
      --plain-http   Registry access via plain HTTP
```

### notation certificate show

```text
Show certificate details given trust store type, named store, and certificate file name. If certificate file contains certificate chain, all certificates in the chain are displayed starting from the leaf. Certificate file on User level is displayed prior to System level.

Usage:
  notation certificate show -t <type> -s <name> <fileName> [flags]

Flags:
  -h, --help           help for show
  -s, --store string   specify named store
  -t, --type string    specify trust store type, options: ca, signingAuthority

Global Flags:
      --plain-http   Registry access via plain HTTP
```

### notation certificate delete

```text
Delete certificates from the trust store. This command only operates on User level.

Usage:
  notation certificate delete -t <type> -s <name> (--all | <cert_filename>) [flags]

Aliases:
  delete, rm

Flags:
  -a, --all            If set to true, remove all certificates in the named store
  -y, --confirm        If yes, do not prompt for confirmation of deletion
  -h, --help           help for delete
  -s, --store string   Specify named store
  -t, --type string    Specify trust store type, options: ca, signingAuthority

Global Flags:
      --plain-http   Registry access via plain HTTP
```

### notation certificate generate-test

```text
Generate a test RSA key and a corresponding self-generated certificate

Usage:
  notation certificate generate-test <host> [flags]

Flags:
  -b, --bits int      RSA key bits (default 2048)
  -d, --default       mark as default
  -h, --help          help for generate-test
  -n, --name string   key and certificate name
      --trust         add the generated certificate to the trust store

Global Flags:
      --plain-http   Registry access via plain HTTP
```

## Usage

### Add certificates to the trust store

```bash
notation certificate add --type <type> --store <name> <cert_path>...
```

If a certificate file contains one certificate, this certificate MUST be self-signed certificate. If a certificate file contains multiple certificates, all these certificates MUST be CA type.

Upon successful adding, the certificate files are added into directory`{NOTATION_CONFIG}/truststore/x509/<type>/<name>/`, and a list of certificate filepaths are printed out. If the adding fails, an error message is printed out by listing which certificate files are successfully added, and which certificate files are not along with detailed reasons.

### List all certificate files stored in the trust store

```bash
notation certificate list
```

Upon successful listing, all the certificate files in the trust store are printed out in a format of absolute filepath. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

### List all certificate files of a certain named store

```bash
notation cert list --store <name>
```

Upon successful listing, all the certificate files in the trust store named `<name>` are printed out in a format of absolute filepath. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

### List all certificate files of a certain type of store

```bash
notation cert list --type <type>
```

Upon successfull listing, all the certificate files in the trust store of type `<type>` are printed out in a format of absolute filepath. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

### List all certificate files of a certain named store of a certain type

```bash
notation cert list --type <type> --store <name>
```

Upon successful listing, all the certificate files in the trust store named `<name>` of type `<type>` are printed out in a format of absolute filepath. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

### Show details of a certain certificate file

```bash
notation cert show --type <type> --store <name> <cert_file_name>
```

Upon successful show, the certificate details are printed out starting from leaf certificate if it's a certificate chain. Here is a list of certificate properties:

* version
* Serial Number
* Signature Algorithm
* Issuer
* Validity
* Subject
* Subject Public Key Info
* X509v3 Key Usage: critical
* X509v3 Extended Key Usage

If the showing fails, an error message is printed out with specific reasons.

### Delete all certificates of a certain named store of a certain type

```bash
notation cert delete --type <type> --store <name> --all
```

A prompt is showed asking user to confirm the deletion. Upon successful deletion, all certificates in trust store named `<name>` of type `<type>` are deleted. If deletion fails, a list of successful deleted certificate files is printed out as well as a list of deletion-failure certificates with specific reasons.

### Delete a specific certificate of a certain named store of a certain type

```bash
notation cert delete --type <type> --store <name> <cert_file_name>
```

A prompt is showed asking user to confirm the deletion. Upon successful deletion, the specific certificate is deleted in trust store named `<name>` of type `<type>`. If deletion fails, an error message with specific reasons is printed out.

### Generate a local RSA key and a corresponding self-generated certificate for testing purpose and add the certificate into trust store

```bash
notation certificate generate-test "wabbit-networks.io" --trust
```

Upon successful execution, a local key file and certificate file named `wabbit-networks.io` are generated and stored in `$XDG_CONFIG_HOME/notation/localkeys/`. `wabbit-networks.io` is also used as certificate subject.CommonName. With `--trust` flag set, the certificate is added into a trust store named `wabbit-networks.io` of type `ca`.
