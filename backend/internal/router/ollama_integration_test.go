package router

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yourname/ai-router/internal/providers"
)

func TestOllamaRoutingIntegration(t *testing.T) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		120*time.Second,
	)
	defer cancel()

	// =====================================================
	// STEP 1 — USER PROMPT
	// =====================================================

	prompt := "Design a complex distributed AI routing architecture with detailed fault tolerance, provider failover, observability, security controls, and scaling strategy."

	fmt.Println("\n========================================")
	fmt.Println("USER PROMPT:")
	fmt.Println(prompt)

	// =====================================================
	// STEP 2 — OLLAMA DECISION ENGINE
	// =====================================================

	decisionEngine := NewOllamaDecisionEngine()

	decision, err := decisionEngine.Decide(
		ctx,
		prompt,
	)

	if err != nil {
		t.Fatalf(
			"Ollama decision failed: %v",
			err,
		)
	}

	fmt.Println("\nOLLAMA DECISION")
	fmt.Println("----------------------------------------")
	fmt.Printf("TASK: %s\n", decision.TaskType)
	fmt.Printf("COMPLEXITY: %s\n", decision.Complexity)
	fmt.Printf("PROVIDER: %s\n", decision.Provider)
	fmt.Printf("MODEL: %s\n", decision.Model)
	fmt.Printf("REASON: %s\n", decision.Reason)

	// =====================================================
	// STEP 3 — VERIFY OLLAMA SELECTED ITSELF
	// =====================================================

	if decision.Provider != "ollama" {
		t.Fatalf(
			"expected Ollama provider, got: %s",
			decision.Provider,
		)
	}

	// =====================================================
	// STEP 4 — CREATE OLLAMA PROVIDER
	// =====================================================

	ollamaProvider := providers.NewOllamaProvider()

	// =====================================================
	// STEP 5 — GENERATE FINAL RESPONSE
	// =====================================================

	result, err := ollamaProvider.CallModel(
		ctx,
		"",
		decision.Model,
		prompt,
	)

	if err != nil {
		t.Fatalf(
			"Ollama response failed: %v",
			err,
		)
	}

	// =====================================================
	// STEP 6 — DISPLAY FINAL RESULT
	// =====================================================

	fmt.Println("\nOLLAMA FINAL RESPONSE")
	fmt.Println("----------------------------------------")
	fmt.Printf("%s\n", result.Content)

	fmt.Println("\nMODEL USED:")
	fmt.Println(result.ModelUsed)

	fmt.Printf("LATENCY: %d ms\n", result.LatencyMs)

	fmt.Println("========================================")

	// =====================================================
	// STEP 7 — BASIC VALIDATION
	// =====================================================

	if result.Content == "" {
		t.Fatal("Ollama returned an empty final response")
	}

	if result.ModelUsed == "" {
		t.Fatal("model used was empty")
	}
}
