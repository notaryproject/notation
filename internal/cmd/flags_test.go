// Package cmd contains common flags and routines for all CLIs.
package cmd

import (
	"reflect"
	"testing"
)

func TestParseKeyValueListFlag(t *testing.T) {
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
		{"spaces", args{"a=b , c=d"}, map[string]string{"a": "b", "c": "d"}, false},
		{"quoted", args{"a=b,\"c\"=d"}, map[string]string{"a": "b", "c": "d"}, false},
		{"quoted comma", args{"a=b,\"c,h\"=d"}, map[string]string{"a": "b", "c,h": "d"}, false},
		{"empty value", args{"a=b,,c=d"}, nil, true},
		{"duplicated", args{"a=b,a=d"}, nil, true},
		{"malformed", args{"a=b,c:d"}, nil, true},
		{"only equal", args{"="}, nil, true},
		{"entry only equal", args{"a=b,="}, nil, true},
		{"entry only equal and space", args{"a=b, = "}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeyValueListFlag(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseKeyValueListFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseKeyValueListFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}
