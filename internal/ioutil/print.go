package ioutil

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/notaryproject/notation-go"
	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/plugin"
	"github.com/notaryproject/notation-go/plugin/proto"
)

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

func PrintPlugins(ctx context.Context, w io.Writer, v []plugin.Plugin) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tVERSION\tCAPABILITIES\t")
	for _, p := range v {
		req := &proto.GetMetadataRequest{}
		metadata, err := p.GetMetadata(ctx, req)
		if err != nil {
			return err
		}
		fmt.Fprintf(tw, "%s\t%s\t%v\t%v\t\n",
			metadata.Name, metadata.Description, metadata.Version, metadata.Capabilities)
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

func PrintVerificationResults(w io.Writer, v []*notation.VerificationOutcome, resultErr error, digest string) error {
	tw := newTabWriter(w)

	if resultErr == nil {
		fmt.Fprintf(tw, "Signature verification succeeded for %s\n", digest)
		// TODO[https://github.com/notaryproject/notation/issues/304]: print out failed validations as warnings.
		return nil
	}

	fmt.Fprintf(tw, "ERROR: %s\n\n", resultErr.Error())
	printOutcomes(tw, v, digest)
	tw.Flush()

	return resultErr
}

func printOutcomes(tw *tabwriter.Writer, outcomes []*notation.VerificationOutcome, digest string) {
	fmt.Printf("Signature verification failed for all the %d signatures associated with digest: %s\n\n", len(outcomes), digest)

	// TODO[https://github.com/notaryproject/notation/issues/304]: print out detailed errors in debug mode.
	for idx, outcome := range outcomes {
		fmt.Printf("Signature #%d : %s\n", idx+1, outcome.Error.Error())
	}
}
