package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/inferroute/inferroute/internal/database"
)

// Usage handles GET /v1/usage — returns rolling request and token usage for the caller.
func Usage(db *database.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		rows, err := db.Query(c.Request.Context(), `
			SELECT date, request_count, total_tokens, total_cost
			FROM usage_metrics
			WHERE user_id = $1
			ORDER BY date DESC
			LIMIT 30
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		defer rows.Close()

		type dayUsage struct {
			Date         string  `json:"date"`
			RequestCount int     `json:"request_count"`
			TotalTokens  int     `json:"total_tokens"`
			TotalCost    float64 `json:"total_cost"`
		}
		var usage []dayUsage
		for rows.Next() {
			var d dayUsage
			if err := rows.Scan(&d.Date, &d.RequestCount, &d.TotalTokens, &d.TotalCost); err != nil {
				continue
			}
			usage = append(usage, d)
		}
		c.JSON(http.StatusOK, gin.H{"usage": usage})
	}
}
