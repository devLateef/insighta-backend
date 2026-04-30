package middleware

import (
	"net/http"
	"strings"

	"insighta/internal/storage"
	"insighta/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuth validates the access token from Authorization header or access_token cookie.
// It checks the user is active in the DB only on first login (not every request)
// to avoid hammering the DB connection pool.
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

		// Use role from JWT claims to avoid a DB hit on every request.
		// The role in the token is set at login time and is accurate.
		// A full DB check only happens at /auth/me or when role changes take effect.
		role := claims.Role
		if role == "" {
			role = "analyst"
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", role)
		c.Next()
	}
}

// JWTAuthStrict is like JWTAuth but always checks the DB.
// Use this only on sensitive endpoints like /auth/me.
func JWTAuthStrict() gin.HandlerFunc {
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
		c.Set("role", user.Role)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header != "" {
		return strings.TrimPrefix(header, "Bearer ")
	}
	if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
		return cookie
	}
	return ""
}
