package utils

import (
	"path/filepath"

	"github.com/notaryproject/notation/internal/osutil"
)

// Defai;tStore is the default trust store name.
const DefaultStore = "teste2e"

// DefaultPolicy is default policy used to verify the integrity of the signature.
// TODO: do we need to use text/template to generate policy?
var DefaultPolicy = `{
    "version": "1.0",
    "trustPolicies": [
        {
            "name": "teste2e",
            "registryScopes": [
                "*"
            ],
            "signatureVerification": {
                "level": "strict"
            },
            "trustStores": [
                "ca:teste2e"
            ],
            "trustedIdentities": [
                "x509.subject: C=US, ST=WA, L=Seattle, O=Notary"
            ]
        }
    ]
}`

// WritePolicy writes policy to the config/notation/trustpolicy.json.
// TODO：After policy cli is ready, we should use policy cli instead.
func WritePolicy(configDir, policy string) error {
	return osutil.WriteFile(filepath.Join(configDir, "config", "notation", "trustpolicy.json"), []byte(policy))
}
