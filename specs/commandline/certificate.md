# notation certificate

## Description

Use ```notation certificate``` command to add/list/delete certificates in notation's trust store. Updating an existing certificate is not allowed since the thumbprint will be inconsistent, which results in a new certificate.

The trust store is in the format of a directory in the filesystem as`x509/<type>/<name>/*.crt|*.cer|*.pem`. Currently two types of trust store are supported:

* `Certificate Authority`: The directory name is `ca`.
* `Signing Authority`: The directory name is `signingAuthority`

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
```

In this example, there are two certificates stored in trust store named `acme-rockets` of type `ca`. There is one certificate stored in trust store named `wabbit-networks` of type `signingAuthority`.

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
  -t, --type string    specify trust store type, options: ca, signingAuthority
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
  -t, --type string    specify trust store type, options: ca, signingAuthority
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
  -t, --type string    specify trust store type, options: ca, signingAuthority
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
  -t, --type string    specify trust store type, options: ca, signingAuthority
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

### Generate a local RSA key and a corresponding self-generated certificate for testing purpose and add the certificate into trust store

```bash
notation certificate generate-test "wabbit-networks.io"
```

Upon successful execution, a local key file and certificate file named `wabbit-networks.io` are generated and stored in `$XDG_CONFIG_HOME/notation/localkeys/`. `wabbit-networks.io` is also used as certificate subject.CommonName.
