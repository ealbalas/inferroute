//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestRateLimiting(t *testing.T) {
	truncateAll(t)
	auth := registerUser(t, "ratelimit@example.com", "password123")
	rawKey, _ := createAPIKey(t, auth.Token, "rate-limit-test")

	doGenerate := func() (int, error) {
		body, _ := json.Marshal(map[string]string{
			"prompt": "test",
			"model":  "inferroute-sim-v1",
		})
		req, _ := http.NewRequest(http.MethodPost, apiBase+"/v1/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+rawKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		return resp.StatusCode, nil
	}

	// Pre-seed the Redis rate-limit counter to 10 (the limit) so the next
	// HTTP request from the middleware increments it to 11 and triggers 429.
	// This avoids relying on bcrypt timing for 11 sequential requests.
	//
	// Ensure we are not at a second boundary before seeding so the window
	// does not flip between the seed and the HTTP request.
	for time.Since(time.Now().Truncate(time.Second)) > 900*time.Millisecond {
		time.Sleep(150 * time.Millisecond)
	}
	windowStart := time.Now().Truncate(time.Second)
	rateLimitKey := fmt.Sprintf("ratelimit:%s:%d", auth.UserID, windowStart.Unix())
	if _, err := testRedis.IncrBy(context.Background(), rateLimitKey, 10, 2*time.Second); err != nil {
		t.Fatalf("seed rate limit counter: %v", err)
	}

	t.Run("11th request returns 429", func(t *testing.T) {
		code, err := doGenerate()
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		if code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 after exceeding rate limit, got %d", code)
		}
	})

	t.Run("rate limit clears after window resets", func(t *testing.T) {
		time.Sleep(1100 * time.Millisecond)
		code, err := doGenerate()
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		if code == http.StatusTooManyRequests {
			t.Fatal("rate limit still active after 1-second window reset")
		}
	})
}
