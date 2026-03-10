package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random hex string ID.
func GenerateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
