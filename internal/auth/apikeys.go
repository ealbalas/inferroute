package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const keyPrefix = "ir_live_"

// GenerateAPIKey creates a new raw API key in the format ir_live_<48 hex chars>.
// Return the raw key to the user once — it is never stored in plaintext.
func GenerateAPIKey() (raw string, err error) {
	b := make([]byte, 24)
	if _, err = rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return keyPrefix + hex.EncodeToString(b), nil
}

// HashAPIKey bcrypt-hashes a raw API key for storage in the database.
func HashAPIKey(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash: %w", err)
	}
	return string(hash), nil
}

// ValidateAPIKey compares a raw key against its stored bcrypt hash.
// Uses constant-time comparison to prevent timing attacks.
func ValidateAPIKey(raw, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))
	return err == nil
}

// FastKeyHash returns a SHA-256 hex digest of the raw key.
// Use this as a fast lookup index — do the bcrypt check only after finding
// the candidate row by this hash.
func FastKeyHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// ConstantTimeEqual compares two strings without leaking length via timing.
func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
