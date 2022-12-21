package notation

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	envRegistryHost        = "NOTATION_E2E_REGISTRY_HOST"
	envRegistryUsername    = "NOTATION_E2E_REGISTRY_USERNAME"
	envRegistryPassword    = "NOTATION_E2E_REGISTRY_PASSWORD"
	envNotationBinPath     = "NOTATION_E2E_BINARY_PATH"
	envNotationE2EKeyPath  = "NOTATION_E2E_KEY_PATH"
	envNotationE2ECertPath = "NOTATION_E2E_CERT_PATH"
	envOCILayoutPath       = "NOTATION_E2E_OCI_LAYOUT_PATH"
	envTestRepo            = "NOTATION_E2E_TEST_REPO"
	envTestTag             = "NOTATION_E2E_TEST_TAG"
	envRegistryStoragePath = "REGISTRY_STORAGE_PATH"
	envGithubWorkSpace     = "GITHUB_WORKSPACE"
)

var (
	NotationBinPath     string
	NotationE2EKeyPath  string
	NotationE2ECertPath string
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
	setUpNotationBinary()
}

func setUpNotationBinary() {
	// set Notation binary path
	p := os.Getenv(envNotationBinPath)
	var err error
	if p != "" && filepath.IsAbs(p) {
		NotationBinPath = p
		fmt.Printf("Testing based on pre-built binary locates in %v\n", p)
	} else if workspacePath := os.Getenv(envGithubWorkSpace); workspacePath != "" && !filepath.IsAbs(p) {
		NotationBinPath = filepath.Join(workspacePath, p)
		NotationBinPath, err = filepath.Abs(NotationBinPath)
		if err != nil {
			panic(fmt.Sprintf("E2E setup failed:%v", err))
		}
		fmt.Printf("Testing based on pre-built binary(github action) locates in %v\n", p)
	}

	// set Notation key and cert path
	setPathValue(envNotationE2EKeyPath, &NotationE2EKeyPath)
	setPathValue(envNotationE2ECertPath, &NotationE2ECertPath)

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
