package notation

import (
	"os"
	"path/filepath"

	"github.com/notaryproject/notation/test/e2e/internal/utils"
)

// Host creates a virtualized notation testing host by modify
// the "XDG_CONFIG_HOME" environment variable of the Executor.
//
// options is the required testing environment options
// fn is the callback function containing the testing logic.
func Host(options []utils.HostOption, fn func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost)) {
	opts := []utils.HostOption{CreateNotationDirOption()}
	opts = append(opts, options...)

	// create a vhost
	vhost, err := utils.NewVirtualHost(NotationBinPath, opts...)
	if err != nil {
		panic(err)
	}

	// generate a repository with an artifact
	artifact := GenerateArtifact()

	// run the main logic
	fn(vhost.Executor, artifact, vhost)

	// remove the generated repository and artifact
	if err := artifact.Remove(); err != nil {
		panic(err)
	}
}

// Opts is a grammar sugar to generate a list of HostOption
func Opts(options ...utils.HostOption) []utils.HostOption {
	return options
}

// BaseOptions returns a list of base Options for a valid notation
// testing environment.
func BaseOptions() []utils.HostOption {
	return Opts(
		AuthOption("", ""),
		AddTestKeyOption(),
		AddTestTrustStoreOption(),
		AddTestTrustPolicyOption(),
	)
}

// CreateNotationDirOption creates the notation directory in temp user dir.
func CreateNotationDirOption() utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		return os.MkdirAll(vhost.UserPath(notationDirName), os.ModePerm)
	}
}

// AuthOption sets the auth environment variables for notation.
func AuthOption(username, password string) utils.HostOption {
	if username == "" {
		username = TestRegistry.Username
	}
	if password == "" {
		password = TestRegistry.Password
	}
	return func(vhost *utils.VirtualHost) error {
		vhost.UpdateEnv(authEnv(username, password))
		return nil
	}
}

// AddTestKeyOption adds the test signingkeys.json, key and cert files to
// the notation directory.
func AddTestKeyOption() utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		return AddTestKeyPairs(vhost.UserPath(notationDirName))
	}
}

// AddTestTrustStoreOption added the test cert to the trust store.
func AddTestTrustStoreOption() utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		vhost.Executor.
			MatchKeyWords("Successfully added following certificates").
			Exec("cert", "add", "--type", "ca", "--store", "e2e", NotationE2ECertPath)
		return nil
	}
}

// AddTestTrustPolicyOption added a valid trust policy for testing
func AddTestTrustPolicyOption() utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		return copyFile(
			filepath.Join(NotationE2ETrustPolicyDir, "trustpolicy.json"),
			vhost.UserPath(notationDirName, notationTrustPolicyName),
		)
	}
}

// authEnv creates an auth info
// (By setting $NOTATION_USERNAME and $NOTATION_PASSWORD)
func authEnv(username, password string) map[string]string {
	env := make(map[string]string)
	env["NOTATION_USERNAME"] = username
	env["NOTATION_PASSWORD"] = password
	return env
}
