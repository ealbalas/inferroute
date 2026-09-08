package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/inferroute/inferroute/internal/database"
)

// Requests handles GET /v1/requests — paginated request history for the caller.
func Requests(db *database.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		limit := 50
		if l := c.Query("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
				limit = n
			}
		}
		offset := 0
		if o := c.Query("offset"); o != "" {
			if n, err := strconv.Atoi(o); err == nil && n >= 0 {
				offset = n
			}
		}

		rows, err := db.Query(c.Request.Context(), `
			SELECT id, worker_id, model, status_code, latency_ms, tokens, cost, cached, created_at
			FROM requests
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, userID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		defer rows.Close()

		type requestRow struct {
			ID         string  `json:"id"`
			WorkerID   string  `json:"worker_id"`
			Model      string  `json:"model"`
			StatusCode int     `json:"status_code"`
			LatencyMS  int     `json:"latency_ms"`
			Tokens     int     `json:"tokens"`
			Cost       float64 `json:"cost"`
			Cached     bool    `json:"cached"`
			CreatedAt  string  `json:"created_at"`
		}
		var reqs []requestRow
		for rows.Next() {
			var r requestRow
			if err := rows.Scan(&r.ID, &r.WorkerID, &r.Model, &r.StatusCode,
				&r.LatencyMS, &r.Tokens, &r.Cost, &r.Cached, &r.CreatedAt); err != nil {
				continue
			}
			reqs = append(reqs, r)
		}
		c.JSON(http.StatusOK, gin.H{"requests": reqs, "limit": limit, "offset": offset})
	}
}
