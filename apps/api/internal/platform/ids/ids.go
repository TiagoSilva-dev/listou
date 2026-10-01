// Package ids generates time-ordered UUIDv7 identifiers and opaque tokens.
package ids

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/google/uuid"
)

// New returns a UUIDv7. It panics only if the system RNG fails, which is unrecoverable.
func New() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

// Parse parses a UUID from user input, reporting ok=false on malformed values.
func Parse(s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	return id, err == nil
}

// Token returns a URL-safe random token with 256 bits of entropy and its SHA-256 hash.
func Token() (token string, hash []byte) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token)
}

func HashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
