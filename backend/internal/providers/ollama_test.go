package providers

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOllamaProvider(t *testing.T) {
	provider := NewOllamaProvider()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	result, err := provider.CallModel(
		ctx,
		"",
		"llama3.2:latest",
		"Explain what a firewall is in simple terms.",
	)

	if err != nil {
		t.Fatalf(
			"Ollama provider failed: %v",
			err,
		)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("PROVIDER: %s\n", provider.Name())
	fmt.Printf("MODEL: %s\n", result.ModelUsed)
	fmt.Printf("LATENCY: %d ms\n", result.LatencyMs)
	fmt.Printf("RESPONSE:\n%s\n", result.Content)
	fmt.Printf("========================================\n")
}
