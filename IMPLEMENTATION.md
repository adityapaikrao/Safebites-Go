# SafeBites Go Backend — Implementation Checklist

Track progress through each phase of the Go backend rewrite.

---

## Phase 1: Project Scaffolding & Database
> **Goal**: Working server that connects to Postgres and runs migrations

- [x] Initialize Go module (`go.mod`)
- [x] Install all dependencies (`go.sum`)
- [x] Create `cmd/server/main.go` (entry point: config → DB → router → serve)
- [x] Create `internal/config/config.go` (env-based config struct)
- [x] Create `docker-compose.yml` (PostgreSQL local dev)
- [x] Create `migrations/001_create_users.up.sql`
- [x] Create `migrations/001_create_users.down.sql`
- [x] Create `migrations/002_create_scans.up.sql`
- [x] Create `migrations/002_create_scans.down.sql`
- [x] Create `migrations/003_create_favorites.up.sql`
- [x] Create `migrations/003_create_favorites.down.sql`
- [x] Create `internal/repository/postgres.go` (pgx pool + migrate runner)
- [x] Create `Makefile` (build, run, test, migrate targets)
- [x] Create `.env.example`
- [x] Create `README.md`
- [x] **Verify**: `make migrate-up` creates all tables in Postgres

---

## Phase 2: Models & Repository Layer
> **Goal**: Full CRUD for users, scans, favorites against real Postgres

- [x] Create `internal/model/user.go` (User, UserPreferences, UserStats structs)
- [x] Create `internal/model/scan.go` (Scan struct)
- [x] Create `internal/model/favorite.go` (Favorite struct)
- [x] Create `internal/model/template.go` (DietaryTemplate static data)
- [x] Create `internal/model/agent.go` (WebSearchResult, ScorerResult, RecommenderResult)
- [x] Create `internal/repository/interfaces.go` (UserRepository, ScanRepository, FavoriteRepository interfaces)
- [x] Create `internal/repository/user_repo.go` (UserRepository implementation)
- [x] Create `internal/repository/scan_repo.go` (ScanRepository implementation)
- [x] Create `internal/repository/favorite_repo.go` (FavoriteRepository implementation)
- [x] Create `internal/repository/user_repo_test.go` (table tests + pgxmock)
- [x] Create `internal/repository/scan_repo_test.go` (table tests + pgxmock)
- [x] Create `internal/repository/favorite_repo_test.go` (table tests + pgxmock)
- [x] **Verify**: `make test` passes all repository tests

---

## Phase 3: HTTP Layer (Handlers + Middleware)
> **Goal**: All REST endpoints responding with correct JSON shapes; auth working

- [x] Create `internal/middleware/cors.go`
- [x] Create `internal/middleware/logging.go`
- [x] Create `internal/middleware/auth.go` (Auth0 JWT + dev bypass; RequireAuth + OptionalAuth)
- [x] Create `internal/handler/health.go` (GET /)
- [x] Create `internal/handler/user.go` (GET/POST /api/users, /api/users/{id}, /me)
- [x] Create `internal/handler/preference.go` (POST /api/users/{id}/preferences)
- [x] Create `internal/handler/template.go` (GET /api/dietary-templates, POST apply-template)
- [x] Create `internal/handler/scan.go` (GET/POST /api/users/{id}/scans, stats)
- [x] Create `internal/handler/favorite.go` (GET/POST/DELETE /api/users/{id}/favorites + check)
- [x] Wire all routes in `cmd/server/main.go`
- [x] Create `internal/handler/user_test.go` (httptest table tests)
- [x] Create `internal/handler/scan_test.go` (httptest table tests)
- [x] Create `internal/handler/favorite_test.go` (httptest table tests)
- [x] Create `internal/middleware/auth_test.go` (token validation tests)
- [x] **Verify**: `make test` passes all handler tests; curl smoke test all endpoints

---

## Phase 4: Agent Layer (Google ADK + Gemini)
> **Goal**: All AI agents working; product name extraction and all pipelines functional

- [x] Create `internal/agent/client.go` (Gemini genai.Client init + interface)
- [x] Create `internal/agent/prompts.go` (port all system prompts from Python)
- [x] Create `internal/agent/vision.go` (ExtractProductName via Gemini vision)
- [x] Create `internal/agent/search.go` (WebSearchAgent with Google Search grounding)
- [x] Create `internal/agent/scorer.go` (ScorerAgent with structured JSON output)
- [x] Create `internal/agent/recommender.go` (RecommenderAgent with Google Search)
- [x] Create `internal/agent/mock_client.go` (mock client for tests)
- [x] Create `internal/agent/vision_test.go` (table tests with mock client)
- [x] Create `internal/agent/search_test.go` (table tests with mock client)
- [x] Create `internal/agent/scorer_test.go` (table tests with mock client)
- [x] Create `internal/agent/recommender_test.go` (table tests with mock client)
- [x] **Verify**: `make test` passes all agent tests; manual test with real GOOGLE_API_KEY

---

## Phase 5: Service Layer & End-to-End Integration
> **Goal**: Full analyze & recommend pipelines wired; all endpoints functional end-to-end

- [x] Create `internal/service/user_service.go` (user CRUD + preference + template logic)
- [x] Create `internal/service/scan_service.go` (scan CRUD + stats)
- [x] Create `internal/service/analyze_service.go` (vision → search → scorer pipeline)
- [x] Create `internal/service/recommend_service.go` (recommender pipeline)
- [x] Create `internal/handler/analyze.go` (POST /api/analyze)
- [x] Create `internal/handler/recommend.go` (GET /api/reccomendations/{name}/{score})
- [x] Create `internal/service/analyze_service_test.go` (integration: real DB + mock agents)
- [x] Create `internal/service/user_service_test.go` (mock repo tests)
- [x] Update `cmd/server/main.go` (wire agents + services into handlers)
- [x] **Verify**: `make test` passes all; full manual test with frontend running

---

## Phase 6: Docker, Docs & Polish
> **Goal**: Production-ready container, clean docs, ready to deploy on Render/Fly

- [ ] Create multi-stage `Dockerfile`
- [ ] Update `docker-compose.yml` to add app service (app + postgres)
- [ ] Update `README.md` with full setup, dev, and deploy instructions
- [ ] Add `render.yaml` (deploy config mirroring Python backend)
- [ ] Final `make test` — all tests green
- [ ] **Verify**: `docker-compose up --build` starts full stack; frontend can connect

---

## Test Coverage Targets

| Package              | Min Coverage |
|----------------------|--------------|
| `internal/repository`| 80%          |
| `internal/handler`   | 80%          |
| `internal/middleware`| 70%          |
| `internal/agent`     | 70%          |
| `internal/service`   | 75%          |

Run coverage: `make test-cover`
