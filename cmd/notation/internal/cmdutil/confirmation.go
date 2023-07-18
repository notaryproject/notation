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

package cmdutil

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// AskForConfirmation prints a propmt to ask for confirmation before doing an
// action and takes user input as response.
func AskForConfirmation(r io.Reader, prompt string, confirmed bool) (bool, error) {
	if confirmed {
		return true, nil
	}

	fmt.Print(prompt, " [y/N] ")

	scanner := bufio.NewScanner(r)
	if ok := scanner.Scan(); !ok {
		return false, scanner.Err()
	}
	response := scanner.Text()

	switch strings.ToLower(response) {
	case "y", "yes":
		return true, nil
	default:
		fmt.Println("Operation cancelled.")
		return false, nil
	}
}
