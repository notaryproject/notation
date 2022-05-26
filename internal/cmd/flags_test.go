// Package cmd contains common flags and routines for all CLIs.
package cmd

import (
	"reflect"
	"testing"
)

func TestParseFlagPluginConfig(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{"empty", args{""}, nil, false},
		{"single", args{"a=b"}, map[string]string{"a": "b"}, false},
		{"multiple", args{"a=b,c=d"}, map[string]string{"a": "b", "c": "d"}, false},
		{"quoted", args{"a=b,\"c\"=d"}, map[string]string{"a": "b", "\"c\"": "d"}, false},
		{"duplicated", args{"a=b,a=d"}, nil, true},
		{"malformed", args{"a=b,c:d"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlagPluginConfig(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlagPluginConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFlagPluginConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
