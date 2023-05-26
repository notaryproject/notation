package version

var (
	// Version shows the current notation version, optionally with pre-release.
	Version = "v1.0.0-rc.7"

	// BuildMetadata stores the build metadata.
	//
	// When execute `make build` command, it will be overridden by
	// environment variable `BUILD_METADATA`. If commit tag was set,
	// BuildMetadata will be empty.
	BuildMetadata = "unreleased"

	// GitCommit stores the git HEAD commit id
	GitCommit = ""
)

// GetVersion returns the version string in SemVer 2.
func GetVersion() string {
	if BuildMetadata == "" {
		return Version
	}
	return Version + "+" + BuildMetadata
}
