# notation certificate

## Description

Use ```notation certificate``` command to add/list/delete certificates in notation's trust store. Updating an existing certificate is not allowed since the thumbprint will be inconsistent, which results in a new certificate.

The trust store is in the format of a directory in the filesystem as`x509/<type>/<name>/*.crt|*.cer|*.pem`. Currently three types of trust store are supported:

* `Certificate Authority`: The directory name is `ca`
* `Signing Authority`: The directory name is `signingAuthority`
* `Timestamping Authority`: The directory name is `tsa`

There could be more trust store types introduced in the future.

Here is an example of trust store directory structure:

```text
$XDG_CONFIG_HOME/notation/truststore
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

        /tsa
            /trusted-tsa
                tsa.crt
```

In this example, there are two certificates stored in trust store named `acme-rockets` of type `ca`. There is one certificate stored in trust store named `wabbit-networks` of type `signingAuthority`. And there is one certificate stored in trust store named `trusted-tsa` of type `tsa`.

## Outline

### notation certificate

```text
Manage certificates in trust store for signature verification.

Usage:
  notation certificate [command]

Aliases:
  certificate, cert

Available Commands:
  add           Add certificates to the trust store.
  cleanup-test  Clean up a test RSA key and its corresponding certificate that were generated using the "generate-test" command.
  delete        Delete certificates from the trust store.
  generate-test Generate a test RSA key and a corresponding self-signed certificate.
  list          List certificates in the trust store.
  show          Show certificate details given trust store type, named store, and certificate file name. If the certificate file contains multiple certificates, then all certificates are displayed.

Flags:
  -h, --help   help for certificate
```

### notation certificate add

```text
Add certificates to the trust store.

Usage:
  notation certificate add --type <type> --store <name> [flags] <cert_path>...

Flags:
  -h, --help           help for add
  -s, --store string   specify named store
  -t, --type string    specify trust store type, options: ca, signingAuthority, tsa
```

### notation certificate list

```text
List certificates in the trust store.

Usage:
  notation certificate list [flags]

Aliases:
  list, ls

Flags:
  -d, --debug          debug mode
  -h, --help           help for list
  -s, --store string   specify named store
  -t, --type string    specify trust store type, options: ca, signingAuthority, tsa
  -v, --verbose        verbose mode
```

### notation certificate show

```text
Show certificate details of given trust store name, trust store type, and certificate file name. If the certificate file contains multiple certificates, then all certificates are displayed.

Usage:
  notation certificate show --type <type> --store <name> [flags] <cert_fileName>

Flags:
  -d, --debug          debug mode
  -h, --help           help for show
  -s, --store string   specify named store
  -t, --type string    specify trust store type, options: ca, signingAuthority, tsa
  -v, --verbose        verbose mode
```

### notation certificate delete

```text
Delete certificates from the trust store.

Usage:
  notation certificate delete --type <type> --store <name> [flags] (--all | <cert_fileName>)

Flags:
  -a, --all            delete all certificates in the named store
  -h, --help           help for delete
  -s, --store string   specify named store
  -t, --type string    specify trust store type, options: ca, signingAuthority, tsa
  -y, --yes            do not prompt for confirmation
```

### notation certificate generate-test

```text
Generate a test RSA key and a corresponding self-signed certificate.

Usage:
  notation certificate generate-test [flags] <common_name>

Flags:
  -b, --bits int   RSA key bits (default 2048)
      --default    mark as default signing key
  -h, --help       help for generate-test
```

### notation certificate cleanup-test

```text
Clean up a test RSA key and its corresponding certificate that were generated using the "generate-test" command.

Usage:
  notation certificate cleanup-test [flags] <common_name>

Flags:
  -h, --help       help for generate-test
  -y, --yes        do not prompt for confirmation
```

## Usage

### Add certificates to the trust store

```bash
notation certificate add --type <type> --store <name> <cert_path>...
```

For each certificate in a certificate file, it MUST be either CA type or self-signed.

Upon successful adding, the certificate files are added into directory`{NOTATION_CONFIG}/truststore/x509/<type>/<name>/`, and a list of certificate filepaths are printed out. If the adding fails, an error message is printed out by listing which certificate files are successfully added, and which certificate files are not along with detailed reasons.

### List all certificate files stored in the trust store

```bash
notation certificate list
```

Upon successful listing, all the certificate files in the trust store are printed out with information of store type, store name and certificate file name. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

An example of the output:
```
STORE TYPE         STORE NAME   CERTIFICATE
ca                 myStore1     cert1.pem
ca                 myStore2     cert2.crt
signingAuthority   myStore1     cert3.crt
signingAuthority   myStore2     cert4.pem
tsa                myTSA        tsa.crt
```
### List all certificate files of a certain named store

