package configutil

import (
	"sync"
	"testing"
)

func TestLoadConfigOnce(t *testing.T) {
	config1, err := LoadConfigOnce()
	if err != nil {
		t.Fatal("LoadConfigOnce failed.")
	}
	config2, err := LoadConfigOnce()
	if err != nil {
		t.Fatal("LoadConfigOnce failed.")
	}
	if config1 != config2 {
		t.Fatal("LoadConfigOnce is invalid.")
	}
	configOnce = sync.Once{}
}

func TestLoadSigningKeysOnce(t *testing.T) {
	signingKeys1, err := LoadSigningkeysOnce()
	if err != nil {
		t.Fatal("LoadSigningkeysOnce failed.")
	}
	signingKeys2, err := LoadSigningkeysOnce()
	if err != nil {
		t.Fatal("LoadSigningkeysOnce failed.")
	}
	if signingKeys1 != signingKeys2 {
		t.Fatal("LoadSigningkeysOnce is invalid.")
	}
	signingKeysOnce = sync.Once{}
}
