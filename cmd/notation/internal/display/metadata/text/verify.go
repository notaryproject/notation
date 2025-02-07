package text

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"text/tabwriter"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
)

// VerifyHandler is a handler for rendering output for verify command in
// human-readable format.
type VerifyHandler struct {
	printer *output.Printer

	outcome         *notation.VerificationOutcome
	digestReference string
}

// NewVerifyHandler creates a VerifyHandler to render verification results in
// human-readable format.
func NewVerifyHandler(printer *output.Printer) *VerifyHandler {
	return &VerifyHandler{
		printer: printer,
	}
}

// OnResolvingTagReference outputs the tag reference warning.
func (h *VerifyHandler) OnResolvingTagReference(reference string) {
	h.printer.ErrorPrintf("Warning: Always verify the artifact using digest(@sha256:...) rather than a tag(:%s) because resolved digest may not point to the same signed artifact, as tags are mutable.\n", reference)
}

// OnVerifySucceeded sets the successful verification result for the handler.
//
// outcomes must not be nil or empty.
func (h *VerifyHandler) OnVerifySucceeded(outcomes []*notation.VerificationOutcome, digestReference string) {
	h.outcome = outcomes[0]
	h.digestReference = digestReference
}

func (h *VerifyHandler) Render() error {
	// write out on success
	// print out warning for any failed result with logged verification action
	for _, result := range h.outcome.VerificationResults {
		if result.Error != nil {
			// at this point, the verification action has to be logged and
			// it's failed
			h.printer.ErrorPrintf("Warning: %v was set to %q and failed with error: %v\n", result.Type, result.Action, result.Error)
		}
	}
	if reflect.DeepEqual(h.outcome.VerificationLevel, trustpolicy.LevelSkip) {
		h.printer.Println("Trust policy is configured to skip signature verification for", h.digestReference)
	} else {
		h.printer.Println("Successfully verified signature for", h.digestReference)
		h.printMetadataIfPresent(h.outcome)
	}
	return nil
}

func (h *VerifyHandler) printMetadataIfPresent(outcome *notation.VerificationOutcome) {
	// the signature envelope is parsed as part of verification.
	// since user metadata is only printed on successful verification,
	// this error can be ignored
	metadata, _ := outcome.UserMetadata()

	if len(metadata) > 0 {
		h.printer.Println("\nThe artifact was signed with the following user metadata.")
		printMetadataMap(os.Stdout, metadata)
	}
}

// printMetadataMap prints out metadata given the metatdata map
//
// The metadata is additional information of text output.
func printMetadataMap(w io.Writer, metadata map[string]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "\nKEY\tVALUE\t")

	for k, v := range metadata {
		fmt.Fprintf(tw, "%v\t%v\t\n", k, v)
	}

	return tw.Flush()
}
