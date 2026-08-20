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

type OpenAIProvider struct {
	httpClient *http.Client
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) CallModel(ctx context.Context, apiKey string, model string, prompt string) (*CallResult, error) {
	start := time.Now()

	reqBody := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call openai: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var parsed openAIResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if parsed.Error != nil {
		return nil, fmt.Errorf("openai error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	return &CallResult{
		Content:    parsed.Choices[0].Message.Content,
		TokensUsed: parsed.Usage.TotalTokens,
		LatencyMs:  time.Since(start).Milliseconds(),
		ModelUsed:  model,
	}, nil
}
