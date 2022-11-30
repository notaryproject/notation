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

func PrintPlugins(ctx context.Context, w io.Writer, v []plugin.Plugin, errors []error) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tVERSION\tCAPABILITIES\tERROR\t")
	for ind, p := range v {
		metaData := proto.GetMetadataResponse{}
		if p != nil {
			req := &proto.GetMetadataRequest{}
			metadata, err := p.GetMetadata(ctx, req)
			if err == nil {
				metaData = *metadata
			}
			errors[ind] = err
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%v\t\n",
			metaData.Name, metaData.Description, metaData.Version, metaData.Capabilities, errors[ind])
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

func PrintVerificationResults(w io.Writer, v []*notation.VerificationOutcome, resultErr error, digest string, isTag bool, tag string) error {
	if isTag {
		fmt.Println("Warning: Always verify artifact using digest(`@sha256:...`) rather than a tag(`:latest`) because tags are mutable and a tag reference can point to a different artifact than the one verified.")
		fmt.Printf("Resolved artifact tag %q to digest %q before verification.\n", tag, digest)
	}

	tw := newTabWriter(w)

	if resultErr == nil {
		fmt.Fprintf(tw, "Successfully verified for %s\n", digest)
		// TODO[https://github.com/notaryproject/notation/issues/304]: print out failed validations as warnings.
		return nil
	}
	fmt.Printf("Signature verification failed for all the signatures associated with digest: %s\n", digest)
	tw.Flush()

	return resultErr
}
