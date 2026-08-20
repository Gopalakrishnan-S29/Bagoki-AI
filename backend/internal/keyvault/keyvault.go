package keyvault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// KeyVault encrypts and decrypts user-supplied provider API keys using
// AES-256-GCM. The master key comes from an environment variable
// (never hardcoded, never committed) and is set once at startup.
type KeyVault struct {
	gcm cipher.AEAD
}

// New creates a KeyVault from a 32-byte master key. Generate one with:
//   openssl rand -base64 32
// and store it as the MASTER_ENCRYPTION_KEY environment variable.
func New(masterKey []byte) (*KeyVault, error) {
	if len(masterKey) != 32 {
		return nil, errors.New("master key must be 32 bytes (AES-256)")
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &KeyVault{gcm: gcm}, nil
}

// Encrypt returns a base64-encoded ciphertext safe to store in the
// api_keys.encrypted_key column.
func (k *KeyVault) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, k.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := k.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt. Called only at the moment a provider call
// is made - the plaintext key should never be logged or persisted
// outside this in-memory use.
func (k *KeyVault) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	nonceSize := k.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := k.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
