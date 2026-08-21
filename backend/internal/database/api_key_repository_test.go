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

func TestSaveAPIKeyReplacesExistingKey(t *testing.T) {
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

	testEmail := "api-key-replacement-test@example.com"

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
		"API Key Replacement Test User",
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

	// Save the first key.
	firstKey, err := apiKeyRepo.SaveAPIKey(
		ctx,
		user.ID,
		"openai",
		"encrypted-first-key",
	)
	if err != nil {
		t.Fatalf("failed to save first API key: %v", err)
	}

	if firstKey.EncryptedKey != "encrypted-first-key" {
		t.Fatalf("expected first key to be stored")
	}

	// Save a second key for the same user/provider.
	secondKey, err := apiKeyRepo.SaveAPIKey(
		ctx,
		user.ID,
		"openai",
		"encrypted-second-key",
	)
	if err != nil {
		t.Fatalf("failed to replace API key: %v", err)
	}

	// The database row should be updated, not duplicated.
	if secondKey.ID != firstKey.ID {
		t.Fatalf(
			"expected existing API key ID %q to be reused, got %q",
			firstKey.ID,
			secondKey.ID,
		)
	}

	if secondKey.EncryptedKey != "encrypted-second-key" {
		t.Fatalf(
			"expected second encrypted key, got %q",
			secondKey.EncryptedKey,
		)
	}

	if !secondKey.Active {
		t.Fatal("expected replaced API key to remain active")
	}

	// Verify retrieval returns the new key.
	retrieved, err := apiKeyRepo.GetAPIKeyByUser(
		ctx,
		user.ID,
		"openai",
	)
	if err != nil {
		t.Fatalf(
			"failed to retrieve replaced API key: %v",
			err,
		)
	}

	if retrieved.ID != firstKey.ID {
		t.Fatalf(
			"expected retrieved ID %q, got %q",
			firstKey.ID,
			retrieved.ID,
		)
	}

	if retrieved.EncryptedKey != "encrypted-second-key" {
		t.Fatalf(
			"expected retrieved new key %q, got %q",
			"encrypted-second-key",
			retrieved.EncryptedKey,
		)
	}

	// Confirm only one row exists for this user/provider.
	var count int

	err = pool.QueryRow(
		ctx,
		`SELECT COUNT(*)
		 FROM api_keys
		 WHERE user_id = $1
		   AND provider = $2`,
		user.ID,
		"openai",
	).Scan(&count)

	if err != nil {
		t.Fatalf(
			"failed to count API key rows: %v",
			err,
		)
	}

	if count != 1 {
		t.Fatalf(
			"expected exactly 1 API key row, got %d",
			count,
		)
	}

	t.Log(
		"successfully replaced existing API key without creating duplicate rows",
	)
}