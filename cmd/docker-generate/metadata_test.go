package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/notaryproject/notation/internal/docker"
)

func TestMetdaDatCommand(t *testing.T) {
	oldStdOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Create pipe for metadata cmd failed: %v", err)

	}
	os.Stdout = w
	cmd := metadataCommand()
	if err := cmd.RunE(cmd, []string{}); err != nil {
		t.Fatalf("Running metadata cmd failed: %v", err)
	}
	w.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Read metadata from stdout failed: %v", err)
	}
	var got docker.PluginMetadata
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal metadata failed: %v", err)
	}
	if got != pluginMetadata {
		t.Fatalf("Expect Metadata: %v, got: %v", data, pluginMetadata)
	}
	defer func() {
		os.Stdout = oldStdOut
		r.Close()
	}()
}
