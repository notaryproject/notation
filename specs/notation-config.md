# Notation Configuration

To enable persisted configuration, simplifying the execution of the `notation` cli, the following configuration file will be available

> Note: there will be a policy based configuration that will come at a later point.

## Location

The default location and file will be stored at: `~/.config/notation/config.json`. The `notation` cli and libraries supports alternate locations through the `XDG_CONFIG_HOME` environment variable.

> TODO: Add Windows and Mac locations

## Properties

| Property                  | Type     | Value                                                                                                                                                     |
| ------------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `insecureRegistries`      | _array_  | a list of registries that may be used without https                                                                                                       |
| `signatureFormat`            | _string_ | `jws` is the default value, available type includes `jws` and `cose`                                                                                      |

## Samples

`~/.config/notation/config.json`

```json
{
    "insecureRegistries": [
        "registry.wabbit-networks.io"
    ],
    "signatureFormat": "jws"
}
```

`~/.config/notation/signingkeys.json`

```json
{
    "default": "wabbit-networks",
    "keys": [
        {
            "name": "wabbit-networks",
            "keyPath": "/home/demo/.config/notation/localkeys/wabbit-networks.key",
            "certPath": "/home/demo/.config/notation/localkeys/wabbit-networks.crt"
        },
        {
            "name": "import.acme-rockets",
            "keyPath": "/home/demo/.config/notation/localkeys/import.acme-rockets.key",
            "certPath": "/home/demo/.config/notation/localkeys/import.acme-rockets.crt"
        }
    ]
}
```
