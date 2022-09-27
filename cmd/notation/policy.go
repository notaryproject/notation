package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/notaryproject/notation-go/dir"
	"github.com/notaryproject/notation-go/verification"
	"github.com/notaryproject/notation/internal/cmd"
	"github.com/notaryproject/notation/internal/ioutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type policyOpts struct {
	configPath string
	name       string
	scopes     []string
	level      string
	override   string
	stores     []string
	identities []string
	certPath   string
}

type deleteOpts struct {
	names     []string
	confirmed bool
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
		policyDeleteCommand(nil),
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
			return runShowCommand(policyName)
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
			return runResolveCommand(scope)
		},
	}
}

func policyDeleteCommand(opts *deleteOpts) *cobra.Command {
	if opts == nil {
		opts = &deleteOpts{}
	}
	command := &cobra.Command{
		Use:   "delete <policy_name> ...",
		Short: "Delete specified policies",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("no policy specified")
			}
			opts.names = append(opts.names, args...)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeleteCommand(opts)
		},
	}
	command.Flags().BoolVarP(&opts.confirmed, "confirm", "y", false, "if yes, do not prompt for confirmation")

	return command
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
notation policy update [--scope <registry>/<repository] [--level <level>] [--level-override <key>=<value>] [--trust-store <type>:<name>] [--identity <identity>] <policy_name>

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
		PreRun: func(cmd *cobra.Command, args []string) {
			trimSpace(opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return updatePolicy(opts)
		},
	}

	opts.applyFlags(command.Flags())
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
notation policy add --name <name> --scope <registry>/<repository> --level <level> --level-override <key>=<value> --trust-store <type>:<name> --identity <identity> [--identity-cert <cert_name>]

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
			if len(args) == 1 {
				opts.configPath = args[0]
			} else if len(args) == 0 {
				return errors.New("no arguments provided")
			} else {
				fmt.Println("multiple files provided, only the first one will be honored")
			}
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			trimSpace(opts)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return addPolicy(opts)
		},
	}

	opts.applyFlags(command.Flags())
	command.Flags().StringVar(&opts.name, "name", "", "policy name")
	// TODO: support add cert to the specified directory.
	command.Flags().StringVar(&opts.certPath, "identity-cert", "", "path to the certificate file")
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

func runShowCommand(policyName string) error {
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

func runResolveCommand(scope string) error {
	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}

	policies := policyDocument.ListPoliciesWithinScope(scope)

	return ioutil.PrintPolicyMap(os.Stdout, policies)
}

func runDeleteCommand(opts *deleteOpts) error {
	prompt := fmt.Sprintf("Are you sure you want to delete policies: %s ?", strings.Join(opts.names, ", "))
	confirmed, err := ioutil.AskForConfirmation(os.Stdin, prompt, opts.confirmed)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}

	if err = policyDocument.DeletePolicies(opts.names, dir.UserLevel); err != nil {
		return err
	}

	ioutil.PrintDeletedPolicyNames(os.Stdout, opts.names)
	return nil
}

func updatePolicy(opts *policyOpts) error {
	opts.prepareForUpdate()
	newPolicies, err := loadPoliciesFromOptions(opts)
	if err != nil {
		return err
	}

	policyDocument, err := verification.LoadDefaultPolicyDocument()
	if err != nil {
		return err
	}

	if err = policyDocument.UpdatePolicies(newPolicies, dir.UserLevel); err != nil {
		return err
	}

	ioutil.PrintPolicyNames(os.Stdout, newPolicies, "Updated Policies:")
	return nil
}

func addPolicy(opts *policyOpts) error {
	policies, err := loadPoliciesFromOptions(opts)
	if err != nil {
		return err
	}

	policyDocument, err := verification.LoadDefaultPolicyDocument()
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
	if len(opts.configPath) == 0 {
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
	if len(opts.name) == 0 {
		return nil, errors.New("no name specified")
	}
	override, err := cmd.ParseFlagPluginConfig(opts.override)
	if err != nil {
		return nil, err
	}

	// TODO: check the default values of each fields once spec is available.

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

func (opts *policyOpts) applyFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&opts.scopes, "scope", nil, "registry scopes")
	fs.StringVar(&opts.level, "level", "", "verification level")
	fs.StringVar(&opts.override, "level-coverride", "", "list of comma-separated {key}={value} pairs that override the behavior of the verification level")
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

func trimSpace(opts *policyOpts) {
	opts.configPath = strings.TrimSpace(opts.configPath)
	opts.name = strings.TrimSpace(opts.name)
}
