package signature

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/notaryproject/nv2/cmd/nv2/common"
	"github.com/notaryproject/nv2/pkg/signature"
	"github.com/notaryproject/nv2/pkg/signature/x509"
	"github.com/urfave/cli/v2"
)

const signerID = "nv2"

// SignCommand defines sign command
var SignCommand = &cli.Command{
	Name:      "sign",
	Usage:     "signs OCI Artifacts",
	ArgsUsage: "[<scheme://reference>]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "method",
			Aliases:  []string{"m"},
			Usage:    "signing method",
			Required: true,
		},
		&cli.StringFlag{
			Name:      "key",
			Aliases:   []string{"k"},
			Usage:     "signing key file [x509]",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "cert",
			Aliases:   []string{"c"},
			Usage:     "signing cert [x509]",
			TakesFile: true,
		},
		&cli.StringSliceFlag{
			Name:    "reference",
			Aliases: []string{"r"},
			Usage:   "original references",
		},
		common.ExpiryFlag,
		common.OutputFlag,
		common.MediaTypeFlag,
		common.UsernameFlag,
		common.PasswordFlag,
		common.InsecureFlag,
	},
	Action: runSign,
}

func runSign(ctx *cli.Context) error {
	// initialize
	scheme, err := getSchemeForSigning(ctx)
	if err != nil {
		return err
	}

	// core process
	claims, err := prepareClaimsForSigning(ctx)
	if err != nil {
		return err
	}
	sig, err := scheme.Sign(signerID, claims)
	if err != nil {
		return err
	}

	// write out
	path := ctx.String(common.OutputFlag.Name)
	if path == "" {
		path = strings.Split(claims.Manifest.Digest, ":")[1] + ".nv2"
	}
	if err := ioutil.WriteFile(path, []byte(sig), 0666); err != nil {
		return err
	}

	fmt.Println(claims.Manifest.Digest)
	return nil
}

func prepareClaimsForSigning(ctx *cli.Context) (signature.Claims, error) {
	manifest, err := common.GetManifestFromContext(ctx)
	if err != nil {
		return signature.Claims{}, err
	}
	now := time.Now()
	nowUnix := now.Unix()
	claims := signature.Claims{
		Manifest: signature.Manifest{
			Descriptor: signature.DescriptorFromReference(manifest.Descriptor),
			References: ctx.StringSlice("reference"),
		},
		IssuedAt: nowUnix,
	}
	if expiry := ctx.Duration(common.ExpiryFlag.Name); expiry != 0 {
		claims.NotBefore = nowUnix
		claims.Expiration = now.Add(expiry).Unix()
	}

	return claims, nil
}

func getSchemeForSigning(ctx *cli.Context) (*signature.Scheme, error) {
	var (
		signer signature.Signer
		err    error
	)
	switch method := ctx.String("method"); method {
	case "x509":
		signer, err = x509.NewSignerFromFiles(ctx.String("key"), ctx.String("cert"))
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", method)
	}
	scheme := signature.NewScheme()
	if err != nil {
		return nil, err
	}
	scheme.RegisterSigner(signerID, signer)
	return scheme, nil
}
