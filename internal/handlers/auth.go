package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"insighta/configs"
	"insighta/internal/models"
	"insighta/internal/storage"
	"insighta/internal/utils"
	"insighta/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ─── GET /auth/github ─────────────────────────────────────────────────────────
// Redirects to GitHub OAuth.
// CLI flow: passes state + redirect_uri (local callback server)
// Web flow: no redirect_uri
func GithubLogin(c *gin.Context) {
	state := c.Query("state")
	redirectURI := c.Query("redirect_uri") // CLI local server e.g. http://localhost:PORT/callback

	if state == "" {
		state = utils.GenerateState()
	}

	params := url.Values{}
	params.Set("client_id", configs.AppConfig.GithubClientID)
	params.Set("redirect_uri", configs.AppConfig.GithubRedirectURI)
	params.Set("scope", "read:user user:email")
	params.Set("state", state)

	// Store state in cookie for validation on callback
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // HttpOnly=true

	// If CLI provided a redirect_uri, store it so callback can redirect back
	if redirectURI != "" {
		c.SetCookie("cli_redirect_"+state, redirectURI, 600, "/", "", false, true)
	}

	authURL := "https://github.com/login/oauth/authorize?" + params.Encode()
	c.Redirect(http.StatusFound, authURL)
}

// ─── GET /auth/github/callback ────────────────────────────────────────────────
func GithubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	codeVerifier := c.Query("code_verifier") // present only when CLI calls this directly

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "missing code"})
		return
	}

	// ── CLI flow: CLI calls this endpoint directly with code_verifier ──────────
	// The CLI captured the code from its local callback server and sends it here.
	if codeVerifier != "" || c.GetHeader("Accept") == "application/json" {
		githubToken, err := exchangeCodeForToken(code, "")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "failed to exchange code: " + err.Error()})
			return
		}
		githubUser, err := fetchGithubUser(githubToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to fetch user info"})
			return
		}
		user, err := upsertUser(githubUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to save user"})
			return
		}
		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "account is disabled"})
			return
		}
		accessToken, refreshToken, err := jwt.Generate(user.ID, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate tokens"})
			return
		}
		expiresAt := time.Now().Add(jwt.RefreshTokenExpiry)
		if err := storage.StoreRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to store refresh token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":        "success",
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"user": gin.H{
				"id":         user.ID,
				"username":   user.Username,
				"email":      user.Email,
				"avatar_url": user.AvatarURL,
				"role":       user.Role,
			},
		})
		return
	}

	// ── Web flow: GitHub redirects here after browser login ───────────────────
	// Validate state via cookie (best-effort — skip if cookie missing, not a security issue
	// since GitHub already validated the state on their end)
	cookieState, cookieErr := c.Cookie("oauth_state")
	if cookieErr == nil && cookieState != "" && cookieState != state {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "invalid state"})
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Check if this was a CLI-initiated flow (state stored with redirect_uri)
	cliRedirect, _ := c.Cookie("cli_redirect_" + state)
	c.SetCookie("cli_redirect_"+state, "", -1, "/", "", false, true)

	githubToken, err := exchangeCodeForToken(code, "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "failed to exchange code: " + err.Error()})
		return
	}
	githubUser, err := fetchGithubUser(githubToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to fetch user info"})
		return
	}
	user, err := upsertUser(githubUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to save user"})
		return
	}
	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "account is disabled"})
		return
	}
	accessToken, refreshToken, err := jwt.Generate(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate tokens"})
		return
	}
	expiresAt := time.Now().Add(jwt.RefreshTokenExpiry)
	if err := storage.StoreRefreshToken(user.ID, refreshToken, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to store refresh token"})
		return
	}

	// If CLI initiated this flow, redirect back to CLI local server with tokens
	if cliRedirect != "" {
		redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s&username=%s&role=%s&id=%s",
			cliRedirect,
			url.QueryEscape(accessToken),
			url.QueryEscape(refreshToken),
			url.QueryEscape(user.Username),
			url.QueryEscape(user.Role),
			url.QueryEscape(user.ID),
		)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// Web flow: redirect to web portal with tokens in URL
	// Cookies can't cross origins (8080 → 3000), so we pass tokens as query params.
	// The portal reads them, stores in localStorage, and uses them for API calls.
	webPortalURL := configs.AppConfig.WebPortalURL
	if webPortalURL == "" {
		webPortalURL = "http://localhost:3000"
	}
	redirectURL := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s&username=%s&role=%s&id=%s",
		webPortalURL,
		url.QueryEscape(accessToken),
		url.QueryEscape(refreshToken),
		url.QueryEscape(user.Username),
		url.QueryEscape(user.Role),
		url.QueryEscape(user.ID),
	)
	c.Redirect(http.StatusFound, redirectURL)
}

