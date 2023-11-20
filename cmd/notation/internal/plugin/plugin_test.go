package plugin

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestGetPluginMetadataIfExist(t *testing.T) {
	_, err := GetPluginMetadataIfExist(context.Background(), "non-exist-plugin")
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist err, got: %v", err)
	}
}
