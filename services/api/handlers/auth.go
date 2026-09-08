package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/inferroute/inferroute/internal/auth"
	"github.com/inferroute/inferroute/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Register handles POST /v1/auth/register.
func Register(db *database.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}

		var userID string
		err = db.QueryRow(c.Request.Context(), `
			INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id
		`, req.Email, string(hash)).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}

		token, _ := auth.SignToken(userID, req.Email, "developer", jwtSecret, 24*time.Hour)
		c.JSON(http.StatusCreated, gin.H{"token": token, "user_id": userID})
	}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /v1/auth/login.
func Login(db *database.Pool, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userID, hash, role string
		err := db.QueryRow(c.Request.Context(), `
			SELECT id, password_hash, role FROM users WHERE email = $1
		`, req.Email).Scan(&userID, &hash, &role)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, _ := auth.SignToken(userID, req.Email, role, jwtSecret, 24*time.Hour)
		c.JSON(http.StatusOK, gin.H{"token": token, "user_id": userID})
	}
}

// ListAPIKeys handles GET /v1/keys.
func ListAPIKeys(db *database.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		rows, err := db.Query(c.Request.Context(), `
			SELECT id, name, last_used_at, created_at FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		defer rows.Close()

		type keyRow struct {
			ID         string  `json:"id"`
			Name       string  `json:"name"`
			LastUsedAt *string `json:"last_used_at"`
			CreatedAt  string  `json:"created_at"`
		}
		var keys []keyRow
		for rows.Next() {
			var k keyRow
			if err := rows.Scan(&k.ID, &k.Name, &k.LastUsedAt, &k.CreatedAt); err == nil {
				keys = append(keys, k)
			}
		}
		c.JSON(http.StatusOK, gin.H{"keys": keys})
	}
}

type createKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateAPIKey handles POST /v1/keys — generates a new API key.
// The raw key is returned once and never stored.
func CreateAPIKey(db *database.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, _ := c.Get("userID")

		raw, err := auth.GenerateAPIKey()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "key generation failed"})
			return
		}
		hashed, err := auth.HashAPIKey(raw)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "key hashing failed"})
			return
		}

		keyID := uuid.New().String()
		_, err = db.Exec(c.Request.Context(), `
			INSERT INTO api_keys (id, user_id, key_hash, name) VALUES ($1, $2, $3, $4)
		`, keyID, userID, hashed, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":   keyID,
			"name": req.Name,
			"key":  raw, // shown once only
		})
	}
}

// DeleteAPIKey handles DELETE /v1/keys/:id.
func DeleteAPIKey(db *database.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		keyID := c.Param("id")

		result, err := db.Exec(c.Request.Context(), `
			DELETE FROM api_keys WHERE id = $1 AND user_id = $2
		`, keyID, userID)
		if err != nil || result.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": keyID})
	}
}