```bash
notation cert list --store <name>
```

Upon successful listing, all the certificate files in the trust store named `<name>` are printed out with information of store type, store name and certificate file name. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

### List all certificate files of a certain type of store

```bash
notation cert list --type <type>
```

Upon successful listing, all the certificate files in the trust store of type `<type>` are printed out with information of store type, store name and certificate file name. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

### List all certificate files of a certain named store of a certain type

```bash
notation cert list --type <type> --store <name>
```

Upon successful listing, all the certificate files in the trust store named `<name>` of type `<type>` are printed out with information of store type, store name and certificate file name. If the listing fails, an error message is printed out with specific reasons. Nothing is printed out if the trust store is empty.

### Show details of a certain certificate file

```bash
notation certificate show --type <type> --store <name> <cert_fileName>
```

Upon successful show, the certificate details are printed out starting from leaf certificate if it's a certificate chain. Here is a list of certificate properties:

* Issuer
* Subject
* Valid from
* Valid to
* IsCA
* Thumbprints

If the showing fails, an error message is printed out with specific reasons.

### Delete all certificates of a certain named store of a certain type

```bash
notation certificate delete --type <type> --store <name> --all
```

A prompt is showed asking user to confirm the deletion. Upon successful deletion, all certificates in trust store named `<name>` of type `<type>` are deleted. If deletion fails, a list of successful deleted certificate files is printed out as well as a list of deletion-failure certificates with specific reasons.

### Delete a specific certificate of a certain named store of a certain type

```bash
notation certificate delete --type <type> --store <name> <cert_fileName>
```

A prompt is displayed, asking the user to confirm the deletion. Upon successful deletion, the specific certificate is deleted from the trust store named `<name>` of type `<type>`. The output message is printed out as following:

```text
Successfully deleted <cert_fileName> from the trust store.
```

If users execute the deletion without specifying required flags using `notation cert delete <cert_fileName>`, the deletion fails and the error output message is printed out as follows:

```text
Error: required flag(s) "store", "type" not set
```

### Generate a local RSA key and a corresponding self-generated certificate for testing purpose

```bash
notation certificate generate-test "wabbit-networks.io"
```

Upon successful execution, a local key file named `wabbit-networks.io.key` and a certificate file named `wabbit-networks.io.crt` are generated and stored in `$XDG_CONFIG_HOME/notation/localkeys/`. `wabbit-networks.io` is also used as the certificate's subject.CommonName. The certificate is added to trust store `wabbit-networks.io` of type `ca`. And the key with name `wabbit-networks.io` is added into `{NOTATION_CONFIG}/signingkeys.json`.

### Clean up a test RSA key and its corresponding certificate that were generated using the "generate-test" command

Use the following command to clean up a test RSA key and its corresponding certificate that were generated using the `generate-test` command.

```bash
notation certificate cleanup-test "wabbit-networks.io"
```

A prompt will be displayed, asking the user to confirm the cleanup.

```text
Are you sure you want to clean up test key <name> and its certificate? [y/N]
```

To suppress the prompt, use the `--yes` or `-y` flag. If the user chooses `y`, the following steps will be executed by the `cleanup-test` command:

- The local certificate file named `wabbit-networks.io.crt` is deleted from the trust store named `wabbit-networks.io` of type `ca`.
- The configuration with local RSA key named `wabbit-networks.io` is removed from `{NOTATION_CONFIG}/signingkeys.json`.
- The local RSA key file `wabbit-networks.io.key` is deleted from the directory "{NOTATION_CONFIG}/localkeys".
- The local certificate file `wabbit-networks.io.crt` is deleted from the directory "{NOTATION_CONFIG}/localkeys".

If any step encounters non-existent conditions, the entire process will not be terminated. This ensures that any previous incomplete cleanup can be addressed.

A sample output for a successful execution:

```text
Successfully deleted certificate <name>.crt from trust store <name> of type ca.
Successfully removed key <name> from signingkeys.json.
Successfully deleted key file: {NOTATION_CONFIG}/localkeys/<name>.key.
Successfully deleted certificate file: {NOTATION_CONFIG}/localkeys/<name>.crt.
Cleanup completed successfully.
```

A sample output for non-existent conditions:

```text
Certificate <name>.crt does not exist in trust store <name> of type ca.
Key <name> does not exist in signingkeys.json.
Key file {NOTATION_CONFIG}/localkeys/<name>.key does not exist.
Successfully deleted certificate file: {NOTATION_CONFIG}/localkeys/<name>.crt.
Cleanup completed successfully.
```

A sample output for failure:

```text
Failed to clean up the test key <name> and its corresponding certificate: <Reason>.
```

