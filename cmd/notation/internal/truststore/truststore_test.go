package truststore

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestEmptyCertFile(t *testing.T) {
	path := filepath.FromSlash("../../../../internal/testdata/Empty.txt")
	expectedErr := errors.New("no valid certificate found in the empty file")
	err := AddCert(path, "ca", "test", false)
	if err == nil || err.Error() != "no valid certificate found in the file" {
		t.Fatalf("expected err: %v, got: %v", expectedErr, err)
	}
}
