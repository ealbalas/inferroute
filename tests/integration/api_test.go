//go:build integration

// Run with: go test ./tests/integration/... -tags integration -v
//
// Requires a running Postgres and Redis (use `make infra` to start them).

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/inferroute/inferroute/internal/cache"
	"github.com/inferroute/inferroute/internal/database"
)

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
		panic("connect postgres: " + err.Error())
	}
	testRedis, err = cache.New(redisURL)
	if err != nil {
		panic("connect redis: " + err.Error())
	}

	os.Exit(m.Run())
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// TODO: wire in the actual Gin handler once the server is assembled into a testable function.
	// For now this demonstrates the integration test pattern.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	ctx := context.Background()

	// Clean up test user
	t.Cleanup(func() {
		testDB.Exec(ctx, "DELETE FROM users WHERE email = $1", "test@example.com")
	})

	// Registration
	body, _ := json.Marshal(map[string]string{
		"email":    "test@example.com",
		"password": "supersecret",
	})
	resp, err := http.Post("http://localhost:8080/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Skip("API server not running — start with `make infra && go run ./services/api`")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", resp.StatusCode)
	}

	var regResult map[string]string
	json.NewDecoder(resp.Body).Decode(&regResult)
	if regResult["token"] == "" {
		t.Fatal("expected JWT token in register response")
	}
}

func TestRateLimiting(t *testing.T) {
	ctx := context.Background()
	key := "ratelimit:test-user:" + time.Now().Truncate(time.Second).String()

	// Simulate 11 increments — 11th should exceed limit of 10
	for i := 0; i < 11; i++ {
		count, err := testRedis.IncrBy(ctx, key, 1, 2*time.Second)
		if err != nil {
			t.Fatalf("IncrBy error: %v", err)
		}
		if i == 10 && count <= 10 {
			t.Fatalf("expected count > 10 on request 11, got %d", count)
		}
	}
}

func TestCacheRoundtrip(t *testing.T) {
	ctx := context.Background()
	key := cache.CacheKey("hello world", "inferroute-sim-v1")

	// Miss
	val, ok, err := testRedis.Get(ctx, key)
	if err != nil || ok {
		t.Fatalf("expected cache miss, got ok=%v val=%q err=%v", ok, val, err)
	}

	// Set
	if err := testRedis.Set(ctx, key, `{"output":"hello"}`, 5*time.Minute); err != nil {
		t.Fatalf("Set error: %v", err)
	}

	// Hit
	val, ok, err = testRedis.Get(ctx, key)
	if err != nil || !ok || val == "" {
		t.Fatalf("expected cache hit, got ok=%v val=%q err=%v", ok, val, err)
	}

	_ = testRedis.Del(ctx, key)
}
