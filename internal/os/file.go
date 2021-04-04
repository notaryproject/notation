package os

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// WriteFile writes to a path with all parent directories created.
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0666)
}
