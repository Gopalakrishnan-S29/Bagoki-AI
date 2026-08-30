package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaDecisionEngine uses a local Ollama model to decide
// how the AI Router should handle a user request.
type OllamaDecisionEngine struct {
	httpClient *http.Client
	baseURL    string
	model      string
}

// NewOllamaDecisionEngine creates the local Ollama decision engine.
func NewOllamaDecisionEngine() *OllamaDecisionEngine {
	return &OllamaDecisionEngine{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "http://localhost:11434",
		model:   "llama3.2:latest",
	}
}

// ollamaGenerateRequest represents a request sent to Ollama.
type ollamaGenerateRequest struct {
	Model  string          `json:"model"`
	Prompt string          `json:"prompt"`
	Stream bool            `json:"stream"`
	Format json.RawMessage `json:"format,omitempty"`
}

// ollamaGenerateResponse represents Ollama's response.
type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// OllamaRoutingDecision is the routing decision returned by Ollama.
type OllamaRoutingDecision struct {
	TaskType   string `json:"task_type"`
	Complexity string `json:"complexity"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Reason     string `json:"reason"`
}

// Decide asks Ollama to classify the user request and select
// the appropriate downstream model/provider.
func (e *OllamaDecisionEngine) Decide(
	ctx context.Context,
	userPrompt string,
) (*OllamaRoutingDecision, error) {

	systemPrompt := `You are the routing decision engine for an AI system.

Your ONLY job is to analyze the user's request and return a routing decision.

You MUST NOT answer, explain, solve, or rewrite the user's request.

Return ONLY one valid JSON object.

TASK CLASSIFICATION RULES:

1. cybersecurity

Use cybersecurity when the SUBJECT or PURPOSE of the request involves:

- cybersecurity
- cyber security
- vulnerability
- exploit or exploitation
- malware
- hacking
- penetration testing
- pentesting
- CVE
- firewall
- attack
- SQL injection
- XSS
- CSRF
- IDOR
- phishing
- OWASP
- authentication security
- JWT security
- network security
- web application security
- vulnerability analysis

IMPORTANT:

If a request contains a specific cybersecurity concept such as
"SQL injection", "XSS", "CSRF", "IDOR", "CVE", "malware",
or "penetration testing", classify it as cybersecurity even
if the user asks to "explain", "summarize", "write", or
"describe" it.

2. coding

Use coding when the main purpose is:

- programming
- source code
- debugging
- software development
- APIs
- algorithms
- scripts

3. image_generation

Use image_generation when the user asks to:

- create an image
- generate an image
- draw an image
- create a picture
- generate a picture
- create an illustration
- design a logo

For image_generation:

provider MUST be openai.

model MUST be image_generation.

4. mathematics

Use mathematics for:

- calculations
- equations
- mathematical problems
- probability
- derivatives
- integrals

5. translation

Use translation when the user asks to translate content.

6. summarization

Use summarization when the main purpose is summarizing
provided content.

7. comparison

Use comparison when the main purpose is comparing two or
more things.

8. reasoning

Use reasoning for:

- complex analysis
- architecture
- system design
- strategy
- evaluation
- difficult multi-step problems

9. writing

Use writing for:

- emails
- letters
- essays
- stories
- poems
- resumes
- professional messages
- general content creation

10. conversation

Use conversation for:

- greetings
- casual conversation
- simple personal interaction

11. general

Use general when none of the above categories clearly applies.

CLASSIFICATION PRIORITY:

Classify based on the SUBJECT and PURPOSE of the request,
not merely individual words.

Examples:

"Explain SQL injection in simple terms."
=> cybersecurity

"Write a Python script to scan a website for SQL injection."
=> cybersecurity

"Write a Python program that calculates factorial."
=> coding

"Generate an image of a cybersecurity SOC."
=> image_generation

"Translate this paragraph into Tamil."
=> translation

"Compare PostgreSQL and MySQL."
=> comparison

COMPLEXITY:

low = simple, direct request

medium = moderately detailed request

high = complex, multi-step, advanced, or demanding request

PROVIDER:

Choose exactly one:

ollama
openai

Use ollama when the local model can reasonably handle
the request.

Use openai when stronger capabilities are needed.

For image_generation ALWAYS use:

provider = openai

MODEL:

For local Ollama use exactly:

llama3.2:latest

For OpenAI text requests choose only from:

gpt-4.1-mini
gpt-4.1
gpt-5

For image_generation use exactly:

image_generation

Do NOT invent model names.

IMPORTANT:

Never return models such as:

"DALL-E 2: latest"
"dall-e"
"gpt-4"
"unknown"
or any model not listed above.

Return this exact JSON structure:

{
  "task_type": "cybersecurity",
  "complexity": "low",
  "provider": "ollama",
  "model": "llama3.2:latest",
  "reason": "The request concerns SQL injection"
}

USER REQUEST:
` + userPrompt

	reqBody := ollamaGenerateRequest{
		Model:  e.model,
		Prompt: systemPrompt,
		Stream: false,
		Format: json.RawMessage(`"json"`),
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal ollama decision request: %w",
			err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		e.baseURL+"/api/generate",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"build ollama decision request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := e.httpClient.Do(req)
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
			strings.TrimSpace(string(respBytes)),
		)
	}

	var parsed ollamaGenerateResponse

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

	var decision OllamaRoutingDecision

	if err := json.Unmarshal(
		[]byte(strings.TrimSpace(parsed.Response)),
		&decision,
	); err != nil {
		return nil, fmt.Errorf(
			"parse ollama decision JSON: %w; raw response: %s",
			err,
			parsed.Response,
		)
	}

	return &decision, nil
}
