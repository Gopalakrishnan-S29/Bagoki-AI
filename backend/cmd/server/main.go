package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourname/ai-router/internal/database"
	"github.com/yourname/ai-router/internal/providers"
	"github.com/yourname/ai-router/internal/router"
)

type chatRequest struct {
	Prompt string `json:"prompt"`
	APIKey string `json:"api_key"` // v1: passed directly by client; move to
	// decrypted server-side lookup once auth + keyvault are wired into
	// the request pipeline (see internal/keyvault).
}

type chatResponse struct {
	Reply      string `json:"reply"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	TaskType   string `json:"task_type"`
	RiskTier   string `json:"risk_tier"`
	Reason     string `json:"reason"`
	TokensUsed int    `json:"tokens_used"`
	LatencyMs  int64  `json:"latency_ms"`
}

func main() {
	// Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, using environment variables")
	}

	// Connect to Neon PostgreSQL
	ctx := context.Background()

	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}
	defer db.Close()

	log.Println("connected to Neon PostgreSQL")

	// Initialize provider registry
	registry := providers.NewRegistry()
	registry.Register(providers.NewOpenAIProvider())
	// registry.Register(providers.NewAnthropicProvider()) // add next

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Chat endpoint
	mux.HandleFunc("/api/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Prompt == "" {
			http.Error(w, "prompt is required", http.StatusBadRequest)
			return
		}

		// 1. Analyze the prompt
		analysis := router.AnalyzePrompt(req.Prompt)

		// 2. Decide which provider/model to use
		choice := router.Decide(analysis)

		// 3. Look up the provider adapter
		provider, ok := registry.Get(choice.Provider)
		if !ok {
			http.Error(
				w,
				"provider not available: "+choice.Provider,
				http.StatusInternalServerError,
			)
			return
		}

		// 4. Call the model with the caller's own API key (BYOK)
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		result, err := provider.CallModel(
			ctx,
			req.APIKey,
			choice.Model,
			req.Prompt,
		)
		if err != nil {
			http.Error(
				w,
				"provider call failed: "+err.Error(),
				http.StatusBadGateway,
			)
			return
		}

		resp := chatResponse{
			Reply:      result.Content,
			Provider:   choice.Provider,
			Model:      choice.Model,
			TaskType:   string(analysis.TaskType),
			RiskTier:   choice.RiskTier,
			Reason:     choice.Reason,
			TokensUsed: result.TokensUsed,
			LatencyMs:  result.LatencyMs,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("ai-router backend listening on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
