package json

import (
	"fmt"
	"time"

	"github.com/notaryproject/notation-core-go/signature"
	"github.com/notaryproject/notation/cmd/notation/internal/display/metadata/model"
	"github.com/notaryproject/notation/cmd/notation/internal/output"
	"github.com/notaryproject/tspclient-go"
)

type inspectOutput struct {
	MediaType  string `json:"mediaType"`
	Signatures []model.Signature
}

type InspectHandler struct {
	output inspectOutput

	// printer is the printer for output.
	printer *output.Printer
}

// NewInspectHandler creates a new InspectHandler.
func NewInspectHandler(printer *output.Printer) *InspectHandler {
	return &InspectHandler{
		printer: printer,
	}
}

// SetReference sets the artifact reference for the handler. It is a no-op for this
// handler.
func (h *InspectHandler) SetReference(_ string) {}

// SetMediaType sets the media type for the handler.
func (h *InspectHandler) SetMediaType(mediaType string) {
	if h.output.MediaType == "" {
		h.output.MediaType = mediaType
	}
}

// AddSignature adds a signature to the handler.
func (h *InspectHandler) AddSignature(digest string, envelopeMediaType string, sigEnvelope signature.Envelope) error {
	sig, err := model.NewSignature(digest, envelopeMediaType, sigEnvelope, formatter)
	if err != nil {
		return err
	}
	h.output.Signatures = append(h.output.Signatures, *sig)
	return nil
}

func (h *InspectHandler) Print() error {
	return output.PrintPrettyJSON(h.printer, h.output)
}

// formatter is the function for formatting the time.Time and tspclient.Timestamp.
func formatter(v any) string {
	switch v := v.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case tspclient.Timestamp:
		return v.Format(time.RFC3339)
	}
	return fmt.Sprintf("%v", v)
}
