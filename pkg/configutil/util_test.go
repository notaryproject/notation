package configutil

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/notaryproject/notation-go/dir"
)

func TestIsRegistryInsecure(t *testing.T) {
	configOnce = sync.Once{}
	// for restore dir
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		configOnce = sync.Once{}
	}(dir.UserConfigDir)
	// update config dir
	dir.UserConfigDir = "testdata"

	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "hit registry", args: args{target: "reg1.io"}, want: true},
		{name: "miss registry", args: args{target: "reg2.io"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRegistryInsecure(tt.args.target); got != tt.want {
				t.Errorf("IsRegistryInsecure() = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestIsRegistryInsecureMissingConfig(t *testing.T) {
	configOnce = sync.Once{}
	// for restore dir
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		configOnce = sync.Once{}
	}(dir.UserConfigDir)
	// update config dir
	dir.UserConfigDir = "./testdata2"

	type args struct {
		target string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "missing config", args: args{target: "reg1.io"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRegistryInsecure(tt.args.target); got != tt.want {
				t.Errorf("IsRegistryInsecure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRegistryInsecureConfigPermissionError(t *testing.T) {
	configDir := "./testdata"
	// for restore dir
	defer func(oldDir string) error {
		// restore permission
		dir.UserConfigDir = oldDir
		configOnce = sync.Once{}
		return os.Chmod(filepath.Join(configDir, "config.json"), 0644)
	}(dir.UserConfigDir)

	// update config dir
	dir.UserConfigDir = configDir

	// forbid reading the file
	if err := os.Chmod(filepath.Join(configDir, "config.json"), 0000); err != nil {
		t.Error(err)
	}

	if IsRegistryInsecure("reg1.io") {
		t.Error("should false because of missing config.json read permission.")
	}
}

func TestResolveKey(t *testing.T) {
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
	}(dir.UserConfigDir)

	t.Run("valid e2e key", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		keySuite, err := ResolveKey("e2e")
		if err != nil {
			t.Fatal(err)
		}
		if keySuite.Name != "e2e" {
			t.Error("key name is not correct.")
		}
	})

	t.Run("key name is empty (using default key)", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		keySuite, err := ResolveKey("")
		if err != nil {
			t.Fatal(err)
		}
		if keySuite.Name != "e2e" {
			t.Error("key name is not correct.")
		}
	})

	t.Run("signingkeys.json without read permission", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		defer func() error {
			// restore the permission
			return os.Chmod(filepath.Join(dir.UserConfigDir, "signingkeys.json"), 0644)
		}()

		// forbid reading the file
		if err := os.Chmod(filepath.Join(dir.UserConfigDir, "signingkeys.json"), 0000); err != nil {
			t.Error(err)
		}
		_, err := ResolveKey("")
		if !strings.Contains(err.Error(), "permission denied") {
			t.Error("should error with permission denied")
		}
	})
}
