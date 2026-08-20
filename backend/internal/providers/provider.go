package providers

import "context"

// CallResult is the normalized response every provider adapter returns,
// regardless of which LLM API it wraps.
type CallResult struct {
	Content     string
	TokensUsed  int
	LatencyMs   int64
	ModelUsed   string
	ProviderErr error
}

// Provider is the common interface every LLM adapter must implement.
// This is what lets the router call any provider through one function
// signature: CallModel(ctx, apiKey, model, prompt).
type Provider interface {
	// Name returns the provider identifier, e.g. "openai", "anthropic".
	Name() string

	// CallModel sends the prompt to the given model using the caller's
	// own API key (BYOK) and returns a normalized result.
	CallModel(ctx context.Context, apiKey string, model string, prompt string) (*CallResult, error)
}

// Registry maps provider names to their adapter implementation.
type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}