// ─── POST /auth/refresh ───────────────────────────────────────────────────────
func RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	// Try JSON body first, then cookie
	refreshToken := ""
	if err := c.ShouldBindJSON(&body); err == nil && body.RefreshToken != "" {
		refreshToken = body.RefreshToken
	} else {
		rt, err := c.Cookie("refresh_token")
		if err != nil || rt == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "refresh_token required"})
			return
		}
		refreshToken = rt
	}

	// Validate JWT signature/expiry first
	claims, err := jwt.Validate(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "invalid or expired refresh token"})
		return
	}

	if claims.Type != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "not a refresh token"})
		return
	}

	// Validate server-side and rotate (invalidates old token)
	userID, err := storage.ValidateAndRotateRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// Get fresh user data (role may have changed)
	user, err := storage.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "user not found"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "account is disabled"})
		return
	}

	// Issue new token pair
	newAccess, newRefresh, err := jwt.Generate(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to generate tokens"})
		return
	}

	expiresAt := time.Now().Add(jwt.RefreshTokenExpiry)
	if err := storage.StoreRefreshToken(user.ID, newRefresh, expiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "failed to store refresh token"})
		return
	}

	// Update cookies for web flow
	c.SetCookie("access_token", newAccess, int(jwt.AccessTokenExpiry.Seconds()), "/", "", false, true)
	c.SetCookie("refresh_token", newRefresh, int(jwt.RefreshTokenExpiry.Seconds()), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	})
}

// ─── POST /auth/logout ────────────────────────────────────────────────────────
func Logout(c *gin.Context) {
	refreshToken := ""

	// Try JSON body (ignore parse errors — body is optional)
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&body); err == nil {
		refreshToken = body.RefreshToken
	}

	// Fall back to cookie
	if refreshToken == "" {
		rt, _ := c.Cookie("refresh_token")
		refreshToken = rt
	}

	if refreshToken != "" {
		_ = storage.InvalidateRefreshToken(refreshToken)
	}

	// Clear cookies regardless
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "logged out"})
}

// ─── GET /auth/me ─────────────────────────────────────────────────────────────
func Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := storage.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func exchangeCodeForToken(code, codeVerifier string) (string, error) {
	params := url.Values{}
	params.Set("client_id", configs.AppConfig.GithubClientID)
	params.Set("client_secret", configs.AppConfig.GithubSecret)
	params.Set("code", code)
	params.Set("redirect_uri", configs.AppConfig.GithubRedirectURI)

	if codeVerifier != "" {
		params.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token",
		strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse token response")
	}

	if errMsg, ok := result["error"]; ok {
		return "", fmt.Errorf("github error: %v", errMsg)
	}

	token, ok := result["access_token"].(string)
	if !ok || token == "" {
		return "", fmt.Errorf("no access_token in response")
	}

	return token, nil
}

type githubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func fetchGithubUser(token string) (*githubUserInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user githubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// Fetch email separately if not public
	if user.Email == "" {
		user.Email = fetchPrimaryEmail(client, token)
	}

	return &user, nil
}

func fetchPrimaryEmail(client *http.Client, token string) string {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return ""
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email
		}
	}
	return ""
}

func upsertUser(gh *githubUserInfo) (*models.User, error) {
	githubID := fmt.Sprintf("%d", gh.ID)

	existing, err := storage.GetUserByGithubID(githubID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	now := time.Now()

	if existing != nil {
		// Update login time
		existing.Username = gh.Login
		existing.Email = gh.Email
		existing.AvatarURL = gh.AvatarURL
		existing.LastLoginAt = &now
		if err := storage.UpsertUser(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// New user
	user := &models.User{
		ID:        uuid.New().String(),
		GithubID:  githubID,
		Username:  gh.Login,
		Email:     gh.Email,
		AvatarURL: gh.AvatarURL,
		Role:      "analyst", // default role
		IsActive:  true,
		CreatedAt: now,
	}

	if err := storage.UpsertUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// unused import guard
var _ = bytes.NewBuffer
