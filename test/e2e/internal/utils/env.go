package utils

import (
	"os"
)

// TempUserDir creates a tmp directory.
//
// Calling cleaner will remove the directory recursively.
func TempUserDir() (userDir string, cleaner func() error, err error) {
	// generate random temp dir for notation
	userDir, err = os.MkdirTemp("", "e2e-")
	if err != nil {
		return "", nil, err
	}
	cleaner = func() error {
		return os.RemoveAll(userDir)
	}

	return
}
