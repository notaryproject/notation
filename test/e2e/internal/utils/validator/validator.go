package validator

import (
	"os"

	. "github.com/onsi/gomega"
)

// CheckFileExist checks file exists.
func CheckFileExist(f string) {
	_, err := os.Stat(f)
	Expect(err).ShouldNot(HaveOccurred())
}

// CheckFileNotExist checks file not exist.
func CheckFileNotExist(f string) {
	_, err := os.Stat(f)
	Expect(err).Should(HaveOccurred())
	Expect(os.IsNotExist(err)).To(BeTrue())
}
