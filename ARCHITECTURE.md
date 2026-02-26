# SafeBites Go Backend — Architecture

## Table of Contents

1. [System Overview](#system-overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Layer-by-Layer Breakdown](#layer-by-layer-breakdown)
4. [AI Agent Pipeline](#ai-agent-pipeline)
5. [Database Design](#database-design)
6. [Design Decisions](#design-decisions)
7. [Go Patterns Used](#go-patterns-used)
8. [Python-to-Go Migration](#python-to-go-migration)
9. [Testing Strategy](#testing-strategy)
10. [Future Additions](#future-additions)

---

## System Overview

SafeBites is a four-layer Go backend that orchestrates a multi-agent AI pipeline for food product safety analysis. A user uploads a product photo, and the system chains four AI agents — Vision OCR, Web Search, Ingredient Scorer, and Recommender — to return a personalized safety assessment and healthier alternatives.

The backend exposes 18 REST endpoints, persists user data and scan history in PostgreSQL, and authenticates via Auth0 JWTs. The AI pipeline is built on top of Google's Agent Development Kit (ADK) for Go, using Gemini 2.5 Flash as the backbone model with Google Search grounding for real-time ingredient lookups.

---

## Architecture Diagram

### Full System Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         Frontend (Next.js)                               │
│                      backendApi.ts (unchanged)                           │
└───────────────────────────┬──────────────────────────────────────────────┘
                            │  HTTP (JSON / multipart)
                            ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                        chi Router + Middleware                            │
│                                                                          │
│   RequestID ──▶ Logging ──▶ Recoverer ──▶ CORS ──▶ OptionalAuth         │
│                                                         │                │
│                                                         ▼                │
│                              ┌──────────────────────────────────┐        │
│                              │          HTTP Handlers           │        │
│                              │                                  │        │
│                              │  analyze.go    recommend.go      │        │
│                              │  user.go       scan.go           │        │
│                              │  favorite.go   template.go       │        │
│                              │  health.go     docs.go           │        │
│                              └──────────┬───────────────────────┘        │
└─────────────────────────────────────────┼────────────────────────────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    ▼                     ▼                     ▼
        ┌───────────────────┐  ┌──────────────────┐  ┌─────────────────┐
        │   Service Layer   │  │   Agent Layer    │  │ Repository Layer│
        │                   │  │                  │  │                 │
        │ analyze_service   │  │  VisionOCR       │  │  UserRepo       │
        │ recommend_service │  │  SearchAgent     │  │  ScanRepo       │
        │ user_service      │  │  ScorerAgent     │  │  FavoriteRepo   │
        │ scan_service      │  │  RecommenderAgent│  │                 │
        │                   │  │  Orchestrator    │  │  (pgx/v5 pool)  │
        └───────────────────┘  └────────┬─────────┘  └────────┬────────┘
                                        │                     │
                                        ▼                     ▼
                               ┌────────────────┐    ┌────────────────┐
                               │  Gemini 2.5    │    │  PostgreSQL 16 │
                               │  Flash (ADK)   │    │  (3 tables)    │
                               │  + Google      │    │                │
                               │    Search      │    │  users         │
                               └────────────────┘    │  scans         │
                                                     │  favorites     │
                                                     └────────────────┘
```

### AI Analysis Pipeline (POST /api/analyze)

```
Image bytes (multipart upload)
        │
        ▼
┌───────────────────────────┐
│     VisionOCR Agent       │   Gemini Vision (direct genai call)
│  Extract product name     │   Input:  image bytes + MIME type
│  from product photo       │   Output: "Coca-Cola Zero Sugar"
└───────────┬───────────────┘
            │
            ▼
┌───────────────────────────┐
│    Search Agent (ADK)     │   Gemini 2.5 Flash + Google Search grounding
│  Find real-time           │   Input:  product name
│  ingredient list          │   Output: WebSearchResult
│  from the web             │     { List_of_ingredients: [{name, description}] }
└───────────┬───────────────┘
            │
            ▼
┌───────────────────────────┐
│    Scorer Agent (ADK)     │   Gemini 2.5 Flash (no tools)
│  Score each ingredient    │   Input:  ingredients JSON + user preferences
│  against user profile     │   Output: ScorerResult
│  (allergies, diet goals,  │     { ingredient_scores: [{name, score, reasoning}],
│   avoid list)             │       overall_score: 0-10 }
└───────────┬───────────────┘
            │
            ▼
      Return to client
      (product_name + scoring_data)
```

### Recommendation Pipeline (GET /api/reccomendations)

```
Product name + overall score
        │
        ▼
┌───────────────────────────┐
│  Recommender Agent (ADK)  │   Gemini 2.5 Flash + Google Search grounding
│  Suggest 3 healthier      │   Input:  product name + current score
│  alternatives             │   Output: RecommenderResult
│                           │     { recommendations: [{product_name,
│                           │       health_score, reason}] }
└───────────┬───────────────┘
            │
            ▼
      Return to client
```

### Full Orchestrator Pipeline (AnalyzeAndImprove — internal)

```
Search ──▶ Score ──▶ [if score < 7.0] ──▶ Loop (max 2 turns) {
                                              Recommend ──▶ Rescore
                                              [break if improved]
                                          }
```

The orchestrator uses ADK's `sequentialagent` for the Search → Score chain and `loopagent` for the refinement cycle. This is built and ready internally but exposed as two separate endpoints to give the frontend control over when to fetch recommendations.

---

## Layer-by-Layer Breakdown

### 1. Handler Layer (`internal/handler/`)

Thin HTTP handlers that parse requests, call services, and return JSON responses. Each handler struct receives only its needed dependencies (interface types). Shared utilities in `common.go` handle JSON encoding, error responses, and `chi.URLParam` extraction.

Handlers are responsible for:
- Request parsing (path params, query params, multipart form data, JSON body)
- Input validation at the HTTP level (missing fields, invalid formats)
- Calling the appropriate service method
- Serializing responses with camelCase JSON keys

The `AnalyzeHandler` accepts multipart image uploads and optionally enriches the analysis with the authenticated user's dietary preferences if a JWT is present.

### 2. Service Layer (`internal/service/`)

Business logic orchestration between repositories and agents. Services own input validation rules and coordinate multi-step operations.

- `AnalyzeService`: Chains VisionOCR → Orchestrator (Search + Score), formats the final response
- `RecommendService`: Wraps the Recommender Agent with validation
- `UserService`: User CRUD + preference management + dietary template application
- `ScanService`: Scan history persistence + statistics aggregation

All four services are defined as interfaces in `internal/service/interfaces.go`, enabling handlers to be tested with mock service implementations.

### 3. Agent Layer (`internal/agent/`)

The AI orchestration layer. Each agent is a Go struct wrapping either a direct `genai` client call (VisionOCR) or an ADK `llmagent` (Search, Scorer, Recommender):

| Agent | SDK | Tools | Purpose |
|-------|-----|-------|---------|
| **VisionOCR** | `genai` (direct) | None | Extract product name from image via Gemini Vision |
| **SearchAgent** | ADK `llmagent` | Google Search | Find product ingredients from the web |
| **ScorerAgent** | ADK `llmagent` (2 variants) | None | Score ingredients or recommendations against user preferences |
| **RecommenderAgent** | ADK `llmagent` | Google Search | Suggest healthier product alternatives |
| **Orchestrator** | ADK `sequentialagent` + `loopagent` | — | Chains agents into analysis/recommendation workflows |

Key design: All agents share a single `model.LLM` instance (Gemini 2.5 Flash) but operate with isolated system prompts defined in `prompts.go`. The `runAgentOnce()` helper in `client.go` creates an in-memory ADK session per invocation, runs the agent, and extracts the last text output. A `stripJSONCodeFences()` utility handles LLM outputs wrapped in markdown code blocks.

### 4. Repository Layer (`internal/repository/`)

SQL-first data access using raw `pgx/v5` queries (no ORM). Each repository is a private struct implementing a public interface:

- `UserRepository` — UPSERT with `ON CONFLICT`, JSONB preference management
- `ScanRepository` — Paginated listing with `LIMIT`, stats aggregation with `COUNT`/`AVG`/`CASE`
- `FavoriteRepository` — Unique constraint enforcement, existence checks

The `DB` struct in `postgres.go` manages the `pgxpool.Pool` and exposes a `pgx.Row`/`pgx.Rows` querier interface. Repositories accept this interface rather than a concrete pool, enabling `pgxmock`-based testing without a running database.

Migrations run automatically at server startup via `golang-migrate`, reading versioned SQL files from the `migrations/` directory.

---

## Database Design

### Entity Relationship

```
users (1) ──────< scans (many)
  │
  └──────────────< favorites (many)

users.id ← TEXT PRIMARY KEY (Auth0 sub, e.g. "auth0|abc123")
scans.user_id → REFERENCES users(id) ON DELETE CASCADE
favorites.user_id → REFERENCES users(id) ON DELETE CASCADE
favorites(user_id, product_name) → UNIQUE constraint
```

### Schema Details

**users** — Stores Auth0 profile data plus dietary preferences as JSONB arrays. Using JSONB for `allergies`, `diet_goals`, and `avoid_ingredients` avoids junction tables and allows flexible, schema-less preference lists that can grow without migrations.

**scans** — Each analysis result is persisted with the full ingredient breakdown as a JSONB column. Indexed on `user_id` and `timestamp DESC` for efficient history queries. The `id` is a UUID generated server-side.

**favorites** — Tracks bookmarked products with a unique constraint on `(user_id, product_name)` to prevent duplicates. Uses `SERIAL` primary key since favorites don't need external UUIDs.

All tables use `TIMESTAMPTZ` for consistent timezone handling and `ON DELETE CASCADE` for referential integrity cleanup.

---

## Design Decisions

### Why Go over Python/FastAPI?

The original SafeBites backend was written in Python with FastAPI and the OpenAI Agents SDK. The rewrite to Go was motivated by:

1. **Type safety** — Go's static typing catches entire categories of bugs at compile time that Python only surfaces at runtime. Agent response parsing, in particular, benefits from typed struct unmarshaling instead of dynamic dict access.
2. **Concurrency model** — Go's goroutines and channels provide lightweight concurrency without the GIL limitations of Python. The HTTP server handles concurrent requests natively without async/await complexity.
3. **Google ADK for Go** — Google's Agent Development Kit has first-class Go support with `sequentialagent`, `loopagent`, and `llmagent` primitives. This provides structured agent orchestration that the OpenAI SDK didn't offer in the Python version.
4. **Single binary deployment** — `go build` produces a single static binary. The Docker image is an Alpine container with just the binary + migration files — no runtime, no virtualenv, no dependency resolution at deploy time.
5. **Performance** — Go's compiled nature delivers lower latency and memory usage. The server starts in milliseconds rather than seconds.

### Why chi over stdlib or Gin?

chi was chosen because it is fully compatible with `net/http` (handlers are standard `http.HandlerFunc`), supports composable middleware chains, and provides `chi.URLParam` for path parameter extraction. Unlike Gin, chi doesn't introduce a custom `Context` type — everything is standard library, so handlers can be tested with `httptest` without framework-specific setup.

### Why raw SQL (pgx) over an ORM?

Using `pgx/v5` with raw SQL provides:
- Full control over queries without ORM-generated SQL surprises
- Direct access to PostgreSQL-specific features (JSONB, `ON CONFLICT DO UPDATE`)
- Lighter dependency footprint than GORM or Ent
- Simpler debugging — the SQL in the code is what runs against the database

### Why JSONB for preferences instead of junction tables?

User preferences (allergies, diet goals, avoid ingredients) are stored as JSONB arrays rather than normalized junction tables. This was intentional:
- Preferences are always read/written as a unit — there's no need to query "all users allergic to peanuts"
- A single `UPDATE` with `$1::jsonb` replaces the preference list atomically
- The dietary template application becomes a single row update instead of multi-table cascade
- Schema evolution (adding new preference types) requires no migration

### Why separate Search + Scorer agents instead of one?

Splitting the analysis into Search and Scorer agents follows the single-responsibility principle at the AI layer:
- The **Search Agent** is optimized for retrieval — it uses Google Search grounding to find real-time ingredient data and doesn't need to reason about safety
- The **Scorer Agent** is optimized for reasoning — it evaluates ingredients against user-specific preferences without needing web access
- This separation makes each agent's prompt smaller and more focused, improving output quality
- It enables independent testing — Search logic can be validated against mocked web results, while Scorer logic can be tested with predefined ingredient lists

### Why auto-run migrations at startup?

The server applies pending migrations on startup via `golang-migrate`. This eliminates a separate migration step in deployment pipelines, ensures the database schema is always consistent with the application version, and simplifies local development (just `make run`). In a production multi-instance scenario, `golang-migrate` uses advisory locks to ensure only one instance runs migrations.

---

## Go Patterns Used

### Repository Pattern

Every data access operation goes through a repository interface. The implementations are private structs, and only the interface is exported. This inverts the dependency — handlers and services depend on abstractions, not concrete database code.

```go
// interfaces.go
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*model.User, error)
    Upsert(ctx context.Context, user *model.User) (*model.User, error)
    UpdatePreferences(ctx context.Context, id string, prefs *model.UserPreferences) (*model.User, error)
}

// user_repo.go — private struct, exported constructor returns the interface
type userRepo struct{ db querier }
func NewUserRepository(db *DB) UserRepository { return &userRepo{db: db.Pool} }
```

### Interface Segregation

Each consumer receives only the interface it needs. Handlers don't see repository internals. Services don't see HTTP details. The `AnalyzeHandler` depends on `AnalyzeService` and `UserService` interfaces, not their concrete implementations. This keeps each component testable in isolation.

### Constructor Injection

All dependency wiring happens in `buildRouter()` in `router.go`. Repositories are constructed from the database pool, agents from the LLM model, services from repositories + agents, and handlers from services. No globals, no `init()` functions, no service locator pattern.

```go
func buildRouter(cfg *config.Config, db *repository.DB) (*chi.Mux, error) {
    userRepo := repository.NewUserRepository(db)
    llm, _ := sbagent.NewGeminiModel(ctx, cfg.GoogleAPIKey, "")
    orchestrator, _ := sbagent.NewOrchestratorFromModel(llm, sbagent.WorkflowConfig{})
    analyzeService := service.NewAnalyzeService(visionOCR, orchestrator)
    analyzeHandler := &handler.AnalyzeHandler{Analyze: analyzeService, Users: userService}
    // ...wire routes...
}
```

### Graceful Shutdown

The server listens for `SIGINT`/`SIGTERM` signals, then calls `srv.Shutdown()` with a 10-second deadline. This allows in-flight requests to complete while refusing new connections — critical for AI-powered endpoints where a single analysis request can take several seconds.

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
srv.Shutdown(shutdownCtx)
```

### Middleware Composition

chi's middleware chain is ordered intentionally: `RequestID` → `Logging` → `Recoverer` → `CORS` → `OptionalAuth`. The logging middleware skips `OPTIONS` preflight requests and the root health check to reduce noise. `OptionalAuth` extracts JWT claims when present but doesn't reject unauthenticated requests; `RequireAuth` is applied selectively per-route (only `/api/users/me`).

### Table-Driven Tests

All 89 test functions follow Go's table-driven pattern with `t.Run()` subtests. Each test case is a struct with named fields for inputs, mock setup, expected outputs, and error expectations. This makes it easy to add edge cases without duplicating test infrastructure.

### FlexibleString for LLM Output Robustness

The `FlexibleString` type in `model/agent.go` implements `UnmarshalJSON` to accept both `"HIGH"` (string) and `8.5` (number) from Gemini outputs. LLMs occasionally return scores as numbers instead of the requested string format — this type handles both transparently without validation errors.

### Embedded Static Assets

The OpenAPI spec (`openapi.json`) is embedded in the handler binary using Go's `//go:embed` directive, eliminating filesystem dependencies at runtime. The Swagger UI HTML is served from a handler function that references the embedded spec path.

---

## Python-to-Go Migration

This backend is a 1:1 API-compatible rewrite of the original Python/FastAPI backend. The frontend `backendApi.ts` requires **zero changes**.

| Concern | Python Backend | Go Backend |
|---------|---------------|------------|
| Framework | FastAPI | chi (stdlib `net/http` compatible) |
| Database | SQLAlchemy ORM + SQLite (dev) | pgx/v5 raw SQL + PostgreSQL always |
| AI SDK | OpenAI Agents SDK + GPT-4.1-mini | Google ADK + Gemini 2.5 Flash |
| Web Search | OpenAI WebSearchTool | Google Search grounding (native to Gemini) |
| Auth | python-jose | golang-jwt/v5 |
| Config | pydantic + dotenv | env vars + godotenv |
| Testing | pytest | go test + testify + pgxmock |
| Deployment | Python runtime + pip install | Single static binary (Alpine container) |

Key differences in the Go version:
- PostgreSQL is used in all environments (not SQLite for dev), ensuring behavior parity between development and production
- Agents use Google Search grounding natively rather than a separate web search tool, reducing API calls
- The ADK `sequentialagent` and `loopagent` provide declarative workflow composition instead of imperative Python orchestration
- JSON responses use `json` struct tags with camelCase naming to match the frontend contract exactly

---

## Testing Strategy

### Test Coverage by Layer

| Layer | Test Files | Tests | Approach |
|-------|-----------|-------|----------|
| Config | 1 | 2 | Direct function tests (dev mode auth, CORS parsing) |
| Middleware | 3 | 4 | `httptest.ResponseRecorder` with crafted requests |
| Handlers | 5 | 22 | `httptest` server + mock services via interfaces |
| Repositories | 3 | 17 | `pgxmock` — mock SQL expectations without a real database |
| Services | 4 | 18 | Mock repositories + mock agents, validate orchestration logic |
| Agents | 4 | 26 | `fakeLLM` + `fakeVisionClient` test doubles, validate prompt construction and JSON parsing |
| **Total** | **20** | **89** | |

### Testing Approach

**Repository tests** use `pgxmock` to set up expected SQL queries and verify that the repository code constructs correct queries and properly maps database rows to Go structs. No running PostgreSQL is needed.

**Handler tests** use Go's `httptest.NewServer` with mock service implementations injected via interfaces. Each test verifies HTTP status codes, response body JSON structure, and error handling paths.

**Agent tests** use a `fakeLLM` struct that implements the `model.LLM` interface and returns predefined responses. This isolates tests from actual Gemini API calls while verifying prompt construction, JSON parsing, and error handling (malformed JSON, code-fenced output, empty responses).

**Service tests** combine mock repositories and mock agents to verify business logic: input validation, orchestration order, and correct data transformation between layers.

### What the Tests Cover

- Happy paths for all CRUD operations
- Input validation (missing fields, invalid formats, out-of-range scores)
- Error propagation (database errors, agent failures, malformed AI responses)
- Edge cases (null safety scores, empty preference lists, user not found still completes analysis)
- LLM output resilience (code-fenced JSON, mixed string/number types)
- Middleware behavior (dev-mode auth bypass, CORS header verification, request logging skip conditions)
- Orchestrator control flow (early stopping when score improves, max iteration limits, high initial score bypass)

---

## Future Additions

### 1. Price-Aware Recommendations

Currently, the Recommender Agent suggests healthier alternatives based purely on health scores. Adding price awareness would constrain recommendations to products within a configurable price range (e.g., ±25% of the scanned product's estimated price).

**Implementation approach:**
- Extend the Search Agent to extract price data alongside ingredients during web search
- Add a `price_range` field to the `RecommenderResult` struct
- Modify the recommender prompt to accept a price constraint and filter alternatives accordingly
- Store price estimates in the `scans` table for historical price tracking
- Expose a `max_price_deviation` query parameter on the recommendations endpoint

This matters because health-only recommendations can suggest premium organic products that are impractical for budget-conscious users. Price-bounded suggestions make the system genuinely useful for everyday shopping decisions.

### 2. Barcode Scanning Support

The current system relies on Gemini Vision to extract product names from images, which can fail on products with unclear labeling or non-English packaging. Adding barcode/UPC scanning would provide a deterministic product identification path.

**Implementation approach:**
- Add a barcode detection step before Vision OCR in the analysis pipeline (using an open barcode database API like Open Food Facts)
- If a barcode is detected and matches a known product, skip Vision OCR entirely and use the database result
- Fall back to Vision OCR for unrecognized barcodes
- Cache barcode → product name mappings in a new `products` table to reduce API calls
- Add a `POST /api/scan-barcode` endpoint for apps that use the device camera's native barcode scanner

This matters because barcode lookup is instantaneous and deterministic (no AI hallucination risk), reducing latency for well-known products while keeping the Vision OCR path for products without barcodes.

### 3. Scan Sharing and Comparison

Users should be able to share their scan results and compare two products side-by-side to make informed choices at the store.

**Implementation approach:**
- Generate shareable links for scan results with a unique token (`GET /api/scans/{scan_id}/share`)
- Add a `POST /api/compare` endpoint that accepts two product names and returns a side-by-side ingredient and safety score comparison
- Store comparison history linked to the user for future reference
- The comparison agent would reuse the existing Scorer Agent to evaluate both products against the same user preferences in a single request

This matters because food safety decisions often happen in the moment — "should I buy Product A or Product B?" — and the current system only evaluates products individually. Side-by-side comparison with shared context (same user preferences applied to both) provides direct decision support.
