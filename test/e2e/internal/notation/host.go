// Copyright The Notary Project Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notation

import (
	"os"
	"path/filepath"

	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
)

// CoreTestFunc is the test function running in a VirtualHost.
//
// notation is an Executor isolated by $XDG_CONFIG_HOME.
// artifact is a generated artifact in a new repository.
// vhost is the VirtualHost instance.
type CoreTestFunc func(notation *utils.ExecOpts, artifact *Artifact, vhost *utils.VirtualHost)

// OCILayoutTestFunc is the test function running in a VirtualHost with isolated
// OCI layout for each test case.
//
// notation is an Executor isolated by $XDG_CONFIG_HOME.
// vhost is the VirtualHost instance.
type OCILayoutTestFunc func(notation *utils.ExecOpts, ocilayout *OCILayout, vhost *utils.VirtualHost)

// Host creates a virtualized notation testing host by modify
// the "XDG_CONFIG_HOME" environment variable of the Executor.
//
// options is the required testing environment options
// fn is the callback function containing the testing logic.
func Host(options []utils.HostOption, fn CoreTestFunc) {
	// create a notation vhost
	vhost, err := createNotationHost(NotationBinPath, options...)
	if err != nil {
		panic(err)
	}

	// generate a repository with an artifact
	artifact := GenerateArtifact("", "")

	// run the main logic
	fn(vhost.Executor, artifact, vhost)
}

// HostInGithubAction only run the test in GitHub Actions.
//
// The booting script will setup TLS reverse proxy and TLS certificate
// for Github Actions environment.
func HostInGithubAction(options []utils.HostOption, fn CoreTestFunc) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		Skip("only run in GitHub Actions")
	}
	Host(options, fn)
}

// HostWithOCILayout creates a virtualized notation testing host by modify
// the "XDG_CONFIG_HOME" environment variable of the Executor. It generates
// isolated OCI layout in the testing host.
//
// options is the required testing environment options
// fn is the callback function containing the testing logic.
func HostWithOCILayout(options []utils.HostOption, fn OCILayoutTestFunc) {
	// create a notation vhost
	vhost, err := createNotationHost(NotationBinPath, options...)
	if err != nil {
		panic(err)
	}

	ocilayout, err := GenerateOCILayout("")
	if err != nil {
		panic(err)
	}

	// run the main logic
	fn(vhost.Executor, ocilayout, vhost)
}

// OldNotation create an old version notation ExecOpts in a VirtualHost
// for testing forward compatibility.
func OldNotation(options ...utils.HostOption) *utils.ExecOpts {
	if len(options) == 0 {
		options = BaseOptions()
	}

	vhost, err := createNotationHost(NotationOldBinPath, options...)
	if err != nil {
		panic(err)
	}

	return vhost.Executor
}

func createNotationHost(path string, options ...utils.HostOption) (*utils.VirtualHost, error) {
	vhost, err := utils.NewVirtualHost(path, CreateNotationDirOption())
	if err != nil {
		return nil, err
	}

	// set additional options
	vhost.SetOption(options...)
	return vhost, nil
}

// Opts is a grammar sugar to generate a list of HostOption.
func Opts(options ...utils.HostOption) []utils.HostOption {
	return options
}

// BaseOptions returns a list of base Options for a valid notation.
// testing environment.
func BaseOptions() []utils.HostOption {
	return Opts(
		AuthOption("", ""),
		AddKeyOption("e2e.key", "e2e.crt"),
		AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "e2e.crt")),
		AddTrustPolicyOption("trustpolicy.json"),
	)
}

func BaseOptionsWithExperimental() []utils.HostOption {
	return Opts(
		AuthOption("", ""),
		AddKeyOption("e2e.key", "e2e.crt"),
		AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "e2e.crt")),
		AddTrustPolicyOption("trustpolicy.json"),
		EnableExperimental(),
	)
}

// TestLoginOptions returns the BaseOptions with removing AuthOption and adding ConfigOption.
// testing environment.
func TestLoginOptions() []utils.HostOption {
	return Opts(
		AddKeyOption("e2e.key", "e2e.crt"),
		AddTrustStoreOption("e2e", filepath.Join(NotationE2ELocalKeysDir, "e2e.crt")),
		AddTrustPolicyOption("trustpolicy.json"),
		AddConfigJsonOption("pass_credential_helper_config.json"),
	)
}

// CreateNotationDirOption creates the notation directory in temp user dir.
func CreateNotationDirOption() utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		return os.MkdirAll(vhost.AbsolutePath(NotationDirName), os.ModePerm)
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

// AddKeyOption adds the test signingkeys.json, key and cert files to
// the notation directory.
func AddKeyOption(keyName, certName string) utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		return AddKeyPairs(vhost.AbsolutePath(NotationDirName), keyName, certName)
	}
}

// AddTrustStoreOption adds the test cert to the trust store.
func AddTrustStoreOption(namedstore string, srcCertPath string) utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		vhost.Executor.
			Exec("cert", "add", "--type", "ca", "--store", namedstore, srcCertPath).
			MatchKeyWords("Successfully added following certificates")
		return nil
	}
}

// AddTrustPolicyOption adds a valid trust policy for testing.
func AddTrustPolicyOption(trustpolicyName string) utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		return copyFile(
			filepath.Join(NotationE2ETrustPolicyDir, trustpolicyName),
			vhost.AbsolutePath(NotationDirName, TrustPolicyName),
		)
	}
}

// AddConfigJsonOption adds a valid config.json for testing.
func AddConfigJsonOption(configJsonName string) utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		return copyFile(
			filepath.Join(NotationE2EConfigJsonDir, configJsonName),
			vhost.AbsolutePath(NotationDirName, ConfigJsonName),
		)
	}
}

// AddPlugin adds a pluginkeys.json config file and installs an e2e-plugin.
func AddPlugin(pluginPath string) utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		// add pluginkeys.json configuration file for e2e-plugin
		saveJSON(
			generatePluginKeys(vhost.AbsolutePath(NotationDirName)),
			vhost.AbsolutePath(NotationDirName, "pluginkeys.json"),
		)

		// install plugin
		e2ePluginDir := vhost.AbsolutePath(NotationDirName, PluginDirName, PluginName)
		if err := os.MkdirAll(e2ePluginDir, 0700); err != nil {
			return err
		}
		return copyFile(
			NotationE2EPluginPath,
			filepath.Join(e2ePluginDir, "notation-"+PluginName),
		)
	}
}

// authEnv creates an auth info.
// (By setting $NOTATION_USERNAME and $NOTATION_PASSWORD)
func authEnv(username, password string) map[string]string {
	return map[string]string{
		"NOTATION_USERNAME": username,
		"NOTATION_PASSWORD": password,
	}
}

// EnableExperimental enables experimental features.
func EnableExperimental() utils.HostOption {
	return func(vhost *utils.VirtualHost) error {
		vhost.UpdateEnv(map[string]string{"NOTATION_EXPERIMENTAL": "1"})
		return nil
	}
}
