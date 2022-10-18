package scenario

import (
	"fmt"

	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
				commands utils.CommandGroup
			)
			BeforeEach(func() {
				Expect(utils.WritePolicy(configDir, utils.DefaultPolicy)).NotTo(HaveOccurred())
				commands = utils.CommandGroup{
					{
						Description: "prepare a key and cert",
						Args: []string{
							"cert", "generate-test", utils.DefaultStore, "--trust",
						},
					},
					{
						Description: "sign",
						Args: []string{
							"sign", reference,
							"-k", utils.DefaultStore,
							"--envelope-type", "cose",
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
			utils.ExecCommandGroup("sign and verify", &commands)
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
						Description: "sign",
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
			utils.ExecCommandGroupInContainer("docker sign and verify", &commands)
		})
	})
})
