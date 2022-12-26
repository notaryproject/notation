package notation

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	NotationDirName         = "notation"
	NotationTrustPolicyName = "trustpolicy.json"
)

const (
	envRegistryHost        = "NOTATION_E2E_REGISTRY_HOST"
	envRegistryUsername    = "NOTATION_E2E_REGISTRY_USERNAME"
	envRegistryPassword    = "NOTATION_E2E_REGISTRY_PASSWORD"
	envNotationBinPath     = "NOTATION_E2E_BINARY_PATH"
	envNotationE2EKeyPath  = "NOTATION_E2E_KEY_PATH"
	envNotationE2ECertPath = "NOTATION_E2E_CERT_PATH"
	envNotationConfigPath  = "NOTATION_E2E_CONFIG_PATH"
	envOCILayoutPath       = "NOTATION_E2E_OCI_LAYOUT_PATH"
	envTestRepo            = "NOTATION_E2E_TEST_REPO"
	envTestTag             = "NOTATION_E2E_TEST_TAG"
	envRegistryStoragePath = "REGISTRY_STORAGE_PATH"
)

var (
	NotationBinPath           string
	NotationE2EKeyPath        string
	NotationE2ECertPath       string
	NotationE2EConfigPath     string
	NotationE2ETrustPolicyDir string
)

var (
	OCILayoutPath       string
	TestRepo            string
	TestTag             string
	RegistryStoragePath string
)

func init() {
	RegisterFailHandler(Fail)
	setUpRegistry()
	setUpNotationValues()
}

func setUpRegistry() {
	setValue(envRegistryHost, &TestRegistry.Host)
	fmt.Printf("Testing using registry host: %s\n", TestRegistry.Host)

	setValue(envRegistryUsername, &TestRegistry.Username)
	fmt.Printf("Testing using registry username: %s\n", TestRegistry.Username)

	setValue(envRegistryPassword, &TestRegistry.Password)
	fmt.Printf("Testing using registry password: %s\n", TestRegistry.Password)

	testImage := &Artifact{
		Registry: &TestRegistry,
		Repo:     testRepo,
		Tag:      testTag,
	}

	if err := testImage.Validate(); err != nil {
		panic(fmt.Sprintf("E2E setup failed: %v", err))
	}
}

func setUpNotationValues() {
	// set Notation binary path
	setPathValue(envNotationBinPath, &NotationBinPath)

	// set Notation configuration paths
	setPathValue(envNotationE2EKeyPath, &NotationE2EKeyPath)
	setPathValue(envNotationE2ECertPath, &NotationE2ECertPath)
	setPathValue(envNotationConfigPath, &NotationE2EConfigPath)
	NotationE2ETrustPolicyDir = filepath.Join(NotationE2EConfigPath, "trustpolicys")

	// set registry values
	setPathValue(envRegistryStoragePath, &RegistryStoragePath)
	setPathValue(envOCILayoutPath, &OCILayoutPath)
	setValue(envTestRepo, &TestRepo)
	setValue(envTestTag, &TestTag)
}

func setPathValue(envKey string, value *string) {
	setValue(envKey, value)
	if !filepath.IsAbs(*value) {
		panic(fmt.Sprintf("env %s=%q is not a absolute path", envKey, *value))
	}
}

func setValue(envKey string, value *string) {
	*value = os.Getenv(envKey)
	if *value == "" {
		panic(fmt.Sprintf("env %s is empty", envKey))
	}
}
