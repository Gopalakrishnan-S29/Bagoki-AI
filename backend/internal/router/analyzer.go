package router

import "strings"

// TaskType represents the classified category of a user prompt.
type TaskType string

const (
	TaskConversation  TaskType = "conversation"
	TaskCoding        TaskType = "coding"
	TaskCybersecurity TaskType = "cybersecurity"
	TaskTranslation   TaskType = "translation"
	TaskMath          TaskType = "mathematics"
	TaskWriting       TaskType = "writing"
	TaskSummarization TaskType = "summarization"
	TaskComparison    TaskType = "comparison"
	TaskReasoning     TaskType = "reasoning"
	TaskImage         TaskType = "image_generation"
	TaskGeneral       TaskType = "general"
)

// PromptAnalysis is the result of classifying a prompt.
type PromptAnalysis struct {
	TaskType   TaskType
	Complexity string // low | medium | high
	RiskTier   string // low | medium | high
}

var codingKeywords = []string{
	"code",
	"coding",
	"function",
	"program",
	"python",
	"javascript",
	"typescript",
	"golang",
	"go program",
	"java",
	"c++",
	"compile",
	"api",
	"script",
	"debug",
	"debugging",
	"bug",
	"algorithm",
	"developer",
}

var securityKeywords = []string{
	"cybersecurity",
	"cyber security",
	"security",
	"vulnerability",
	"vulnerabilities",
	"exploit",
	"malware",
	"hack",
	"hacking",
	"cve",
	"penetration testing",
	"pentest",
	"firewall",
	"attack",
	"sql injection",
	"xss",
	"cross site scripting",
	"csrf",
	"idor",
	"phishing",
	"owasp",
	"jwt security",
	"authentication security",
}

var translationKeywords = []string{
	"translate",
	"translation",
	"translate this",
	"in french",
	"in tamil",
	"in hindi",
	"in spanish",
	"in german",
}

var mathKeywords = []string{
	"calculate",
	"calculation",
	"solve",
	"equation",
	"integral",
	"derivative",
	"mathematics",
	"math",
	"percentage",
	"probability",
	"multiply",
	"divide",
	"addition",
	"subtraction",
}

var writingKeywords = []string{
	"write",
	"writing",
	"email",
	"e-mail",
	"letter",
	"essay",
	"blog post",
	"story",
	"poem",
	"speech",
	"resume",
	"cover letter",
	"application",
	"rewrite",
	"professional message",
}

var summarizationKeywords = []string{
	"summarize",
	"summarise",
	"summary",
	"summarization",
	"key points",
	"brief summary",
	"shorten this",
	"give me the main points",
}

var comparisonKeywords = []string{
	"compare",
	"comparison",
	"difference between",
	"differences between",
	"versus",
	" vs ",
	"which is better",
	"pros and cons",
}

var reasoningKeywords = []string{
	"analyze",
	"analyse",
	"deep analysis",
	"deeply",
	"architecture",
	"design a system",
	"system design",
	"strategy",
	"reason",
	"reasoning",
	"complex problem",
	"step by step",
	"evaluate",
}

var imageKeywords = []string{
	"create an image",
	"generate an image",
	"generate image",
	"make an image",
	"create a picture",
	"generate a picture",
	"draw an image",
	"draw a picture",
	"image of",
	"picture of",
	"illustration of",
	"logo design",
}

// AnalyzePrompt classifies a raw prompt.
func AnalyzePrompt(prompt string) PromptAnalysis {
	lower := strings.ToLower(strings.TrimSpace(prompt))

	task := classify(lower)
	complexity := estimateComplexity(prompt, task)
	risk := assessRiskTier(task, lower)

	return PromptAnalysis{
		TaskType:   task,
		Complexity: complexity,
		RiskTier:   risk,
	}
}

func classify(lower string) TaskType {
	switch {
	case containsAny(lower, imageKeywords):
		return TaskImage

	case containsAny(lower, securityKeywords):
		return TaskCybersecurity

	case containsAny(lower, codingKeywords):
		return TaskCoding

	case containsAny(lower, translationKeywords):
		return TaskTranslation

	case containsAny(lower, mathKeywords):
		return TaskMath

	case containsAny(lower, summarizationKeywords):
		return TaskSummarization

	case containsAny(lower, comparisonKeywords):
		return TaskComparison

	case containsAny(lower, reasoningKeywords):
		return TaskReasoning

	case containsAny(lower, writingKeywords):
		return TaskWriting

	case isConversation(lower):
		return TaskConversation

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

func isConversation(lower string) bool {
	conversationKeywords := []string{
		"hi",
		"hello",
		"hey",
		"good morning",
		"good afternoon",
		"good evening",
		"how are you",
		"who are you",
		"what can you do",
		"thanks",
		"thank you",
		"goodbye",
		"bye",
	}

	return containsAny(lower, conversationKeywords)
}

// estimateComplexity combines prompt length with task type.
func estimateComplexity(prompt string, task TaskType) string {
	wordCount := len(strings.Fields(prompt))

	if task == TaskReasoning {
		return "high"
	}

	if task == TaskCybersecurity && wordCount > 15 {
		return "high"
	}

	if task == TaskCoding && wordCount > 20 {
		return "high"
	}

	if wordCount > 80 {
		return "high"
	}

	if wordCount > 25 {
		return "medium"
	}

	return "low"
}

func assessRiskTier(task TaskType, lower string) string {
	agenticSignals := []string{
		"autonomously",
		"without asking",
		"run this on",
		"execute on the internet",
		"access the internet",
	}

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
