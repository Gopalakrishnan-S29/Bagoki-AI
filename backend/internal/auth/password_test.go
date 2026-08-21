package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("expected password hash, got empty string")
	}

	if hash == password {
		t.Fatal("password must not be stored as plain text")
	}

	if !strings.HasPrefix(hash, "$2") {
		t.Fatalf("expected bcrypt hash, got: %s", hash)
	}

	if !CheckPassword(password, hash) {
		t.Fatal("expected correct password to pass verification")
	}

	if CheckPassword("WrongPassword123!", hash) {
		t.Fatal("expected incorrect password to fail verification")
	}

	t.Log("successfully hashed and verified password")
}