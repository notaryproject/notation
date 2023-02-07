package version

import "testing"

func TestGetVersion(t *testing.T) {
	t.Run("BuildMetadata is empty", func(t *testing.T) {
		Version = "1.0"
		BuildMetadata = ""
		v := GetVersion()
		if Version != v {
			t.Errorf("Should return Version = %s, got %s", Version, v)
		}
	})

	t.Run("BuildMetadata is not empty", func(t *testing.T) {
		Version = "1.0"
		BuildMetadata = "unreleased"
		v := GetVersion()
		want := "1.0+unreleased"
		if want != v {
			t.Errorf("Should return Version = %s, got %s", want, v)
		}
	})
}
