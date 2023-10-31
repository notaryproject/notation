package plugin

import (
	"context"
	"runtime"
	"testing"

	"github.com/notaryproject/notation-go/dir"
)

func TestCheckPluginExistence(t *testing.T) {
	dir.UserConfigDir = "testdata"
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows")
	}
	exist, err := CheckPluginExistence(context.Background(), "non-exist-plugin")
	if exist || err != nil {
		t.Fatalf("expected exist to be false with nil err, got: %v, %s", exist, err)
	}

	exist, err = CheckPluginExistence(context.Background(), "test-plugin")
	if !exist || err != nil {
		t.Fatalf("expected exist to be true with nil err, got: %v, %s", exist, err)
	}
}
