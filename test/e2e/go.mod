module github.com/notaryproject/notation/test/e2e

go 1.21

require (
	github.com/notaryproject/notation-core-go v1.1.0-beta.1
	github.com/onsi/ginkgo/v2 v2.11.0
	github.com/onsi/gomega v1.27.10
	github.com/opencontainers/image-spec v1.1.0
	oras.land/oras-go/v2 v2.4.0
)

require (
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/pprof v0.0.0-20230510103437-eeec1cb781c3 // indirect
	github.com/notaryproject/tspclient-go v0.1.1-0.20240715235637-df25ef8d2172 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/veraison/go-cose v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.9.3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/notaryproject/notation/test/e2e/plugin => ./plugin
