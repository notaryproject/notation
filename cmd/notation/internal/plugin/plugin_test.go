package plugin

import (
	"context"
	"testing"
)

func TestCheckPluginExistence(t *testing.T) {
	exist, err := CheckPluginExistence(context.Background(), "non-exist-plugin")
	if exist || err != nil {
		t.Fatalf("expected exist to be false with nil err, got: %v, %s", exist, err)
	}
}
