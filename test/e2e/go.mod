module github.com/notaryproject/notation/test/e2e

go 1.20

require (
	github.com/notaryproject/notation-core-go v1.0.0-rc.1
	github.com/notaryproject/notation/test/e2e/plugin v1.0.0
	github.com/onsi/ginkgo/v2 v2.3.0
	github.com/onsi/gomega v1.22.1
	github.com/opencontainers/image-spec v1.1.0-rc2
	oras.land/oras-go/v2 v2.0.0-rc.6
)

require (
	github.com/fxamacker/cbor/v2 v2.4.0 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/veraison/go-cose v1.0.0-rc.2 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/notaryproject/notation/test/e2e/plugin => ./plugin
