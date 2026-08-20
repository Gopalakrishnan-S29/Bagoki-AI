
package database

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestUsageRepository(t *testing.T) {
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

	user, err := userRepo.CreateUser(
		ctx,
		"Usage Test User",
		"usage-test@example.com",
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

	repo := NewUsageRepository(pool)

	usageLog, err := repo.CreateUsageLog(
		ctx,
		user.ID,
		"openai",
		"gpt-4.1-mini",
		42,
		275,
	)
	if err != nil {
		t.Fatalf("failed to create usage log: %v", err)
	}

	if usageLog.ID == "" {
		t.Fatal("expected usage log ID, got empty ID")
	}

	if usageLog.UserID != user.ID {
		t.Fatalf(
			"expected user ID %q, got %q",
			user.ID,
			usageLog.UserID,
		)
	}

	if usageLog.Provider != "openai" {
		t.Fatalf(
			"expected provider %q, got %q",
			"openai",
			usageLog.Provider,
		)
	}

	if usageLog.ModelUsed != "gpt-4.1-mini" {
		t.Fatalf(
			"expected model %q, got %q",
			"gpt-4.1-mini",
			usageLog.ModelUsed,
		)
	}

	if usageLog.TokensUsed != 42 {
		t.Fatalf(
			"expected tokens used %d, got %d",
			42,
			usageLog.TokensUsed,
		)
	}

	if usageLog.LatencyMs != 275 {
		t.Fatalf(
			"expected latency %d, got %d",
			275,
			usageLog.LatencyMs,
		)
	}

	t.Logf(
		"successfully created usage log: ID=%s, Provider=%s, Model=%s, Tokens=%d, Latency=%dms",
		usageLog.ID,
		usageLog.Provider,
		usageLog.ModelUsed,
		usageLog.TokensUsed,
		usageLog.LatencyMs,
	)
}

