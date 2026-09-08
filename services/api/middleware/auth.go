package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/inferroute/inferroute/internal/auth"
)

const userIDKey = "userID"

// APIKeyAuth validates Bearer ir_live_* tokens against the database.
// It sets "userID" in the Gin context on success.
//
// In a real implementation this queries the api_keys table using FastKeyHash
// for the index lookup, then bcrypt-validates the full key.
func APIKeyAuth() gin.HandlerFunc {
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

		// TODO: look up api_keys by FastKeyHash(raw), then call auth.ValidateAPIKey(raw, row.KeyHash)
		_ = auth.FastKeyHash(raw)

		// Placeholder: set a synthetic user ID until DB lookup is wired up
		c.Set(userIDKey, "user_placeholder")
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
