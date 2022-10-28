package cert

import (
	"reflect"
	"testing"
)

func TestCertListCommand(t *testing.T) {
	opts := &certListOpts{}
	cmd := certListCommand(opts)
	expected := &certListOpts{
		storeType:  "ca",
		namedStore: "test",
	}
	if err := cmd.ParseFlags([]string{
		"-t", "ca",
		"-s", "test"}); err != nil {
		t.Fatalf("Parse Flag failed: %v", err)
	}
	if !reflect.DeepEqual(*expected, *opts) {
		t.Fatalf("Expect cert list opts: %v, got: %v", expected, opts)
	}
}
