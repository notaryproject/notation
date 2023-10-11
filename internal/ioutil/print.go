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
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"text/tabwriter"

	"github.com/notaryproject/notation-go/config"
)

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

// PrintKeyMap prints out key information given array of KeySuite
func PrintKeyMap(w io.Writer, target *string, v []config.KeySuite) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tKEY PATH\tCERTIFICATE PATH\tID\tPLUGIN NAME\t")
	for _, key := range v {
		name := key.Name
		if target != nil && key.Name == *target {
			name = "* " + name
		}
		kp := key.X509KeyPair
		if kp == nil {
			kp = &config.X509KeyPair{}
		}
		ext := key.ExternalKey
		if ext == nil {
			ext = &config.ExternalKey{}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t\n", name, kp.KeyPath, kp.CertificatePath, ext.ID, ext.PluginName)
	}
	return tw.Flush()
}

// PrintMetadataMap prints out metadata given the metatdata map
func PrintMetadataMap(w io.Writer, metadata map[string]string) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "\nKEY\tVALUE\t")

	for k, v := range metadata {
		fmt.Fprintf(tw, "%v\t%v\t\n", k, v)
	}

	return tw.Flush()
}

// PrintCertMap lists certificate files in the trust store given array of cert
// paths
func PrintCertMap(w io.Writer, certPaths []string) error {
	if len(certPaths) == 0 {
		return nil
	}
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "STORE TYPE\tSTORE NAME\tCERTIFICATE\t")
	for _, cert := range certPaths {
		fileName := filepath.Base(cert)
		dir := filepath.Dir(cert)
		namedStore := filepath.Base(dir)
		dir = filepath.Dir(dir)
		storeType := filepath.Base(dir)
		fmt.Fprintf(tw, "%s\t%s\t%s\t\n", storeType, namedStore, fileName)
	}
	return tw.Flush()
}

// PrintObjectAsJSON takes an interface and prints it as an indented JSON string
func PrintObjectAsJSON(i interface{}) error {
	jsonBytes, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

	return nil
}
