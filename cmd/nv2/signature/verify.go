package signature

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/notaryproject/nv2/cmd/nv2/common"
	"github.com/notaryproject/nv2/pkg/signature"
	x509nv2 "github.com/notaryproject/nv2/pkg/signature/x509"
	"github.com/urfave/cli/v2"
)

// VerifyCommand defines verify command
var VerifyCommand = &cli.Command{
	Name:      "verify",
	Usage:     "verifies OCI Artifacts",
	ArgsUsage: "[<scheme://reference>]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "signature",
			Aliases:   []string{"s", "f"},
			Usage:     "signature file",
			Required:  true,
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:      "cert",
			Aliases:   []string{"c"},
			Usage:     "certs for verification [x509]",
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:      "ca-cert",
			Usage:     "CA certs for verification [x509]",
			TakesFile: true,
		},
		common.MediaTypeFlag,
		common.UsernameFlag,
		common.PasswordFlag,
		common.InsecureFlag,
	},
	Action: runVerify,
}

func runVerify(ctx *cli.Context) error {
	// initialize
	scheme, err := getSchemeForVerification(ctx)
	if err != nil {
		return err
	}
	sig, err := readSignatrueFile(ctx.String("signature"))
	if err != nil {
		return err
	}

	// core process
	claims, err := scheme.Verify(sig)
	if err != nil {
		return fmt.Errorf("verification failure: %v", err)
	}
	manifest, err := common.GetManifestFromContext(ctx)
	if err != nil {
		return err
	}
	descriptor := signature.DescriptorFromReference(manifest.Descriptor)

	if descriptor != claims.Manifest.Descriptor {
		return fmt.Errorf("verification failure: %s: ", descriptor.Digest)
	}

	// write out
	fmt.Println(descriptor.Digest)
	return nil
}

func readSignatrueFile(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func getSchemeForVerification(ctx *cli.Context) (*signature.Scheme, error) {
	scheme := signature.NewScheme()

	// add x509 verifier
	verifier, err := getX509Verifier(ctx)
	if err != nil {
		return nil, err
	}
	scheme.RegisterVerifier(verifier)

	return scheme, nil
}

func getX509Verifier(ctx *cli.Context) (signature.Verifier, error) {
	roots := x509.NewCertPool()

	var certs []*x509.Certificate
	for _, path := range ctx.StringSlice("cert") {
		bundledCerts, err := x509nv2.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		certs = append(certs, bundledCerts...)
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}
	for _, path := range ctx.StringSlice("ca-cert") {
		bundledCerts, err := x509nv2.ReadCertificateFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}

	return x509nv2.NewVerifier(certs, roots)
}
