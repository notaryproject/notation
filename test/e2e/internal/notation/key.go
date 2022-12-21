package notation

import (
	"encoding/json"
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
	f, err := os.Create(filepath.Join(dir, signingKeysName))
	if err != nil {
		return err
	}
	sk, err := json.Marshal(signingKey(dir))
	if err != nil {
		return err
	}
	f.Write(sk)
	f.Close()

	// create localkeys directory
	localKeysDir := filepath.Join(dir, localkeysDirName)
	os.MkdirAll(localKeysDir, os.ModePerm)
	copyFile(NotationE2EKeyPath, filepath.Join(localKeysDir, "e2e.key"))
	copyFile(NotationE2ECertPath, filepath.Join(localKeysDir, "e2e.crt"))
	return nil
}

func signingKey(dir string) *SigningKeys {
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
