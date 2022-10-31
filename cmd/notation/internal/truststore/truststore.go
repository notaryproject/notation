package truststore

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verification"
	"github.com/notaryproject/notation/cmd/notation/internal/cmdutil"
	"github.com/notaryproject/notation/internal/osutil"
)

// AddCert adds a single cert file at path to the trust store
// under dir truststore/x509/storeType/namedStore
func AddCert(path, storeType, namedStore string, display bool) error {
	// initialize
	certPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if storeType == "" {
		return errors.New("store type cannot be empty")
	}
	if !IsValidStoreType(storeType) {
		return fmt.Errorf("unsupported store type: %s", storeType)
	}
	if !IsValidFileName(namedStore) {
		return errors.New("named store name needs to follow [a-zA-Z0-9_.-]+ format")
	}

	// check if the target path is a cert (support PEM and DER formats)
	if _, err := corex509.ReadCertificateFile(certPath); err != nil {
		return err
	}

	// core process
	// get the trust store path
	trustStorePath, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
	if err := CheckNonErrNotExistError(err); err != nil {
		return err
	}
	// check if certificate already in the trust store
	if _, err := os.Stat(filepath.Join(trustStorePath, filepath.Base(certPath))); err == nil {
		return errors.New("certificate already exists in the Trust Store")
	}
	// add cert to trust store
	_, err = osutil.CopyToDir(certPath, trustStorePath)
	if err != nil {
		return err
	}

	// write out
	if display {
		fmt.Printf("Successfully added %s to named store %s of type %s\n", filepath.Base(certPath), namedStore, storeType)
	}

	return nil
}

// ListCerts walks through root and lists all x509 certificates in it,
// sub-dirs are ignored.
func ListCerts(root string, depth int) error {
	maxDepth := strings.Count(root, string(os.PathSeparator)) + depth

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && strings.Count(path, string(os.PathSeparator)) > maxDepth {
			return fs.SkipDir
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			if _, err := corex509.ReadCertificateFile(path); err != nil {
				return err
			}
			fmt.Println(path)
		}
		return nil
	})
}

// ShowCerts writes out details of certificates
func ShowCerts(certs []*x509.Certificate) {
	fmt.Println("Certificate details")
	fmt.Println("--------------------------------------------------------------------------------")
	for ind, cert := range certs {
		showCert(cert)
		if ind != len(certs)-1 {
			fmt.Println("--------------------------------------------------------------------------------")
		}
	}
}

// showCert displays details of a certificate
func showCert(cert *x509.Certificate) {
	fmt.Println("Issuer:", cert.Issuer)
	fmt.Println("Subject:", cert.Subject)
	fmt.Println("Valid from:", cert.NotBefore)
	fmt.Println("Valid to:", cert.NotAfter)
	fmt.Println("IsCA:", cert.IsCA)

	h := sha1.Sum(cert.Raw)
	fmt.Println("SHA1 Thumbprint:", strings.ToLower(hex.EncodeToString(h[:])))
}

// DeleteAllCerts deletes all certificate files from the trust store
// under dir truststore/x509/storeType/namedStore
func DeleteAllCerts(storeType, namedStore string, confirmed bool, errorSlice []error) []error {
	path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
	if err == nil {
		prompt := fmt.Sprintf("Are you sure you want to delete all certificate in %q of type %q?", namedStore, storeType)
		confirmed, err := cmdutil.AskForConfirmation(os.Stdin, prompt, confirmed)
		if err != nil {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
			return errorSlice
		}
		if !confirmed {
			return errorSlice
		}

		if err = os.RemoveAll(path); err != nil {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
		} else {
			// write out on success
			fmt.Printf("Successfully deleted %s\n", path)
			return nil
		}
	} else {
		errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
	}
	return errorSlice
}

// DeleteCert deletes a specific certificate file from the
// trust store, namely truststore/x509/storeType/namedStore/cert
func DeleteCert(storeType, namedStore, cert string, confirmed bool, errorSlice []error) []error {
	path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore, cert)
	if err == nil {
		prompt := fmt.Sprintf("Are you sure you want to delete %q in %q of type %q?", cert, namedStore, storeType)
		confirmed, err := cmdutil.AskForConfirmation(os.Stdin, prompt, confirmed)
		if err != nil {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
			return errorSlice
		}
		if !confirmed {
			return errorSlice
		}

		if err = os.RemoveAll(path); err != nil {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
		} else {
			// write out on success
			fmt.Printf("Successfully deleted %s\n", path)
			return nil
		}
	} else {
		errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
	}

	return errorSlice
}

// CheckNonErrNotExistError returns nil when no err or err is fs.ErrNotExist
func CheckNonErrNotExistError(err error) error {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// IsValidStoreType checks if storeType is supported
func IsValidStoreType(storeType string) bool {
	for _, t := range verification.TrustStorePrefixes {
		if storeType == string(t) {
			return true
		}
	}
	return false
}

// IsValidFileName checks if a file name is cross-platform compatible
func IsValidFileName(fileName string) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`).MatchString(fileName)
}
