package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

// ExtractSingleFileFromZip extracts a single file from a zip archive.
func ExtractSingleFileFromZip(zipFilePath, fileName, targetPath string) error {
	// Open the zip file
	reader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Look for the file in the archive
	var fileToExtract *zip.File
	for _, file := range reader.File {
		if file.Name == fileName {
			fileToExtract = file
			break
		}
	}

	if fileToExtract == nil {
		return fmt.Errorf("file '%s' not found in archive", fileName)
	}

	// Open the file from the archive
	src, err := fileToExtract.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create the target file
	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the contents
	_, err = io.Copy(dst, src)
	return err
}
