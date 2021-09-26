package ioutil

import (
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"
)

// ReadAllVerified reads from r until an error or EOF and returns the data it read
// if the data matches the expected digest.
// A successful call returns err == nil not err == EOF.
func ReadAllVerified(r io.Reader, expected digest.Digest) ([]byte, error) {
	digester := expected.Algorithm().Digester()
	content, err := io.ReadAll(io.TeeReader(r, digester.Hash()))
	if err != nil {
		return nil, err
	}
	if actual := digester.Digest(); actual != expected {
		return nil, fmt.Errorf("mismatch digest: expect %v: got %v", expected, actual)
	}
	return content, nil
}
