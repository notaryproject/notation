package gpg

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
)

func findEntityFromFile(path, name string) (*openpgp.Entity, string, error) {
	list, err := readKeyRingFromFile(path)
	if err != nil {
		return nil, "", err
	}
	return findEntity(list, name)
}

func readKeyRingFromFile(path string) (openpgp.EntityList, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return openpgp.ReadKeyRing(file)
}

func findEntity(list openpgp.EntityList, name string) (*openpgp.Entity, string, error) {
	var candidate *openpgp.Entity
	var candidateIdentity string
	for _, entity := range list {
		for identity := range entity.Identities {
			if identity == name {
				return entity, identity, nil
			}
			if strings.Contains(identity, name) {
				if candidate != nil {
					return nil, "", errors.New("ambiguous identity: " + name)
				}
				candidate = entity
				candidateIdentity = identity
			}
		}
	}
	if candidate == nil {
		return nil, "", errors.New("identity not found: " + name)
	}
	return candidate, candidateIdentity, nil
}
