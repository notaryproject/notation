package certificate

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	corex509 "github.com/notaryproject/notation-core-go/x509"
	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/notaryproject/notation/internal/osutil"
)

// KeyUsageNameMap is a map of x509.Certificate KeyUsage map used in
// 'notation cert show' command
var KeyUsageNameMap = map[x509.KeyUsage]string{
	x509.KeyUsageDigitalSignature:  "Digital Signature",
	x509.KeyUsageContentCommitment: "Non Repudiation",
	x509.KeyUsageKeyEncipherment:   "Key Encipherment",
	x509.KeyUsageDataEncipherment:  "Data Encipherment",
	x509.KeyUsageKeyAgreement:      "Key Agreement",
	x509.KeyUsageCertSign:          "Certificate Sign",
	x509.KeyUsageCRLSign:           "CRL Sign",
	x509.KeyUsageEncipherOnly:      "Encipher Only",
	x509.KeyUsageDecipherOnly:      "Decipher Only",
}

// ExtKeyUsagesNameMap is a map of x509.Certificate ExtKeyUsages map used
// in 'notation cert show' command
var ExtKeyUsagesNameMap = map[x509.ExtKeyUsage]string{
	x509.ExtKeyUsageAny:                            "Any",
	x509.ExtKeyUsageServerAuth:                     "TLS Web Server Authentication",
	x509.ExtKeyUsageClientAuth:                     "TLS Web Client Authentication",
	x509.ExtKeyUsageCodeSigning:                    "Code Signing",
	x509.ExtKeyUsageEmailProtection:                "E-mail Protection",
	x509.ExtKeyUsageIPSECEndSystem:                 "IPSec End System",
	x509.ExtKeyUsageIPSECTunnel:                    "IPSec Tunnel",
	x509.ExtKeyUsageIPSECUser:                      "IPSec User",
	x509.ExtKeyUsageTimeStamping:                   "Time Stamping",
	x509.ExtKeyUsageOCSPSigning:                    "OCSP Signing",
	x509.ExtKeyUsageMicrosoftServerGatedCrypto:     "Microsoft Server Gated Crypto",
	x509.ExtKeyUsageNetscapeServerGatedCrypto:      "Netscape Server Gated Crypto",
	x509.ExtKeyUsageMicrosoftCommercialCodeSigning: "Microsoft Commercial Code Signing",
	x509.ExtKeyUsageMicrosoftKernelCodeSigning:     "Microsoft Kernel Code Signing",
}

// AddCertCore adds a single cert file at path to the trust store
// under dir truststore/x509/storeType/namedStore
func AddCertCore(path, storeType, namedStore string, display bool) error {
	// initialize
	certPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	storeType = strings.TrimSpace(storeType)
	if storeType == "" {
		return errors.New("store type cannot be empty or contain only whitespaces")
	}
	namedStore = strings.TrimSpace(namedStore)
	if namedStore == "" {
		return errors.New("named store cannot be empty or contain only whitespaces")
	}

	// check if the target path is a cert (support PEM and DER formats)
	if _, err := corex509.ReadCertificateFile(certPath); err != nil {
		return err
	}

	// core process
	// get the trust store path
	trustStorePath, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
	if err := CheckError(err); err != nil {
		return err
	}
	// check if certificate already in the trust store
	if _, err := os.Stat(filepath.Join(trustStorePath, filepath.Base(certPath))); err == nil {
		return errors.New("certificate already exists in the Trust Store")
	}
	// add cert to trust store
	_, err = osutil.Copy(certPath, trustStorePath)
	if err != nil {
		return err
	}

	// write out
	if display {
		fmt.Printf("Successfully added %s to named store %s of type %s\n", filepath.Base(certPath), namedStore, storeType)
	}

	return nil
}

// ListCertsCore walks through root and lists all x509 certificates in it
func ListCertsCore(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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

// ShowCertsCore writes out details of certificates
func ShowCertsCore(certs []*x509.Certificate) {
	fmt.Println("Display certificate details. Starting from leaf certificate if it's a certificate chain.")
	fmt.Println("-----------------------------------------------------------------------------------------")
	for ind, cert := range certs {
		showCert(cert)
		if ind != len(certs)-1 {
			fmt.Println("-----------------------------------------------------------------------------------------")
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
	fmt.Println("Thumbprints:", hex.EncodeToString(h[:]))
}

// RemoveAllCerts deletes all certificate files from the trust store
// under dir truststore/x509/storeType/namedStore
func RemoveAllCerts(storeType, namedStore string, confirmed bool, errorSlice []error) []error {
	path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore)
	if err == nil {
		prompt := fmt.Sprintf("Are you sure you want to remove all certificate files under dir: %q?", path)
		confirmed, err := ioutil.AskForConfirmation(os.Stdin, prompt, confirmed)
		if err != nil {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
			return errorSlice
		}
		if !confirmed {
			return errorSlice
		}

		if err = osutil.CleanDir(path); err != nil {
			errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
		}
	} else {
		errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
	}
	return errorSlice
}

// RemoveCert deletes a specific certificate file from the
// trust store, namely truststore/x509/storeType/namedStore/cert
func RemoveCert(storeType, namedStore, cert string, confirmed bool, errorSlice []error) []error {
	path, err := dir.Path.UserConfigFS.GetPath(dir.TrustStoreDir, "x509", storeType, namedStore, cert)
	if err == nil {
		prompt := fmt.Sprintf("Are you sure you want to delete: %q?", path)
		confirmed, err := ioutil.AskForConfirmation(os.Stdin, prompt, confirmed)
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
			return []error{}
		}
	} else {
		errorSlice = append(errorSlice, fmt.Errorf("%s with error %q", path, err.Error()))
	}

	return errorSlice
}

// CheckError returns nil when no err or err is fs.ErrNotExist
func CheckError(err error) error {
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}
