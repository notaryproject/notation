package notation

import (
	"os"

	"github.com/notaryproject/notation/test/e2e/internal/utils"
)

// Host creates a virtualized notation testing host by modify
// the "XDG_CONFIG_HOME" environment variable of the Executor.
//
// options is the required testing environment options
// fn is the callback function containing the testing logic.
func Host(options []utils.Option, fn func(notation *utils.ExecOpts, vhost *utils.VirtualHost)) {
	opts := []utils.Option{CreateNotationDirOption()}
	opts = append(opts, options...)
	vhost, err := utils.NewVirtualHost(NotationBinPath, opts...)
	if err != nil {
		panic(err)
	}

	defer vhost.CleanDirFunc()

	fn(vhost.Executor, vhost)
}

func Setting(options ...utils.Option) []utils.Option {
	return options
}

// BaseOptions returns the a list of base Options for a valid notation
// testing environment.
func BaseOptions() []utils.Option {
	return Setting(
		AuthOption("", ""),
		AddTestKeyOption(),
		AddTestCertOption(),
	)
}

// CreateNotationDir creates the notation directory in temp user dir.
func CreateNotationDirOption() utils.Option {
	return func(vhost *utils.VirtualHost) error {
		notationDir := NotationDir(vhost.UserDir)
		return os.MkdirAll(notationDir, os.ModePerm)
	}
}

// AuthOption sets the auth environment variables for notation.
func AuthOption(username, password string) utils.Option {
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
func AddTestKeyOption() utils.Option {
	return func(vhost *utils.VirtualHost) error {
		return AddTestKeyPairs(NotationDir(vhost.UserDir))
	}
}

func AddTestCertOption() utils.Option {
	return func(vhost *utils.VirtualHost) error {
		vhost.Executor.
			MatchKeyWords("Successfully added following certificates").
			Exec("cert", "add", "--type", "ca", "--store", "e2e", NotationE2ECertPath)
		return nil
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
