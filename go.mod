module github.com/notaryproject/notation

go 1.22

require (
	github.com/notaryproject/notation-core-go v1.0.3-0.20240325061945-807a3386734e
	github.com/notaryproject/notation-go v1.1.1-0.20240327165254-57ff8e68a0a8
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/term v0.20.0
	oras.land/oras-go/v2 v2.5.0
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/fxamacker/cbor/v2 v2.6.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.5 // indirect
	github.com/go-ldap/ldap/v3 v3.4.8 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/notaryproject/notation-plugin-framework-go v1.0.0 // indirect
	github.com/notaryproject/tspclient-go v0.0.0-20240122083733-a373599795a2 // indirect
	github.com/veraison/go-cose v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)

replace github.com/notaryproject/tspclient-go => github.com/Two-Hearts/tspclient-go v0.0.0-20240510080813-e58c4f362fa4

replace github.com/notaryproject/notation-core-go => github.com/Two-Hearts/notation-core-go v0.0.0-20240510081223-bf89fbfde06f

replace github.com/notaryproject/notation-go => github.com/Two-Hearts/notation-go v0.0.0-20240510081650-cde5adb39abf
