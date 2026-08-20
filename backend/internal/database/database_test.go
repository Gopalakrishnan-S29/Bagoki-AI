package database

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
)

func TestConnect(t *testing.T) {
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

	t.Log("successfully connected to Neon PostgreSQL")
}
