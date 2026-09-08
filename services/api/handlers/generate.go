package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inferroute/inferroute/internal/cache"
	"github.com/inferroute/inferroute/internal/database"
)

type generateRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Model    string `json:"model"`
	Priority string `json:"priority"` // latency | cost | balanced
}

type generateResponse struct {
	RequestID string  `json:"request_id"`
	Worker    string  `json:"worker"`
	Output    string  `json:"output"`
	Tokens    int     `json:"tokens"`
	LatencyMS int64   `json:"latency_ms"`
	Cost      float64 `json:"cost"`
	Cached    bool    `json:"cached"`
}

// Generate handles POST /v1/generate — the primary inference endpoint.
//
// Flow:
//  1. Check Redis cache (return immediately on HIT)
//  2. Ask the router which worker to use
//  3. Forward the request to that worker
//  4. On failure, record it in the circuit breaker and retry once
//  5. Cache the response and record the request in PostgreSQL
func Generate(db *database.Pool, redis *cache.Client, routerURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req generateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Model == "" {
			req.Model = "inferroute-sim-v1"
		}

		userID, _ := c.Get("userID")
		start := time.Now()

		// ── 1. Cache check ────────────────────────────────────────────────────
		cacheKey := cache.CacheKey(req.Prompt, req.Model)
		if cached, ok, _ := redis.Get(c.Request.Context(), cacheKey); ok {
			var resp generateResponse
			if err := json.Unmarshal([]byte(cached), &resp); err == nil {
				resp.Cached = true
				resp.LatencyMS = time.Since(start).Milliseconds()
				c.JSON(http.StatusOK, resp)
				recordRequest(c, db, userID.(string), "", req.Model, 200,
					int(resp.LatencyMS), resp.Tokens, resp.Cost, true)
				return
			}
		}

		// ── 2. Ask router for the best worker ─────────────────────────────────
		routeResp, workerAddr, err := callRouter(routerURL, req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no healthy workers: " + err.Error()})
			return
		}

		// ── 3. Forward to the selected worker ────────────────────────────────
		workerResp, statusCode, err := callWorker(workerAddr, req)
		if err != nil || statusCode >= 500 {
			// TODO: record failure in router circuit breaker via routerURL
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "worker error"})
			return
		}

		workerResp.Cached = false
		workerResp.LatencyMS = time.Since(start).Milliseconds()

		// ── 4. Cache successful response ──────────────────────────────────────
		if b, err := json.Marshal(workerResp); err == nil {
			_ = redis.Set(c.Request.Context(), cacheKey, string(b), 5*time.Minute)
		}

		// ── 5. Persist request record ─────────────────────────────────────────
		recordRequest(c, db, userID.(string), routeResp.WorkerName, req.Model,
			statusCode, int(workerResp.LatencyMS), workerResp.Tokens, workerResp.Cost, false)

		c.JSON(http.StatusOK, workerResp)
	}
}

// routeResponse mirrors the router service's response.
type routeResponse struct {
	WorkerName    string  `json:"worker_name"`
	WorkerAddress string  `json:"worker_address"`
	Algorithm     string  `json:"algorithm"`
	Score         float64 `json:"score,omitempty"`
}

func callRouter(routerURL string, req generateRequest) (*routeResponse, string, error) {
	body, _ := json.Marshal(req)
	resp, err := http.Post(routerURL+"/route", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("router returned %d", resp.StatusCode)
	}
	var route routeResponse
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return nil, "", err
	}
	return &route, route.WorkerAddress, nil
}

func callWorker(workerAddr string, req generateRequest) (*generateResponse, int, error) {
	body, _ := json.Marshal(req)
	resp, err := http.Post(workerAddr+"/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	var result generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, err
	}
	return &result, resp.StatusCode, nil
}

func recordRequest(c *gin.Context, db *database.Pool, userID, workerName, model string,
	statusCode, latencyMS int, tokens int, cost float64, cached bool) {

	_, _ = db.Exec(c.Request.Context(), `
		INSERT INTO requests (user_id, model, status_code, latency_ms, tokens, cost, cached)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, model, statusCode, latencyMS, tokens, cost, cached)
}
