package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yourname/ai-router/internal/auth"
	"github.com/yourname/ai-router/internal/database"
	"github.com/yourname/ai-router/internal/keyvault"
	"github.com/yourname/ai-router/internal/providers"
	"github.com/yourname/ai-router/internal/router"
)

type chatRequest struct {
	Prompt string `json:"prompt"`
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
	// Load environment variables from .env.
	if err := godotenv.Load(); err != nil {
		log.Println(
			"warning: .env file not found, using environment variables",
		)
	}

	// Connect to Neon PostgreSQL.
	ctx := context.Background()

	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}
	defer db.Close()

	log.Println("connected to Neon PostgreSQL")

	// Initialize repositories.
	userRepository := database.NewUserRepository(db)
	apiKeyRepository := database.NewAPIKeyRepository(db)

	// Initialize JWT manager.
	tokenManager, err := auth.NewTokenManager()
	if err != nil {
		log.Fatal(
			"authentication initialization failed: ",
			err,
		)
	}

	// Initialize Key Vault.
	masterEncryptionKey := os.Getenv("MASTER_ENCRYPTION_KEY")

	if masterEncryptionKey == "" {
		log.Fatal("MASTER_ENCRYPTION_KEY is not set")
	}

	vault, err := keyvault.New(masterEncryptionKey)
	if err != nil {
		log.Fatal(
			"key vault initialization failed: ",
			err,
		)
	}

	// Initialize API key service.
	apiKeyService := keyvault.NewAPIKeyService(
		apiKeyRepository,
		vault,
	)

	// Initialize API key handler.
	apiKeyHandler := keyvault.NewAPIKeyHandler(
		apiKeyService,
	)

	// Initialize authentication handler.
	authHandler := auth.NewAuthHandler(
		userRepository,
		tokenManager,
	)

	// Initialize provider registry.
	registry := providers.NewRegistry()

	registry.Register(
		providers.NewOpenAIProvider(),
	)

	// Add additional providers here later.
	// registry.Register(providers.NewAnthropicProvider())

	mux := http.NewServeMux()

	// =========================================================
	// Health endpoint
	// =========================================================

	mux.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		},
	)

	// =========================================================
	// Authentication endpoints
	// =========================================================

	mux.HandleFunc(
		"/api/auth/signup",
		authHandler.Signup,
	)

	mux.HandleFunc(
		"/api/auth/login",
		authHandler.Login,
	)

	// =========================================================
	// Protected API-key endpoint
	// =========================================================

	mux.Handle(
		"/api/keys",
		tokenManager.Middleware(
			http.HandlerFunc(apiKeyHandler.SaveAPIKey),
		),
	)

	// =========================================================
	// Protected Chat endpoint
	// =========================================================

	chatHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodPost {
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
				return
			}

			var req chatRequest

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(
					w,
					"invalid request body",
					http.StatusBadRequest,
				)
				return
			}

			if req.Prompt == "" {
				http.Error(
					w,
					"prompt is required",
					http.StatusBadRequest,
				)
				return
			}

			// Get authenticated user ID from JWT middleware.
			userID, ok := auth.UserIDFromContext(
				r.Context(),
			)

			if !ok {
				http.Error(
					w,
					"authenticated user not found",
					http.StatusUnauthorized,
				)
				return
			}

			log.Printf(
				"chat request from authenticated user: %s",
				userID,
			)

			// -------------------------------------------------
			// 1. Analyze the prompt.
			// -------------------------------------------------

			analysis := router.AnalyzePrompt(
				req.Prompt,
			)

			// -------------------------------------------------
			// 2. Decide provider and model.
			// -------------------------------------------------

			choice := router.Decide(
				analysis,
			)

			// -------------------------------------------------
			// 3. Find provider adapter.
			// -------------------------------------------------

			provider, ok := registry.Get(
				choice.Provider,
			)

			if !ok {
				http.Error(
					w,
					"provider not available: "+choice.Provider,
					http.StatusInternalServerError,
				)
				return
			}

			// -------------------------------------------------
			// 4. Retrieve user's encrypted API key.
			//    APIKeyService decrypts it only in memory.
			// -------------------------------------------------

			apiKey, err := apiKeyService.GetAPIKey(
				r.Context(),
				userID,
				choice.Provider,
			)

			if err != nil {
				log.Printf(
					"failed to retrieve API key for user %s and provider %s: %v",
					userID,
					choice.Provider,
					err,
				)

				http.Error(
					w,
					"provider API key is not configured",
					http.StatusBadRequest,
				)
				return
			}

			// -------------------------------------------------
			// 5. Call selected provider.
			// -------------------------------------------------

			providerCtx, cancel := context.WithTimeout(
				r.Context(),
				60*time.Second,
			)
			defer cancel()

			result, err := provider.CallModel(
				providerCtx,
				apiKey,
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

			// -------------------------------------------------
			// 6. Build normalized response.
			// -------------------------------------------------

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

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			if err := json.NewEncoder(w).Encode(resp); err != nil {
				log.Printf(
					"failed to encode chat response: %v",
					err,
				)
			}
		},
	)

	// Require JWT authentication for /api/chat.
	mux.Handle(
		"/api/chat",
		tokenManager.Middleware(
			chatHandler,
		),
	)

	// =========================================================
	// Start server
	// =========================================================

	log.Println(
		"ai-router backend listening on :8080",
	)

	if err := http.ListenAndServe(
		":8080",
		mux,
	); err != nil {
		log.Fatal(err)
	}
}