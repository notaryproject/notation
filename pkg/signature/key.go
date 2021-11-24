package signature

import (
	"errors"
	"fmt"

	"github.com/notaryproject/notation/pkg/config"
)

func AddKeyCore(cfg *config.File, name, keyPath, certPath string, markDefault bool) (bool, error) {
	if ok := cfg.SigningKeys.Keys.Append(name, keyPath, certPath); !ok {
		return false, errors.New(name + ": already exists")
	}
	if markDefault {
		cfg.SigningKeys.Default = name
	}
	return cfg.SigningKeys.Default == name, nil
}

func PrintKeySet(target string, s config.KeyMap) {
	if len(s) == 0 {
		fmt.Println("NAME\tPATH")
		return
	}

	var maxNameSize, maxKeyPathSize int
	for _, ref := range s {
		if len(ref.Name) > maxNameSize {
			maxNameSize = len(ref.Name)
		}
		if len(ref.KeyPath) > maxKeyPathSize {
			maxKeyPathSize = len(ref.KeyPath)
		}
	}
	format := fmt.Sprintf("%%c %%-%ds\t%%-%ds\t%%s\n", maxNameSize, maxKeyPathSize)
	fmt.Printf(format, ' ', "NAME", "KEY PATH", "CERTIFICATE PATH")
	for _, ref := range s {
		mark := ' '
		if ref.Name == target {
			mark = '*'
		}
		fmt.Printf(format, mark, ref.Name, ref.KeyPath, ref.CertificatePath)
	}
}
