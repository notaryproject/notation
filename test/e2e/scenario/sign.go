package scenario

import (
	"path/filepath"

	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("notation user", func() {
	var (
		configDir    string
		reference    string
		err          error
		imageCleaner func()
		// fileCleaner   func()
	)
	BeforeEach(func() {
		configDir, _, err = utils.SetUpUserDir()
		Expect(err).ShouldNot(HaveOccurred())
		reference, imageCleaner, err = utils.TestRegistry.PushRandomImage()
		Expect(err).ShouldNot(HaveOccurred())
		DeferCleanup(func() {
			imageCleaner()
			// fileCleaner()
		})
	})
	Context("signs", func() {
		When("using cose envelope", func() {
			var (
				sigPath  string
				commands utils.CommandGroup
			)
			BeforeEach(func() {
				sigPath = filepath.Join(configDir, "sign-cose.sig")
				commands = utils.CommandGroup{
					{
						Description: "sign and save the signature to a local file",
						Args: []string{
							"sign", reference,
							"--key-file", utils.NotationE2EKeyPath,
							"--cert-file", utils.NotationE2ECertPath,
							"--envelope-type", "cose",
							"-o", sigPath,
						},
						Checker: func(c utils.CommandOpts, s *gexec.Session) {
							utils.CheckSignatureFormatCose(sigPath)
						},
					},
					{
						Description: "verify signature from remote registry",
						Args: []string{
							"verify", reference,
							"--cert-file", utils.NotationE2ECertPath,
						},
					},
				}
				commands = utils.NewCommandGroup(
					commands,
					utils.WithAuth(utils.TestRegistry.Username, utils.TestRegistry.Password),
					utils.WithUserDir(configDir),
				)

			})
			utils.ExecCommandGroup("sign pull verify", &commands)
			utils.ExecCommandGroupInUserEnv("host sign pull verify", &commands)
		})
	})
	Context("test in docker", func() {
		When("using jws envelope", func() {
			var (
				commands utils.CommandGroup
				certName string
				sigPath  string
			)
			BeforeEach(func() {
				certName = "hello"
				sigPath = filepath.Join(configDir, "sign-jws.sig")
				commands = utils.CommandGroup{
					{
						Description: "prepare a key and cert",
						Args: []string{
							"cert", "generate-test", certName,
						},
					},
					{
						Description: "sign and save the signature to a local file",
						Args: []string{
							"sign", reference,
							"--key", certName,
							"--envelope-type", "jws",
							"-o", sigPath,
							"-p", utils.TestRegistry.Password,
							"-u", utils.TestRegistry.Username,
						},
					},
					{
						Description: "verify signature from remote registry",
						Args: []string{
							"verify", reference,
							"--cert", certName,
							"-p", utils.TestRegistry.Password,
							"-u", utils.TestRegistry.Username,
						},
					},
				}
			})
			utils.ExecCommandGroupInContainer("docker sign pull verify", &commands)
		})
	})
})
