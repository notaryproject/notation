module github.com/notaryproject/nv2

go 1.16

require (
	github.com/docker/cli v20.10.5+incompatible
	github.com/docker/distribution v0.0.0-20210206161202-6200038bc715
	github.com/docker/docker v20.10.5+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/notaryproject/notary/v2 v2.0.0-00010101000000-000000000000
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sys v0.0.0-20210330210617-4fbd30eecc44 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace (
	github.com/notaryproject/notary/v2 => github.com/shizhMSFT/notary/v2 v2.0.0-20210331100429-b6fe3e90e3d0
	github.com/opencontainers/artifacts => github.com/aviral26/artifacts v0.0.0-20210331072334-fac4f3a4da4e
)
