# InferRoute

**Distributed AI Inference Gateway and Monitoring Platform**

A production-style platform that routes AI inference requests across multiple model workers based on latency, load, availability, cost, and error rate.

> The AI model is not the main accomplishment. The infrastructure around it is.

---

## Quick Start

```bash
# Start Postgres + Redis
make infra

# Start all services
make up

# Health check
curl http://localhost:8080/health

# Frontend
cd frontend && npm install && npm run dev
# → http://localhost:5173
```

---

## Architecture

```
React Frontend (5173)
       │
       ▼
API Service (8080)
  ├── Auth middleware (JWT / API key)
  ├── Rate limiter (Redis sliding window)
  └── Cache check (Redis)
       │
       ▼
Router Service (8081)
  ├── Round Robin
  └── Weighted Score (0.40×latency + 0.30×load + 0.20×error + 0.10×cost)
       │
  ┌────┼────┐
  ▼    ▼    ▼
Worker A  Worker B  Worker C
(9001)    (9002)    (9003)
```

---

## Services

| Service | Port | Description |
|---------|------|-------------|
| `api` | 8080 | API gateway — auth, rate limiting, caching, request routing |
| `router` | 8081 | Routing engine — worker selection, circuit breaker, health checks |
| `worker` | 9001–9003 | Simulated inference servers |
| `metrics` | 9100 | Prometheus metrics + SQS consumer (Phase 3) |
| `postgres` | 5432 | Primary datastore |
| `redis` | 6379 | Rate limiting + response cache |
| `prometheus` | 9090 | Metrics scraper |
| `grafana` | 3001 | Metrics dashboard |

---

## API

```bash
# Register
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"yourpassword"}'

# Create API key (use JWT from register response)
curl -X POST http://localhost:8080/v1/keys \
  -H "Authorization: Bearer <jwt>" \
  -d '{"name":"my-key"}'

# Generate inference
curl -X POST http://localhost:8080/v1/generate \
  -H "Authorization: Bearer ir_live_..." \
  -d '{"prompt":"Explain TCP congestion control","priority":"latency"}'
```

---

## Development Phases

| Phase | What you build |
|-------|---------------|
| **1 — MVP** | React UI, API, Postgres, 3 simulated workers, round-robin routing, Docker Compose |
| **2 — Distributed Systems** | Redis rate limiting + caching, health checks, weighted routing, circuit breakers, retry/failover |
| **3 — Production Engineering** | AWS (ECS/RDS/Redis/SQS), Terraform, GitHub Actions CI/CD, OpenTelemetry, Prometheus, Grafana |
| **4 — Advanced Features** | Load testing (k6), multiple routing algorithms, real inference models, regional workers |

---

## Testing

```bash
# Unit tests
make test

# Integration tests (requires running infra)
make test-integration

# Load test (requires k6 and a running API key)
k6 run -e API_KEY=ir_live_... tests/load/script.js

# E2E (requires running frontend + API)
cd tests/e2e && npx playwright test
```

---

## Metrics

_Fill in as you complete each phase:_

| Metric | Value |
|--------|-------|
| Max throughput | — req/s |
| P50 latency | — ms |
| P95 latency | — ms |
| P99 latency | — ms |
| Cache hit rate | — % |
| Failover time | — ms |
| Test coverage | — % |

---

## Stack

**Go · TypeScript · React · PostgreSQL · Redis · Docker · AWS ECS · RDS · SQS · Terraform · GitHub Actions · OpenTelemetry · Prometheus · Grafana · k6**
