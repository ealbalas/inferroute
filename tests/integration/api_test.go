//go:build integration

// Run with: go test ./tests/integration/... -tags integration -v
//
// Requires a running Postgres and Redis (use `make infra`) and the full
// API stack at http://localhost:8080 for HTTP tests.

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/inferroute/inferroute/internal/cache"
	"github.com/inferroute/inferroute/internal/database"
)

const apiBase = "http://localhost:8080"

var (
	testDB    *database.Pool
	testRedis *cache.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://inferroute:inferroute@localhost:5432/inferroute?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	var err error
	testDB, err = database.Connect(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect postgres: %v\n", err)
		os.Exit(1)
	}

	testRedis, err = cache.New(redisURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect redis: %v\n", err)
		os.Exit(1)
	}

	for _, path := range []string{
		"../../internal/database/migrations/001_initial.sql",
		"../../internal/database/migrations/002_add_key_hash_fast.sql",
	} {
		sql, readErr := os.ReadFile(path)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "read migration %s: %v\n", path, readErr)
			os.Exit(1)
		}
		if _, execErr := testDB.Exec(ctx, string(sql)); execErr != nil {
			fmt.Fprintf(os.Stderr, "exec migration %s: %v\n", path, execErr)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

// truncateAll resets all mutable tables before a test, keeping workers seeded by migrations.
func truncateAll(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testDB.Exec(ctx, `
		TRUNCATE TABLE routing_decisions, requests, usage_metrics, api_keys, incidents, users
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncateAll: %v", err)
	}
}

type authResult struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// registerUser creates a test user via the API and returns the JWT + user ID.
// Skips the test if the API server is unreachable.
func registerUser(t *testing.T, email, password string) authResult {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	resp, err := http.Post(apiBase+"/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skipf("API server not running: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: expected 201, got %d", email, resp.StatusCode)
	}
	var result authResult
	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		t.Fatalf("decode register response: %v", decErr)
	}
	return result
}

// createAPIKey creates a new API key for the JWT-authenticated user.
// Returns the raw key and key ID.
func createAPIKey(t *testing.T, jwtToken, name string) (rawKey, keyID string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"name": name})
	req, _ := http.NewRequest(http.MethodPost, apiBase+"/v1/keys", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("API server not running: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createAPIKey %q: expected 201, got %d", name, resp.StatusCode)
	}
	var result map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result["key"], result["id"]
}

// TestCacheRoundtrip verifies Redis get/set/del round-trip.
func TestCacheRoundtrip(t *testing.T) {
	ctx := context.Background()
	key := cache.CacheKey("integration-test-prompt", "inferroute-sim-v1")
	t.Cleanup(func() { _ = testRedis.Del(ctx, key) })

	_, ok, err := testRedis.Get(ctx, key)
	if err != nil || ok {
		t.Fatalf("expected cache miss, got ok=%v err=%v", ok, err)
	}

	if setErr := testRedis.Set(ctx, key, `{"output":"hello"}`, 5*time.Minute); setErr != nil {
		t.Fatalf("Set: %v", setErr)
	}

	val, ok, err := testRedis.Get(ctx, key)
	if err != nil || !ok || val == "" {
		t.Fatalf("expected cache hit, got ok=%v val=%q err=%v", ok, val, err)
	}
}
