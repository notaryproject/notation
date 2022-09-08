package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verification"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type policyOpts struct {
	configPath string
	name       string
	scopes     []string
	level      string
	override   []string
	stores     []string
	identities []string
	certPath   string
}

func policyCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "policy",
		Short: "Manage trust policy for verification",
	}
	command.AddCommand(
		policyListCommand(),
		policyShowCommand(),
		policyResolveCommand(),
		policyDeleteCommand(),
		policyUpdateCommand(nil),
		policyAddCommand(nil),
	)
	return command
}

func policyListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all configured trust policies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listPolicies()
		},
	}
}

func policyShowCommand() *cobra.Command {
	var policyName string
	return &cobra.Command{
		Use:   "show <policy_name>",
		Short: "Inspect a policy by name",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no policy name specified")
			}
			policyName = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowCommand(cmd, policyName)
		},
	}
}

func policyResolveCommand() *cobra.Command {
	var scope string
	return &cobra.Command{
		Use:   "resolve <scope>",
		Short: "Present policies under the given scope",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no scope specified")
			}
			scope = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResolveCommand(cmd, scope)
		},
	}
}

func policyDeleteCommand() *cobra.Command {
	var policyNames []string
	return &cobra.Command{
		Use:   "delete <policy_name> ...",
		Short: "Delete specified policies",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no policy specified")
			}
			policyNames = append(policyNames, args...)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteCommand(cmd, policyNames)
		},
	}
}

func policyUpdateCommand(opts *policyOpts) *cobra.Command {
	if opts == nil {
		opts = &policyOpts{}
	}
	command := &cobra.Command{
		Use:   "update [<options>] [<json_file_path>|<policy_name>]",
		Short: "Update existing policies",
		Long: `Update existing policies
notation policy update path/to/config/file
notation policy update [--scope <registry>/<repository] [--level <level>] [--level-override <key>:<value>] [--trust-store <type>:<name>] [--identity <identity>] <policy_name>

Example - Update policies from a config file
	notation policy update "/home/user/trustpolicy.json"

Example - Update policy from options
	notation policy update wabbit-networks-images \
		--scope registry.acme-rockets.io/software/net-monitor \
		--scope registry.acme-rockets.io/software/net-logger \
		--level strict \
		--trust-store ca:wabbit-networks \
		--identity "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Security Tools"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no policy name or file path provided")
			}
			opts.configPath = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return updatePolicy(cmd, opts)
		},
	}

	opts.ApplyFlags(command.Flags())
	return command
}

func policyAddCommand(opts *policyOpts) *cobra.Command {
	if opts == nil {
		opts = &policyOpts{}
	}
	command := &cobra.Command{
		Use:   "add [<json_file_path>]",
		Short: "Add new policies",
		Long: `Add new policies
notation policy add path/to/config/file
notation policy add --name <name> --scope <registry>/<repository> --level <level> --level-override <key>:<value> --trust-store <type>:<name> --identity <identity> [--identity-cert <cert_name>]

Example - Add policies from a config file
	notation policy add "/home/user/trustpolicy.json"

Example - Add policy from options
	notation policy add --name wabbit-networks-images \
		--scope registry.acme-rockets.io/software/net-monitor \
		--scope registry.acme-rockets.io/software/net-logger \
		--level strict \
		--trust-store ca:wabbit-networks \
		--identity "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Security Tools"
		--identity-cert acme_root.crt`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.configPath = args[0]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addPolicy(cmd, opts)
		},
	}

	opts.ApplyFlags(command.Flags())
	command.Flags().StringVar(&opts.name, "name", "", "policy name")
	// TODO: support add cert to the specified directory.
	command.Flags().StringVar(&opts.certPath, "identity-cert", "", "path to the root certificate file")
	return command
}

func listPolicies() error {
	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}
	policies := policyDocument.ListPolicies()

	return ioutil.PrintPolicyMap(os.Stdout, policies)
}

func runShowCommand(cmd *cobra.Command, policyName string) error {
	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}

	policy, err := policyDocument.GetPolicy(policyName)
	if err != nil {
		return err
	}

	return ioutil.PrintPolicyMap(os.Stdout, []*verification.TrustPolicy{policy})
}

func runResolveCommand(cmd *cobra.Command, scope string) error {
	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}

	policies := policyDocument.ListPoliciesWithinScope(scope)

	return ioutil.PrintPolicyMap(os.Stdout, policies)
}

func runDeleteCommand(cmd *cobra.Command, policyNames []string) error {
	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}

	if err = policyDocument.DeletePolicies(policyNames, dir.UserLevel); err != nil {
		return err
	}

	ioutil.PrintDeletedPolicyNames(os.Stdout, policyNames)
	return nil
}

func updatePolicy(cmd *cobra.Command, opts *policyOpts) error {
	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}
	opts.prepareForUpdate()

	newPolicies, err := loadPoliciesFromOptions(opts)
	if err != nil {
		return err
	}

	if err = policyDocument.UpdatePolicies(newPolicies, dir.UserLevel); err != nil {
		return err
	}

	ioutil.PrintPolicyNames(os.Stdout, newPolicies, "Updated Policies:")
	return nil
}

func addPolicy(cmd *cobra.Command, opts *policyOpts) error {
	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}

	policies, err := loadPoliciesFromOptions(opts)
	if err != nil {
		return err
	}

	if err = policyDocument.AddPolicies(policies, dir.UserLevel); err != nil {
		return err
	}

	ioutil.PrintPolicyNames(os.Stdout, policies, "Added Policies:")
	return nil
}

func loadPoliciesFromOptions(opts *policyOpts) ([]*verification.TrustPolicy, error) {
	if opts.configPath == "" {
		// create a policy from options.
		policy, err := newPolicyFromOptions(opts)
		if err != nil {
			return nil, err
		}
		return []*verification.TrustPolicy{policy}, nil
	}
	// load policies from provided JSON file.
	return verification.LoadTrustPolicies(opts.configPath)
}

func newPolicyFromOptions(opts *policyOpts) (*verification.TrustPolicy, error) {
	override := make(map[string]string)
	// initialize override map from the input if provided.
	for _, config := range opts.override {
		idx := strings.Index(config, ":")
		if idx == -1 {
			return nil, fmt.Errorf("invalid override config: %s", config)
		}
		override[config[0:idx]] = config[idx+1:]
	}

	return &verification.TrustPolicy{
		Name:           opts.name,
		RegistryScopes: opts.scopes,
		SignatureVerification: verification.SignatureVerification{
			Level:    opts.level,
			Override: override,
		},
		TrustStores:       opts.stores,
		TrustedIdentities: opts.identities,
	}, nil
}

func (opts *policyOpts) ApplyFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&opts.scopes, "scope", nil, "registry scopes")
	fs.StringVar(&opts.level, "level", "", "verification level")
	fs.StringSliceVar(&opts.override, "level-override", nil, "config of custom verification level")
	fs.StringSliceVar(&opts.stores, "trust-store", nil, "trust stores containing trusted roots")
	fs.StringArrayVar(&opts.identities, "identity", nil, "trusted identities")
}

// prepareForUpdate checks if the update works on a file or via typed options.
func (opts *policyOpts) prepareForUpdate() {
	if len(opts.scopes) != 0 || opts.level != "" || len(opts.override) != 0 || len(opts.stores) != 0 || len(opts.identities) != 0 {
		opts.name = opts.configPath
		opts.configPath = ""
	}
}
