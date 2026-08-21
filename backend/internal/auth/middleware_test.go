package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMiddleware(t *testing.T) {
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

	tokenManager, err := NewTokenManager()
	if err != nil {
		t.Fatalf("failed to create token manager: %v", err)
	}

	userID := "middleware-test-user"

	token, err := tokenManager.GenerateToken(userID)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	protectedHandler := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		extractedUserID, ok := UserIDFromContext(r.Context())

		if !ok {
			http.Error(
				w,
				"user ID not found in context",
				http.StatusInternalServerError,
			)
			return
		}

		if extractedUserID != userID {
			http.Error(
				w,
				"unexpected user ID",
				http.StatusInternalServerError,
			)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("protected endpoint reached"))
	})

	middleware := tokenManager.Middleware(protectedHandler)

	t.Run("ValidToken", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/protected",
			nil,
		)

		req.Header.Set(
			"Authorization",
			"Bearer "+token,
		)

		recorder := httptest.NewRecorder()

		middleware.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusOK,
				recorder.Code,
			)
		}

		t.Log("successfully accepted valid JWT")
	})

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/protected",
			nil,
		)

		recorder := httptest.NewRecorder()

		middleware.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusUnauthorized,
				recorder.Code,
			)
		}

		t.Log("successfully rejected request without JWT")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/protected",
			nil,
		)

		req.Header.Set(
			"Authorization",
			"Bearer invalid-token",
		)

		recorder := httptest.NewRecorder()

		middleware.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusUnauthorized,
				recorder.Code,
			)
		}

		t.Log("successfully rejected invalid JWT")
	})

	t.Run("InvalidAuthorizationFormat", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/protected",
			nil,
		)

		req.Header.Set(
			"Authorization",
			"InvalidFormat",
		)

		recorder := httptest.NewRecorder()

		middleware.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusUnauthorized,
				recorder.Code,
			)
		}

		t.Log("successfully rejected invalid authorization format")
	})

	t.Run("WrongScheme", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/protected",
			nil,
		)

		req.Header.Set(
			"Authorization",
			"Basic "+token,
		)

		recorder := httptest.NewRecorder()

		middleware.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status %d, got %d",
				http.StatusUnauthorized,
				recorder.Code,
			)
		}

		t.Log("successfully rejected non-Bearer authentication")
	})
}