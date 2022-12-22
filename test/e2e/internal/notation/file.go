package notation

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copyDir copies the source directory to the destination directory
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// generate the dst path
		relPath, err := filepath.Rel(src, srcPath)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, os.ModePerm)
		}
		return copyFile(srcPath, dstPath)
	})
}

// copyFile copies the source file to the destination file
func copyFile(src, dst string) error {
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

	si, err := in.Stat()
	if err != nil {
		return err
	}
	return out.Chmod(si.Mode())
}

// saveJSON marshals the data and save to the given path.
func saveJSON(data any, path string) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0644)
}
