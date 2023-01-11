package scenario_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScenario(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scenario Suite")
}
