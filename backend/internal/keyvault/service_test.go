package keyvault

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/yourname/ai-router/internal/database"
)

func TestAPIKeyService(t *testing.T) {
	// Load environment variables from backend/.env.
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}

	ctx := context.Background()

	// Connect to Neon PostgreSQL.
	db, err := database.Connect(ctx)
	if err != nil {
		t.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	// Read the Base64-encoded master encryption key.
	masterKey := os.Getenv("MASTER_ENCRYPTION_KEY")

	if masterKey == "" {
		t.Fatal("MASTER_ENCRYPTION_KEY is not set")
	}

	// Create KeyVault.
	vault, err := New(masterKey)
	if err != nil {
		t.Fatalf("failed to create key vault: %v", err)
	}

	// Create repositories.
	userRepository := database.NewUserRepository(db)
	apiKeyRepository := database.NewAPIKeyRepository(db)

	// Create API key service.
	service := NewAPIKeyService(
		apiKeyRepository,
		vault,
	)

	// Test data.
	email := "keyvault-service-test@example.com"
	provider := "openai"

	// Fake test API key only.
	// Never use a real provider API key here.
	plainAPIKey := "sk-test-keyvault-service"

	// Clean up previous test data.
	_, err = db.Exec(
		ctx,
		"DELETE FROM api_keys WHERE user_id IN (SELECT id FROM users WHERE email = $1)",
		email,
	)
	if err != nil {
		t.Fatalf(
			"failed to clean previous API key test data: %v",
			err,
		)
	}

	_, err = db.Exec(
		ctx,
		"DELETE FROM users WHERE email = $1",
		email,
	)
	if err != nil {
		t.Fatalf(
			"failed to clean previous test user: %v",
			err,
		)
	}

	// Create test user.
	passwordHash := "$2a$10$7EqJtq98hPqEX7fNZaFWoO9L7xjY2M4Qm4x3e1X2L8N8L1z9L6X7K"

	user, err := userRepository.CreateUser(
		ctx,
		"Key Vault Test User",
		email,
		passwordHash,
	)
	if err != nil {
		t.Fatalf(
			"failed to create test user: %v",
			err,
		)
	}

	// Clean up after the test.
	defer func() {
		_, err := db.Exec(
			ctx,
			"DELETE FROM api_keys WHERE user_id = $1",
			user.ID,
		)
		if err != nil {
			t.Logf(
				"warning: failed to clean API key: %v",
				err,
			)
		}

		_, err = db.Exec(
			ctx,
			"DELETE FROM users WHERE id = $1",
			user.ID,
		)
		if err != nil {
			t.Logf(
				"warning: failed to clean test user: %v",
				err,
			)
		}
	}()

	// Save the fake API key.
	savedKey, err := service.SaveAPIKey(
		ctx,
		user.ID,
		provider,
		plainAPIKey,
	)
	if err != nil {
		t.Fatalf(
			"failed to save API key: %v",
			err,
		)
	}

	if savedKey.ID == "" {
		t.Fatal("expected API key ID")
	}

	if savedKey.EncryptedKey == "" {
		t.Fatal("expected encrypted API key")
	}

	// Confirm plaintext was not stored.
	if savedKey.EncryptedKey == plainAPIKey {
		t.Fatal("API key must not be stored as plaintext")
	}

	if savedKey.Provider != provider {
		t.Fatalf(
			"expected provider %q, got %q",
			provider,
			savedKey.Provider,
		)
	}

	t.Logf(
		"successfully stored encrypted API key: ID=%s, Provider=%s",
		savedKey.ID,
		savedKey.Provider,
	)

	// Retrieve and decrypt the API key.
	retrievedKey, err := service.GetAPIKey(
		ctx,
		user.ID,
		provider,
	)
	if err != nil {
		t.Fatalf(
			"failed to retrieve API key: %v",
			err,
		)
	}

	if retrievedKey != plainAPIKey {
		t.Fatalf(
			"expected decrypted API key %q, got %q",
			plainAPIKey,
			retrievedKey,
		)
	}

	t.Log("successfully retrieved and decrypted API key")
}