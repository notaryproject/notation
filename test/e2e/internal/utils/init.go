package utils

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/notaryproject/notation-core-go/signature/cose"
	_ "github.com/notaryproject/notation-core-go/signature/jws"
	"github.com/notaryproject/notation-go/dir"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	notationRegistryHost     = "NOTATION_E2E_REGISTRY_HOST"
	notationRegistryUsername = "NOTATION_E2E_REGISTRY_USERNAME"
	notationRegistryPassword = "NOTATION_E2E_REGISTRY_PASSWORD"
	notationBinaryPath       = "NOTATION_E2E_BINARY_PATH"
	githubWorkSpace          = "GITHUB_WORKSPACE"
	targetArtifact           = "NOTATION_E2E_TAGET_ARTIFACT"
	notationBinaryImage      = "NOTATION_BINARY_IMAGE"
)

var TestRegistry registry = registry{
	Host:     "localhost:5000",
	Username: "user",
	Password: "password",
	Artifact: "localhost:5000/net-monitor:v1",
}

var (
	NotationBinaryPath  string
	NotationE2EKeyPath  string
	NotationE2ECertPath string
	NotationBinaryImage string = "notation-e2e"
)

func setUpNotationBinary() {
	p := os.Getenv(notationBinaryPath)
	var err error
	if p != "" && filepath.IsAbs(p) {
		NotationBinaryPath = p
		fmt.Printf("Testing based on pre-built binary locates in %v\n", p)
	} else if workspacePath := os.Getenv(githubWorkSpace); workspacePath != "" && !filepath.IsAbs(p) {
		NotationBinaryPath = filepath.Join(workspacePath, p)
		NotationBinaryPath, err = filepath.Abs(NotationBinaryPath)
		if err != nil {
			panic(fmt.Sprintf("E2E setup failed:%v", err))
		}
		fmt.Printf("Testing based on pre-built binary(github action) locates in %v\n", p)
	} else {
		// TODO: do we need this if
		NotationBinaryPath, err = gexec.Build("github.com/notaryproject/notation/cmd/notation")
		if err != nil {
			panic(fmt.Sprintf("E2E setup failed:%v", err))
		}
		fmt.Printf("Testing based on pre-built binary(github source code) locates in %v\n", notationBinaryPath)
	}
}

func setUpRegistry() {
	if host := os.Getenv(notationRegistryHost); host != "" {
		TestRegistry.Host = host
		fmt.Printf("Testing using $%v as e2e registry host. Hostname: %v\n", notationRegistryHost, host)
	}

	if username := os.Getenv(notationRegistryUsername); username != "" {
		TestRegistry.Username = username
		fmt.Printf("Testing using $%v as e2e registry username. Username: %v\n", notationRegistryUsername, username)
	}

	if password := os.Getenv(notationRegistryPassword); password != "" {
		TestRegistry.Password = password
		fmt.Printf("Testing using $%v as e2e registry password. Password: %v\n", notationRegistryPassword, password)
	}

	if artifact := os.Getenv(targetArtifact); artifact != "" {
		TestRegistry.Artifact = artifact
		fmt.Printf("Testing using $%v as artifact. Artifact: %v\n", targetArtifact, artifact)
	}

	if err := TestRegistry.Validate(); err != nil {
		panic(fmt.Sprintf("E2E setup failed: %v", err))
	}
}

func setUpKeyCerts() {
	NotationE2EKeyPath, NotationE2ECertPath = dir.Path.Localkey("e2e")
	_, err := Exec("rm", ExecOpts{}, NotationE2EKeyPath, NotationE2ECertPath)
	if err != nil {
		panic(fmt.Sprintf("E2E set up certs failed: %v", err))
	}
	_, err = Exec(NotationBinaryPath, ExecOpts{}, "cert", "generate-test", "e2e")
	if err != nil {
		panic(fmt.Sprintf("E2E set up certs failed: %v", err))
	}
	fmt.Printf("Testing based on private key locates in %v\n", NotationE2EKeyPath)
	fmt.Printf("Testing based on certificate chain locates in %v\n", NotationE2ECertPath)
}

func setUpNotationImage() {
	if image := os.Getenv(notationBinaryImage); image != "" {
		NotationBinaryImage = image
		fmt.Printf("Testing using $%v as notation image. Image: %v\n", notationBinaryImage, NotationBinaryImage)
	}
}

func init() {
	RegisterFailHandler(Fail)
	setUpRegistry()
	setUpNotationBinary()
	setUpKeyCerts()
	setUpNotationImage()
}
