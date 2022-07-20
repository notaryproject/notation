package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/notaryproject/notation/pkg/config"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"oras.land/oras-go/v2/registry"
)

type cacheListOpts struct {
	RemoteFlagOpts
	reference string
}
type cachePruneOpts struct {
	RemoteFlagOpts
	references []string
	all        bool
	purge      bool
	force      bool
}

type cacheRemoveOpts struct {
	RemoteFlagOpts
	reference  string
	sigDigests []string
}

func cacheCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "cache",
		Short: "Manage signature cache",
	}
	command.AddCommand(cacheListCommand(nil), cachePruneCommand(nil), cacheRemoveCommand(nil))
	return command
}

func cacheListCommand(opts *cacheListOpts) *cobra.Command {
	if opts == nil {
		opts = &cacheListOpts{}
	}
	command := &cobra.Command{
		Use:     "list [reference|manifest_digest]",
		Aliases: []string{"ls"},
		Short:   "List signatures in cache",
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				opts.reference = args[0]
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCachedSignatures(cmd, opts)
		},
	}
	opts.ApplyFlag(command.Flags())
	return command
}

func cachePruneCommand(opts *cachePruneOpts) *cobra.Command {
	if opts == nil {
		opts = &cachePruneOpts{}
	}
	command := &cobra.Command{
		Use:   "prune [reference|manifest_digest]...",
		Short: "Prune signature from cache",
		Args:  cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.references = args
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return pruneCachedSignatures(cmd, opts)
		},
	}
	command.Flags().BoolVarP(&opts.all, "all", "a", false, "prune all cached signatures")
	command.Flags().BoolVar(&opts.purge, "purge", false, "remove the signature directory, combined with --all")
	command.Flags().BoolVarP(&opts.force, "force", "f", false, "do not prompt for confirmation")
	opts.ApplyFlag(command.Flags())
	return command
}

func cacheRemoveCommand(opts *cacheRemoveOpts) *cobra.Command {
	if opts == nil {
		opts = &cacheRemoveOpts{}
	}
	command := &cobra.Command{
		Use:     "remove [reference|manifest_digest] [signature_digest]...",
		Aliases: []string{"rm"},
		Short:   "Remove signature from cache",
		Args:    cobra.MinimumNArgs(2),
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.reference = args[0]
			opts.sigDigests = args[1:]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return removeCachedSignatures(cmd, opts)
		},
	}
	opts.ApplyFlag(command.Flags())
	return command
}

func listCachedSignatures(command *cobra.Command, opts *cacheListOpts) error {
	if command.Flags().NArg() == 0 {
		return listManifestsWithCachedSignature()
	}

	manifestDigest, err := getManifestDigestFromContext(command.Context(), &opts.RemoteFlagOpts, opts.reference)
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

func pruneCachedSignatures(command *cobra.Command, opts *cachePruneOpts) error {
	if opts.all {
		if !opts.force {
			fmt.Println("WARNING! This will remove:")
			fmt.Println("- all cached signatures")
			if opts.purge {
				fmt.Println("- all files in the cache signature directory")
			}
			fmt.Println()
			if confirmed := promptConfirmation(); !confirmed {
				return nil
			}
		}
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
		if opts.purge {
			return os.RemoveAll(config.SignatureStoreDirPath)
		}
		return nil
	}

	if len(opts.references) == 0 {
		return errors.New("nothing to prune")
	}
	refs := opts.references
	if !opts.force {
		fmt.Println("WARNING! This will remove cached signatures for manifests below:")
		for _, ref := range refs {
			fmt.Println("-", ref)
		}
		fmt.Println()
		if confirmed := promptConfirmation(); !confirmed {
			return nil
		}
	}
	for _, ref := range refs {
		manifestDigest, err := getManifestDigestFromContext(command.Context(), &opts.RemoteFlagOpts, ref)
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

func removeCachedSignatures(command *cobra.Command, opts *cacheRemoveOpts) error {
	// initialize
	manifestDigest, err := getManifestDigestFromContext(command.Context(), &opts.RemoteFlagOpts, opts.reference)
	if err != nil {
		return err
	}

	// core process
	for _, sigDigest := range opts.sigDigests {
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

func getManifestDigestFromContext(ctx context.Context, opts *RemoteFlagOpts, ref string) (manifestDigest digest.Digest, err error) {
	manifestDigest, err = digest.Parse(ref)
	if err == nil {
		return
	}

	reference, err := registry.ParseReference(ref)
	if err != nil {
		return
	}
	manifestDigest, err = reference.Digest()
	if err == nil {
		return
	}

	manifest, err := getManifestDescriptorFromContextWithReference(ctx, opts, ref)
	if err != nil {
		return
	}
	manifestDigest = digest.Digest(manifest.Digest)
	return
}

func promptConfirmation() bool {
	fmt.Printf("Are you sure you want to continue? [y/N]: ")
	scanner := bufio.NewScanner(os.Stdin)
	return scanner.Scan() && strings.EqualFold(scanner.Text(), "y")
}
