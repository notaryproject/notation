package ioutil

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/config"
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

func PrintVerificationResults(w io.Writer, v []*notation.VerificationOutcome, resultErr error, digest string) error {
	tw := newTabWriter(w)

	if resultErr == nil {
		fmt.Fprintf(tw, "Successfully verified for %s\n", digest)
		// TODO[https://github.com/notaryproject/notation/issues/304]: print out failed validations as warnings.
		return nil
	}
	tw.Flush()

	return resultErr
}
