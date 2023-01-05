package notation

import (
	"encoding/json"
	"io"
	"os"
)

// copyFile copies the source file to the destination file
func copyFile(src, dst string) error {
	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}
	return out.Chmod(si.Mode())
}

// saveJSON marshals the data and save to the given path.
func saveJSON(data any, path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(data)
}
