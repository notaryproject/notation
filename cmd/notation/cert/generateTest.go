// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	keyDefaultFlag = &pflag.Flag{
		Name:  "default",
		Usage: "mark as default signing key",
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
		Long: `Generate a test RSA key and a corresponding self-signed certificate

Example - Generate a test RSA key and a corresponding self-signed certificate named "wabbit-networks.io":
  notation cert generate-test "wabbit-networks.io"

Example - Generate a test RSA key and a corresponding self-signed certificate, set RSA key as a default signing key:
  notation cert generate-test --default "wabbit-networks.io"
`,
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
	if !truststore.IsValidFileName(name) {
		return errors.New("name needs to follow [a-zA-Z0-9_.-]+ format")
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
	relativeKeyPath, relativeCertPath := dir.LocalKeyPath(name)
	configFS := dir.ConfigFS()
	keyPath, err := configFS.SysPath(relativeKeyPath)
	if err != nil {
		return err
	}
	certPath, err := configFS.SysPath(relativeCertPath)
	if err != nil {
		return err
	}
	if err := osutil.WriteFileWithPermission(keyPath, keyBytes, 0600, false); err != nil {
		return fmt.Errorf("failed to write key file: %v", err)
	}
	fmt.Println("wrote key:", keyPath)

	// write the self-signed certificate
	if err := osutil.WriteFileWithPermission(certPath, certBytes, 0644, false); err != nil {
		return fmt.Errorf("failed to write certificate file: %v", err)
	}
	fmt.Println("wrote certificate:", certPath)

	// update signingkeys.json config
	exec := func(s *config.SigningKeys) error {
		return s.Add(opts.name, keyPath, certPath, opts.isDefault)
	}
	if err := config.LoadExecSaveSigningKeys(exec); err != nil {
		return err
	}

	// Add to the trust store
	if err := truststore.AddCert(certPath, "ca", name, true); err != nil {
		return err
	}

	// write out
	fmt.Printf("%s: added to the key list\n", name)
	if opts.isDefault {
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
