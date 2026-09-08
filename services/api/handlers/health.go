package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health handles GET /health — used by load balancers and Docker health checks.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "0.1.0",
		"service": "api",
	})
}
