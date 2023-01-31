package configutil

import (
	"strings"
	"sync"
	"testing"

	"github.com/notaryproject/notation-go/dir"
)

func TestIsRegistryInsecure(t *testing.T) {
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

func TestResolveKey(t *testing.T) {
	defer func(oldDir string) {
		dir.UserConfigDir = oldDir
		signingKeysOnce = sync.Once{}
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
		signingKeysOnce = sync.Once{}
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
		signingKeysOnce = sync.Once{}
	})

	t.Run("key name doesn't exist", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/valid_signingkeys"
		_, err := ResolveKey("e2e2")
		if !strings.Contains(err.Error(), "signing key not found") {
			t.Error("should error")
		}
		signingKeysOnce = sync.Once{}
	})

	t.Run("key name is empty (no default key)", func(t *testing.T) {
		dir.UserConfigDir = "./testdata/no_default_key_signingkeys"
		_, err := ResolveKey("")
		if !strings.Contains(err.Error(), "default signing key not set.") {
			t.Error("should error")
		}
		signingKeysOnce = sync.Once{}
	})
}