package tuf

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/docker/go/canonical/json"
	"github.com/notaryproject/nv2/cmd/nv2/common"
	"github.com/notaryproject/nv2/pkg/tuf"
	"github.com/notaryproject/nv2/pkg/tuf/local"
	"github.com/theupdateframework/notary/tuf/utils"
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
			Usage:     "certs for verification",
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:      "ca-cert",
			Usage:     "CA certs for verification",
			TakesFile: true,
		},
		&cli.IntFlag{
			Name:    "min-version",
			Aliases: []string{"m"},
			Usage:   "min version of the signature",
		},
		&cli.StringFlag{
			Name:    "reference",
			Aliases: []string{"r"},
			Usage:   "original reference",
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
	verifier, err := getVerifier(ctx)
	if err != nil {
		return err
	}
	sig, err := readSignatrueFile(ctx.String("signature"))
	if err != nil {
		return err
	}

	// core process
	minVer := ctx.Int("min-version")
	targets, err := tuf.VerifyTargets(ctx.Context, verifier, sig, minVer)
	if err != nil {
		return fmt.Errorf("verification failure: %v", err)
	}
	manifest, err := common.GetManifestFromContext(ctx)
	if err != nil {
		return err
	}
	if reference := ctx.String("reference"); reference != "" {
		manifest.Name = reference
	}
	if !tuf.IsManifestInTargets(manifest, targets) {
		return fmt.Errorf("verification failure: %s: not found", manifest.Digests[0])
	}

	// write out
	fmt.Println(manifest.Digests[0])
	return nil
}

func readSignatrueFile(path string) (*tuf.Signed, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	signed := new(tuf.Signed)
	if err := json.Unmarshal(bytes, signed); err != nil {
		return nil, err
	}
	return signed, nil
}

func getVerifier(ctx *cli.Context) (tuf.Verifier, error) {
	roots := x509.NewCertPool()

	var certs []*x509.Certificate
	for _, path := range ctx.StringSlice("cert") {
		bundledCerts, err := utils.LoadCertBundleFromFile(path)
		if err != nil {
			return nil, err
		}
		certs = append(certs, bundledCerts...)
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}
	for _, path := range ctx.StringSlice("ca-cert") {
		bundledCerts, err := utils.LoadCertBundleFromFile(path)
		if err != nil {
			return nil, err
		}
		for _, cert := range bundledCerts {
			roots.AddCert(cert)
		}
	}

	return local.NewVerifier(certs, roots)
}
