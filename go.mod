module github.com/notaryproject/notation

go 1.22

require (
	github.com/notaryproject/notation-core-go v1.0.3
	github.com/notaryproject/notation-go v1.1.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.8.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/term v0.21.0
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
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
)

replace github.com/notaryproject/tspclient-go => github.com/Two-Hearts/tspclient-go v0.0.0-20240618021928-8938258a8bd9

replace github.com/notaryproject/notation-core-go => github.com/Two-Hearts/notation-core-go v0.0.0-20240620100717-817296a010e2

replace github.com/notaryproject/notation-go => github.com/Two-Hearts/notation-go v0.0.0-20240620100629-0697044fdb65
