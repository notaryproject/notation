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

package table

import (
	"fmt"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
)

// KeyListHandler is the handler to print the metadata of keys in table format.
type KeyListHandler struct {
	printer *output.Printer
}

// NewKeyListHandler creates a new KeyListHandler.
func NewKeyListHandler(printer *output.Printer) *KeyListHandler {
	return &KeyListHandler{
		printer: printer,
	}
}

// PrintKeys prints the metadata of keys in table format.
func (h *KeyListHandler) PrintKeys(defaultKeyName *string, keySuite []config.KeySuite) error {
	tw := newTabWriter(h.printer)
	fmt.Fprintln(tw, "NAME\tKEY PATH\tCERTIFICATE PATH\tID\tPLUGIN NAME\t")
	for _, key := range keySuite {
		name := key.Name
		if defaultKeyName != nil && key.Name == *defaultKeyName {
			name = "* " + name
		}
		kp := key.X509KeyPair
		if kp == nil {
			kp = &config.X509KeyPair{}
		}
		ext := key.ExternalKey
		if ext == nil {
			ext = &config.ExternalKey{}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t\n", name, kp.KeyPath, kp.CertificatePath, ext.ID, ext.PluginName)
	}
	return tw.Flush()
}
