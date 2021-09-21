# Notation Configuration

To enable persisted configuration, simplifying the execution of the `notation` cli, the following configuration file will be available

> Note: there will be a policy based configuration that will come at a later point.

## Location

The default location and file will be stored at: `~/.notation/config.json`. The `notation` cli and libraries will support alternate locations through a `config-location` parameters.

> TODO: Add Windows and Mac locations

## Properties

Property | Type |  Value
------ | ------ | ---
`verificationCerts.certs`|_array_|collection of name/value pairs for a collection of public certs that are used for verification. These may be replaced with a future policy configuration.
`cert.name`|_string_|a named reference to the certificate
`cert.path`|_string_|a location by which the certificate can be found by the notation cli or notation libraries
`signing-keys.keys`|_array_|a collection of name/value pairs of signing keys.
`key.name`|_string_|a named reference to the key
`key.path`|_string_|a location by which the key can be found by the notation cli or notation libraries
`signing-keys.default`|_string_|the signing key to be used when `notation sign` is called without `--name`
`insecureRegistries`|_array_|a list of registries that may be used without https

## Samples 

`~/.notary/notation-config.json`

```json
{
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
  "signingKeys": {
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
