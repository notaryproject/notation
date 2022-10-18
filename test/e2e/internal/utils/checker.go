package utils

import (
	"os"

	. "github.com/onsi/gomega"
)

// CheckFileExist checks file exists.
func CheckFileExist(f string) {
	_, err := os.Stat(f)
	Expect(err).ShouldNot(HaveOccurred())
}

// CheckFileNotExist checks file not exist.
func CheckFileNotExist(f string) {
	_, err := os.Stat(f)
	Expect(err).Should(HaveOccurred())
	Expect(os.IsNotExist(err)).To(BeTrue())
}

// // CheckSignatureFormatSupported checks signature from sigPath is supported by notation.
// //
// // Currently only cose and jws is supported.
// func CheckSignatureFormatSupported(sigPath string) string {
// 	sig, err := os.ReadFile(sigPath)
// 	Expect(err).ShouldNot(HaveOccurred())
// 	for _, mediaType := range signature.RegisteredEnvelopeTypes() {
// 		if _, err := signature.ParseEnvelope(mediaType, sig); err == nil {
// 			return mediaType
// 		}
// 	}
// 	Expect(true).To(BeFalse())
// 	return ""
// }

// // CheckSignatureFormatJWS checks whether signature is jws format.
// func CheckSignatureFormatJWS(sigPath string) {
// 	t := CheckSignatureFormatSupported(sigPath)
// 	Expect(t).To(Equal("application/jws"))
// }

// // CheckSignatureFormatJWS checks whether signature is cose format.
// func CheckSignatureFormatCose(sigPath string) {
// 	t := CheckSignatureFormatSupported(sigPath)
// 	Expect(t).To(Equal("application/cose"))
// }
