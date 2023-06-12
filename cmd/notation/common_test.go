package main

import (
	"reflect"
	"testing"

	"oras.land/oras-go/v2/registry/remote/auth"
)

func TestSecureFlagOpts_Credential(t *testing.T) {
	tests := []struct {
		name string
		opts *SecureFlagOpts
		want auth.Credential
	}{
		{
			name: "Username and password",
			opts: &SecureFlagOpts{
				Username: "username",
				Password: "password",
			},
			want: auth.Credential{
				Username: "username",
				Password: "password",
			},
		},
		{
			name: "Username only",
			opts: &SecureFlagOpts{
				Username: "username",
			},
			want: auth.Credential{
				Username: "username",
			},
		},
		{
			name: "Password only",
			opts: &SecureFlagOpts{
				Password: "token",
			},
			want: auth.Credential{
				RefreshToken: "token",
			},
		},
		{
			name: "Empty username and password",
			opts: &SecureFlagOpts{
				Username: "",
				Password: "",
			},
			want: auth.EmptyCredential,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.Credential(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SecureFlagOpts.Credential() = %v, want %v", got, tt.want)
			}
		})
	}
}
