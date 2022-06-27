module github.com/notaryproject/notation

go 1.18

require (
	github.com/distribution/distribution/v3 v3.0.0-20210804104954-38ab4c606ee3
	github.com/docker/cli v20.10.17+incompatible
	github.com/docker/docker-credential-helpers v0.6.4
	github.com/notaryproject/notation-core-go v0.0.0-20220712013708-3c4b3efa03c5
	github.com/notaryproject/notation-go v0.9.0-alpha.1.0.20220712175603-962d79cd4090
	github.com/opencontainers/go-digest v1.0.0
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/urfave/cli/v2 v2.11.0
	oras.land/oras-go/v2 v2.0.0-20220620164807-8b2a54608a94 // TODO: upgrade to v2.0.0-rc.1 in the next PR
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/docker/docker v20.10.8+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.4.2 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/oras-project/artifacts-spec v1.0.0-rc.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)
