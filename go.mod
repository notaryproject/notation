module github.com/notaryproject/nv2

go 1.16

require (
	github.com/docker/cli v20.10.5+incompatible
	github.com/docker/distribution v0.0.0-20210206161202-6200038bc715
	github.com/docker/docker v20.10.5+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/notaryproject/notary/v2 v2.0.0-20210414032403-d1367cc13db7
	github.com/opencontainers/artifacts v0.0.0-20210209205009-a282023000bd
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sys v0.0.0-20210330210617-4fbd30eecc44 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace github.com/opencontainers/artifacts => github.com/notaryproject/artifacts v0.0.0-20210414030140-c7c701eff45d
