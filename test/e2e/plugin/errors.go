package main

import "github.com/notaryproject/notation-go/plugin/proto"

const (
	ErrorCodeInvalidKeyID       proto.ErrorCode = "INVALID_KEY_ID"
	ErrorCodeInvalidCertificate proto.ErrorCode = "INVALID_CERTIFICATE"
	ErrorCodeConfigError        proto.ErrorCode = "CONFIG_ERROR"
	ErrorCodeSigningError       proto.ErrorCode = "SIGNING_ERROR"
)
