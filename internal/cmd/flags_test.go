// Package cmd contains common flags and routines for all CLIs.
package cmd

import (
	"reflect"
	"testing"
)

func TestParseFlagPluginConfig(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{"nil", args{nil}, nil, false},
		{"empty", args{[]string{}}, nil, false},
		{"single", args{[]string{"a=b"}}, map[string]string{"a": "b"}, false},
		{"multiple", args{[]string{"a=b", "c=d"}}, map[string]string{"a": "b", "c": "d"}, false},
		{"quoted", args{[]string{"a=b", "\"c\"=d"}}, map[string]string{"a": "b", "\"c\"": "d"}, false},
		{"duplicated", args{[]string{"a=b", "a=d"}}, nil, true},
		{"malformed", args{[]string{"a=b", "c:d"}}, nil, true},
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
