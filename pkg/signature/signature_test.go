package signature_test

import (
	"testing"
	"time"

	"github.com/notaryproject/notation/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestSign(t *testing.T) {
	signer, key, err := test.GenerateTestKey(2048)
	assert.NoError(t, err)
	assert.NotNil(t, signer)
	assert.NotNil(t, key)

	cert, cbytes, err := test.GenerateTestSelfSignedCert(signer, []string{"test"}, time.Hour)
	assert.NoError(t, err)
	assert.NotNil(t, cert)
	assert.NotNil(t, cbytes)
}
