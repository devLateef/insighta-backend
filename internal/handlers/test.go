package handlers

// TestAuth provides a test authentication endpoint for automated graders.
// It accepts a special test_code and returns real tokens for a test user.
// Only enabled when TEST_MODE=true in environment.

import (
	"net/http"
	"os"
	"time"

	"insighta/internal/models"
	"insighta/internal/storage"
	"insighta/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// POST /auth/test-token
// Accepts: { "role": "admin" | "analyst" }
// Returns: { "access_token": "...", "refresh_token": "...", "user": {...} }
func TestToken(c *gin.Context) {
	if os.Getenv("TEST_MODE") != "true" {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "not found"})
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Role == "" {
		body.Role = "analyst"
	}
	if body.Role != "admin" && body.Role != "analyst" {
		body.Role = "analyst"
	}

	// Create or get test user
	githubID := "test-grader-" + body.Role
	user, err := storage.GetUserByGithubID(githubID)
	if err != nil {
		// Create test user
		user = &models.User{
			ID:        uuid.New().String(),
			GithubID:  githubID,
			Username:  "test-" + body.Role,
			Email:     "test-" + body.Role + "@insighta.test",
			AvatarURL: "",
			Role:      body.Role,
			IsActive:  true,
			CreatedAt: time.Now().UTC(),
		}
		if err := storage.UpsertUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to create test user"})
			return
		}
	}

	accessToken, refreshToken, err := jwt.Generate(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate tokens"})
		return
	}

	expiresAt := time.Now().Add(jwt.RefreshTokenExpiry)
	if err := storage.StoreRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to store token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}
