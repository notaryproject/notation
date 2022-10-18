package osutil

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// WriteFile writes to a path with all parent directories created.
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0666)
}

// WriteFileWithPermission writes to a path with all parent directories created.
func WriteFileWithPermission(path string, data []byte, perm fs.FileMode, overwrite bool) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	flag := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_EXCL
	}
	file, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

// Copy the src file to dst. Existing file will be overwritten.
func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	if err := os.MkdirAll(dst, 0700); err != nil {
		return 0, err
	}
	certFile := filepath.Join(dst, filepath.Base(src))
	destination, err := os.Create(certFile)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	return io.Copy(destination, source)
}

// CleanDir removes all regular files under dir
func CleanDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Printf("dir \"%s\" is empty\n", dir)
	}
	var success []string
	var failure []string
	var errorSlice []error
	for _, name := range names {
		path := filepath.Join(dir, name)
		// Validate path
		fileInfo, err := os.Stat(path)
		if err != nil {
			failure = append(failure, path)
			errorSlice = append(errorSlice, err)
			continue
		}
		if !fileInfo.Mode().IsRegular() {
			failure = append(failure, path)
			errorSlice = append(errorSlice, errors.New("not a regular file"))
			continue
		}

		err = os.RemoveAll(path)
		if err != nil {
			failure = append(failure, path)
			errorSlice = append(errorSlice, err)
			continue
		}

		success = append(success, path)
	}

	// write out
	if len(success) != 0 {
		fmt.Println("Successfully deleted following certificates from Trust Store:")
		for _, p := range success {
			fmt.Println(p)
		}
	}
	if len(failure) != 0 {
		fmt.Println("Failed to delete following certificates from Trust Store:")

		for ind := range failure {
			fmt.Printf("%s, with error \"%s\"\n", failure[ind], errorSlice[ind])
		}
	}

	return nil
}
