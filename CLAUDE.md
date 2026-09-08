# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Local development — start Postgres + Redis only (run services with `go run`)
make infra

# Build all service binaries to bin/
make build

# Run a single service locally (after `make infra`)
go run ./services/api
go run ./services/router
go run ./services/worker

# Start everything via Docker Compose
make up

# Unit tests (all packages)
make test                          # go test ./... -race -count=1

# Run tests for a single package
go test ./internal/cache/... -v

# Integration tests (require running Postgres + Redis)
make test-integration              # -tags integration

# Run a single integration test
go test ./tests/integration/... -tags integration -run TestRateLimiting -v

# Lint
make vet                           # go vet
make lint                          # golangci-lint (must be installed)

# Migrations (run against the Docker Compose postgres)
make migrate

# Frontend
cd frontend && npm install && npm run dev   # http://localhost:5173
cd frontend && npm run build
```

## Architecture

**Single Go module** (`github.com/inferroute/inferroute`) at the repo root. All four services share this module and import shared code from `internal/`.

### Request flow

```
POST /v1/generate
  │
  ▼ services/api (8080)
  ├── middleware/auth.go      — extracts Bearer token; API keys use FastKeyHash (SHA-256 index) + bcrypt verify
  ├── middleware/ratelimit.go — Redis sliding-window; IncrBy per user per 1-second window
  ├── handlers/generate.go   — cache lookup → call router → call worker → cache write → DB record
  │
  ▼ services/router (8081)   POST /route
  ├── router.go              — WorkerRegistry (thread-safe map), circuit breaker (5 failures → 30 s cooldown)
  ├── algorithms/weighted_score.go — score = 0.40×latency + 0.30×load + 0.20×errors + 0.10×cost; lowest wins
  ├── algorithms/round_robin.go   — atomic counter mod len(healthy)
  └── health_checker.go      — goroutine polling every worker GET /health every 5 s
  │
  ▼ services/worker (9001–9003)   POST /generate
  └── main.go                — simulated inference; jitter ±30%; configurable via WORKER_LATENCY_MS / WORKER_ERROR_RATE
```

### Internal packages

| Package | Key types / functions |
|---------|----------------------|
| `internal/auth` | `GenerateAPIKey()`, `HashAPIKey()`, `FastKeyHash()`, `SignToken()`, `ParseToken()` |
| `internal/database` | `Connect()` — pgx pool with 10-retry startup loop |
| `internal/cache` | `Client` wrapping go-redis; `CacheKey(prompt, model)` = SHA-256 hex; `RateLimitKey(userID, window)` |
| `internal/observability` | `NewLogger()` — slog JSON; `InitTracer()` — OTLP gRPC exporter |

### Auth model

Two separate auth paths share the same `Authorization: Bearer` header:
- **API consumers** send `ir_live_*` keys → `middleware/auth.go` → DB lookup via `FastKeyHash` index + bcrypt
- **Dashboard users** send JWTs → `middleware/auth.go` (`JWTAuth`) → `internal/auth.ParseToken()`

The Gin context key `"userID"` is set by both paths and read by all handlers.

### Database

Schema lives in `internal/database/migrations/001_initial.sql`. Docker Compose mounts this as an init script so Postgres runs it automatically on first start. For subsequent runs use `make migrate`.

Seven tables: `users`, `api_keys`, `workers`, `requests`, `routing_decisions`, `usage_metrics`, `incidents`. Workers are seeded with the three Docker Compose worker addresses.

### Frontend

Vite + React + TypeScript. `src/api/client.ts` is an axios instance that injects the JWT from `localStorage` and redirects to `/login` on 401. All API calls go through `/v1` which Vite's dev proxy forwards to `localhost:8080`.

### Infrastructure (Phase 3)

Terraform files in `infrastructure/` have all AWS resources commented out. Uncomment when beginning cloud deployment. State backend config is also commented out in `main.tf`. Sensitive values (`db_password`, `jwt_secret`) must be passed via `TF_VAR_*` env vars.
