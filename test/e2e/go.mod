module github.com/notaryproject/notation/test/e2e

go 1.24.0

require (
	github.com/notaryproject/notation-core-go v1.2.0
	github.com/notaryproject/notation-go v1.2.0-beta.1.0.20250122072255-6eb53a50d69e
	github.com/onsi/ginkgo/v2 v2.22.2
	github.com/onsi/gomega v1.36.2
	github.com/opencontainers/image-spec v1.1.0
	oras.land/oras-go/v2 v2.5.0
)

require (
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20241210010833-40e02aabc2ad // indirect
	github.com/notaryproject/tspclient-go v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/veraison/go-cose v1.3.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/notaryproject/notation/test/e2e/plugin => ./plugin
