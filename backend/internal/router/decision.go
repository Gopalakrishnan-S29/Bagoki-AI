package router

// ModelChoice is the routing decision returned to the caller.
type ModelChoice struct {
	Provider string
	Model    string
	Reason   string
	RiskTier string
}

// routingTable maps task types to preferred OpenAI models.
//
// For the current demo we use OpenAI as the only provider,
// but Bagoki demonstrates different model-selection decisions.
var routingTable = map[TaskType]ModelChoice{

	TaskConversation: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Fast and efficient for simple conversations",
	},

	TaskGeneral: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Fast model for general questions",
	},

	TaskWriting: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Efficient model for emails, letters and general writing",
	},

	TaskTranslation: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Fast model suitable for translation tasks",
	},

	TaskSummarization: {
		Provider: "openai",
		Model:    "gpt-4.1-mini",
		Reason:   "Efficient model for summarization",
	},

	TaskCoding: {
		Provider: "openai",
		Model:    "gpt-4.1",
		Reason:   "Stronger model selected for programming tasks",
	},

	TaskCybersecurity: {
		Provider: "openai",
		Model:    "gpt-4.1",
		Reason:   "Stronger reasoning capability for cybersecurity analysis",
	},

	TaskMath: {
		Provider: "openai",
		Model:    "gpt-5",
		Reason:   "Stronger reasoning capability for mathematical problems",
	},

	TaskComparison: {
		Provider: "openai",
		Model:    "gpt-4.1",
		Reason:   "Strong model for structured comparison and analysis",
	},

	TaskReasoning: {
		Provider: "openai",
		Model:    "gpt-5",
		Reason:   "Advanced reasoning model for complex analysis",
	},

	TaskImage: {
		Provider: "openai",
		Model:    "gpt-4.1",
		Reason:   "Image-generation requests require an image-capable workflow",
	},
}

// Decide applies the routing table to a PromptAnalysis.
func Decide(analysis PromptAnalysis) ModelChoice {

	choice, ok := routingTable[analysis.TaskType]

	if !ok {
		choice = routingTable[TaskGeneral]
	}

	// Escalate simple models for high-complexity requests.
	if analysis.Complexity == "high" &&
		choice.Provider == "openai" &&
		choice.Model == "gpt-4.1-mini" {

		choice.Model = "gpt-5"
		choice.Reason += " (escalated: high complexity)"
	}

	choice.RiskTier = analysis.RiskTier

	return choice
}
