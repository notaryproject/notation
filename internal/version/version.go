package version

var (
	// Version shows the current notation version, optionally with pre-release.
	Version = "v0.10.0-alpha.3"

	// BuildMetadata stores the build metadata.
	BuildMetadata = "unreleased"
)

// GetVersion returns the version string in SemVer 2.
func GetVersion() string {
	if BuildMetadata == "" {
		return Version
	}
	return Version + "+" + BuildMetadata
}
