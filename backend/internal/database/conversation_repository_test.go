
package database

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestConversationRepository(t *testing.T) {
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

	testEmail := "conversation-test@example.com"

	user, err := userRepo.CreateUser(
		ctx,
		"Conversation Test User",
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

	repo := NewConversationRepository(pool)

	conversation, err := repo.CreateConversation(
		ctx,
		user.ID,
		"Test Conversation",
	)
	if err != nil {
		t.Fatalf("failed to create conversation: %v", err)
	}

	if conversation.ID == "" {
		t.Fatal("expected conversation ID, got empty ID")
	}

	if conversation.UserID != user.ID {
		t.Fatalf(
			"expected user ID %q, got %q",
			user.ID,
			conversation.UserID,
		)
	}

	if conversation.Title != "Test Conversation" {
		t.Fatalf(
			"expected title %q, got %q",
			"Test Conversation",
			conversation.Title,
		)
	}

	t.Logf(
		"successfully created conversation: ID=%s",
		conversation.ID,
	)

	foundConversation, err := repo.GetConversationByID(
		ctx,
		conversation.ID,
	)
	if err != nil {
		t.Fatalf(
			"failed to get conversation by ID: %v",
			err,
		)
	}

	if foundConversation.ID != conversation.ID {
		t.Fatalf(
			"expected conversation ID %q, got %q",
			conversation.ID,
			foundConversation.ID,
		)
	}

	if foundConversation.UserID != user.ID {
		t.Fatalf(
			"expected user ID %q, got %q",
			user.ID,
			foundConversation.UserID,
		)
	}

	if foundConversation.Title != "Test Conversation" {
		t.Fatalf(
			"expected title %q, got %q",
			"Test Conversation",
			foundConversation.Title,
		)
	}

	t.Logf(
		"successfully retrieved conversation: ID=%s, Title=%s",
		foundConversation.ID,
		foundConversation.Title,
	)
}

