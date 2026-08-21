package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/yourname/ai-router/internal/database"
)

func TestAuthHandlers(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatalf("failed to load .env: %v", err)
	}

	originalSecret := os.Getenv("JWT_SECRET")

	testSecret := "test-jwt-secret-key-that-is-at-least-32-characters-long"

	if err := os.Setenv("JWT_SECRET", testSecret); err != nil {
		t.Fatalf("failed to set JWT_SECRET: %v", err)
	}

	defer func() {
		if originalSecret == "" {
			_ = os.Unsetenv("JWT_SECRET")
		} else {
			_ = os.Setenv("JWT_SECRET", originalSecret)
		}
	}()

	ctx := context.Background()

	pool, err := database.Connect(ctx)
	if err != nil {
		t.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	userRepo := database.NewUserRepository(pool)

	tokenManager, err := NewTokenManager()
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	handler := NewAuthHandler(userRepo, tokenManager)

	email := "auth-handler-test@example.com"

	// Remove leftover test data from previous runs.
	_, err = pool.Exec(
		ctx,
		"DELETE FROM users WHERE email = $1",
		email,
	)
	if err != nil {
		t.Fatalf("failed to clean previous test user: %v", err)
	}

	t.Run("Signup", func(t *testing.T) {
		body := `{
			"name": "Authentication Test User",
			"email": "auth-handler-test@example.com",
			"password": "TestPassword123!"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/auth/signup",
			strings.NewReader(body),
		)

		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()

		handler.Signup(recorder, req)

		if recorder.Code != http.StatusCreated {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusCreated,
				recorder.Code,
				recorder.Body.String(),
			)
		}

		var response authResponse

		if err := json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		); err != nil {
			t.Fatalf("failed to decode signup response: %v", err)
		}

		if response.UserID == "" {
			t.Fatal("expected user ID in signup response")
		}

		if response.Email != email {
			t.Fatalf(
				"expected email %q, got %q",
				email,
				response.Email,
			)
		}

		if response.Token != "" {
			t.Fatal("signup response should not contain a JWT token")
		}

		t.Logf(
			"successfully created user: ID=%s, Email=%s",
			response.UserID,
			response.Email,
		)
	})

	defer func() {
		_, err := pool.Exec(
			ctx,
			"DELETE FROM users WHERE email = $1",
			email,
		)
		if err != nil {
			t.Logf("warning: failed to clean up test user: %v", err)
		}
	}()

	t.Run("Login", func(t *testing.T) {
		body := `{
			"email": "auth-handler-test@example.com",
			"password": "TestPassword123!"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/auth/login",
			strings.NewReader(body),
		)

		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()

		handler.Login(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d: %s",
				http.StatusOK,
				recorder.Code,
				recorder.Body.String(),
			)
		}

		var response authResponse

		if err := json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		); err != nil {
			t.Fatalf("failed to decode login response: %v", err)
		}

		if response.UserID == "" {
			t.Fatal("expected user ID in login response")
		}

		if response.Token == "" {
			t.Fatal("expected JWT token in login response")
		}

		if response.Email != email {
			t.Fatalf(
				"expected email %q, got %q",
				email,
				response.Email,
			)
		}

		claims, err := tokenManager.ValidateToken(response.Token)
		if err != nil {
			t.Fatalf(
				"failed to validate returned JWT: %v",
				err,
			)
		}

		if claims.UserID != response.UserID {
			t.Fatalf(
				"expected token user ID %q, got %q",
				response.UserID,
				claims.UserID,
			)
		}

		t.Logf(
			"successfully logged in user and received JWT: UserID=%s",
			response.UserID,
		)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		body := `{
			"email": "auth-handler-test@example.com",
			"password": "WrongPassword123!"
		}`

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/auth/login",
			strings.NewReader(body),
		)

		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()

		handler.Login(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusUnauthorized,
				recorder.Code,
			)
		}

		t.Log("successfully rejected incorrect password")
	})
}