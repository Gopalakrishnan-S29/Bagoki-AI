# AI Model Router

Intelligent multi-LLM orchestration platform. Users bring their own provider
API keys (BYOK); the router analyzes each prompt and picks the best model.

## Stack
- Backend: Go
- Frontend: Next.js (React)
- Database: PostgreSQL (Neon free tier)

## Project layout
```
backend/
  cmd/server/main.go       entry point, HTTP routes
  internal/
    auth/                  JWT + password hashing (TODO)
    router/                prompt analyzer + decision engine  <- core logic
    providers/              LLM adapters (OpenAI done, others TODO)
    keyvault/               AES-256-GCM encryption for stored API keys
    usage/                  usage logging (TODO)
    db/                     database access layer (TODO)
  migrations/001_init.sql  PostgreSQL schema
frontend/                  Next.js app (TODO)
```

## Setup (backend)
1. Install Go 1.22+
2. `cd backend && go mod tidy`
3. Copy `.env.example` to `.env` and fill in:
   - `DATABASE_URL` from your Neon project
   - `MASTER_ENCRYPTION_KEY`: generate with `openssl rand -base64 32`
4. Run the migration in `migrations/001_init.sql` against your Neon database
   (via Neon's SQL editor or `psql $DATABASE_URL -f migrations/001_init.sql`)
5. `go run cmd/server/main.go`
6. Test: `curl http://localhost:8080/health`

## Test the chat endpoint (before auth is wired in)
```bash
curl -X POST http://localhost:8080/api/chat \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Write a Python function to reverse a string", "api_key": "sk-..."}'
```
This should route to the coding task type and call OpenAI/whichever
provider is configured, returning the model's reply plus routing metadata
(task_type, risk_tier, reason).

## Build order (recommended)
1. ✅ Router core (prompt analyzer + decision engine) - no dependencies, testable standalone
2. ✅ OpenAI provider adapter
3. ✅ Key vault (encryption)
4. ⬜ Database layer (users, api_keys, conversations, messages tables)
5. ⬜ Auth (signup/login, JWT middleware)
6. ⬜ Wire key vault into /api/chat (fetch + decrypt user's stored key instead of accepting it in the request body)
7. ⬜ Usage logging
8. ⬜ Additional provider adapters (Anthropic, Gemini)
9. ⬜ Frontend: auth pages, chat UI, settings/API key page

## Notes
- The routing table in `internal/router/decision.go` is rule-based (v1).
  This is the seam to swap in a learned/adaptive router later.
- `risk_tier` on each routing decision is a placeholder for constrained
  execution (limited tool access, human review) on higher-risk prompts -
  not yet enforced, just classified.
