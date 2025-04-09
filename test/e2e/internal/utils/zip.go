// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
