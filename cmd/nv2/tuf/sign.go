package tuf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/notaryproject/nv2/cmd/nv2/common"
	"github.com/notaryproject/nv2/pkg/tuf"
	"github.com/notaryproject/nv2/pkg/tuf/local"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/urfave/cli/v2"
)

// SignCommand defines sign command
var SignCommand = &cli.Command{
	Name:      "sign",
	Usage:     "signs OCI Artifacts",
	ArgsUsage: "[<scheme://reference>]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "key",
			Aliases:   []string{"k"},
			Usage:     "signing key file",
			TakesFile: true,
			Required:  true,
		},
		&cli.StringFlag{
			Name:      "cert",
			Aliases:   []string{"c"},
			Usage:     "signing cert",
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:      "signature",
			Aliases:   []string{"s", "f"},
			Usage:     "base signature file",
			TakesFile: true,
		},
		&cli.StringFlag{
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
	signer, err := local.NewSignerFromFiles(ctx.String("key"), ctx.String("cert"))
	if err != nil {
		return err
	}

	// core process
	targets, manifestDigest, err := prepareTargetsForSigning(ctx)
	if err != nil {
		return err
	}
	signed, err := tuf.SignTargets(ctx.Context, signer, targets)
	if err != nil {
		return err
	}
	// non-canonical JSON marshal to match Docker Notary 0.6.0 implementation
	sig, err := json.Marshal(signed)
	if err != nil {
		return err
	}

	// write out
	path := ctx.String(common.OutputFlag.Name)
	if path == "" {
		path = strings.Split(manifestDigest, ":")[1] + ".nv2"
	}
	if err := ioutil.WriteFile(path, []byte(sig), 0666); err != nil {
		return err
	}

	fmt.Println(manifestDigest)
	return nil
}

func prepareTargetsForSigning(ctx *cli.Context) (*data.Targets, string, error) {
	manifest, err := common.GetManifestFromContext(ctx)
	if err != nil {
		return nil, "", err
	}
	if reference := ctx.String("reference"); reference != "" {
		manifest.Name = reference
	}
	if manifest.Name == "" {
		return nil, "", errors.New("manifest is not referenced")
	}
	target, err := tuf.NewTarget(manifest)
	if err != nil {
		return nil, "", err
	}

	var targets *data.Targets
	if path := ctx.String("signature"); path != "" {
		targets, err = readTargetsFromFile(path)
		if err != nil {
			return nil, "", err
		}
	}
	targets = tuf.AddTargets(targets, target)

	if expiry := ctx.Duration(common.ExpiryFlag.Name); expiry != 0 {
		targets.Expires = time.Now().UTC().Add(expiry)
	}

	return targets, manifest.Digests[0].String(), nil
}

func readTargetsFromFile(path string) (*data.Targets, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	signed := new(data.Signed)
	if err := json.Unmarshal(raw, signed); err != nil {
		return nil, err
	}

	signedTargets, err := data.TargetsFromSigned(signed, data.CanonicalTargetsRole)
	if err != nil {
		return nil, err
	}
	return &signedTargets.Signed, nil
}
