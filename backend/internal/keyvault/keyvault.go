package keyvault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// KeyVault encrypts and decrypts provider API keys using AES-256-GCM.
//
// The master key is provided as a Base64-encoded 32-byte value.
// The decoded key must contain exactly 32 bytes.
type KeyVault struct {
	gcm cipher.AEAD
}

// New creates a KeyVault from a Base64-encoded 32-byte master key.
//
// The environment variable MASTER_ENCRYPTION_KEY should contain
// the output of:
//
//	openssl rand -base64 32
func New(masterKey string) (*KeyVault, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return nil, errors.New("master key must be valid base64")
	}

	if len(decodedKey) != 32 {
		return nil, errors.New(
			"master key must decode to exactly 32 bytes (AES-256)",
		)
	}

	block, err := aes.NewCipher(decodedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &KeyVault{
		gcm: gcm,
	}, nil
}

// Encrypt returns a Base64-encoded ciphertext safe to store
// in the api_keys.encrypted_key column.
func (k *KeyVault) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, k.gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := k.gcm.Seal(
		nonce,
		nonce,
		[]byte(plaintext),
		nil,
	)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt.
//
// The plaintext API key should exist only in memory while
// it is being used by the provider.
func (k *KeyVault) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	nonceSize := k.gcm.NonceSize()

	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := k.gcm.Open(
		nil,
		nonce,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}