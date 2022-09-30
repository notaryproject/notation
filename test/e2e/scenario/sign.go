package scenario

import (
	"fmt"
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
				Expect(utils.WritePolicy(configDir, utils.DefaultPolicy)).NotTo(HaveOccurred())
				commands = utils.CommandGroup{
					{
						Description: "prepare a key and cert",
						Args: []string{
							"cert", "generate-test", utils.DefaultStore, "--trust",
						},
					},
					{
						Description: "sign and save the signature to a local file",
						Args: []string{
							"sign", reference,
							"-k", utils.DefaultStore,
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
		})
	})
	Context("test in docker", func() {
		When("using jws envelope", func() {
			var commands utils.CommandGroup
			BeforeEach(func() {
				commands = utils.CommandGroup{
					{
						Description: "prepare a key and cert",
						Args: []string{
							"cert", "generate-test", "hello", "--name", utils.DefaultStore, "--trust",
						},
					},
					{
						// For executing in docker, create a new policy to test
						Description: "create a default policy",
						Binary:      "sh",
						Args: []string{
							"-c", fmt.Sprintf("cat <<EOF >> ~/.config/notation/trustpolicy.json\n%v\nEOF", utils.DefaultPolicy),
						},
					},
					{
						Description: "sign and save the signature to a local file",
						Args: []string{
							"sign", reference,
							"--key", utils.DefaultStore,
							"--envelope-type", "jws",
							"-p", utils.TestRegistry.Password,
							"-u", utils.TestRegistry.Username,
						},
					},
					{
						Description: "verify signature from remote registry",
						Args: []string{
							"verify", reference,
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
