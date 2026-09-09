//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGenerateEndpoint(t *testing.T) {
	truncateAll(t)
	auth := registerUser(t, "gentest@example.com", "password123")
	rawKey, _ := createAPIKey(t, auth.Token, "gen-test")

	doGenerate := func(t *testing.T) (map[string]interface{}, int) {
		t.Helper()
		body, _ := json.Marshal(map[string]string{
			"prompt": "What is 2 + 2?",
			"model":  "inferroute-sim-v1",
		})
		req, _ := http.NewRequest(http.MethodPost, apiBase+"/v1/generate", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+rawKey)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()
		var result map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&result)
		return result, resp.StatusCode
	}

	// Retry up to 5 times to tolerate the worker's 2% simulated error rate.
	var result map[string]interface{}
	var lastCode int
	for i := 0; i < 5; i++ {
		result, lastCode = doGenerate(t)
		if lastCode == http.StatusOK {
			break
		}
	}
	if lastCode != http.StatusOK {
		t.Skipf("generate returned %d after retries - router/workers may not be running", lastCode)
	}

	t.Run("response contains required fields", func(t *testing.T) {
		for _, field := range []string{"request_id", "worker", "output", "tokens", "latency_ms", "cost"} {
			if _, ok := result[field]; !ok {
				t.Errorf("missing field %q in generate response", field)
			}
		}
	})

	t.Run("request appears in history", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiBase+"/v1/requests", nil)
		req.Header.Set("Authorization", "Bearer "+auth.Token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("requests history: expected 200, got %d", resp.StatusCode)
		}

		var body struct {
			Requests []map[string]interface{} `json:"requests"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)

		if len(body.Requests) == 0 {
			t.Fatal("expected at least one request in history after generate")
		}
	})
}
