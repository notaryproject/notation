package slices

import (
	"reflect"
	"testing"
)

type iss string

func (i iss) Is(v string) bool { return string(i) == v }

func TestIndex(t *testing.T) {
	tests := []struct {
		s    []iss
		v    string
		want int
	}{
		{nil, "", -1},
		{[]iss{}, "", -1},
		{[]iss{"1", "2", "3"}, "2", 1},
		{[]iss{"1", "2", "2", "3"}, "2", 1},
		{[]iss{"1", "2", "3", "2"}, "2", 1},
	}
	for _, tt := range tests {
		if got := Index(tt.s, tt.v); got != tt.want {
			t.Errorf("Index() = %v, want %v", got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s    []iss
		v    string
		want bool
	}{
		{nil, "", false},
		{[]iss{}, "", false},
		{[]iss{"1", "2", "3"}, "2", true},
		{[]iss{"1", "2", "2", "3"}, "2", true},
		{[]iss{"1", "2", "3", "2"}, "2", true},
	}
	for _, tt := range tests {
		if got := Contains(tt.s, tt.v); got != tt.want {
			t.Errorf("Index() = %v, want %v", got, tt.want)
		}
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		s    []iss
		i    int
		want []iss
	}{
		{[]iss{"1", "2", "3"}, 1, []iss{"1", "3"}},
		{[]iss{"1", "2", "2", "3"}, 2, []iss{"1", "2", "3"}},
		{[]iss{"1", "2", "3", "2"}, 2, []iss{"1", "2", "2"}},
	}
	for _, tt := range tests {
		if got := Delete(tt.s, tt.i); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Delete() = %v, want %v", got, tt.want)
		}
	}
}
