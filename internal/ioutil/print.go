package ioutil

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation/pkg/config"
)

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

func PrintPlugins(w io.Writer, v []*plugin.Plugin) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tVERSION\tURL\tSUPPORTED CONTRACTS\tCAPABILITIES\tERROR\tPATH\t")
	for _, p := range v {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%v\t%v\t%v\t%s\t\n",
			p.Name, p.Description, p.Version, p.URL, p.SupportedContractVersions, p.Capabilities, p.Err, p.Path)
	}
	return tw.Flush()
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

func PrintCertificateMap(w io.Writer, v []config.CertificateReference) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tPATH\t")
	for _, cert := range v {
		fmt.Fprintf(tw, "%s\t%s\t\n", cert.Name, cert.Path)
	}
	return tw.Flush()
}
