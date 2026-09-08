package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inferroute/inferroute/internal/cache"
)

// RateLimit enforces a sliding-window rate limit using Redis.
// limit is the max requests allowed per window duration.
func RateLimit(redis *cache.Client, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(userIDKey)
		uid, _ := userID.(string)
		if uid == "" {
			c.Next()
			return
		}

		// Sliding window key: bucket by window start
		windowStart := time.Now().Truncate(window)
		key := cache.RateLimitKey(uid, windowStart)

		count, err := redis.IncrBy(c.Request.Context(), key, 1, window*2)
		if err != nil {
			// Fail open on Redis errors — don't block legitimate traffic
			c.Next()
			return
		}

		if count > limit {
			c.Header("X-RateLimit-Limit", "10")
			c.Header("X-RateLimit-Remaining", "0")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Header("X-RateLimit-Limit", "10")
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))
		c.Next()
	}
}
