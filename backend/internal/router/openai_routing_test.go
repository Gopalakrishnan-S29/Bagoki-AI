package router

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOpenAIRoutingDecision(t *testing.T) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	// =====================================================
	// USER PROMPT
	// =====================================================

	prompt := "Create a realistic image of a futuristic cybersecurity operations center."

	fmt.Println("\n========================================")
	fmt.Println("USER PROMPT:")
	fmt.Println(prompt)

	// =====================================================
	// STEP 1 — OLLAMA DECISION ENGINE
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

	// =====================================================
	// STEP 2 — DISPLAY DECISION
	// =====================================================

	fmt.Println("\nOLLAMA DECISION")
	fmt.Println("----------------------------------------")
	fmt.Printf("TASK: %s\n", decision.TaskType)
	fmt.Printf("COMPLEXITY: %s\n", decision.Complexity)
	fmt.Printf("PROVIDER: %s\n", decision.Provider)
	fmt.Printf("MODEL: %s\n", decision.Model)
	fmt.Printf("REASON: %s\n", decision.Reason)

	// =====================================================
	// STEP 3 — VALIDATE ROUTING DECISION
	// =====================================================

	if decision.TaskType != string(TaskImage) {
		t.Fatalf(
			"expected task type %q, got %q",
			TaskImage,
			decision.TaskType,
		)
	}

	if decision.Provider != "openai" {
		t.Fatalf(
			"expected provider %q, got %q",
			"openai",
			decision.Provider,
		)
	}

	// =====================================================
	// SUCCESS
	// =====================================================

	fmt.Println("----------------------------------------")
	fmt.Println("OPENAI ROUTING DECISION: PASS")
	fmt.Println("========================================")
}
