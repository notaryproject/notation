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

// Package display provides the display handlers to render information for
// commands.
//
// - It includes the metadata, content and status packages for handling
// different types of information.
// - It includes the output package for writing information to the output.
package display

import (
	"fmt"

	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata"
	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata/inspect"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
)

// NewMetadataInpsectHandler creates a new InspectHandler based on the output
// format.
func NewMetadataInpsectHandler(printer *output.Printer, format option.Format) (metadata.InspectHandler, error) {
	switch option.FormatType(format.CurrentFormat) {
	case option.FormatTypeJSON:
		return inspect.NewJSONHandler(printer), nil
	case option.FormatTypeText:
		return inspect.NewTreeHandler(printer), nil
	}
	return nil, fmt.Errorf("unrecognized output format %s", format.CurrentFormat)
}
