package keyvault

import (
	"encoding/json"
	"net/http"

	"github.com/yourname/ai-router/internal/auth"
)

type APIKeyHandler struct {
	service *APIKeyService
}

func NewAPIKeyHandler(service *APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{
		service: service,
	}
}

type saveAPIKeyRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

type saveAPIKeyResponse struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Active   bool   `json:"active"`
}

func (h *APIKeyHandler) SaveAPIKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

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

	var req saveAPIKeyRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	if req.Provider == "" {
		http.Error(
			w,
			"provider is required",
			http.StatusBadRequest,
		)
		return
	}

	if req.APIKey == "" {
		http.Error(
			w,
			"api_key is required",
			http.StatusBadRequest,
		)
		return
	}

	apiKey, err := h.service.SaveAPIKey(
		r.Context(),
		userID,
		req.Provider,
		req.APIKey,
	)
	if err != nil {
		http.Error(
			w,
			"failed to save API key",
			http.StatusInternalServerError,
		)
		return
	}

	response := saveAPIKeyResponse{
		ID:       apiKey.ID,
		Provider: apiKey.Provider,
		Active:   apiKey.Active,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}