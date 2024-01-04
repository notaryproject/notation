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
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NotationDirName   = "notation"
	TrustPolicyName   = "trustpolicy.json"
	TrustStoreDirName = "truststore"
	TrustStoreTypeCA  = "ca"
	PluginDirName     = "plugins"
	PluginName        = "e2e-plugin"
	ConfigJsonName    = "config.json"
)

const (
	envKeyRegistryHost                      = "NOTATION_E2E_REGISTRY_HOST"
	envKeyRegistryUsername                  = "NOTATION_E2E_REGISTRY_USERNAME"
	envKeyRegistryPassword                  = "NOTATION_E2E_REGISTRY_PASSWORD"
	envKeyDomainRegistryHost                = "NOTATION_E2E_DOMAIN_REGISTRY_HOST"
	envKeyNotationBinPath                   = "NOTATION_E2E_BINARY_PATH"
	envKeyNotationOldBinPath                = "NOTATION_E2E_OLD_BINARY_PATH"
	envKeyNotationPluginPath                = "NOTATION_E2E_PLUGIN_PATH"
	envKeyNotationPluginTarGzPath           = "NOTATION_E2E_PLUGIN_TAR_GZ_PATH"
	envKeyNotationMaliciouPluginArchivePath = "NOTATION_E2E_MALICIOUS_PLUGIN_ARCHIVE_PATH"
	envKeyNotationConfigPath                = "NOTATION_E2E_CONFIG_PATH"
	envKeyOCILayoutPath                     = "NOTATION_E2E_OCI_LAYOUT_PATH"
	envKeyTestRepo                          = "NOTATION_E2E_TEST_REPO"
	envKeyTestTag                           = "NOTATION_E2E_TEST_TAG"
)

var (
	// NotationBinPath is the notation binary path.
	NotationBinPath string
	// NotationOldBinPath is the path of an old version notation binary for
	// testing forward compatibility.
	NotationOldBinPath                    string
	NotationE2EPluginPath                 string
	NotationE2EPluginTarGzPath            string
	NotationE2EMaliciousPluginArchivePath string
	NotationE2EConfigPath                 string
	NotationE2ELocalKeysDir               string
	NotationE2ETrustPolicyDir             string
	NotationE2EConfigJsonDir              string
)

var (
	OCILayoutPath       string
	TestRepoUri         string
	TestTag             string
	RegistryStoragePath string
)

func init() {
	RegisterFailHandler(Fail)
	setUpRegistry()
	setUpNotationValues()
}

func setUpRegistry() {
	setValue(envKeyRegistryHost, &TestRegistry.Host)
	setValue(envKeyRegistryUsername, &TestRegistry.Username)
	setValue(envKeyRegistryPassword, &TestRegistry.Password)
	setValue(envKeyDomainRegistryHost, &TestRegistry.DomainHost)

	setPathValue(envKeyOCILayoutPath, &OCILayoutPath)
	setValue(envKeyTestRepo, &TestRepoUri)
	setValue(envKeyTestTag, &TestTag)
}

func setUpNotationValues() {
	// set Notation binary path
	setPathValue(envKeyNotationBinPath, &NotationBinPath)
	setPathValue(envKeyNotationOldBinPath, &NotationOldBinPath)

	// set Notation e2e-plugin path
	setPathValue(envKeyNotationPluginPath, &NotationE2EPluginPath)
	setPathValue(envKeyNotationPluginTarGzPath, &NotationE2EPluginTarGzPath)
	setPathValue(envKeyNotationMaliciouPluginArchivePath, &NotationE2EMaliciousPluginArchivePath)

	// set Notation configuration paths
	setPathValue(envKeyNotationConfigPath, &NotationE2EConfigPath)
	NotationE2ETrustPolicyDir = filepath.Join(NotationE2EConfigPath, "trustpolicies")
	NotationE2ELocalKeysDir = filepath.Join(NotationE2EConfigPath, LocalKeysDirName)
	NotationE2EConfigJsonDir = filepath.Join(NotationE2EConfigPath, LocalConfigJsonsDirName)
}

func setPathValue(envKey string, value *string) {
	setValue(envKey, value)
	if !filepath.IsAbs(*value) {
		panic(fmt.Sprintf("env %s=%q is not a absolute path", envKey, *value))
	}
}

func setValue(envKey string, value *string) {
	if *value = os.Getenv(envKey); *value == "" {
		panic(fmt.Sprintf("env %s is empty", envKey))
	}
	fmt.Printf("set test value $%s=%s\n", envKey, *value)
}
