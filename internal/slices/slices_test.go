package slices

import (
	"testing"
)

func TestContainerElement(t *testing.T) {
	tests := []struct {
		c    []string
		v    string
		want bool
	}{
		{nil, "", false},
		{[]string{}, "", false},
		{[]string{"1", "2", "3"}, "4", false},
		{[]string{"1", "2", "3"}, "2", true},
		{[]string{"1", "2", "2", "3"}, "2", true},
		{[]string{"1", "2", "3", "2"}, "2", true},
	}
	for _, tt := range tests {
		if got := Contains(tt.c, tt.v); got != tt.want {
			t.Errorf("ContainerElement() = %v, want %v", got, tt.want)
		}
	}
}
