package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/yourname/ai-router/internal/database"
)

type AuthHandler struct {
	users  *database.UserRepository
	tokens *TokenManager
}

func NewAuthHandler(
	users *database.UserRepository,
	tokens *TokenManager,
) *AuthHandler {
	return &AuthHandler{
		users:  users,
		tokens: tokens,
	}
}

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Token  string `json:"token,omitempty"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req signupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(
			w,
			"password must be at least 8 characters",
			http.StatusBadRequest,
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	_, err := h.users.GetUserByEmail(ctx, req.Email)
	if err == nil {
		http.Error(w, "email already registered", http.StatusConflict)
		return
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		http.Error(
			w,
			"failed to hash password",
			http.StatusInternalServerError,
		)
		return
	}

	user, err := h.users.CreateUser(
		ctx,
		req.Name,
		req.Email,
		passwordHash,
	)
	if err != nil {
		http.Error(
			w,
			"failed to create user",
			http.StatusInternalServerError,
		)
		return
	}

	response := authResponse{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" || req.Password == "" {
		http.Error(
			w,
			"email and password are required",
			http.StatusBadRequest,
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	user, err := h.users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		http.Error(
			w,
			"invalid email or password",
			http.StatusUnauthorized,
		)
		return
	}

	if !CheckPassword(req.Password, user.PasswordHash) {
		http.Error(
			w,
			"invalid email or password",
			http.StatusUnauthorized,
		)
		return
	}

	token, err := h.tokens.GenerateToken(user.ID)
	if err != nil {
		http.Error(
			w,
			"failed to generate token",
			http.StatusInternalServerError,
		)
		return
	}

	response := authResponse{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Token:  token,
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return
	}
}

var ErrUnauthorized = errors.New("unauthorized")