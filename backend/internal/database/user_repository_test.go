
package database

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestGetUserByID(t *testing.T) {
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

	repo := NewUserRepository(pool)

	user, err := repo.CreateUser(
		ctx,
		"Get By ID Test User",
		"get-by-id-test@example.com",
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

	foundUser, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to get user by ID: %v", err)
	}

	if foundUser.ID != user.ID {
		t.Fatalf(
			"expected ID %q, got %q",
			user.ID,
			foundUser.ID,
		)
	}

	if foundUser.Email != user.Email {
		t.Fatalf(
			"expected email %q, got %q",
			user.Email,
			foundUser.Email,
		)
	}

	if foundUser.Name != user.Name {
		t.Fatalf(
			"expected name %q, got %q",
			user.Name,
			foundUser.Name,
		)
	}

	t.Logf(
		"successfully retrieved user by ID: ID=%s, Email=%s",
		foundUser.ID,
		foundUser.Email,
	)
}
