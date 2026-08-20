
package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsageLog struct {
	ID         string
	UserID     string
	Provider   string
	ModelUsed  string
	TokensUsed int
	LatencyMs  int64
	CreatedAt  time.Time
}

type UsageRepository struct {
	db *pgxpool.Pool
}

func NewUsageRepository(db *pgxpool.Pool) *UsageRepository {
	return &UsageRepository{
		db: db,
	}
}

func (r *UsageRepository) CreateUsageLog(
	ctx context.Context,
	userID string,
	provider string,
	modelUsed string,
	tokensUsed int,
	latencyMs int64,
) (*UsageLog, error) {
	usageLog := &UsageLog{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO usage_logs (
			user_id,
			provider,
			model_used,
			tokens_used,
			latency_ms
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			user_id,
			provider,
			model_used,
			tokens_used,
			latency_ms,
			created_at`,
		userID,
		provider,
		modelUsed,
		tokensUsed,
		latencyMs,
	).Scan(
		&usageLog.ID,
		&usageLog.UserID,
		&usageLog.Provider,
		&usageLog.ModelUsed,
		&usageLog.TokensUsed,
		&usageLog.LatencyMs,
		&usageLog.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return usageLog, nil
}
