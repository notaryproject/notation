package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation/pkg/config"
	"github.com/notaryproject/notation/pkg/registry"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli/v2"
)

var (
	cacheCommand = &cli.Command{
		Name:  "cache",
		Usage: "Manage signature cache",
		Subcommands: []*cli.Command{
			cacheListCommand,
			cachePruneCommand,
			cacheRemoveCommand,
		},
	}

	cacheListCommand = &cli.Command{
		Name:    "list",
		Usage:   "List signatures in cache",
		Aliases: []string{"ls"},
		Flags: []cli.Flag{
			localFlag,
			usernameFlag,
			passwordFlag,
			plainHTTPFlag,
		},
		ArgsUsage: "[reference|manifest_digest]",
		Action:    listCachedSignatures,
	}

	cachePruneCommand = &cli.Command{
		Name:      "prune",
		Usage:     "Prune signature from cache",
		ArgsUsage: "[reference|manifest_digest] ...",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "prune all cached signatures",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "prune entire directory, combined with --all",
			},
			localFlag,
			usernameFlag,
			passwordFlag,
			plainHTTPFlag,
		},
		Action: pruneCachedSignatures,
	}

	cacheRemoveCommand = &cli.Command{
		Name:      "remove",
		Usage:     "Remove signature from cache",
		Aliases:   []string{"rm"},
		ArgsUsage: "<reference|manifest_digest> <signature_digest> ...",
		Flags: []cli.Flag{
			localFlag,
			usernameFlag,
			passwordFlag,
			plainHTTPFlag,
		},
		Action: removeCachedSignatures,
	}
)

func listCachedSignatures(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return listManifestsWithCachedSignature()
	}

	manifestDigest, err := getManifestDigestFromContext(ctx, ctx.Args().First())
	if err != nil {
		return err
	}

	fmt.Println("SIGNATURE")
	return walkCachedSignatureTree(
		config.SignatureRootPath(manifestDigest),
		func(algorithm string, value fs.DirEntry) error {
			if strings.HasSuffix(value.Name(), config.SignatureExtension) {
				encoded := strings.TrimSuffix(value.Name(), config.SignatureExtension)
				fmt.Printf("%s:%s\n", algorithm, encoded)
			}
			return nil
		})
}

func listManifestsWithCachedSignature() error {
	fmt.Println("MANIFEST")
	return walkCachedSignatureTree(
		config.SignatureStoreDirPath,
		func(algorithm string, value fs.DirEntry) error {
			if value.IsDir() {
				fmt.Printf("%s:%s\n", algorithm, value.Name())
			}
			return nil
		})
}

func pruneCachedSignatures(ctx *cli.Context) error {
	if ctx.Bool("all") {
		if err := walkCachedSignatureTree(
			config.SignatureStoreDirPath,
			func(algorithm string, value fs.DirEntry) error {
				if !value.IsDir() {
					return nil
				}
				manifestDigest := digest.NewDigestFromEncoded(digest.Algorithm(algorithm), value.Name())
				if err := os.RemoveAll(config.SignatureRootPath(manifestDigest)); err != nil {
					return err
				}

				// write out
				fmt.Println(manifestDigest)
				return nil
			},
		); err != nil {
			return err
		}
		if ctx.Bool("force") {
			return os.RemoveAll(config.SignatureStoreDirPath)
		}
		return nil
	}

	if !ctx.Args().Present() {
		return errors.New("nothing to prune")
	}
	for _, ref := range ctx.Args().Slice() {
		manifestDigest, err := getManifestDigestFromContext(ctx, ref)
		if err != nil {
			return err
		}
		if err := os.RemoveAll(config.SignatureRootPath(manifestDigest)); err != nil {
			return err
		}

		// write out
		fmt.Println(manifestDigest)

	}
	return nil
}

func removeCachedSignatures(ctx *cli.Context) error {
	// initialize
	sigDigests := ctx.Args().Slice()
	if len(sigDigests) == 0 {
		return errors.New("missing target manifest")
	}
	sigDigests = sigDigests[1:]
	if len(sigDigests) == 0 {
		return errors.New("no signature specified")
	}

	manifestDigest, err := getManifestDigestFromContext(ctx, ctx.Args().First())
	if err != nil {
		return err
	}

	// core process
	for _, sigDigest := range sigDigests {
		path := config.SignaturePath(manifestDigest, digest.Digest(sigDigest))
		if err := os.Remove(path); err != nil {
			return err
		}

		// write out
		fmt.Println(sigDigest)
	}

	return nil
}

func walkCachedSignatureTree(root string, fn func(algorithm string, encodedEntry fs.DirEntry) error) error {
	algorithms, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, alg := range algorithms {
		if !alg.IsDir() {
			continue
		}
		encodedEntries, err := os.ReadDir(filepath.Join(root, alg.Name()))
		if err != nil {
			return err
		}
		for _, encodedEntry := range encodedEntries {
			if err := fn(alg.Name(), encodedEntry); err != nil {
				return err
			}
		}
	}
	return nil
}

func getManifestDigestFromContext(ctx *cli.Context, ref string) (manifestDigest digest.Digest, err error) {
	manifestDigest, err = digest.Parse(ref)
	if err == nil {
		return
	}

	reference := registry.ParseReference(ref)
	manifestDigest, err = reference.Digest()
	if err == nil {
		return
	}

	manifest, err := getManifestFromContextWithReference(ctx, ref)
	if err != nil {
		return
	}
	manifestDigest = digest.Digest(manifest.Digest)
	return
}
