package middleware

import (
	"net/http"
	"strings"

	"insighta/internal/storage"
	"insighta/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuth validates the access token from Authorization header or access_token cookie.
// It also checks that the user is still active in the database.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "authentication required",
			})
			return
		}

		claims, err := jwt.Validate(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "invalid or expired token",
			})
			return
		}

		if claims.Type != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "not an access token",
			})
			return
		}

		// Check user is still active
		user, err := storage.GetUserByID(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "user not found",
			})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "account is disabled",
			})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", user.Role) // always use DB role, not token role
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// 1. Authorization header
	header := c.GetHeader("Authorization")
	if header != "" {
		return strings.TrimPrefix(header, "Bearer ")
	}

	// 2. HTTP-only cookie (web portal)
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}

	return ""
}
