package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger attaches a request_id to every request and emits a structured
// log line on completion.
func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		userID, _ := c.Get(userIDKey)
		log.Info("request",
			"request_id", requestID,
			"user_id", userID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", latency.Milliseconds(),
			"ip", c.ClientIP(),
		)
	}
}
