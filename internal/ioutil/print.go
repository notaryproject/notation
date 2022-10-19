package ioutil

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/notaryproject/notation-go/config"
	"github.com/notaryproject/notation-go/plugin/manager"
	"github.com/notaryproject/notation-go/verification"
)

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

func PrintPlugins(w io.Writer, v []*manager.Plugin) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tDESCRIPTION\tVERSION\tCAPABILITIES\tERROR\t")
	for _, p := range v {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%v\t%v\t\n",
			p.Name, p.Description, p.Version, p.Capabilities, p.Err)
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

func PrintVerificationResults(w io.Writer, v []*verification.SignatureVerificationOutcome, resultErr error, digest string) error {
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

func printOutcomes(tw *tabwriter.Writer, outcomes []*verification.SignatureVerificationOutcome, digest string) {
	fmt.Printf("Signature verification failed for all the %d signatures associated with digest: %s\n\n", len(outcomes), digest)

	// TODO[https://github.com/notaryproject/notation/issues/304]: print out detailed errors in debug mode.
	for idx, outcome := range outcomes {
		fmt.Printf("Signature #%d : %s\n", idx+1, outcome.Error.Error())
	}
}

func PrintPolicyMap(w io.Writer, v []*verification.TrustPolicy) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "NAME\tSCOPE\tVERIFICATION_LEVEL\tTRUST_STORE\tTRUSTED_IDENTITY\t")
	for _, policy := range v {
		fmt.Fprintf(
			tw,
			"%s\t%s\t%+v\t%s\t%s\n",
			policy.Name,
			policy.RegistryScopes,
			policy.SignatureVerification,
			policy.TrustStores,
			policy.TrustedIdentities,
		)
	}
	return tw.Flush()
}

func PrintPolicyNames(w io.Writer, v []*verification.TrustPolicy, header string) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, header)
	for _, policy := range v {
		fmt.Fprintln(tw, policy.Name)
	}
	return tw.Flush()
}

func PrintDeletedPolicyNames(w io.Writer, v []string) error {
	tw := newTabWriter(w)
	fmt.Fprintln(tw, "Deleted Policies:")
	for _, name := range v {
		fmt.Fprintln(tw, name)
	}
	return tw.Flush()
}
