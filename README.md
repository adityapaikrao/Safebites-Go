# SafeBites Go Backend

[![Go](https://img.shields.io/badge/Go_1.25-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Google ADK](https://img.shields.io/badge/Google%20ADK-4285F4?logo=google&logoColor=white)](https://google.github.io/adk-docs/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL_16-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![Auth0](https://img.shields.io/badge/Auth0_JWT-EB5424?logo=auth0&logoColor=white)](https://auth0.com)
[![Tests](https://img.shields.io/badge/tests-89_passed-brightgreen)]()
[![Endpoints](https://img.shields.io/badge/API_endpoints-18-blue)]()

SafeBites is a production-grade, AI-powered backend that helps users with dietary restrictions make safer food choices. Users snap a photo of a food product, and the system identifies it, looks up its ingredients, scores each ingredient against the user's personal dietary profile (allergies, diet goals, avoidances), and recommends healthier alternatives — all in a single request.

Built in Go and orchestrated through Google's Agent Development Kit (ADK), the backend manages a multi-agent AI pipeline (Vision → Search → Scorer → Recommender) backed by Gemini 2.5 Flash with Google Search grounding, ensuring ingredient data is real-time and web-sourced rather than static.

This is a ground-up rewrite of the original Python/FastAPI + OpenAI Agents backend. The Go version was chosen for stronger type safety, lower latency under concurrent load, and native support for Google's ADK agent orchestration primitives (`sequentialagent`, `loopagent`).

## Project at a Glance

| Metric | Value |
|--------|-------|
| Language | Go 1.25 |
| Production code | ~3,000 lines |
| Test code | ~2,100 lines |
| Test functions | 89 across 21 test files |
| API endpoints | 18 (REST) |
| AI agents | 4 (Vision OCR, Search, Scorer, Recommender) |
| Database tables | 3 (users, scans, favorites) |
| SQL migrations | 6 (3 up + 3 down) |
| Dietary templates | 7 built-in (Vegan, Keto, Gluten-Free, etc.) |

## Core Features

**AI-Powered Product Analysis** — Upload a product image and receive a full safety breakdown. The pipeline chains four AI agents: Gemini Vision extracts the product name, a Search Agent with Google Search grounding retrieves real-time ingredient data, and a Scorer Agent evaluates each ingredient against the user's dietary profile, producing per-ingredient safety ratings and an overall score (0–10).

**Smart Recommendations** — For products scoring below a configurable threshold, the Recommender Agent suggests 3 healthier alternatives in the same product category, each scored and justified. The orchestrator runs a refinement loop (up to 2 iterations) to ensure recommendations genuinely improve on the original product's score.

**Personalized Dietary Profiles** — Users configure allergies, diet goals, and ingredients to avoid. 7 built-in dietary templates (Vegan, Vegetarian, Gluten-Free, Keto, Dairy-Free, Nut-Free, Paleo) can be applied with a single API call, or preferences can be set manually.

**Scan History & Favorites** — Every analysis is persisted as a scan record with full ingredient breakdowns. Users can browse scan history, view stats (total scans, daily counts, average safety scores), and bookmark products as favorites.

**Auth0 Integration** — JWT-based authentication with optional and required middleware variants. Development mode bypasses Auth0 when credentials are not configured, enabling local development without external dependencies.

## Tech Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| Language | Go 1.25 | Type safety, compiled performance, first-class concurrency |
| Router | [chi v5](https://github.com/go-chi/chi) | `net/http` compatible, composable middleware, zero dependencies |
| Database | PostgreSQL 16 + [pgx/v5](https://github.com/jackc/pgx) | Raw SQL (no ORM), connection pooling, JSONB for flexible schema |
| AI/LLM | Gemini 2.5 Flash via [Google ADK](https://google.github.io/adk-docs/) + [genai SDK](https://pkg.go.dev/google.golang.org/genai) | Agent orchestration primitives, Google Search grounding, vision support |
| Auth | Auth0 + [golang-jwt/v5](https://github.com/golang-jwt/jwt) | Industry-standard JWT verification with dev-mode bypass |
| Migrations | [golang-migrate/v4](https://github.com/golang-migrate/migrate) | Versioned SQL files, auto-applied at startup |
| Testing | [testify](https://github.com/stretchr/testify) + [pgxmock](https://github.com/pashagolub/pgxmock) | Assertion helpers, database mocking without a running Postgres |
| Containerization | Docker multi-stage build + Docker Compose | Minimal Alpine runtime image, one-command local stack |

## API Endpoints

### Infrastructure
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/` | Health check |
| `GET` | `/docs` | Swagger UI |
| `GET` | `/docs/openapi.json` | OpenAPI 3.0 spec |

### Analysis & Recommendations
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/analyze` | Optional | Upload product image → AI pipeline → safety scoring |
| `GET` | `/api/reccomendations/{product_name}/{overall_score}` | No | Get healthier alternative recommendations |

### Users & Preferences
| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/users/me` | **Required** | Get authenticated user profile |
| `GET` | `/api/users/{user_id}` | No | Get user by ID |
| `POST` | `/api/users` | No | Create or update user |
| `POST` | `/api/users/{user_id}/preferences` | No | Update dietary preferences |
| `GET` | `/api/dietary-templates` | No | List all 7 dietary templates |
| `POST` | `/api/users/{user_id}/apply-template/{template_key}` | No | Apply a dietary template |

### Scans & Favorites
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/users/{user_id}/scans` | List scan history (paginated) |
| `POST` | `/api/users/{user_id}/scans` | Create scan record |
| `GET` | `/api/users/{user_id}/stats` | Scan statistics (totals, averages) |
| `GET` | `/api/users/{user_id}/favorites` | List favorited products |
| `POST` | `/api/users/{user_id}/favorites` | Add to favorites |
| `DELETE` | `/api/users/{user_id}/favorites/{favorite_id}` | Remove from favorites |
| `GET` | `/api/users/{user_id}/favorites/check/{product_name}` | Check if product is favorited |

## Architecture Overview

```
Frontend (Next.js)  ──HTTP──▶  chi Router + Middleware
                                      │
                        ┌─────────────┼─────────────┐
                        ▼             ▼              ▼
                   Service Layer     Agent Layer   Repository Layer
                   (business logic)  (Gemini ADK)  (pgx/v5 → PostgreSQL)
```

The backend follows a **four-layer architecture**: Handlers accept HTTP requests and delegate to Services, which orchestrate Repositories (database) and Agents (AI). Every layer communicates through Go interfaces, enabling comprehensive testing with mocks at every boundary.

The AI Agent Layer uses Google ADK's `sequentialagent` to chain Search → Scorer steps, and a `loopagent` for the recommendation refinement cycle. All agents share a single Gemini model instance but operate with isolated system prompts and tool configurations.

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed architecture diagrams, design decisions, Go patterns, and testing strategy.

## Getting Started

### Prerequisites

- Go 1.25+
- Docker + Docker Compose
- Google API Key (Gemini access)

### Quick Start

```bash
# Clone and configure
cp .env.example .env
# Set DATABASE_URL and GOOGLE_API_KEY in .env

# Start PostgreSQL (migrations run automatically on startup)
make docker-postgres

# Run the server
make run
# → listening on :8080
```

### Run Tests

```bash
make test          # All 89 tests with race detection
make test-cover    # Generate HTML coverage report
```

### Docker (Full Stack)

```bash
make docker-up     # Builds app + starts Postgres, runs migrations
make docker-down   # Tear down
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Compile to `bin/server` |
| `make run` | Run server locally |
| `make test` | Run all tests (`-race -count=1`) |
| `make test-cover` | Tests with HTML coverage report |
| `make lint` | Run `golangci-lint` |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Rollback one migration |
| `make docker-up` | Full Docker Compose stack |
| `make docker-postgres` | Start only Postgres |

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `GOOGLE_API_KEY` | Yes | — | Gemini API key |
| `PORT` | No | `8080` | Server port |
| `ENV` | No | `development` | `development` or `production` |
| `AUTH0_DOMAIN` | No | — | Auth0 tenant domain (omit for dev bypass) |
| `AUTH0_API_AUDIENCE` | No | — | Auth0 API audience (omit for dev bypass) |
| `CORS_ORIGINS` | No | `http://localhost:3000` | Comma-separated allowed origins |
| `MIGRATIONS_PATH` | No | `migrations` | Path to SQL migration files |

## Project Structure

```
cmd/server/          Entry point + router wiring
internal/
  config/            Environment-based configuration
  middleware/        CORS, request logging, JWT auth (optional + required)
  handler/           HTTP handlers, Swagger docs, OpenAPI spec
  model/             Domain structs (User, Scan, Favorite, Agent results)
  repository/        PostgreSQL data access (interfaces + pgx implementations)
  service/           Business logic orchestration, input validation
  agent/             AI pipeline (Vision, Search, Scorer, Recommender, Orchestrator)
migrations/          Versioned SQL (3 tables: users, scans, favorites)
```
