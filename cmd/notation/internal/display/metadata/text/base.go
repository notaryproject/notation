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

// Package text provides the text output in human-readable format for metadata
// information.
package text

import (
	"fmt"
	"reflect"
	"text/tabwriter"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/verifier/trustpolicy"
	"github.com/notaryproject/notation/cmd/notation/internal/display/output"
)

// PrintVerificationSuccess prints out messages when verification succeeds
func PrintVerificationSuccess(printer *output.Printer, outcome *notation.VerificationOutcome, printout string, hasWarning bool) error {
	// write out on success
	// print out warning for any failed result with logged verification action
	for _, result := range outcome.VerificationResults {
		if result.Error != nil {
			// at this point, the verification action has to be logged and
			// it's failed
			printer.PrintErrorf("Warning: %v was set to %q and failed with error: %v\n", result.Type, result.Action, result.Error)
			hasWarning = true
		}
	}
	if hasWarning {
		// print a newline to separate the warning from the final message
		printer.Println()
	}
	if reflect.DeepEqual(outcome.VerificationLevel, trustpolicy.LevelSkip) {
		printer.Println("Trust policy is configured to skip signature verification for", printout)
	} else {
		printer.Println("Successfully verified signature for", printout)
		PrintUserMetadataIfPresent(printer, outcome)
	}
	return nil
}

// PrintUserMetadataIfPresent prints out user metadata if present
func PrintUserMetadataIfPresent(printer *output.Printer, outcome *notation.VerificationOutcome) {
	// the signature envelope is parsed as part of verification.
	// since user metadata is only printed on successful verification,
	// this error can be ignored.
	metadata, _ := outcome.UserMetadata()
	if len(metadata) > 0 {
		printer.Println("\nThe artifact was signed with the following user metadata.")
		printUserMetadataMap(printer, metadata)
	}
}

// printUserMetadataMap prints out user metadata map
func printUserMetadataMap(printer *output.Printer, metadata map[string]string) error {
	tw := tabwriter.NewWriter(printer, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "\nKEY\tVALUE\t")
	for k, v := range metadata {
		fmt.Fprintf(tw, "%v\t%v\t\n", k, v)
	}
	return tw.Flush()
}
