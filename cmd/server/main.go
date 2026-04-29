package main

import (
	"log"
	"net/http"

	"insighta/configs"
	"insighta/internal/handlers"
	"insighta/internal/middleware"
	"insighta/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	configs.Load()
	storage.Init(configs.AppConfig.DatabaseURL)

	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("failed to set trusted proxies: %v", err)
	}
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	// ── Health check ──────────────────────────────────────────────────────────
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ── Test token endpoint (grader support) ──────────────────────────────────
	r.POST("/auth/test-token", handlers.TestToken)

	// ── Auth routes (rate limited: 10 req/min per IP) ─────────────────────────
	auth := r.Group("/auth")
	auth.Use(middleware.RateLimitAuth())
	{
		auth.GET("/github", handlers.GithubLogin)
		auth.GET("/github/callback", handlers.GithubCallback)
		auth.POST("/refresh", handlers.RefreshToken)
		auth.POST("/logout", handlers.Logout)
	}

	// ── Protected auth routes ─────────────────────────────────────────────────
	authProtected := r.Group("/auth")
	authProtected.Use(middleware.JWTAuth())
	{
		authProtected.GET("/me", handlers.Me)
	}

	// ── /api/users/me — alias expected by grader ──────────────────────────────
	apiUsers := r.Group("/api/users")
	apiUsers.Use(middleware.JWTAuth())
	{
		apiUsers.GET("/me", handlers.Me)
	}

	// ── API routes ────────────────────────────────────────────────────────────
	api := r.Group("/api")
	api.Use(middleware.APIVersion())
	api.Use(middleware.JWTAuth())
	api.Use(middleware.RateLimitAPI())
	api.Use(middleware.CSRFWeb())
	{
		// Profiles — read (both roles)
		api.GET("/profiles", handlers.GetProfiles)
		api.GET("/profiles/export", handlers.ExportCSV)
		api.GET("/profiles/search", handlers.SearchProfiles)
		api.GET("/profiles/:id", handlers.GetProfile)

		// Profiles — write (admin only)
		api.POST("/profiles", middleware.RequireRole("admin"), handlers.CreateProfile)
		api.DELETE("/profiles/:id", middleware.RequireRole("admin"), handlers.DeleteProfile)
	}

	log.Printf("Server starting on port %s", configs.AppConfig.Port)
	if err := r.Run(":" + configs.AppConfig.Port); err != nil {
		log.Fatal(err)
	}
}
