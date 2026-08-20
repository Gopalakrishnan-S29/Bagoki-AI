
package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type APIKey struct {
	ID           string
	UserID       string
	Provider     string
	EncryptedKey string
	Active       bool
	CreatedAt    time.Time
}

type APIKeyRepository struct {
	db *pgxpool.Pool
}

func NewAPIKeyRepository(db *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{
		db: db,
	}
}

func (r *APIKeyRepository) SaveAPIKey(
	ctx context.Context,
	userID string,
	provider string,
	encryptedKey string,
) (*APIKey, error) {
	apiKey := &APIKey{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO api_keys (user_id, provider, encrypted_key, active)
		 VALUES ($1, $2, $3, TRUE)
		 RETURNING id, user_id, provider, encrypted_key, active, created_at`,
		userID,
		provider,
		encryptedKey,
	).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Provider,
		&apiKey.EncryptedKey,
		&apiKey.Active,
		&apiKey.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (r *APIKeyRepository) GetAPIKeyByUser(
	ctx context.Context,
	userID string,
	provider string,
) (*APIKey, error) {
	apiKey := &APIKey{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, user_id, provider, encrypted_key, active, created_at
		 FROM api_keys
		 WHERE user_id = $1
		   AND provider = $2
		   AND active = TRUE
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID,
		provider,
	).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Provider,
		&apiKey.EncryptedKey,
		&apiKey.Active,
		&apiKey.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

