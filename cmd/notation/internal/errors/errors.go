package errors

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

// ErrorOCILayoutTarForSign is used when signing local content in oci layout
// tarball, but failed to get an oci.ReadOnlyStorage
type ErrorOCILayoutTarForSign struct {
	Msg string
}

func (e ErrorOCILayoutTarForSign) Error() string {
	if e.Msg != "" {
		return "failed to create ReadOnlyStorage from tar: " + e.Msg
	}
	return "failed to create ReadOnlyStorage from tar"
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
	return "reference is missing digest or tag"
}
