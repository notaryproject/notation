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

package osutil

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(file)
	const bytes int64 = 10485760 //10MB in bytes
	limitedReader := &io.LimitedReader{R: reader, N: bytes}
	contents, _ := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}
	if len(contents) == 0 {
		return nil, fmt.Errorf("file is empty")
	}
	if limitedReader.N == 0 {
		return nil, fmt.Errorf("unable to read as file size was greater than %v bytes", bytes)
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}
	return contents, nil
}
