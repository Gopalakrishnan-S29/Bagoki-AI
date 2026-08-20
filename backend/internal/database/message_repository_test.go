
package database

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestMessageRepository(t *testing.T) {
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
		"Message Test User",
		"message-test@example.com",
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

	conversationRepo := NewConversationRepository(pool)

	conversation, err := conversationRepo.CreateConversation(
		ctx,
		user.ID,
		"Message Test Conversation",
	)
	if err != nil {
		t.Fatalf("failed to create test conversation: %v", err)
	}

	messageRepo := NewMessageRepository(pool)

	message, err := messageRepo.CreateMessage(
		ctx,
		conversation.ID,
		"user",
		"Hello, this is a test message.",
		"openai",
		"gpt-4.1-mini",
		25,
		150,
	)
	if err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	if message.ID == "" {
		t.Fatal("expected message ID, got empty ID")
	}

	if message.ConversationID != conversation.ID {
		t.Fatalf(
			"expected conversation ID %q, got %q",
			conversation.ID,
			message.ConversationID,
		)
	}

	if message.Role != "user" {
		t.Fatalf(
			"expected role %q, got %q",
			"user",
			message.Role,
		)
	}

	if message.Content != "Hello, this is a test message." {
		t.Fatalf(
			"unexpected message content: %q",
			message.Content,
		)
	}

	if message.Provider != "openai" {
		t.Fatalf(
			"expected provider %q, got %q",
			"openai",
			message.Provider,
		)
	}

	if message.ModelUsed != "gpt-4.1-mini" {
		t.Fatalf(
			"expected model %q, got %q",
			"gpt-4.1-mini",
			message.ModelUsed,
		)
	}

	if message.TokensUsed != 25 {
		t.Fatalf(
			"expected tokens used %d, got %d",
			25,
			message.TokensUsed,
		)
	}

	if message.LatencyMs != 150 {
		t.Fatalf(
			"expected latency %d, got %d",
			150,
			message.LatencyMs,
		)
	}

	t.Logf(
		"successfully created message: ID=%s, Role=%s, Provider=%s, Model=%s",
		message.ID,
		message.Role,
		message.Provider,
		message.ModelUsed,
	)
}

