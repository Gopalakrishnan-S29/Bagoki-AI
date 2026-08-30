package router

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOllamaDecision(t *testing.T) {
	engine := NewOllamaDecisionEngine()

	prompt := "Create an image of a futuristic cybersecurity operations center."

	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	decision, err := engine.Decide(ctx, prompt)
	if err != nil {
		t.Fatalf(
			"Ollama decision failed for %q: %v",
			prompt,
			err,
		)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("PROMPT: %s\n", prompt)
	fmt.Printf("TASK: %s\n", decision.TaskType)
	fmt.Printf("COMPLEXITY: %s\n", decision.Complexity)
	fmt.Printf("PROVIDER: %s\n", decision.Provider)
	fmt.Printf("MODEL: %s\n", decision.Model)
	fmt.Printf("REASON: %s\n", decision.Reason)
	fmt.Printf("========================================\n")
}
