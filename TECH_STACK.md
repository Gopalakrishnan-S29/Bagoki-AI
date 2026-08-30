# AI Router — Technology Stack

## 1. Project Overview

AI Router is an AI model routing system that analyzes a user's prompt,
determines the task type and complexity, selects a suitable model,
and generates the response through the configured AI provider.

### Current Demo Architecture

User
  ↓
React + TypeScript Frontend
  ↓
Go Backend API
  ↓
Prompt Analyzer
  ↓
Decision Engine
  ↓
Ollama
  ↓
Llama 3.2
  ↓
Response
  ↓
Frontend

The current demonstration uses Ollama for the actual AI response.

No image-generation provider is included in the current architecture.

---

# 2. Backend Technology

## Go

Version:

Go 1.25.0

Purpose:

- Backend API
- HTTP server
- Routing logic
- Authentication
- AI provider integration
- Database communication

---

# 3. Backend Libraries

## pgx

Package:

github.com/jackc/pgx/v5

Version:

5.10.0

Purpose:

- PostgreSQL database driver
- Database connections
- Database queries

Database:

Neon PostgreSQL

---

## golang-jwt

Package:

github.com/golang-jwt/jwt/v5

Version:

5.3.1

Purpose:

- JWT token creation
- JWT token validation
- Authentication

---

## godotenv

Package:

github.com/joho/godotenv

Version:

1.5.1

Purpose:

- Load environment variables from `.env`
- Local development configuration

Important:

The `.env` file must never be committed to Git.

---

## golang.org/x/crypto

Version:

0.55.0

Purpose:

- Cryptographic functionality
- Password security and related security operations

---

# 4. Go Indirect Dependencies

The project also uses the following dependencies indirectly:

- github.com/jackc/pgpassfile
- github.com/jackc/pgservicefile
- github.com/jackc/puddle/v2
- golang.org/x/sync
- golang.org/x/text

These are dependency requirements of the Go packages used by the project.

They normally do not need to be installed manually.

---

# 5. AI / Local Model

## Ollama

Purpose:

- Local AI model execution
- Local model inference
- AI response generation

Ollama server:

http://localhost:11434

Current model:

llama3.2:latest

The current demo uses Ollama instead of requiring a paid
external AI API for the actual response generation.

---

# 6. AI Router

Location:

backend/internal/router/

Responsibilities:

- Analyze user prompts
- Identify task type
- Determine complexity
- Select provider
- Select model
- Provide routing decisions

Current flow:

Prompt
  ↓
Prompt Analysis
  ↓
Task Type
  ↓
Complexity
  ↓
Provider / Model Decision
  ↓
Ollama
  ↓
Llama 3.2
  ↓
Final Response

---

# 7. Frontend Technology

## React

Version:

19.2.8

Purpose:

- User interface
- Component-based development
- Chat interface
- Dynamic UI updates

---

## React DOM

Version:

19.2.8

Purpose:

- Render React components in the browser

---

## TypeScript

Version:

6.0.2

Purpose:

- Type-safe frontend development
- Better code maintainability
- Static type checking

---

## Vite

Version:

8.2.2

Purpose:

- Frontend development server
- Fast development environment
- Hot Module Replacement
- Production build

Development server:

http://localhost:5173

---

# 8. Frontend Development Dependencies

## @vitejs/plugin-react

Version:

6.1.0

Purpose:

- React integration with Vite
- React development and build support

---

## @types/node

Version:

24.13.3

Purpose:

- TypeScript type definitions for Node.js

---

## @types/react

Version:

19.2.18

Purpose:

- TypeScript type definitions for React

---

## @types/react-dom

Version:

19.2.4

Purpose:

- TypeScript type definitions for React DOM

---

# 9. Code Quality

## Oxlint

Version:

1.79.0

Purpose:

- JavaScript and TypeScript linting
- Detect potential coding problems
- Maintain frontend code quality

Run:

npm run lint

---

# 10. Frontend Commands

Install dependencies:

npm install

Start development server:

npm run dev

Build frontend:

npm run build

Run linter:

npm run lint

Preview production build:

npm run preview

---

# 11. Backend Commands

Download Go dependencies:

go mod download

