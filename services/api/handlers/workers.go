package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Workers handles GET /v1/workers — proxies the router's worker list.
func Workers(routerURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := http.Get(routerURL + "/workers")
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "router unavailable"})
			return
		}
		defer resp.Body.Close()

		var workers any
		if err := json.NewDecoder(resp.Body).Decode(&workers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decode error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"workers": workers})
	}
}
