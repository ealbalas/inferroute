//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestUsageEndpoint(t *testing.T) {
	truncateAll(t)
	auth := registerUser(t, "usagetest@example.com", "password123")

	// Seed a usage_metrics row directly - the daily rollup is performed by a
	// background aggregator, so we populate it here to test the endpoint itself.
	_, err := testDB.Exec(context.Background(), `
		INSERT INTO usage_metrics (user_id, date, request_count, total_tokens, total_cost)
		VALUES ($1, CURRENT_DATE, 5, 1200, 0.0024)
	`, auth.UserID)
	if err != nil {
		t.Fatalf("seed usage_metrics: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, apiBase+"/v1/usage", nil)
	req.Header.Set("Authorization", "Bearer "+auth.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Skipf("API server not running: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("usage: expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Usage []struct {
			Date         string  `json:"date"`
			RequestCount int     `json:"request_count"`
			TotalTokens  int     `json:"total_tokens"`
			TotalCost    float64 `json:"total_cost"`
		} `json:"usage"`
	}
	if decErr := json.NewDecoder(resp.Body).Decode(&body); decErr != nil {
		t.Fatalf("decode usage response: %v", decErr)
	}

	if len(body.Usage) == 0 {
		t.Fatal("expected at least one usage row")
	}

	row := body.Usage[0]
	if row.Date == "" {
		t.Error("expected non-empty date in usage row")
	}
	if row.TotalTokens <= 0 {
		t.Errorf("expected positive total_tokens, got %d", row.TotalTokens)
	}
}
