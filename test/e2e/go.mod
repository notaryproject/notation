module github.com/notaryproject/notation/test/e2e

go 1.24.0

require (
	github.com/notaryproject/notation-core-go v1.3.0
	github.com/notaryproject/notation-go v1.2.0-beta.1.0.20250512015818-2bc67e7695ef
	github.com/onsi/ginkgo/v2 v2.25.1
	github.com/onsi/gomega v1.38.1
	github.com/opencontainers/image-spec v1.1.1
	oras.land/oras-go/v2 v2.6.0
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/fxamacker/cbor/v2 v2.8.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250820193118-f64d9cf942d6 // indirect
	github.com/notaryproject/tspclient-go v1.0.1-0.20250306063739-4f55b14d9f01 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/veraison/go-cose v1.3.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
)

replace github.com/notaryproject/notation/test/e2e/plugin => ./plugin
