package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/notaryproject/notation-core-go/testhelper"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/notaryproject/notation/internal/slices"
	"github.com/notaryproject/notation/pkg/configutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	keyDefaultFlag = &pflag.Flag{
		Name:      "default",
		Shorthand: "d",
		Usage:     "mark as default signing key",
	}
	setKeyDefaultFlag = func(fs *pflag.FlagSet, p *bool) {
		fs.BoolVarP(p, keyDefaultFlag.Name, keyDefaultFlag.Shorthand, false, keyDefaultFlag.Usage)
	}
)

type certGenerateTestOpts struct {
	name      string
	bits      int
	isDefault bool
}

func certGenerateTestCommand(opts *certGenerateTestOpts) *cobra.Command {
	if opts == nil {
		opts = &certGenerateTestOpts{}
	}
	command := &cobra.Command{
		Use:   "generate-test [flags] <common_name>",
		Short: "Generate a test RSA key and a corresponding self-signed certificate.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate common_name")
			}
			opts.name = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return generateTestCert(opts)
		},
	}

	command.Flags().IntVarP(&opts.bits, "bits", "b", 2048, "RSA key bits")
	setKeyDefaultFlag(command.Flags(), &opts.isDefault)
	return command
}

func generateTestCert(opts *certGenerateTestOpts) error {
	// initialize
	name := opts.name
	if name == "" {
		return errors.New("certificate common_name cannot be empty")
	}

	// generate RSA private key
	bits := opts.bits
	fmt.Println("generating RSA Key with", bits, "bits")
	key, keyBytes, err := generateTestKey(bits)
	if err != nil {
		return err
	}

	rsaCertTuple, certBytes, err := generateSelfSignedCert(key, name)
	if err != nil {
		return err
	}
	fmt.Println("generated certificate expiring on", rsaCertTuple.Cert.NotAfter.Format(time.RFC3339))

	// write private key
	keyPath, certPath := dir.Path.Localkey(name)
	if err := osutil.WriteFileWithPermission(keyPath, keyBytes, 0600, false); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	fmt.Println("wrote key:", keyPath)

	// write the self-signed certificate
	if err := osutil.WriteFileWithPermission(certPath, certBytes, 0644, false); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	fmt.Println("wrote certificate:", certPath)

	// update config
	signingKeys, err := configutil.LoadSigningkeysOnce()
	if err != nil {
		return err
	}
	isDefault := opts.isDefault
	keySuite := config.KeySuite{
		Name: name,
		X509KeyPair: &config.X509KeyPair{
			KeyPath:         keyPath,
			CertificatePath: certPath,
		},
	}
	err = addKeyToSigningKeys(signingKeys, keySuite, isDefault)
	if err != nil {
		return err
	}

	// Add to the trust store
	if err := truststore.AddCert(certPath, "ca", name, true); err != nil {
		return err
	}

	// Save to the SigningKeys.json
	if err := signingKeys.Save(); err != nil {
		return err
	}

	// write out
	fmt.Printf("%s: added to the key list\n", name)
	if isDefault {
		fmt.Printf("%s: mark as default signing key\n", name)
	}
	return nil
}

func generateTestKey(bits int) (*rsa.PrivateKey, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	return key, keyPEM, nil
}

func generateCertPEM(rsaCertTuple *testhelper.RSACertTuple) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rsaCertTuple.Cert.Raw})
}

// generateTestCert generates a self-signed non-CA certificate
func generateSelfSignedCert(privateKey *rsa.PrivateKey, name string) (testhelper.RSACertTuple, []byte, error) {
	rsaCertTuple := testhelper.GetRSASelfSignedCertTupleWithPK(privateKey, name)
	return rsaCertTuple, generateCertPEM(&rsaCertTuple), nil
}

func addKeyToSigningKeys(signingKeys *config.SigningKeys, key config.KeySuite, markDefault bool) error {
	if slices.Contains(signingKeys.Keys, key.Name) {
		return errors.New(key.Name + ": already exists")
	}
	signingKeys.Keys = append(signingKeys.Keys, key)
	if markDefault {
		signingKeys.Default = key.Name
	}
	return nil
}
