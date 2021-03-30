module github.com/notaryproject/nv2

go 1.14

require (
	github.com/notaryproject/notary/v2 v2.0.0-00010101000000-000000000000
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/urfave/cli/v2 v2.2.0
)

replace github.com/notaryproject/notary/v2 => github.com/shizhMSFT/notary/v2 v2.0.0-20210330091034-dc6e56acc97a
