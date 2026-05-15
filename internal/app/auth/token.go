package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const (
	registerVerificationKeyPrefix = "register_verification"
	registerSessionKeyPrefix      = "register_session"
)

func GenerateSecureToken(byteSize int) (string, error) {
	if byteSize <= 0 {
		return "", fmt.Errorf("token byte size must be positive")
	}

	token := make([]byte, byteSize)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(token), nil
}

func RegisterVerificationKey(token string) string {
	return fmt.Sprintf("%s:%s", registerVerificationKeyPrefix, token)
}

func RegisterSessionKey(session string) string {
	return fmt.Sprintf("%s:%s", registerSessionKeyPrefix, session)
}
