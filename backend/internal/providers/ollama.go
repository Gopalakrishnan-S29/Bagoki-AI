package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider connects to the local Ollama server.
type OllamaProvider struct {
	httpClient *http.Client
	baseURL    string
}

// NewOllamaProvider creates a new Ollama provider.
func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "http://localhost:11434",
	}
}

// Name returns the provider identifier.
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// ollamaProviderRequest represents a request sent to Ollama.
type ollamaProviderRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// ollamaProviderResponse represents Ollama's response.
type ollamaProviderResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// CallModel sends the user's prompt to the local Ollama model.
func (p *OllamaProvider) CallModel(
	ctx context.Context,
	apiKey string,
	model string,
	prompt string,
) (*CallResult, error) {

	start := time.Now()

	// Ollama is running locally, so apiKey is intentionally unused.
	_ = apiKey

	// Use the model selected by the router.
	if model == "" {
		model = "llama3.2:latest"
	}

	reqBody := ollamaProviderRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal ollama request: %w",
			err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/api/generate",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"build ollama request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"call ollama: %w",
			err,
		)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"read ollama response: %w",
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"ollama returned status %d: %s",
			resp.StatusCode,
			string(respBytes),
		)
	}

	var parsed ollamaProviderResponse

	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return nil, fmt.Errorf(
			"parse ollama response: %w",
			err,
		)
	}

	if parsed.Error != "" {
		return nil, fmt.Errorf(
			"ollama error: %s",
			parsed.Error,
		)
	}

	if parsed.Response == "" {
		return nil, fmt.Errorf(
			"ollama returned empty response",
		)
	}

	return &CallResult{
		Content:    parsed.Response,
		TokensUsed: 0,
		LatencyMs:  time.Since(start).Milliseconds(),
		ModelUsed:  model,
	}, nil
}
