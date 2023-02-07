# Notation E2E Plugin
The package implement a simple plugin for Notation plugin extensibility E2E test.

## Commands
It support following commands:
- get-plugin-metadata (the plugin information)
- describe-key (the keySpec)
- generate-signature (signing a payload with protected data)
- generate-envelope (signing a payload with a descriptor and generate the envelope)
- verify-signature (fake verifier for testing the plugin protocol)

the `SIGNATURE_GENERATOR.RAW` and `SIGNATURE_GENERATOR.ENVELOPE` capabilities are hidden by default. `SIGNATURE_VERIFIER.TRUSTED_IDENTITY` and `SIGNATURE_VERIFIER.REVOCATION_CHECK` are shown all the time in the response of `get-plugin-metadata` command.

You can enable one of signing capability by passing the enable flag in PluginConfig for `get-plugin-metadata` command:
```json
{
    "pluginConfig": {
        "SIGNATURE_GENERATOR.RAW": "true"
    }
}
```
> Notation only uses one of `SIGNATURE_GENERATOR.RAW` or `SIGNATURE_GENERATOR.ENVELOPE` capability for a signing operation.

## Config
It reads the $NOTATION_CONFIG_DIR/pluginkeys.json to get the key and cert info.

example pluginkey.json file
```json
{
    "keys": [
        {
            "name": "e2e",
            "id": "keyid",
            "keyPath": "/home/username/.config/notation/localkeys/e2e.key",
            "certPath": "/home/username/.config/notation/localkeys/e2e.crt"
        }
    ]
}
```