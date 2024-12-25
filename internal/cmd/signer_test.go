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

package cmd

import (
	"testing"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/signer"
)

func TestGenericSignerImpl(t *testing.T) {
	g := &signer.GenericSigner{}
	if _, ok := interface{}(g).(notation.Signer); !ok {
		t.Fatal("GenericSigner does not implement notation.Signer")
	}

	if _, ok := interface{}(g).(notation.BlobSigner); !ok {
		t.Fatal("GenericSigner does not implement notation.BlobSigner")
	}
}

func TestPluginSignerImpl(t *testing.T) {
	p := &signer.PluginSigner{}
	if _, ok := interface{}(p).(notation.Signer); !ok {
		t.Fatal("PluginSigner does not implement notation.Signer")
	}

	if _, ok := interface{}(p).(notation.BlobSigner); !ok {
		t.Fatal("PluginSigner does not implement notation.BlobSigner")
	}
}
