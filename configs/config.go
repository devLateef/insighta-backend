package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	JWTSecret         string
	GithubClientID    string
	GithubSecret      string
	BaseURL           string
	DatabaseURL       string
	CallbackURL       string
	GithubRedirectURI string
	WebPortalURL      string
	TestMode          bool
}

var AppConfig *Config

func Load() {
	_ = godotenv.Load()

	AppConfig = &Config{
		Port:           getEnv("PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "supersecret-change-in-production"),
		GithubClientID: os.Getenv("GITHUB_CLIENT_ID"),
		GithubSecret:   os.Getenv("GITHUB_SECRET"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/insighta?sslmode=disable"),
		WebPortalURL:   getEnv("WEB_PORTAL_URL", "http://localhost:3000"),
		TestMode:       os.Getenv("TEST_MODE") == "true",
	}

	AppConfig.GithubRedirectURI = AppConfig.BaseURL + "/auth/github/callback"

	log.Println("Config loaded")
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
