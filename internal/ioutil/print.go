package ioutil

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/notaryproject/notation-go/config"
)

const (
	OutputPlaintext = "text"
	OutputJson      = "json"
)

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

func PrintKeyMap(w io.Writer, target string, v []config.KeySuite) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tKEY PATH\tCERTIFICATE PATH\tID\tPLUGIN NAME\t")
	for _, key := range v {
		name := key.Name
		if key.Name == target {
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

func PrintMetadataMap(w io.Writer, metadata map[string]string) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "\nKEY\tVALUE\t")

	for k, v := range metadata {
		fmt.Fprintf(tw, "%v\t%v\t\n", k, v)
	}

	return tw.Flush()
}

func PrintObjectAsJson(i interface{}) error {
	jsonBytes, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonBytes))

	return nil
}