Run tests:

go test ./...

Build backend:

go build -o server.exe ./cmd/server

Run backend:

go run ./cmd/server

---

# 12. Required Software

A team member cloning this project should have:

1. Git
2. Go 1.25.0
3. Node.js
4. npm
5. Ollama
6. PostgreSQL-compatible database access through the configured Neon database

---

# 13. Ollama Setup

Check Ollama:

ollama --version

Check installed models:

ollama list

Install the required model if it is not available:

ollama pull llama3.2:latest

Test the model:

ollama run llama3.2:latest "Say hello in one sentence."

The expected response should come directly from the local Ollama model.

---

# 14. Environment Configuration

Backend environment configuration:

backend/.env

Example structure:

DATABASE_URL=your_database_url
JWT_SECRET=your_jwt_secret
MASTER_ENCRYPTION_KEY=your_encryption_key

Never place real credentials in this documentation.

Never commit:

- API keys
- Database passwords
- JWT secrets
- Encryption keys
- `.env` files
- Personal credentials

---

# 15. Running the Complete Project

## Terminal 1 — Ollama

Make sure Ollama is running.

Verify:

ollama list

---

## Terminal 2 — Backend

cd backend

go run ./cmd/server

Backend:

http://localhost:8080

---

## Terminal 3 — Frontend

cd frontend

npm install

npm run dev

Frontend:

http://localhost:5173

---

# 16. Backend API

Main chat endpoint:

POST

/api/router/chat

Example request:

{
  "prompt": "Explain what a firewall is in simple terms."
}

Example response structure:

{
  "reply": "AI generated response...",
  "provider": "ollama",
  "model": "llama3.2:latest",
  "task_type": "cybersecurity",
  "complexity": "low",
  "reason": "The request concerns firewall",
  "tokens_used": 0,
  "latency_ms": 3491
}

---

# 17. Project Structure

AI Router/

├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── auth/
│   │   ├── database/
│   │   ├── keyvault/
│   │   ├── providers/
│   │   │   └── ollama.go
│   │   └── router/
│   │
│   ├── go.mod
│   └── .env
│
├── frontend/
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── vite.config.*
│
├── frontend-old/
│
├── .gitignore
├── README.md
└── TECH_STACK.md

---

# 18. Current Provider Architecture

The project contains a provider-based architecture.

Current actual response provider:

Ollama

Current model:

llama3.2:latest

The routing architecture is designed so that providers can be
expanded in the future.

For the current demonstration, the actual AI response is generated
locally through Ollama.

---

# 19. Security Notes

Sensitive configuration must remain outside Git.

The following must not be committed:

.env
API keys
Database credentials
JWT secrets
Encryption keys
Private credentials

The `.gitignore` file is configured to ignore `.env`.

Verify with:

git check-ignore -v backend/.env

---

# 20. Verification Checklist

After cloning the repository:

### Backend

go test ./...

Expected:

All backend packages should pass.

### Backend Build

go build -o server.exe ./cmd/server

### Frontend

npm install

npm run build

### Frontend Lint

npm run lint

### Ollama

ollama list

Verify:

llama3.2:latest

### API

Start the backend and test:

POST /api/router/chat

---

# 21. Technology Summary

| Layer | Technology |
|---|---|
| Backend Language | Go 1.25.0 |
| Backend HTTP | Go standard library |
| Database | Neon PostgreSQL |
| PostgreSQL Driver | pgx v5 |
| Authentication | JWT |
| Password/Crypto | x/crypto |
| Environment | godotenv |
| Local AI Runtime | Ollama |
| AI Model | Llama 3.2 |
| Frontend | React 19 |
| Frontend Language | TypeScript 6 |
| Frontend Build Tool | Vite 8 |
| Frontend Linting | Oxlint |
| Version Control | Git |

---

# 22. Important Architecture Note

The current project is intentionally designed so that the AI response
does not depend on a paid OpenAI API key.

The current demonstration flow is:

User
  ↓
React Frontend
  ↓
Go Backend
  ↓
AI Router
  ↓
Ollama
  ↓
Llama 3.2
  ↓
Response

This allows the team to demonstrate the AI routing concept using
a locally running model.