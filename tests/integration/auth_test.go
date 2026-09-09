//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestAuthFlow(t *testing.T) {
	truncateAll(t)

	t.Run("register returns 201 and JWT", func(t *testing.T) {
		result := registerUser(t, "authflow_register@example.com", "password123")
		if result.Token == "" {
			t.Fatal("expected non-empty JWT token in register response")
		}
		if result.UserID == "" {
			t.Fatal("expected non-empty user_id in register response")
		}
	})

	t.Run("login with correct password returns 200 and JWT", func(t *testing.T) {
		registerUser(t, "authflow_login@example.com", "password123")

		body, _ := json.Marshal(map[string]string{
			"email":    "authflow_login@example.com",
			"password": "password123",
		})
		resp, err := http.Post(apiBase+"/v1/auth/login", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("login: expected 200, got %d", resp.StatusCode)
		}
		var result authResult
		_ = json.NewDecoder(resp.Body).Decode(&result)
		if result.Token == "" {
			t.Fatal("expected non-empty JWT token in login response")
		}
	})

	t.Run("login with wrong password returns 401", func(t *testing.T) {
		registerUser(t, "authflow_wrongpass@example.com", "password123")

		body, _ := json.Marshal(map[string]string{
			"email":    "authflow_wrongpass@example.com",
			"password": "notthepassword",
		})
		resp, err := http.Post(apiBase+"/v1/auth/login", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Skipf("API server not running: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 for wrong password, got %d", resp.StatusCode)
		}
	})
}
