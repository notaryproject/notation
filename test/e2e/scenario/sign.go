package scenario

import (
	. "github.com/notaryproject/notation/test/e2e/internal/notation"
	"github.com/notaryproject/notation/test/e2e/internal/utils"
	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
)

var _ = Describe("notation sign", func() {
	It("test basic sign", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			img := GenImage()
			defer img.ClearImage()
			notation.Exec("sign", img.GUN())
		})
	})
	It("test basic sign", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			img := GenImage()
			defer img.ClearImage()
			notation.Exec("sign", img.GUN())
		})
	})
	It("test basic sign", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			img := GenImage()
			defer img.ClearImage()
			notation.Exec("sign", img.GUN())
		})
	})
	It("test basic sign", func() {
		Host(BaseOptions(), func(notation *utils.ExecOpts, vhost *utils.VirtualHost) {
			img := GenImage()
			defer img.ClearImage()
			notation.Exec("sign", img.GUN())
		})
	})
})
