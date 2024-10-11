// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copied and adapted from oras (https://github.com/oras-project/oras)
/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trace

import (
	"errors"
	"net/http"
	"testing"
)

type TransportMock struct {
	resp    *http.Response
	respErr error
}

// RoundTrip calls base roundtrip while keeping track of the current request.
func (t *TransportMock) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp = t.resp
	err = t.respErr
	return
}

func TestTransport_RoundTrip(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		req := &http.Request{}
		transport := NewTransport(&TransportMock{})
		_, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatal("should have no error")
		}
	})

	t.Run("error response", func(t *testing.T) {
		errMsg := "response error"
		req := &http.Request{}
		transport := NewTransport(&TransportMock{respErr: errors.New(errMsg)})
		_, err := transport.RoundTrip(req)
		if err.Error() != errMsg {
			t.Fatalf("want err = %s, got %s", errMsg, err)
		}
	})

	t.Run("valid response", func(t *testing.T) {
		req := &http.Request{}
		transport := NewTransport(&TransportMock{resp: &http.Response{}})
		_, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatal("should have no error")
		}
	})

	t.Run("request header with Authorization", func(t *testing.T) {
		header := http.Header{}
		header.Add("Authorization", "abc")
		req := &http.Request{Header: header}
		transport := NewTransport(&TransportMock{resp: &http.Response{}})
		_, err := transport.RoundTrip(req)
		if err != nil {
			t.Fatal("should have no error")
		}
	})
}
