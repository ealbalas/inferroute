//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAPIKeyLifecycle(t *testing.T) {
	truncateAll(t)
	auth := registerUser(t, "keylifecycle@example.com", "password123")

	// Create.
	rawKey, keyID := createAPIKey(t, auth.Token, "lifecycle-test")

	t.Run("created key has ir_live_ prefix", func(t *testing.T) {
		if !strings.HasPrefix(rawKey, "ir_live_") {
			t.Fatalf("expected ir_live_ prefix, got %q", rawKey)
		}
		if keyID == "" {
			t.Fatal("expected non-empty key ID")
		}
	})

	t.Run("key appears in list with created_at", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiBase+"/v1/keys", nil)
		req.Header.Set("Authorization", "Bearer "+auth.Token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("list keys: expected 200, got %d", resp.StatusCode)
		}

		var body struct {
			Keys []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				CreatedAt string `json:"created_at"`
			} `json:"keys"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)

		var found bool
		for _, k := range body.Keys {
			if k.ID == keyID {
				found = true
				if k.CreatedAt == "" {
					t.Error("expected non-empty created_at on listed key")
				}
			}
		}
		if !found {
			t.Fatalf("created key %s not found in listing", keyID)
		}
	})

	// Delete.
	t.Run("delete returns 200", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, apiBase+"/v1/keys/"+keyID, nil)
		req.Header.Set("Authorization", "Bearer "+auth.Token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("delete key: expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("deleted key no longer appears in list", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiBase+"/v1/keys", nil)
		req.Header.Set("Authorization", "Bearer "+auth.Token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()

		var body struct {
			Keys []struct {
				ID string `json:"id"`
			} `json:"keys"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)

		for _, k := range body.Keys {
			if k.ID == keyID {
				t.Fatalf("deleted key %s still appears in listing", keyID)
			}
		}
	})
}

func TestAPIKeyAuthentication(t *testing.T) {
	truncateAll(t)
	auth := registerUser(t, "apikeyauth@example.com", "password123")
	rawKey, _ := createAPIKey(t, auth.Token, "auth-test")

	doGenerate := func(t *testing.T, authHeaderValue string) int {
		t.Helper()
		body, _ := json.Marshal(map[string]string{
			"prompt": "hello",
			"model":  "inferroute-sim-v1",
		})
		req, _ := http.NewRequest(http.MethodPost, apiBase+"/v1/generate", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeaderValue)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	t.Run("valid API key is accepted (auth passes)", func(t *testing.T) {
		code := doGenerate(t, "Bearer "+rawKey)
		// Auth passed when status is not 401; may be 503 if router/worker is down.
		if code == http.StatusUnauthorized {
			t.Fatalf("valid API key was rejected with 401")
		}
	})

	t.Run("unknown API key returns 401", func(t *testing.T) {
		code := doGenerate(t, "Bearer ir_live_0000000000000000000000000000000000000000")
		if code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for unknown key, got %d", code)
		}
	})

	t.Run("token without ir_live_ prefix returns 401", func(t *testing.T) {
		code := doGenerate(t, "Bearer notanirliveprefixedtoken")
		if code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for misformatted token, got %d", code)
		}
	})
}
