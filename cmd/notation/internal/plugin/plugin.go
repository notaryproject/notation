package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
	notationerrors "github.com/notaryproject/notation/cmd/notation/internal/errors"
	"github.com/notaryproject/notation/internal/osutil"
	"github.com/opencontainers/go-digest"
)

// Type is an enum for plugin source types.
type Type string

const (
	TypeFile Type = "file"
	TypeURL  Type = "url"
)

var (
	Types = []Type{
		TypeFile,
		TypeURL,
	}
)

// CheckPluginExistence returns true if a plugin already exists
func CheckPluginExistence(ctx context.Context, pluginName string) (bool, error) {
	mgr := plugin.NewCLIManager(dir.PluginFS())
	_, err := mgr.Get(ctx, pluginName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ValidateInstallSource returns true if source is in Types
func ValidateInstallSource(source string) bool {
	for _, t := range Types {
		if strings.ToLower(source) == string(t) {
			return true
		}
	}
	return false
}

// ValidateCheckSum returns nil if SHA256 of file at path equals to checkSum.
func ValidateCheckSum(path string, checkSum string) error {
	r, err := os.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()
	dgst, err := digest.FromReader(r)
	if err != nil {
		return err
	}
	enc := dgst.Encoded()
	if enc != checkSum {
		return fmt.Errorf("plugin checkSum does not match user input. User input is %s, got %s", checkSum, enc)
	}
	return nil
}

// ValidatePluginMetadata validates plugin metadata given plugin name and path
// returns the plugin version on success
func ValidatePluginMetadata(ctx context.Context, pluginName, path string) (string, error) {
	plugin, err := plugin.NewCLIPlugin(ctx, pluginName, path)
	if err != nil {
		return "", err
	}
	metadata, err := plugin.GetMetadata(ctx, &proto.GetMetadataRequest{})
	if err != nil {
		return "", err
	}
	return metadata.Version, nil
}

// ExtractPluginNameFromExecutableFileName gets plugin name from plugin
// executable file name based on spec: https://github.com/notaryproject/specifications/blob/main/specs/plugin-extensibility.md#installation
func ExtractPluginNameFromExecutableFileName(execFileName string) (string, error) {
	fileName := osutil.FileNameWithoutExtension(execFileName)
	_, pluginName, found := strings.Cut(fileName, "-")
	if !found || !strings.HasPrefix(fileName, proto.Prefix) {
		return "", notationerrors.ErrorInvalidPluginName{Msg: fmt.Sprintf("invalid plugin executable file name. file name requires format notation-{plugin-name}, got %s", fileName)}
	}
	return pluginName, nil
}
