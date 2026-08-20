
package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Message struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	Provider       string
	ModelUsed      string
	TokensUsed     int
	LatencyMs      int64
	CreatedAt      time.Time
}

type MessageRepository struct {
	db *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

func (r *MessageRepository) CreateMessage(
	ctx context.Context,
	conversationID string,
	role string,
	content string,
	provider string,
	modelUsed string,
	tokensUsed int,
	latencyMs int64,
) (*Message, error) {
	message := &Message{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO messages (
			conversation_id,
			role,
			content,
			provider,
			model_used,
			tokens_used,
			latency_ms
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			conversation_id,
			role,
			content,
			provider,
			model_used,
			tokens_used,
			latency_ms,
			created_at`,
		conversationID,
		role,
		content,
		provider,
		modelUsed,
		tokensUsed,
		latencyMs,
	).Scan(
		&message.ID,
		&message.ConversationID,
		&message.Role,
		&message.Content,
		&message.Provider,
		&message.ModelUsed,
		&message.TokensUsed,
		&message.LatencyMs,
		&message.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return message, nil
}
