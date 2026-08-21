package auth

import (
	"os"
	"testing"
)

func TestTokenManager(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")

	testSecret := "test-jwt-secret-key-that-is-at-least-32-characters-long"

	if err := os.Setenv("JWT_SECRET", testSecret); err != nil {
		t.Fatalf("failed to set JWT_SECRET: %v", err)
	}

	defer func() {
		if originalSecret == "" {
			_ = os.Unsetenv("JWT_SECRET")
		} else {
			_ = os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	manager, err := NewTokenManager()
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	userID := "test-user-id"

	token, err := manager.GenerateToken(userID)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("expected JWT token, got empty string")
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf(
			"expected user ID %q, got %q",
			userID,
			claims.UserID,
		)
	}

	if claims.Subject != userID {
		t.Fatalf(
			"expected subject %q, got %q",
			userID,
			claims.Subject,
		)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected token expiration time")
	}

	if claims.IssuedAt == nil {
		t.Fatal("expected token issued-at time")
	}

	if !claims.ExpiresAt.After(claims.IssuedAt.Time) {
		t.Fatal("expected expiration time to be after issued-at time")
	}

	t.Logf(
		"successfully generated and validated JWT for user: %s",
		userID,
	)
}

func TestTokenManagerRejectsInvalidToken(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")

	testSecret := "test-jwt-secret-key-that-is-at-least-32-characters-long"

	if err := os.Setenv("JWT_SECRET", testSecret); err != nil {
		t.Fatalf("failed to set JWT_SECRET: %v", err)
	}

	defer func() {
		if originalSecret == "" {
			_ = os.Unsetenv("JWT_SECRET")
		} else {
			_ = os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	manager, err := NewTokenManager()
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	_, err = manager.ValidateToken("this-is-not-a-valid-jwt")
	if err == nil {
		t.Fatal("expected invalid token to be rejected")
	}

	t.Log("successfully rejected invalid JWT")
}