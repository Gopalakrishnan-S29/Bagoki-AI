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
	Complexity string `json:"complexity"`
	RiskTier   string `json:"risk_tier"`
	Reason     string `json:"reason"`
	TokensUsed int    `json:"tokens_used"`
	LatencyMs  int64  `json:"latency_ms"`
}

func main() {

	// =========================================================
	// Environment
	// =========================================================

	if err := godotenv.Load(); err != nil {
		log.Println("warning: .env file not found, using environment variables")
	}

	// =========================================================
	// Database
	// =========================================================

	ctx := context.Background()

	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}
	defer db.Close()

	log.Println("connected to Neon PostgreSQL")

	userRepository := database.NewUserRepository(db)
	apiKeyRepository := database.NewAPIKeyRepository(db)

	// =========================================================
	// Authentication
	// =========================================================

	tokenManager, err := auth.NewTokenManager()
	if err != nil {
		log.Fatal("authentication initialization failed: ", err)
	}

	// =========================================================
	// Key Vault
	// =========================================================

	masterEncryptionKey := os.Getenv("MASTER_ENCRYPTION_KEY")

	if masterEncryptionKey == "" {
		log.Fatal("MASTER_ENCRYPTION_KEY is not set")
	}

	vault, err := keyvault.New(masterEncryptionKey)
	if err != nil {
		log.Fatal("key vault initialization failed: ", err)
	}

	apiKeyService := keyvault.NewAPIKeyService(
		apiKeyRepository,
		vault,
	)

	apiKeyHandler := keyvault.NewAPIKeyHandler(
		apiKeyService,
	)

	authHandler := auth.NewAuthHandler(
		userRepository,
		tokenManager,
	)

	// =========================================================
	// Provider Registry
	//
	// Ollama is the actual response provider.
	// OpenAI is registered only because the project keeps
	// the provider architecture ready for future expansion.
	// =========================================================

	registry := providers.NewRegistry()

	registry.Register(
		providers.NewOpenAIProvider(),
	)

	registry.Register(
		providers.NewOllamaProvider(),
	)

	// =========================================================
	// Ollama Decision Engine
	// =========================================================

	ollamaDecisionEngine := router.NewOllamaDecisionEngine()

	log.Println("Ollama decision engine initialized")

	// =========================================================
	// HTTP Router
	// =========================================================

	mux := http.NewServeMux()

	// Serve frontend files.
	mux.Handle(
		"/",
		http.FileServer(http.Dir("../frontend")),
	)

	// =========================================================
	// Health
	// =========================================================

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {

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
	})

	// =========================================================
	// Authentication
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
	// API Keys
	// =========================================================

	mux.Handle(
		"/api/keys",
		tokenManager.Middleware(
			http.HandlerFunc(apiKeyHandler.SaveAPIKey),
		),
	)

	// =========================================================
	// AI ROUTER CHAT
	//
	// Current architecture:
	//
	// User Prompt
	//      ↓
	// Ollama Decision Engine
	//      ↓
	// Task / Complexity / Model Decision
	//      ↓
	// Ollama Provider
	//      ↓
	// Actual Response
	//
	// No image generation.
	// No OpenAI API call for the actual response.
	// =========================================================

	chatHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

		// =====================================================
		// STEP 1 — OLLAMA ROUTING DECISION
		// =====================================================

		decisionCtx, cancel := context.WithTimeout(
			r.Context(),
			60*time.Second,
		)
		defer cancel()

		decision, err := ollamaDecisionEngine.Decide(
			decisionCtx,
			req.Prompt,
		)

		if err != nil {
			http.Error(
				w,
				"routing decision failed: "+err.Error(),
				http.StatusBadGateway,
			)
			return
		}

		log.Printf(
			"ROUTING | task=%s | complexity=%s | provider=%s | model=%s",
			decision.TaskType,
			decision.Complexity,
			decision.Provider,
			decision.Model,
		)

		// =====================================================
		// STEP 2 — GET SELECTED PROVIDER
		// =====================================================

		provider, ok := registry.Get(
			decision.Provider,
		)

		if !ok {
			http.Error(
				w,
				"provider not available: "+decision.Provider,
				http.StatusInternalServerError,
			)
			return
		}

		// =====================================================
		// STEP 3 — CURRENT DEMO ARCHITECTURE
		//
		// The actual answer is generated by Ollama.
		//
		// Even if the decision engine mentions an OpenAI model,
		// we do not call OpenAI for the current demo.
		// =====================================================

		ollamaProvider, ok := provider.(*providers.OllamaProvider)

		if !ok {
			http.Error(
				w,
				"current demo requires Ollama as the response provider",
				http.StatusInternalServerError,
			)
			return
		}

		providerCtx, cancelProvider := context.WithTimeout(
			r.Context(),
			120*time.Second,
		)
		defer cancelProvider()

		result, err := ollamaProvider.CallModel(
			providerCtx,
			"",
			decision.Model,
			req.Prompt,
		)

		if err != nil {
			http.Error(
				w,
				"ollama response failed: "+err.Error(),
				http.StatusBadGateway,
			)
			return
		}

		// =====================================================
		// STEP 4 — RESPONSE
		// =====================================================

		resp := chatResponse{
			Reply:      result.Content,
			Provider:   "ollama",
			Model:      decision.Model,
			TaskType:   decision.TaskType,
			Complexity: decision.Complexity,
			RiskTier:   "",
			Reason:     decision.Reason,
			TokensUsed: result.TokensUsed,
			LatencyMs:  result.LatencyMs,
		}

		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.Handle(
		"/api/router/chat",
		chatHandler,
	)

	// =========================================================
	// CORS
	// =========================================================

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set(
			"Access-Control-Allow-Origin",
			"*",
		)

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		mux.ServeHTTP(w, r)
	})

	// =========================================================
	// START SERVER
	// =========================================================

	log.Println("ai-router backend listening on :8080")

	if err := http.ListenAndServe(
		":8080",
		corsHandler,
	); err != nil {
		log.Fatal(err)
	}
}
