package keyvault

import (
	"context"
	"errors"

	"github.com/yourname/ai-router/internal/database"
)

type APIKeyService struct {
	repository *database.APIKeyRepository
	vault      *KeyVault
}

func NewAPIKeyService(
	repository *database.APIKeyRepository,
	vault *KeyVault,
) *APIKeyService {
	return &APIKeyService{
		repository: repository,
		vault:      vault,
	}
}

func (s *APIKeyService) SaveAPIKey(
	ctx context.Context,
	userID string,
	provider string,
	plainAPIKey string,
) (*database.APIKey, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	if provider == "" {
		return nil, errors.New("provider is required")
	}

	if plainAPIKey == "" {
		return nil, errors.New("API key is required")
	}

	encryptedKey, err := s.vault.Encrypt(plainAPIKey)
	if err != nil {
		return nil, err
	}

	return s.repository.SaveAPIKey(
		ctx,
		userID,
		provider,
		encryptedKey,
	)
}

func (s *APIKeyService) GetAPIKey(
	ctx context.Context,
	userID string,
	provider string,
) (string, error) {
	if userID == "" {
		return "", errors.New("user ID is required")
	}

	if provider == "" {
		return "", errors.New("provider is required")
	}

	apiKey, err := s.repository.GetAPIKeyByUser(
		ctx,
		userID,
		provider,
	)
	if err != nil {
		return "", err
	}

	if !apiKey.Active {
		return "", errors.New("API key is inactive")
	}

	plainAPIKey, err := s.vault.Decrypt(
		apiKey.EncryptedKey,
	)
	if err != nil {
		return "", err
	}

	return plainAPIKey, nil
}