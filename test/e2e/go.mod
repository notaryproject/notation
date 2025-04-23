module github.com/notaryproject/notation/test/e2e

go 1.23.0

require (
	github.com/notaryproject/notation-core-go v1.3.0
	github.com/notaryproject/notation-go v1.3.2
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.37.0
	github.com/opencontainers/image-spec v1.1.1
	oras.land/oras-go/v2 v2.5.0
)

require (
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/notaryproject/tspclient-go v1.0.1-0.20250306063739-4f55b14d9f01 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/veraison/go-cose v1.3.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/notaryproject/notation/test/e2e/plugin => ./plugin
