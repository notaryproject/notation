package display

import (
	"fmt"

	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata"
	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata/json"
	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata/tree"
	"github.com/notaryproject/notation/cmd/notation/internal/option"
	"github.com/notaryproject/notation/cmd/notation/internal/output"
)

func NewInpsectHandler(printer *output.Printer, format option.Format) (metadata.InspectHandler, error) {
	switch option.FormatType(format.FormatFlag) {
	case option.FormatTypeJSON:
		return json.NewInspectHandler(printer), nil
	case option.FormatTypeText:
		return tree.NewInspectHandler(printer), nil
	}
	return nil, fmt.Errorf("unrecognized output format %s", format.FormatFlag)
}
