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
package display

import (
	"fmt"

	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata"
	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata/json"
	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata/tree"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/cmd/notation/internal/output"
)

// NewInpsectHandler creates a new InspectHandler based on the output format.
func NewInpsectHandler(printer *output.Printer, format option.Format) (metadata.InspectHandler, error) {
	switch option.FormatType(format.FormatFlag) {
	case option.FormatTypeJSON:
		return json.NewInspectHandler(printer), nil
	case option.FormatTypeText:
		return tree.NewInspectHandler(printer), nil
	}
	return nil, fmt.Errorf("unrecognized output format %s", format.FormatFlag)
}
