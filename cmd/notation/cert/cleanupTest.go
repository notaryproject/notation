package cert

import (
	"errors"
	"fmt"
	"os"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/spf13/cobra"
)

type certCleanupTestOpts struct {
	keyName string
}

func certCleanupTestCommand(opts *certCleanupTestOpts) *cobra.Command {
	if opts == nil {
		opts = &certCleanupTestOpts{}
	}
	command := &cobra.Command{
		Use:   "cleanup-test <common_name>",
		Short: "[Experimental] Clean up test key and its corresponding self-signed certificate created by the generate-test command.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("missing certificate common_name")
			}
			opts.keyName = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cleanupTestCert(opts)
		},
	}
	return command
}

func cleanupTestCert(opts *certCleanupTestOpts) error {
	keyName := opts.keyName
	var finalError []error
	err := config.LoadExecSaveSigningKeys(func(keys *config.SigningKeys) error {
		keySuite, err := keys.Get(keyName)
		if err != nil {
			return err
		}
		if keySuite.ExternalKey != nil {
			return errors.New("cleanup-test can only apply to non-external keys created by the generate-test command")
		}
		if keySuite.X509KeyPair == nil {
			return errors.New("cleanup-test requires key pair files created by the generate-test command")
		}
		// delete the key file from keyPath
		keyPath := keySuite.X509KeyPair.KeyPath
		err = osutil.DeleteFile(keyPath)
		if err != nil {
			finalError = append(finalError, fmt.Errorf("cannot delete the key file from %q during cleanup-test, %v", keyPath, err))
		} else {
			fmt.Printf("Successfully deleted the key file from %q\n", keyPath)
		}
		// delete the certificate file from certPath
		certPath := keySuite.X509KeyPair.CertificatePath
		err = osutil.DeleteFile(certPath)
		if err != nil {
			finalError = append(finalError, fmt.Errorf("cannot delete the certificate file from %q during cleanup-test, %v", certPath, err))
		} else {
			fmt.Printf("Successfully deleted the certificate file from %q\n", certPath)
		}
		// delete the certificate from trust store
		certFile := keyName + dir.LocalCertificateExtension
		err = truststore.DeleteCert("ca", keyName, certFile, true)
		if err != nil {
			finalError = append(finalError, fmt.Errorf("cannot delete certificate from truststore/ca/%s/%s during cleanup-test, %v", keyName, certFile, err))
		}
		// remove the key from Notation's signing key list
		_, err = keys.Remove(keyName)
		if err != nil {
			finalError = append(finalError, fmt.Errorf("cannot remove %q from notation signing key list during cleanup-test, %v", keyName, err))
		} else {
			fmt.Printf("Successfully removed %q from notation signing key list\n", keyName)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// check if there is any error in the process of cleanup-test
	if len(finalError) > 0 {
		for _, err := range finalError {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return errors.New("failed to complete cleanup-test, manual deletion may be required")
	}
	return nil
}
