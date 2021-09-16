package docker

// PluginMetadataCommandName is the internal command name for docker CLI plugin metadata.
const PluginMetadataCommandName = "docker-cli-plugin-metadata"

// PluginMetadata presents the plugin metadata to the docker CLI.
type PluginMetadata struct {
	SchemaVersion    string `json:"SchemaVersion,omitempty"`
	Vendor           string `json:"Vendor,omitempty"`
	Version          string `json:"Version,omitempty"`
	ShortDescription string `json:"ShortDescription,omitempty"`
	URL              string `json:"URL,omitempty"`
	Experimental     bool   `json:"Experimental,omitempty"`
}
