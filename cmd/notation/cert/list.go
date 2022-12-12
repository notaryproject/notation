package cert

import (
	"context"
	"fmt"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/log"
	notationgoTruststore "github.com/notaryproject/notation-go/verifier/truststore"
	"github.com/notaryproject/notation/cmd/notation/internal/truststore"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/spf13/cobra"
)

type certListOpts struct {
	cmd.LoggingFlagOpts
	storeType  string
	namedStore string
}

func certListCommand(opts *certListOpts) *cobra.Command {
	if opts == nil {
		opts = &certListOpts{}
	}
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List certificates in the trust store.",
		Long: `List certificates in trust store

Example - List all certificate files stored in the trust store
  notation cert ls

Example - List all certificate files of trust store "acme-rockets"
  notation cert ls --store "acme-rockets"

Example - List all certificate files from trust store of type "ca"
  notation cert ls --type ca

Example - List all certificate files from trust store "wabbit-networks" of type "signingAuthority"
  notation cert ls --type signingAuthority --store "wabbit-networks"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listCerts(cmd.Context(), opts)
		},
	}
	opts.LoggingFlagOpts.ApplyFlags(command.Flags())
	command.Flags().StringVarP(&opts.storeType, "type", "t", "", "specify trust store type, options: ca, signingAuthority")
	command.Flags().StringVarP(&opts.namedStore, "store", "s", "", "specify named store")
	return command
}

func listCerts(ctx context.Context, opts *certListOpts) error {
	// set log level
	ctx = opts.LoggingFlagOpts.SetLoggerLevel(ctx)
	logger := log.GetLogger(ctx)

	namedStore := opts.namedStore
	storeType := opts.storeType
	configFS := dir.ConfigFS()

	// List all certificates under truststore/x509, display empty if there's
	// no certificate yet
	if namedStore == "" && storeType == "" {
		path, err := configFS.SysPath(dir.TrustStoreDir, "x509")
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			return err
		}
		if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 2)); err != nil {
			logger.Debugln("Failed to complete list at path:", path)
			return fmt.Errorf("failed to complete listing all certificates in the trust store, with error: %s", err.Error())
		}

		return nil
	}

	// List all certificates under truststore/x509/storeType/namedStore,
	// display empty if there's no such certificate
	if namedStore != "" && storeType != "" {
		path, err := configFS.SysPath(dir.TrustStoreDir, "x509", storeType, namedStore)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			return err
		}
		if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 0)); err != nil {
			logger.Debugln("Failed to complete list at path:", path)
			return fmt.Errorf("failed to complete listing certificates in the named store %s of type %s, with error: %s", namedStore, storeType, err.Error())
		}

		return nil
	}

	// List all certificates under x509/storeType, display empty if
	// there's no certificate yet
	if storeType != "" {
		path, err := configFS.SysPath(dir.TrustStoreDir, "x509", storeType)
		if err := truststore.CheckNonErrNotExistError(err); err != nil {
			return err
		}
		if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 1)); err != nil {
			logger.Debugln("Failed to complete list at path:", path)
			return fmt.Errorf("failed to complete listing certificates of type %s, with error: %s", storeType, err.Error())
		}
	} else {
		// List all certificates under named store namedStore, display empty if
		// there's no such certificate
		for _, t := range notationgoTruststore.Types {
			path, err := configFS.SysPath(dir.TrustStoreDir, "x509", string(t), namedStore)
			if err := truststore.CheckNonErrNotExistError(err); err != nil {
				return err
			}
			if err := truststore.CheckNonErrNotExistError(truststore.ListCerts(path, 0)); err != nil {
				logger.Debugln("Failed to complete list at path:", path)
				return fmt.Errorf("failed to complete listing certificates in the named store %s, with error: %s", namedStore, err.Error())
			}
		}
	}

	return nil
}
