package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateState generates a random state string for OAuth CSRF protection.
func GenerateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateVerifier generates a PKCE code_verifier (random 32-byte base64url string).
func GenerateVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateChallenge derives the PKCE code_challenge from a verifier using S256.
func GenerateChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
