package configutil

import (
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
}
