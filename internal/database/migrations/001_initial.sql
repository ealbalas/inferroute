-- InferRoute initial schema
-- Run once on a fresh database. Idempotent via IF NOT EXISTS.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── Users ─────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL DEFAULT 'developer' CHECK (role IN ('admin', 'developer', 'viewer')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── API Keys ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash     TEXT NOT NULL UNIQUE,        -- bcrypt hash of the raw key
    name         TEXT NOT NULL,
    last_used_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id  ON api_keys(user_id);

-- ── Workers ───────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS workers (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL UNIQUE,
    region           TEXT NOT NULL,
    address          TEXT NOT NULL,            -- http://host:port
    status           TEXT NOT NULL DEFAULT 'unknown' CHECK (status IN ('healthy', 'degraded', 'offline', 'unknown')),
    current_load     FLOAT NOT NULL DEFAULT 0,
    average_latency  FLOAT NOT NULL DEFAULT 0, -- milliseconds
    error_rate       FLOAT NOT NULL DEFAULT 0, -- fraction 0.0–1.0
    last_heartbeat   TIMESTAMPTZ
);

-- ── Requests ──────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS requests (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    worker_id   UUID REFERENCES workers(id),
    model       TEXT,
    prompt_hash TEXT,                          -- SHA-256 for cache key, not the raw prompt
    status_code INT NOT NULL,
    latency_ms  INT NOT NULL,
    tokens      INT NOT NULL DEFAULT 0,
    cost        NUMERIC(10, 6) NOT NULL DEFAULT 0,
    cached      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_requests_user_id    ON requests(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_worker_id  ON requests(worker_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at DESC);

-- ── Routing Decisions ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS routing_decisions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id     UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    worker_id      UUID NOT NULL REFERENCES workers(id),
    algorithm      TEXT NOT NULL,              -- round_robin | weighted_score | least_connections
    routing_score  FLOAT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_routing_decisions_request_id ON routing_decisions(request_id);

-- ── Usage Metrics (daily rollup) ──────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS usage_metrics (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date           DATE NOT NULL,
    request_count  INT NOT NULL DEFAULT 0,
    total_tokens   INT NOT NULL DEFAULT 0,
    total_cost     NUMERIC(12, 6) NOT NULL DEFAULT 0,
    UNIQUE (user_id, date)
);

-- ── Incidents ─────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS incidents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    worker_id    UUID REFERENCES workers(id),
    type         TEXT NOT NULL,                -- timeout | error_spike | offline | degraded
    description  TEXT NOT NULL,
    resolved_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Seed Workers ──────────────────────────────────────────────────────────────
-- These match the docker-compose worker services.

INSERT INTO workers (name, region, address, status) VALUES
    ('us-west-1', 'us-west',  'http://worker-a:9001', 'unknown'),
    ('us-west-2', 'us-west',  'http://worker-b:9002', 'unknown'),
    ('us-east-1', 'us-east',  'http://worker-c:9003', 'unknown')
ON CONFLICT (name) DO NOTHING;
