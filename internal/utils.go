package internal

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func IsProduction(env string) bool {
	return env == "prod" || env == "production"
}

func VerifyPassword(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

func HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
