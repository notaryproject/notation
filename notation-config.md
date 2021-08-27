# Notation Configuration

To enable persisted configuration, simplifying the execution of the `notation` cli, the following configuration file will be available

> Note: there will be a policy based configuration that will come at a later point.

## Location

The default location and file will be stored at: `~/.notary/notation-config.json`. The `notation` cli and libraries will support alternate locations through a `config-location` parameters.

## Properties

- `enabled` - bool, which acts as an on/off switch for default notation behavior. --enabled false would disable all automatic validations
- `verificationCerts` - collection of name/value pairs for a collection of public certs that are used for verification. These may be replaced with a future policy configuration.
  - `name` - a named reference to the certificate
  - `path` - a location by which the certificate can be found by the notation cli or notation libraries
- `signing-keys` - a collection of name/value pairs of signing keys.
  - `name` - a named reference to the key
  - `path` - a location by which the key can be found by the notation cli or notation libraries
  - `default` - the signing key to be used when `notation sign` is called without `--name`
- `insecureRegistries` - a list of registries that may be used without https

## Samples 

`~/.notary/notation-config.json`

```json
{
  "enabled": true,
  "verificationCerts": {
    "certs": [
      {
        "name": "wabbit-networks.io",
        "path": "~/./notary/keys/wabbit-networks.crt"
      },
      {
        "name": "import.acme-rockets.io",
        "path": "~/./notary/keys/import-acme-rockets.crt"
      }
    ]
  },
  "signing-keys": {
    "default": "wabbit-networks.io",
    "keys": [
      {
        "name": "wabbit-networks.io",
        "path": "~/./notary/keys/wabbit-networks.key"
      },
      {
        "name": "import.acme-rockets.io",
        "path": "~/./notary/keys/import-acme-rockets.key"
      }
    ]
  },
  "insecureRegistries": [
    "registry.wabbit-networks.io"
  ]
}
```
