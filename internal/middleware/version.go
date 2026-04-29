package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIVersion enforces the X-API-Version header on all /api/* routes.
func APIVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := c.GetHeader("X-API-Version")
		if version == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "API version header required",
			})
			return
		}
		c.Next()
	}
}
