package trustpolicy

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTrustPolicy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trust Policy Suite")
}
