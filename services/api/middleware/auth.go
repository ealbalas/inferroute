package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inferroute/inferroute/internal/auth"
	"github.com/inferroute/inferroute/internal/database"
)

const userIDKey = "userID"

// APIKeyAuth validates Bearer ir_live_* tokens against the database.
// It performs a fast SHA-256 index lookup then bcrypt verification,
// and sets "userID" in the Gin context on success.
func APIKeyAuth(db *database.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		raw := strings.TrimPrefix(header, "Bearer ")
		if !strings.HasPrefix(raw, "ir_live_") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key format"})
			return
		}

		fastHash := auth.FastKeyHash(raw)

		var keyID, userID, storedHash string
		err := db.QueryRow(c.Request.Context(), `
			SELECT id, user_id, key_hash FROM api_keys WHERE key_hash_fast = $1
		`, fastHash).Scan(&keyID, &userID, &storedHash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		if !auth.ValidateAPIKey(raw, storedHash) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}

		c.Set(userIDKey, userID)

		// Best-effort, non-blocking update of last_used_at.
		go func() {
			_, _ = db.Exec(context.Background(), `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, keyID)
		}()

		c.Next()
	}
}

// JWTAuth validates dashboard user JWTs. Sets "userID", "email", "role".
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		raw := strings.TrimPrefix(header, "Bearer ")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := auth.ParseToken(raw, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}
