package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Conversation struct {
	ID        string
	UserID    string
	Title     string
	CreatedAt time.Time
}

type ConversationRepository struct {
	db *pgxpool.Pool
}

func NewConversationRepository(db *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{
		db: db,
	}
}

func (r *ConversationRepository) CreateConversation(
	ctx context.Context,
	userID string,
	title string,
) (*Conversation, error) {
	conversation := &Conversation{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO conversations (user_id, title)
		 VALUES ($1, $2)
		 RETURNING id, user_id, title, created_at`,
		userID,
		title,
	).Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.Title,
		&conversation.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (r *ConversationRepository) GetConversationByID(
	ctx context.Context,
	id string,
) (*Conversation, error) {
	conversation := &Conversation{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, user_id, title, created_at
		 FROM conversations
		 WHERE id = $1`,
		id,
	).Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.Title,
		&conversation.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return conversation, nil
}

