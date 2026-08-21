package keyvault

import (
	"encoding/base64"
	"testing"
)

func TestKeyVault(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(
		[]byte("12345678901234567890123456789012"),
	)

	vault, err := New(masterKey)
	if err != nil {
		t.Fatalf("failed to create key vault: %v", err)
	}

	plaintext := "sk-test-provider-api-key"

	encrypted, err := vault.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt API key: %v", err)
	}

	if encrypted == "" {
		t.Fatal("expected encrypted value, got empty string")
	}

	if encrypted == plaintext {
		t.Fatal("encrypted value must not equal plaintext")
	}

	decrypted, err := vault.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt API key: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf(
			"expected decrypted value %q, got %q",
			plaintext,
			decrypted,
		)
	}

	t.Log("successfully encrypted and decrypted API key")
}

func TestKeyVaultWrongKey(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(
		[]byte("12345678901234567890123456789012"),
	)

	vault, err := New(masterKey)
	if err != nil {
		t.Fatalf("failed to create key vault: %v", err)
	}

	plaintext := "sk-test-provider-api-key"

	encrypted, err := vault.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt API key: %v", err)
	}

	// Different 32-byte key.
	wrongKey := base64.StdEncoding.EncodeToString(
		[]byte("abcdefghijklmnopqrstuvwxyz123456"),
	)

	wrongVault, err := New(wrongKey)
	if err != nil {
		t.Fatalf(
			"failed to create vault with wrong key: %v",
			err,
		)
	}

	_, err = wrongVault.Decrypt(encrypted)
	if err == nil {
		t.Fatal("expected decryption with wrong key to fail")
	}

	t.Log("successfully rejected decryption with wrong master key")
}

func TestKeyVaultInvalidCiphertext(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(
		[]byte("12345678901234567890123456789012"),
	)

	vault, err := New(masterKey)
	if err != nil {
		t.Fatalf("failed to create key vault: %v", err)
	}

	_, err = vault.Decrypt("not-valid-encrypted-data")
	if err == nil {
		t.Fatal("expected invalid ciphertext to be rejected")
	}

	t.Log("successfully rejected invalid ciphertext")
}

func TestKeyVaultInvalidMasterKeyLength(t *testing.T) {
	invalidKey := base64.StdEncoding.EncodeToString(
		[]byte("too-short"),
	)

	_, err := New(invalidKey)
	if err == nil {
		t.Fatal("expected invalid master key length to be rejected")
	}

	t.Log("successfully rejected invalid master key length")
}

func TestKeyVaultUsesRandomNonce(t *testing.T) {
	masterKey := base64.StdEncoding.EncodeToString(
		[]byte("12345678901234567890123456789012"),
	)

	vault, err := New(masterKey)
	if err != nil {
		t.Fatalf("failed to create key vault: %v", err)
	}

	plaintext := "sk-test-provider-api-key"

	encrypted1, err := vault.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt first value: %v", err)
	}

	encrypted2, err := vault.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt second value: %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Fatal(
			"encrypting the same plaintext should produce different ciphertext",
		)
	}

	t.Log("successfully verified random nonce behavior")
}