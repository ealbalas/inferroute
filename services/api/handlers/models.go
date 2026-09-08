package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type modelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxTokens   int    `json:"max_tokens"`
}

// Models handles GET /v1/models — returns the list of available inference models.
func Models(c *gin.Context) {
	models := []modelInfo{
		{
			ID:          "inferroute-sim-v1",
			Name:        "InferRoute Simulated v1",
			Description: "Simulated inference for development and load testing.",
			MaxTokens:   4096,
		},
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}
