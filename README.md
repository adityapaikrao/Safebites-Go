# SafeBites Go Backend

[![Go Version](https://img.shields.io/github/go-mod/go-version/adityapaikrao/safebites-go)](https://golang.org)
[![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Google ADK](https://img.shields.io/badge/Google%20ADK-4285F4?logo=google&logoColor=white)](https://google.github.io/adk-docs/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![Auth0](https://img.shields.io/badge/Auth0-EB5424?logo=auth0&logoColor=white)](https://auth0.com)
[![chi](https://img.shields.io/badge/chi-router-00ADD8?logo=go&logoColor=white)](https://github.com/go-chi/chi)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

SafeBites is a high-performance, AI-driven backend service designed to empower users with dietary restrictions. Built with Go and powered by Google's Gemini models, SafeBites analyzes food products in real-time to determine safety based on personalized health profiles and dietary preferences.

This project is a high-availability rewrite of the original Python/FastAPI backend, optimized for performance, type safety, and seamless AI agent orchestration using the **Google Agent Development Kit (ADK) for Go**.

## 🚀 Key Features

- **AI-Powered Product Scanning**: Utilizes **Gemini 2.0 Flash Vision** to accurately extract product information from images.
- **Intelligent Ingredient Analysis**: Multi-agent workflow that leverages **Google Search grounding** to cross-reference ingredients against known allergens and health risks.
- **Personalized Safety Scoring**: Provides a granular health score and detailed breakdown based on user-defined dietary templates (e.g., Vegan, Gluten-Free, Keto).
- **Smart Recommendations**: Recommends safe alternatives for products flagged as unhealthy or unsafe for the user's specific profile.
- **Robust CRUD Operations**: Efficient management of user profiles, scan histories, and favorite products using **PostgreSQL**.

## 🛠 Tech Stack

- **Language**: Go 1.25+ (optimized for concurrency and performance)
- **API Framework**: [chi](https://github.com/go-chi/chi) (standard-library compatible router)
- **Database**: [PostgreSQL](https://www.postgresql.org/) with [pgx/v5](https://github.com/jackc/pgx) (high-performance driver)
- **AI/LLM**: [Google Gemini 2.0 Flash](https://ai.google.dev/models/gemini) & [Google ADK for Go](https://github.com/google/ai-development-kit)
- **Authentication**: Auth0 (JWT verification)
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate) for versioned database schema management
- **Containerization**: Docker & Docker Compose for local development and deployment

## 🏗 Architecture

SafeBites follows a clean, modular architecture:

- **Handler Layer**: RESTful API endpoints and middleware.
- **Service Layer**: Core business logic and agent orchestration.
- **Agent Layer**: Specialized AI agents (Vision, Search, Scorer, Recommender).
- **Repository Layer**: Data persistence with SQL-first approach.

## 🏁 Getting Started

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
