package signature

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/notaryproject/notation/pkg/config"
)

func AddCertCore(cfg *config.File, name, path string) error {
	if ok := cfg.VerificationCertificates.Certificates.Append(name, path); !ok {
		return errors.New(name + ": already exists")
	}
	return nil
}

func PrintCertificateSet(s config.CertificateMap) {
	maxNameSize := 0
	for _, ref := range s {
		if len(ref.Name) > maxNameSize {
			maxNameSize = len(ref.Name)
		}
	}
	format := fmt.Sprintf("%%-%ds\t%%s\n", maxNameSize)
	fmt.Printf(format, "NAME", "PATH")
	for _, ref := range s {
		fmt.Printf(format, ref.Name, ref.Path)
	}
}

func NameFromPath(path string) string {
	base := filepath.Base(path)
	name := base[:len(base)-len(filepath.Ext(base))]
	if name == "" {
		return base
	}
	return name
}
