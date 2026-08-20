package router

// ModelChoice is the routing decision returned to the caller: which
// provider/model to call, and how much autonomy to allow it.
type ModelChoice struct {
	Provider   string
	Model      string
	Reason     string
	RiskTier   string
}

// routingTable maps task type to a preferred provider+model.
// This mirrors the "Example Routing Rules" table from the project plan.
// Extend or replace with a learned/adaptive router later.
var routingTable = map[TaskType]ModelChoice{
	TaskCoding: {
    Provider: "openai",
    Model:    "gpt-4.1-mini",
    Reason:   "Fast and capable for coding tasks",
},
	TaskCybersecurity: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Fast and capable for cybersecurity tasks",
	},
	TaskTranslation: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Fast and inexpensive for language tasks",
	},
	TaskMath: {
		Provider: "openai",
		Model:    "gpt-5",
		Reason:   "Reliable multi-step reasoning",
	},
	TaskWriting: {
		Provider: "openai",
		Model:    "gpt-5",
		Reason:   "Strong creative and long-form writing",
	},
	TaskGeneral: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Fast, cheap default for simple queries",
	},
}

// Decide applies the routing table to a PromptAnalysis and returns the
// chosen provider/model. It also carries the assessed risk tier so the
// execution layer can decide how much autonomy/tool access to grant.
func Decide(analysis PromptAnalysis) ModelChoice {
	choice, ok := routingTable[analysis.TaskType]
	if !ok {
		choice = routingTable[TaskGeneral]
	}

	// Escalate to a stronger model when complexity is high, regardless
	// of task type - a simple v1 cost/capability tradeoff.
	if analysis.Complexity == "high" && choice.Provider == "openai" && choice.Model == "gpt-4.1-mini" {
		choice.Model = "gpt-5"
		choice.Reason += " (escalated: high complexity)"
	}

	choice.RiskTier = analysis.RiskTier
	return choice
}
