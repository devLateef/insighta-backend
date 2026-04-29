package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole enforces that the authenticated user has one of the allowed roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")

		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "insufficient permissions",
		})
	}
}
