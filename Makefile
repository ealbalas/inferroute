.PHONY: up down build test migrate lint clean

# ── Local dev ──────────────────────────────────────────────────────────────────

up:
	docker compose up --build

down:
	docker compose down -v

# Start only the backing services (postgres + redis) for local Go development
infra:
	docker compose up -d postgres redis

# ── Build ──────────────────────────────────────────────────────────────────────

build:
	go build -o bin/api      ./services/api
	go build -o bin/router   ./services/router
	go build -o bin/worker   ./services/worker
	go build -o bin/metrics  ./services/metrics

# ── Database ───────────────────────────────────────────────────────────────────

migrate:
	@echo "Running migrations..."
	@docker compose exec -T postgres psql -U inferroute -d inferroute \
		< internal/database/migrations/001_initial.sql
	@echo "Migrations complete."

# ── Testing ────────────────────────────────────────────────────────────────────

test:
	go test ./... -race -count=1

test-integration:
	go test ./tests/integration/... -tags integration -v

# ── Code quality ───────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./...

vet:
	go vet ./...

# ── Cleanup ────────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/
	docker compose down -v --remove-orphans
