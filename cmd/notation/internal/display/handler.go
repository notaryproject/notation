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

package display

import (
	"fmt"

	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/metadata"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/metadata/json"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/metadata/text"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/metadata/tree"
	"github.com/notaryproject/notation/v2/cmd/notation/internal/display/output"
)

// NewInspectHandler creates a new metadata InspectHandler based on the output
// format.
func NewInspectHandler(printer *output.Printer, format output.Format) (metadata.InspectHandler, error) {
	switch format {
	case output.FormatJSON:
		return json.NewInspectHandler(printer), nil
	case output.FormatTree:
		return tree.NewInspectHandler(printer), nil
	}
	return nil, fmt.Errorf("unrecognized output format %s", format)
}

// NewBlobInspectHandler creates a new metadata BlobInspectHandler based on the
// output format.
func NewBlobInspectHandler(printer *output.Printer, format output.Format) (metadata.BlobInspectHandler, error) {
	switch format {
	case output.FormatJSON:
		return json.NewBlobInspectHandler(printer), nil
	case output.FormatTree:
		return tree.NewBlobInspectHandler(printer), nil
	}
	return nil, fmt.Errorf("unrecognized output format %s", format)
}

// NewVerifyHandler creates a new metadata VerifyHandler for printing
// verification result and warnings.
func NewVerifyHandler(printer *output.Printer) metadata.VerifyHandler {
	return text.NewVerifyHandler(printer)
}

// NewBlobVerifyHandler creates a new metadata BlobVerifyHandler for printing
// blob verification result and warnings.
func NewBlobVerifyHandler(printer *output.Printer) metadata.BlobVerifyHandler {
	return text.NewBlobVerifyHandler(printer)
}

// NewListHandler creates a new metadata ListHandler for rendering signature
// metadata information in a tree format.
func NewListHandler(printer *output.Printer) metadata.ListHandler {
	return tree.NewListHandler(printer)
}
