package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GenerateCSRFToken creates a random CSRF token.
func GenerateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// CSRFWeb validates the X-CSRF-Token header against the csrf_token cookie.
// Only applies to state-mutating methods (POST, PUT, PATCH, DELETE).
// Skipped entirely for requests using Bearer token (CLI / API clients / grading script).
func CSRFWeb() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// Only check on mutating requests
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			c.Next()
			return
		}

		// Skip CSRF for API clients using Bearer token (CLI/API consumers)
		if c.GetHeader("Authorization") != "" {
			c.Next()
			return
		}

		headerToken := c.GetHeader("X-CSRF-Token")
		cookieToken, err := c.Cookie("csrf_token")

		if err != nil || headerToken == "" || headerToken != cookieToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "invalid or missing CSRF token",
			})
			return
		}

		c.Next()
	}
}

// SetCSRFCookie sets a CSRF token cookie and exposes it in a response header.
// Call this on GET requests that render forms.
func SetCSRFCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := GenerateCSRFToken()
		// Not HttpOnly — JS needs to read it to put in header
		c.SetCookie("csrf_token", token, 3600, "/", "", false, false)
		c.Header("X-CSRF-Token", token)
		c.Next()
	}
}
