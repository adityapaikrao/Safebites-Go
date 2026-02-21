# SafeBites Go Backend — Architecture & Implementation Plan

## Table of Contents

1. [Overview](#overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Current vs New: Side-by-Side](#current-vs-new-side-by-side)
4. [Tech Stack](#tech-stack)
5. [Project Structure](#project-structure)
6. [Database Schema](#database-schema)
7. [API Contract (1:1 with Python backend)](#api-contract)
8. [Agent Architecture (Google ADK + Gemini)](#agent-architecture)
9. [Authentication Flow](#authentication-flow)
10. [Configuration](#configuration)
11. [Implementation Phases](#implementation-phases)
12. [Testing Strategy](#testing-strategy)
13. [Running Locally](#running-locally)

---

## Overview

Rewrite the SafeBites Python/FastAPI backend in **Go** using:

- **Google ADK (Agent Development Kit) for Go** — agent orchestration with Gemini models
- **Gemini 2.0 Flash** — vision (product name extraction) + agent LLM backbone
- **PostgreSQL** — persistent storage via `pgx/v5`
- **chi router** — lightweight, idiomatic HTTP routing
- **Auth0 JWT** — same auth flow as current backend

The frontend (`backendApi.ts`) remains **unchanged** — all API routes, request/response shapes, and status codes are preserved 1:1.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Next.js)                       │
│                    backendApi.ts (unchanged)                     │
└─────────────────────┬───────────────────────────────────────────┘
                      │  HTTP (JSON / multipart)
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                     API Gateway (chi router)                    │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐  ┌──────────────┐  │
│  │  CORS    │→ │  Logger  │→ │  Auth/JWT  │→ │   Handlers   │  │
│  │Middleware│  │Middleware │  │ Middleware │  │              │  │
│  └──────────┘  └──────────┘  └───────────┘  └──────┬───────┘  │
└─────────────────────────────────────────────────────┼──────────┘
                                                      │
                      ┌───────────────────────────────┼──────────┐
                      │                               ▼          │
                      │  ┌─────────────────────────────────────┐ │
                      │  │         Service Layer               │ │
                      │  │  (business logic, orchestration)    │ │
                      │  └────────┬──────────────┬─────────────┘ │
                      │           │              │               │
                      │           ▼              ▼               │
                      │  ┌──────────────┐  ┌──────────────────┐ │
                      │  │  Repository  │  │  Agent Layer     │ │
                      │  │  (pgx/v5)    │  │  (Google ADK)    │ │
                      │  │              │  │                  │ │
                      │  │ • UserRepo   │  │ • Vision         │ │
                      │  │ • ScanRepo   │  │   (Gemini Flash) │ │
                      │  │ • FavRepo    │  │ • SearchAgent    │ │
                      │  │              │  │   (Gemini+GSearch)│ │
                      │  └──────┬───────┘  │ • ScorerAgent    │ │
                      │         │          │   (Gemini)       │ │
                      │         ▼          │ • RecommendAgent │ │
                      │  ┌──────────────┐  │   (Gemini+GSearch)│ │
                      │  │  PostgreSQL   │  └──────────────────┘ │
                      │  │  (pgx pool)   │                       │
                      │  └──────────────┘                        │
                      │                  Internal                 │
                      └──────────────────────────────────────────┘
```

---

## Current vs New: Side-by-Side

| Concern              | Python Backend                        | Go Backend                                |
|----------------------|---------------------------------------|-------------------------------------------|
| Framework            | FastAPI                               | chi router (stdlib `net/http` compatible) |
| Language             | Python 3.11+                          | Go 1.22+                                  |
| Database driver      | SQLAlchemy ORM                        | pgx/v5 (raw SQL, no ORM)                 |
| Database             | SQLite (dev) / PostgreSQL (prod)      | PostgreSQL always (Docker for local)      |
| AI SDK               | OpenAI Agents SDK                     | Google ADK for Go                         |
| LLM models           | GPT-4.1-mini, GPT-5-mini             | Gemini 2.0 Flash (all agents)            |
| Image → product name | Gemini 2.0 Flash (google-genai)       | Gemini 2.0 Flash (genai Go SDK)          |
| Web search           | OpenAI WebSearchTool                  | Gemini Google Search grounding            |
| Auth                 | python-jose (JWT)                     | golang-jwt/jwt/v5                         |
| Config               | pydantic + dotenv                     | env vars + `.env` file                    |
| Testing              | pytest                                | go test + testify + pgxmock               |
| Migrations           | SQLAlchemy auto-create                | golang-migrate (versioned SQL files)      |

---

## Tech Stack

| Library                              | Purpose                                | Version  |
|--------------------------------------|----------------------------------------|----------|
| `github.com/go-chi/chi/v5`          | HTTP router & middleware               | v5       |
| `github.com/go-chi/cors`            | CORS middleware                        | latest   |
| `github.com/jackc/pgx/v5`           | PostgreSQL driver & connection pool    | v5       |
| `github.com/golang-jwt/jwt/v5`      | JWT parsing & validation               | v5       |
| `github.com/google/genai`           | Google Gen AI SDK (Gemini calls)       | latest   |
| `github.com/joho/godotenv`          | Load `.env` file                       | latest   |
| `github.com/golang-migrate/migrate/v4` | Database migrations                 | v4       |
| `github.com/stretchr/testify`       | Test assertions & mocks                | latest   |
| `github.com/pashagolub/pgxmock/v4`  | pgx mock for repository tests          | v4       |

> **Note on Google ADK for Go**: If `github.com/google/adk-go` is stable at implementation time, we use it for agent orchestration. Otherwise, we implement a lightweight agent abstraction over the `github.com/google/genai` SDK directly (same Gemini models, Google Search grounding, structured output via JSON schema). Either way, the agent layer is isolated behind interfaces and swappable.

---

## Project Structure

```
backend-go/
├── cmd/
│   └── server/
│       └── main.go                  # Entry point: config → DB → agents → router → serve
│
├── internal/
│   ├── config/
│   │   └── config.go                # Env-based configuration struct
│   │
│   ├── middleware/
│   │   ├── auth.go                  # Auth0 JWT verification (+ dev bypass)
│   │   ├── cors.go                  # CORS settings
│   │   └── logging.go               # Request/response logging
│   │
│   ├── handler/
│   │   ├── analyze.go               # POST /api/analyze
│   │   ├── recommend.go             # GET  /api/reccomendations/:name/:score
│   │   ├── user.go                  # GET/POST /api/users, /api/users/:id, /me
│   │   ├── preference.go            # POST /api/users/:id/preferences
│   │   ├── scan.go                  # GET/POST /api/users/:id/scans, stats
│   │   ├── favorite.go              # GET/POST/DELETE /api/users/:id/favorites
│   │   ├── template.go              # GET /api/dietary-templates, POST apply-template
│   │   └── health.go                # GET /
│   │
│   ├── model/
│   │   ├── user.go                  # User, UserPreferences structs
│   │   ├── scan.go                  # Scan struct
│   │   ├── favorite.go              # Favorite struct
│   │   ├── template.go              # DietaryTemplate definitions (static data)
│   │   └── agent.go                 # WebSearchResult, ScorerResult, RecommenderResult
│   │
│   ├── repository/
│   │   ├── interfaces.go            # UserRepository, ScanRepository, FavoriteRepository interfaces
│   │   ├── postgres.go              # NewPostgresDB, pool management
│   │   ├── user_repo.go             # UserRepository implementation
│   │   ├── scan_repo.go             # ScanRepository implementation
│   │   └── favorite_repo.go         # FavoriteRepository implementation
│   │
│   ├── service/
│   │   ├── analyze_service.go       # Orchestrates vision → search → scorer pipeline
│   │   ├── recommend_service.go     # Orchestrates recommender agent
│   │   ├── user_service.go          # User CRUD + preference logic
│   │   └── scan_service.go          # Scan + stats logic
│   │
│   └── agent/
│       ├── client.go                # Gemini client initialization
│       ├── vision.go                # ExtractProductName (Gemini Vision)
│       ├── search.go                # WebSearchAgent (Gemini + Google Search grounding)
│       ├── scorer.go                # ScorerAgent (Gemini structured output)
│       ├── recommender.go           # RecommenderAgent (Gemini + Google Search)
│       └── prompts.go               # All system prompts (ported from Python)
│
├── migrations/
│   ├── 001_create_users.up.sql
│   ├── 001_create_users.down.sql
│   ├── 002_create_scans.up.sql
│   ├── 002_create_scans.down.sql
│   ├── 003_create_favorites.up.sql
│   └── 003_create_favorites.down.sql
│
├── go.mod
├── go.sum
├── Makefile                         # build, run, test, migrate commands
├── Dockerfile
├── docker-compose.yml               # PostgreSQL + app for local dev
├── .env.example
├── ARCHITECTURE.md                  # This file
└── README.md
```

---

## Database Schema

### Table: `users`

```sql
CREATE TABLE users (
    id                TEXT PRIMARY KEY,          -- Auth0 sub (e.g. "auth0|abc123")
    email             TEXT NOT NULL,
    name              TEXT,
    picture           TEXT,
    allergies         JSONB DEFAULT '[]'::jsonb,
    diet_goals        JSONB DEFAULT '[]'::jsonb,
    avoid_ingredients JSONB DEFAULT '[]'::jsonb,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);
```

### Table: `scans`

```sql
CREATE TABLE scans (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_name  TEXT NOT NULL,
    brand         TEXT DEFAULT '',
    image         TEXT,                           -- base64 or URL
    safety_score  INTEGER NOT NULL,
    is_safe       BOOLEAN NOT NULL,
    ingredients   JSONB DEFAULT '[]'::jsonb,
    timestamp     TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_scans_user_id ON scans(user_id);
CREATE INDEX idx_scans_timestamp ON scans(timestamp DESC);
```

### Table: `favorites`

```sql
CREATE TABLE favorites (
    id            SERIAL PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_name  TEXT NOT NULL,
    brand         TEXT DEFAULT '',
    safety_score  INTEGER,
    image         TEXT,
    added_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, product_name)
);
CREATE INDEX idx_favorites_user_id ON favorites(user_id);
```

---

## API Contract

All routes match the existing Python backend exactly. The frontend `backendApi.ts` requires **zero changes**.

### Health

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET    | `/`  | No   | Health check → `{ message, status, version }` |

### Product Analysis

| Method | Path | Auth | Request | Response |
|--------|------|------|---------|----------|
| POST   | `/api/analyze` | Optional (Bearer) | `multipart/form-data` with `image` file | `{ status, product_name, scoring_data }` |
| GET    | `/api/reccomendations/{product_name}/{overall_score}` | No | Path params | `{ status, reccomender_data }` |

### Users

| Method | Path | Auth | Request | Response |
|--------|------|------|---------|----------|
| GET    | `/api/users/me` | Required | — | `{ user }` |
| GET    | `/api/users/{user_id}` | No | — | `{ user }` |
| POST   | `/api/users` | No | JSON body: `{ id, email, name, picture?, allergies?, dietGoals?, avoidIngredients? }` | `{ user }` |
| POST   | `/api/users/{user_id}/preferences` | No | JSON body: `{ allergies?, dietGoals?, avoidIngredients? }` | `{ user }` |

### Dietary Templates

| Method | Path | Auth | Request | Response |
|--------|------|------|---------|----------|
| GET    | `/api/dietary-templates` | No | — | `{ templates }` |
| POST   | `/api/users/{user_id}/apply-template/{template_key}` | No | — | `{ user, template }` |

### Scans

| Method | Path | Auth | Request | Response |
|--------|------|------|---------|----------|
| GET    | `/api/users/{user_id}/scans` | No | Query: `limit` (optional) | `{ scans }` |
| POST   | `/api/users/{user_id}/scans` | No | JSON body: scan data | `{ scan, status }` |
| GET    | `/api/users/{user_id}/stats` | No | — | `{ stats }` |

### Favorites

| Method | Path | Auth | Request | Response |
|--------|------|------|---------|----------|
| GET    | `/api/users/{user_id}/favorites` | No | — | `{ favorites }` |
| POST   | `/api/users/{user_id}/favorites` | No | JSON body: `{ productName, brand?, safetyScore?, image? }` | `{ favorite, status }` |
| DELETE | `/api/users/{user_id}/favorites/{favorite_id}` | No | — | `{ status }` |
| GET    | `/api/users/{user_id}/favorites/check/{product_name}` | No | — | `{ isFavorite }` |

### Response JSON Field Naming

All JSON responses use **camelCase** keys to match the existing frontend expectations:

```json
// User
{ "id", "email", "name", "picture", "allergies", "dietGoals", "avoidIngredients", "createdAt" }

// Scan
{ "id", "productName", "brand", "image", "safetyScore", "isSafe", "ingredients", "timestamp" }

// Favorite
{ "id", "productName", "brand", "safetyScore", "image", "addedAt" }

// Stats
{ "totalScans", "todayScans", "safeToday", "riskyToday", "averageScore" }
```

---

## Agent Architecture (Google ADK + Gemini)

The current Python backend uses 3 OpenAI-based agents. The Go backend replaces them with Gemini-powered equivalents using Google's Gen AI SDK.

### Analysis Pipeline Flow (same as Python)

```
Image bytes
    │
    ▼
┌──────────────────────┐
│  Gemini Vision       │  Extract product name from image
│  (gemini-2.0-flash)  │  Input: image bytes
│                      │  Output: product name string
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Search Agent        │  Find ingredient list via web search
│  (gemini-2.0-flash   │  Input: product name
│   + Google Search    │  Output: WebSearchResult
│     grounding)       │    { List_of_ingredients: [{name, description}] }
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Scorer Agent        │  Score each ingredient for safety
│  (gemini-2.0-flash)  │  Input: ingredients JSON + user preferences
│                      │  Output: ScorerResult
│                      │    { ingredient_scores: [{name, score, reasoning}],
│                      │      overall_score: float }
└──────────┬───────────┘
           │
           ▼
    Return to client
```

### Recommendation Pipeline (separate endpoint)

```
Product name + overall score
    │
    ▼
┌──────────────────────┐
│  Recommender Agent   │  Suggest healthier alternatives
│  (gemini-2.0-flash   │  Input: product name + score
│   + Google Search    │  Output: RecommenderResult
│     grounding)       │    { recommendations: [{product_name, health_score, reason}] }
└──────────┬───────────┘
           │
           ▼
    Return to client
```

### Agent Implementation Approach

Each agent is a Go function that:
1. Constructs a prompt (system + user input)
2. Calls `genai.Client.Models.GenerateContent()` with the appropriate model
3. Uses **Google Search grounding** tool config where web search is needed
4. Parses structured JSON output into typed Go structs
5. Returns the result

```go
// Pseudo-code for Search Agent
func (a *SearchAgent) Run(ctx context.Context, productName string) (*model.WebSearchResult, error) {
    resp, err := a.client.Models.GenerateContent(ctx, "gemini-2.0-flash", 
        genai.Text(productName),
        &genai.GenerateContentConfig{
            SystemInstruction: genai.NewContentFromText(searchPrompt, "user"),
            Tools: []*genai.Tool{{GoogleSearch: &genai.GoogleSearch{}}},
            ResponseMIMEType: "application/json",
            ResponseSchema: webSearchResultSchema,
        },
    )
    // parse resp.Text() into WebSearchResult
}
```

### Agent Output Structs (Go)

```go
// WebSearchResult — matches Python WebSearchResult
type Ingredient struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}
type WebSearchResult struct {
    ListOfIngredients []Ingredient `json:"List_of_ingredients"`
}

// ScorerResult — matches Python ScorerResult
type IngredientScore struct {
    IngredientName string `json:"ingredient_name"`
    SafetyScore    string `json:"safety_score"` // "LOW", "MEDIUM", "HIGH"
    Reasoning      string `json:"reasoning"`
}
type ScorerResult struct {
    IngredientScores []IngredientScore `json:"ingredient_scores"`
    OverallScore     float64           `json:"overall_score"`
}

// RecommenderResult — matches Python ReccomenderResult
type Recommendation struct {
    ProductName string `json:"product_name"`
    HealthScore string `json:"health_score"`
    Reason      string `json:"reason"`
}
type RecommenderResult struct {
    Recommendations []Recommendation `json:"recommendations"`
}
```

---

## Authentication Flow

Same as current Python backend. Auth0 JWT verification with dev-mode bypass.

```
Request with `Authorization: Bearer <token>`
    │
    ▼
┌─────────────────────────┐
│  Auth Middleware         │
│                         │
│  1. Extract Bearer token│
│  2. If DEV_MODE && no   │
│     token → pass through│
│     (optional user)     │
│  3. Fetch JWKS from     │
│     Auth0 (cached)      │
│  4. Verify JWT signature│
│     + claims            │
│  5. Extract `sub` claim │
│     as user_id          │
│  6. Set user_id in      │
│     request context     │
└─────────────────────────┘
```

Two middleware variants (matching Python):

- **`RequireAuth`** — returns 401 if no valid token; dev fallback to `"dev-test-user"`
- **`OptionalAuth`** — extracts user_id if token present, nil otherwise

---

## Configuration

All config from environment variables (loaded from `.env` in dev):

```bash
# .env.example
PORT=8080
ENV=development                          # "development" | "production"

# Database
DATABASE_URL=postgres://safebites:safebites@localhost:5432/safebites?sslmode=disable

# Google AI
GOOGLE_API_KEY=your-gemini-api-key

# Auth0
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_API_AUDIENCE=https://your-api-audience

# CORS
CORS_ORIGINS=http://localhost:3000
```

Go config struct:

```go
type Config struct {
    Port             string   // default "8080"
    Env              string   // default "development"
    DatabaseURL      string   // required
    GoogleAPIKey     string   // required
    Auth0Domain      string   // empty = dev mode
    Auth0APIAudience string
    CORSOrigins      []string // comma-separated
}
```

---

## Implementation Phases

### Phase 1: Project Scaffolding & Database
> **Goal**: Working server that connects to Postgres and runs migrations

- [ ] Initialize Go module, install dependencies
- [ ] Create `cmd/server/main.go` entry point
- [ ] Create `internal/config/config.go`
- [ ] Create `docker-compose.yml` (PostgreSQL)
- [ ] Create SQL migration files (users, scans, favorites)
- [ ] Create `internal/repository/postgres.go` (connection pool + migrate)
- [ ] Create `Makefile` with build/run/migrate targets
- [ ] **Test**: Verify DB connects, tables created

### Phase 2: Models & Repository Layer
> **Goal**: Full CRUD for users, scans, favorites against real Postgres

- [ ] Create `internal/model/` structs (User, Scan, Favorite, Template, Agent models)
- [ ] Create `internal/repository/interfaces.go` (repository interfaces)
- [ ] Implement `internal/repository/user_repo.go`
- [ ] Implement `internal/repository/scan_repo.go`
- [ ] Implement `internal/repository/favorite_repo.go`
- [ ] **Test**: Table-driven tests for each repo method with pgxmock

### Phase 3: HTTP Layer (Handlers + Middleware)
> **Goal**: All REST endpoints wired up, responding with correct JSON shapes

- [ ] Create `internal/middleware/cors.go`
- [ ] Create `internal/middleware/logging.go`
- [ ] Create `internal/middleware/auth.go` (JWT verification + dev bypass)
- [ ] Create `internal/handler/health.go`
- [ ] Create `internal/handler/user.go`
- [ ] Create `internal/handler/preference.go`
- [ ] Create `internal/handler/template.go`
- [ ] Create `internal/handler/scan.go`
- [ ] Create `internal/handler/favorite.go`
- [ ] Wire up router in `main.go`
- [ ] **Test**: Table-driven handler tests with httptest + mock repos

### Phase 4: Agent Layer (Google ADK + Gemini)
> **Goal**: All three AI agents working with Gemini models

- [ ] Create `internal/agent/client.go` (Gemini client init)
- [ ] Create `internal/agent/prompts.go` (port all system prompts)
- [ ] Create `internal/agent/vision.go` (product name extraction)
- [ ] Create `internal/agent/search.go` (web search agent)
- [ ] Create `internal/agent/scorer.go` (scorer agent)
- [ ] Create `internal/agent/recommender.go` (recommender agent)
- [ ] **Test**: Mock Gemini client, verify prompt construction & response parsing

### Phase 5: Service Layer & Integration
> **Goal**: Analysis pipeline wired end-to-end; analyze endpoint works

- [ ] Create `internal/service/analyze_service.go` (vision → search → scorer)
- [ ] Create `internal/service/recommend_service.go`
- [ ] Create `internal/service/user_service.go`
- [ ] Create `internal/service/scan_service.go`
- [ ] Create `internal/handler/analyze.go` (POST /api/analyze)
- [ ] Create `internal/handler/recommend.go` (GET /api/reccomendations)
- [ ] **Test**: Integration tests with real DB (docker-compose) + mocked agents

### Phase 6: Docker, Docs & Polish
> **Goal**: Production-ready container, README, env example

- [ ] Create `Dockerfile` (multi-stage build)
- [ ] Update `docker-compose.yml` to include app service
- [ ] Create `README.md` with setup/run instructions
- [ ] Create `.env.example`
- [ ] End-to-end manual test against frontend

---

## Testing Strategy

### Test Types

| Type | Tool | Scope | When |
|------|------|-------|------|
| **Unit (table tests)** | `go test` + `testify` | Individual functions, model methods | Every phase |
| **Repository mock tests** | `pgxmock` | SQL queries without real DB | Phase 2 |
| **Handler tests** | `httptest` + mock services | HTTP request/response, status codes, JSON shape | Phase 3 |
| **Agent mock tests** | Custom mock `genai.Client` | Prompt construction, response parsing | Phase 4 |
| **Integration tests** | `go test` + Docker Postgres | Full stack with real DB, mocked agents | Phase 5 |

### Test Conventions

```
internal/
  repository/
    user_repo.go
    user_repo_test.go        ← pgxmock-based tests
  handler/
    user.go
    user_test.go             ← httptest-based tests
  agent/
    search.go
    search_test.go           ← mock client tests
  service/
    analyze_service.go
    analyze_service_test.go  ← integration tests
```

### Table Test Example

```go
func TestUserRepo_GetUser(t *testing.T) {
    tests := []struct {
        name     string
        userID   string
        mockSetup func(mock pgxmock.PgxPoolIface)
        want     *model.User
        wantErr  bool
    }{
        {
            name:   "existing user",
            userID: "auth0|123",
            mockSetup: func(mock pgxmock.PgxPoolIface) {
                rows := pgxmock.NewRows([]string{"id","email","name",...}).
                    AddRow("auth0|123", "test@test.com", "Test", ...)
                mock.ExpectQuery("SELECT").WithArgs("auth0|123").WillReturnRows(rows)
            },
            want: &model.User{ID: "auth0|123", Email: "test@test.com", ...},
        },
        {
            name:   "user not found",
            userID: "auth0|999",
            mockSetup: func(mock pgxmock.PgxPoolIface) {
                mock.ExpectQuery("SELECT").WithArgs("auth0|999").
                    WillReturnError(pgx.ErrNoRows)
            },
            want:    nil,
            wantErr: false, // Not found is not an error, returns nil
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock, _ := pgxmock.NewPool()
            defer mock.Close()
            tt.mockSetup(mock)
            repo := repository.NewUserRepo(mock)
            got, err := repo.GetUser(context.Background(), tt.userID)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Interfaces for Testability

All major components expose interfaces so handlers/services can be tested with mocks:

```go
// Repository interfaces
type UserRepository interface {
    GetUser(ctx context.Context, id string) (*model.User, error)
    CreateOrUpdateUser(ctx context.Context, user *model.User) (*model.User, error)
    UpdatePreferences(ctx context.Context, id string, prefs *model.UserPreferences) (*model.User, error)
    ApplyTemplate(ctx context.Context, id string, templateKey string) (*model.User, error)
}

type ScanRepository interface {
    GetUserScans(ctx context.Context, userID string, limit *int) ([]model.Scan, error)
    AddScan(ctx context.Context, userID string, scan *model.Scan) (*model.Scan, error)
    GetUserStats(ctx context.Context, userID string) (*model.UserStats, error)
}

type FavoriteRepository interface {
    GetFavorites(ctx context.Context, userID string) ([]model.Favorite, error)
    AddFavorite(ctx context.Context, userID string, fav *model.Favorite) (*model.Favorite, error)
    RemoveFavorite(ctx context.Context, userID string, favID int) error
    IsFavorite(ctx context.Context, userID string, productName string) (bool, error)
}

// Agent interfaces
type VisionAgent interface {
    ExtractProductName(ctx context.Context, imageBytes []byte) (string, error)
}

type SearchAgent interface {
    Search(ctx context.Context, productName string) (*model.WebSearchResult, error)
}

type ScorerAgent interface {
    Score(ctx context.Context, ingredients string, prefs *model.UserPreferences) (*model.ScorerResult, error)
}

type RecommenderAgent interface {
    Recommend(ctx context.Context, productName string, score float64) (*model.RecommenderResult, error)
}
```

---

## Running Locally

```bash
# 1. Start PostgreSQL
docker-compose up -d postgres

# 2. Run migrations
make migrate-up

# 3. Set env vars
cp .env.example .env
# Edit .env with your GOOGLE_API_KEY

# 4. Run server
make run
# → Server listening on :8080

# 5. Run tests
make test

# 6. Run with Docker (full stack)
docker-compose up --build
```

### Makefile Targets

```makefile
build:          go build -o bin/server ./cmd/server
run:            go run ./cmd/server
test:           go test ./... -v -count=1
test-cover:     go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
migrate-up:     migrate -path migrations -database "$(DATABASE_URL)" up
migrate-down:   migrate -path migrations -database "$(DATABASE_URL)" down
lint:           golangci-lint run
docker-build:   docker build -t safebites-backend .
```
