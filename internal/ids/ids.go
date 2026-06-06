package ids

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
)

var validID = regexp.MustCompile(`^[a-f0-9]{8,32}$`)

func New() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func Valid(id string) bool {
	return validID.MatchString(id)
}
