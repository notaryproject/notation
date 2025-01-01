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

package ioutil

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPrintObjectAsTree(t *testing.T) {
	tests := []struct {
		name     string
		buildObj func() *Node
		want     string
	}{
		{
			name: "Detailed signature inspection",
			buildObj: func() *Node {
				root := New("docker.io/holiodin01/net-monitor@sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208")
				sigNode := root.Add("application/vnd.cncf.notary.signature")

				// First signature
				sig1 := sigNode.Add("sha256:efaea8b26ef007bfbdaecab63b068320e508921b7687e38b892fd4ce71a92f4e")
				sig1.AddPair("media type", "application/jose+json")
				sig1.AddPair("signature algorithm", "RSASSA-PSS-SHA-256")

				// Signed attributes for sig1
				signedAttr1 := sig1.Add("signed attributes")
				signedAttr1.AddPair("signingScheme", "notary.x509")
				signedAttr1.AddPair("signingTime", "Sat Dec 28 11:11:16 2024")

				// User defined attributes for sig1
				userAttr1 := sig1.Add("user defined attributes")
				userAttr1.Add("(empty)")

				// Unsigned attributes for sig1
				unsignedAttr1 := sig1.Add("unsigned attributes")
				unsignedAttr1.AddPair("signingAgent", "notation-go/1.3.0+unreleased")

				// Certificates for sig1
				certs1 := sig1.Add("certificates")
				cert1 := certs1.Add("SHA256 fingerprint: 1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e")
				cert1.AddPair("issued to", "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US")
				cert1.AddPair("issued by", "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US")
				cert1.AddPair("expiry", "Sun Dec 29 05:40:12 2024")

				// Signed artifact for sig1
				artifact1 := sig1.Add("signed artifact")
				artifact1.AddPair("media type", "application/vnd.docker.distribution.manifest.v2+json")
				artifact1.AddPair("digest", "sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208")
				artifact1.AddPair("size", "942")

				// Second signature with same structure
				sig2 := sigNode.Add("sha256:7612778026574769361c778fa8c7b8935f273ad85425b637561167c6629ce369")
				sig2.AddPair("media type", "application/jose+json")
				sig2.AddPair("signature algorithm", "RSASSA-PSS-SHA-256")

				signedAttr2 := sig2.Add("signed attributes")
				signedAttr2.AddPair("signingScheme", "notary.x509")
				signedAttr2.AddPair("signingTime", "Sat Dec 28 11:13:47 2024")

				userAttr2 := sig2.Add("user defined attributes")
				userAttr2.Add("(empty)")

				unsignedAttr2 := sig2.Add("unsigned attributes")
				unsignedAttr2.AddPair("signingAgent", "notation-go/1.3.0+unreleased")

				certs2 := sig2.Add("certificates")
				cert2 := certs2.Add("SHA256 fingerprint: 1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e")
				cert2.AddPair("issued to", "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US")
				cert2.AddPair("issued by", "CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US")
				cert2.AddPair("expiry", "Sun Dec 29 05:40:12 2024")

				artifact2 := sig2.Add("signed artifact")
				artifact2.AddPair("media type", "application/vnd.docker.distribution.manifest.v2+json")
				artifact2.AddPair("digest", "sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208")
				artifact2.AddPair("size", "942")

				return root
			},
			want: `docker.io/holiodin01/net-monitor@sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208
└── application/vnd.cncf.notary.signature
    ├── sha256:efaea8b26ef007bfbdaecab63b068320e508921b7687e38b892fd4ce71a92f4e
    │   ├── media type: application/jose+json
    │   ├── signature algorithm: RSASSA-PSS-SHA-256
    │   ├── signed attributes
    │   │   ├── signingScheme: notary.x509
    │   │   └── signingTime: Sat Dec 28 11:11:16 2024
    │   ├── user defined attributes
    │   │   └── (empty)
    │   ├── unsigned attributes
    │   │   └── signingAgent: notation-go/1.3.0+unreleased
    │   ├── certificates
    │   │   └── SHA256 fingerprint: 1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e
    │   │       ├── issued to: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
    │   │       ├── issued by: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
    │   │       └── expiry: Sun Dec 29 05:40:12 2024
    │   └── signed artifact
    │       ├── media type: application/vnd.docker.distribution.manifest.v2+json
    │       ├── digest: sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208
    │       └── size: 942
    └── sha256:7612778026574769361c778fa8c7b8935f273ad85425b637561167c6629ce369
        ├── media type: application/jose+json
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signed attributes
        │   ├── signingScheme: notary.x509
        │   └── signingTime: Sat Dec 28 11:13:47 2024
        ├── user defined attributes
        │   └── (empty)
        ├── unsigned attributes
        │   └── signingAgent: notation-go/1.3.0+unreleased
        ├── certificates
        │   └── SHA256 fingerprint: 1487c63cf13ae92488c2bd3d9bcb95687c4bb8e90d501a2f15956e42e47ba06e
        │       ├── issued to: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
        │       ├── issued by: CN=valid-example,O=Notary,L=Seattle,ST=WA,C=US
        │       └── expiry: Sun Dec 29 05:40:12 2024
        └── signed artifact
            ├── media type: application/vnd.docker.distribution.manifest.v2+json
            ├── digest: sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208
            └── size: 942`,
		},
		{
			name: "Simple signature list",
			buildObj: func() *Node {
				root := New("docker.io/holiodin01/net-monitor@sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208")
				sigNode := root.Add("application/vnd.cncf.notary.signature")
				sigNode.Add("sha256:efaea8b26ef007bfbdaecab63b068320e508921b7687e38b892fd4ce71a92f4e")
				sigNode.Add("sha256:7612778026574769361c778fa8c7b8935f273ad85425b637561167c6629ce369")
				return root
			},
			want: `docker.io/holiodin01/net-monitor@sha256:984a85875f80346ec12aab68fefa1d816dd3916175cf18938f498d2f0677c208
└── application/vnd.cncf.notary.signature
    ├── sha256:efaea8b26ef007bfbdaecab63b068320e508921b7687e38b892fd4ce71a92f4e
    └── sha256:7612778026574769361c778fa8c7b8935f273ad85425b637561167c6629ce369`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the function
			if err := PrintObjectAsTree(tt.buildObj()); err != nil {
				t.Errorf("PrintObjectAsTree() error = %v", err)
			}

			// Restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			got := strings.TrimSpace(buf.String())

			// Compare with expected output
			if got != tt.want {
				t.Errorf("PrintObjectAsTree() output mismatch:\ngot:\n%s\n\nwant:\n%s", got, tt.want)
			}
		})
	}
}
