package notation

import (
	"os"
	"path/filepath"
)

const (
	signingKeysName  = "signingkeys.json"
	localkeysDirName = "localkeys"
)

// X509KeyPair contains the paths of a public/private key pair files.
type X509KeyPair struct {
	KeyPath         string `json:"keyPath,omitempty"`
	CertificatePath string `json:"certPath,omitempty"`
}

// KeySuite is a named key suite.
type KeySuite struct {
	Name string `json:"name"`
	*X509KeyPair
}

// SigningKeys reflects the signingkeys.json file.
type SigningKeys struct {
	Default string     `json:"default"`
	Keys    []KeySuite `json:"keys"`
}

// AddTestKeyPairs creates the signingkeys.json file and the localkeys directory
// with e2e.key and e2e.crt
func AddTestKeyPairs(dir string) error {
	// create signingkeys.json files
	if err := saveJSON(
		genTestSigningKey(dir),
		filepath.Join(dir, signingKeysName)); err != nil {
		return err
	}

	// create localkeys directory
	localKeysDir := filepath.Join(dir, localkeysDirName)
	os.MkdirAll(localKeysDir, os.ModePerm)
	if err := copyFile(NotationE2EKeyPath, filepath.Join(localKeysDir, "e2e.key")); err != nil {
		return err
	}
	return copyFile(NotationE2ECertPath, filepath.Join(localKeysDir, "e2e.crt"))
}

func genTestSigningKey(dir string) *SigningKeys {
	return &SigningKeys{
		Default: "e2e",
		Keys: []KeySuite{
			{
				Name: "e2e",
				X509KeyPair: &X509KeyPair{
					KeyPath:         filepath.Join(dir, "localkeys", "e2e.key"),
					CertificatePath: filepath.Join(dir, "localkeys", "e2e.crt"),
				},
			},
		},
	}
}
