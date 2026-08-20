package router

import "strings"

// TaskType represents the classified category of a user prompt.
type TaskType string

const (
	TaskCoding        TaskType = "coding"
	TaskCybersecurity TaskType = "cybersecurity"
	TaskTranslation   TaskType = "translation"
	TaskMath          TaskType = "mathematics"
	TaskWriting       TaskType = "writing"
	TaskGeneral       TaskType = "general"
)

// PromptAnalysis is the result of classifying a prompt, used by the
// decision engine to select a model and risk tier.
type PromptAnalysis struct {
	TaskType   TaskType
	Complexity string // "low" | "medium" | "high"
	RiskTier   string // "low" | "medium" | "high" - see decision.go
}

// keyword sets used for v1 rule-based classification.
// This is intentionally simple to start; swap in an ML/embedding
// classifier later without changing the interface below.
var codingKeywords = []string{"code", "function", "bug", "python", "javascript", "compile", "api", "script", "debug"}
var securityKeywords = []string{"vulnerability", "exploit", "malware", "hack", "cve", "penetration", "firewall", "attack", "sql injection", "xss", "phishing"}
var translationKeywords = []string{"translate", "translation", "in french", "in tamil", "in hindi", "in spanish"}
var mathKeywords = []string{"solve", "calculate", "equation", "integral", "derivative", "sum of"}
var writingKeywords = []string{"write an essay", "blog post", "write a letter", "write a story", "poem"}

// AnalyzePrompt classifies a raw prompt into a TaskType and a rough
// complexity estimate. v1 uses keyword matching; this is the seam
// where a trained classifier can be swapped in later.
func AnalyzePrompt(prompt string) PromptAnalysis {
	lower := strings.ToLower(prompt)

	task := classify(lower)
	complexity := estimateComplexity(prompt)
	risk := assessRiskTier(task, lower)

	return PromptAnalysis{
		TaskType:   task,
		Complexity: complexity,
		RiskTier:   risk,
	}
}

func classify(lower string) TaskType {
	switch {
	case containsAny(lower, securityKeywords):
		return TaskCybersecurity
	case containsAny(lower, codingKeywords):
		return TaskCoding
	case containsAny(lower, translationKeywords):
		return TaskTranslation
	case containsAny(lower, mathKeywords):
		return TaskMath
	case containsAny(lower, writingKeywords):
		return TaskWriting
	default:
		return TaskGeneral
	}
}

func containsAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

// estimateComplexity uses prompt length as a rough proxy for v1.
// NOTE: length alone is a weak signal (see project notes) - this is
// a placeholder until reasoning-requirement detection is added.
func estimateComplexity(prompt string) string {
	wordCount := len(strings.Fields(prompt))
	switch {
	case wordCount > 80:
		return "high"
	case wordCount > 25:
		return "medium"
	default:
		return "low"
	}
}

// assessRiskTier flags prompts that may warrant constrained execution
// (no autonomous tool use, human review) rather than just picking a
// capable model. Cybersecurity/agentic-sounding prompts default to a
// higher tier - inspired by real incidents where permissive execution
// of security-related agent tasks caused unintended behaviour.
func assessRiskTier(task TaskType, lower string) string {
	agenticSignals := []string{"autonomously", "without asking", "run this on", "execute on the internet", "access the internet"}

	if task == TaskCybersecurity {
		return "high"
	}
	if containsAny(lower, agenticSignals) {
		return "high"
	}
	if task == TaskCoding {
		return "medium"
	}
	return "low"
}
