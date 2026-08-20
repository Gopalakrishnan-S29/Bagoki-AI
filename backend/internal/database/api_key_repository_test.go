
package database

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetAPIKeyByUser(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}

	ctx := context.Background()

	pool, err := Connect(ctx)
	if err != nil {
		t.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	userRepo := NewUserRepository(pool)

	testEmail := "api-key-retrieval-test@example.com"

	// Remove any leftover test data from previous runs.
	_, err = pool.Exec(
		ctx,
		`DELETE FROM users WHERE email = $1`,
		testEmail,
	)
	if err != nil {
		t.Fatalf("failed to clean previous test user: %v", err)
	}

	user, err := userRepo.CreateUser(
		ctx,
		"API Key Retrieval Test User",
		testEmail,
		"test-password-hash",
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	defer func() {
		_, err := pool.Exec(
			ctx,
			"DELETE FROM users WHERE id = $1",
			user.ID,
		)
		if err != nil {
			t.Logf("warning: failed to clean up test user: %v", err)
		}
	}()

	apiKeyRepo := NewAPIKeyRepository(pool)

	_, err = apiKeyRepo.SaveAPIKey(
		ctx,
		user.ID,
		"openai",
		"encrypted-retrieval-test-key",
	)
	if err != nil {
		t.Fatalf("failed to save test API key: %v", err)
	}

	apiKey, err := apiKeyRepo.GetAPIKeyByUser(
		ctx,
		user.ID,
		"openai",
	)
	if err != nil {
		t.Fatalf("failed to get API key by user: %v", err)
	}

	if apiKey.ID == "" {
		t.Fatal("expected API key ID, got empty ID")
	}

	if apiKey.UserID != user.ID {
		t.Fatalf(
			"expected user ID %q, got %q",
			user.ID,
			apiKey.UserID,
		)
	}

	if apiKey.Provider != "openai" {
		t.Fatalf(
			"expected provider %q, got %q",
			"openai",
			apiKey.Provider,
		)
	}

	if apiKey.EncryptedKey != "encrypted-retrieval-test-key" {
		t.Fatalf(
			"expected encrypted key %q, got %q",
			"encrypted-retrieval-test-key",
			apiKey.EncryptedKey,
		)
	}

	if !apiKey.Active {
		t.Fatal("expected API key to be active")
	}

	t.Logf(
		"successfully retrieved API key: ID=%s, Provider=%s",
		apiKey.ID,
		apiKey.Provider,
	)
}

