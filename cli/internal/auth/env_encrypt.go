package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

const (
	envEncCipher = "ed25519+aes-256-gcm-v1"
)

// EncryptEnvBlob encrypts plaintext bytes using a per-installation symmetric key.
// This is client-side encryption for "opaque blob" storage.
func EncryptEnvBlob(plain []byte) (cipherName string, b64Ciphertext string, size int, err error) {
	key, err := getOrCreateSessionKey()
	if err != nil {
		return "", "", 0, err
	}
	if len(key) != 32 {
		return "", "", 0, fmt.Errorf("invalid encryption key length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", 0, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", 0, err
	}

	ct := gcm.Seal(nil, nonce, plain, nil)
	out := append(nonce, ct...)
	return envEncCipher, base64.RawURLEncoding.EncodeToString(out), len(out), nil
}
