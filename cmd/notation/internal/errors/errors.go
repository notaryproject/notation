package errors

import "fmt"

// ErrorReferrersAPINotSupported is used when the target registry does not
// support the Referrers API
type ErrorReferrersAPINotSupported struct {
	Msg string
}

func (e ErrorReferrersAPINotSupported) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return "referrers API not supported"
}

// ErrorOCILayoutMissingReference is used when signing local content in oci
// layout folder but missing input tag or digest.
type ErrorOCILayoutMissingReference struct {
	Msg string
}

func (e ErrorOCILayoutMissingReference) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return "reference is missing either digest or tag"
}

// ErrorExceedMaxSignatures is used when the number of signatures has surpassed
// the maximum limit that can be evaluated.
type ErrorExceedMaxSignatures struct {
	MaxSignatures int
}

func (e ErrorExceedMaxSignatures) Error() string {
	return fmt.Sprintf("exceeded configured limit of max signatures %d to examine", e.MaxSignatures)
}
