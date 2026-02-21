# SafeBites Go Backend

Go backend for SafeBites with PostgreSQL, chi router, and Gemini-powered agent workflows.

## Phase 1 (Scaffolding + Database)

### Prerequisites
- Go 1.25+
- Docker + Docker Compose
- `migrate` CLI (for local migration commands)

### Setup
1. Copy `.env.example` to `.env`
2. Fill in required values (`DATABASE_URL`, `GOOGLE_API_KEY`)

### Run locally
- Start Postgres: `make docker-postgres`
- Run service: `make run`

### Database migrations
- Apply all pending migrations: `make migrate-up`
- Roll back one migration step: `make migrate-down`

You can override migration folder path at runtime with `MIGRATIONS_PATH`.

### Run tests
- All tests: `make test`

## Notes
- `AUTH0_DOMAIN` and `AUTH0_API_AUDIENCE` are optional in local development.
- Current JWT parsing is intentionally minimal and should be replaced with full Auth0 JWKS verification before production rollout.
