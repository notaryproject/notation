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
	"path/filepath"

	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
)

// CertificateListHandler is the handler to print the metadata of certificatesin table
// format.
type CertificateListHandler struct {
	printer *output.Printer
}

// NewCertificateListHandler creates a new CertListHandler.
func NewCertificateListHandler(printer *output.Printer) *CertificateListHandler {
	return &CertificateListHandler{
		printer: printer,
	}
}

// PrintCertificates prints the metadata of certificates in table format.
func (h *CertificateListHandler) PrintCertificates(certificatePaths []string) error {
	if len(certificatePaths) == 0 {
		return nil
	}
	tw := newTabWriter(h.printer)
	fmt.Fprintln(tw, "STORE TYPE\tSTORE NAME\tCERTIFICATE\t")
	for _, cert := range certificatePaths {
		fileName := filepath.Base(cert)
		dir := filepath.Dir(cert)
		namedStore := filepath.Base(dir)
		dir = filepath.Dir(dir)
		storeType := filepath.Base(dir)
		fmt.Fprintf(tw, "%s\t%s\t%s\t\n", storeType, namedStore, fileName)
	}
	return tw.Flush()
}
