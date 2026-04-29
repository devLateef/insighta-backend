package jwt_test

import (
	"testing"
	"time"

	"insighta/configs"
	"insighta/pkg/jwt"
)

func init() {
	configs.AppConfig = &configs.Config{
		JWTSecret: "test-secret-key",
	}
}

func TestGenerateAndValidate(t *testing.T) {
	access, refresh, err := jwt.Generate("user-123", "analyst")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if access == "" || refresh == "" {
		t.Fatal("expected non-empty tokens")
	}

	// Validate access token
	claims, err := jwt.Validate(access)
	if err != nil {
		t.Fatalf("Validate access failed: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected user-123, got %s", claims.UserID)
	}
	if claims.Role != "analyst" {
		t.Errorf("expected analyst, got %s", claims.Role)
	}
	if claims.Type != "access" {
		t.Errorf("expected access type, got %s", claims.Type)
	}

	// Validate refresh token
	rclaims, err := jwt.Validate(refresh)
	if err != nil {
		t.Fatalf("Validate refresh failed: %v", err)
	}
	if rclaims.Type != "refresh" {
		t.Errorf("expected refresh type, got %s", rclaims.Type)
	}
}

func TestTokenExpiry(t *testing.T) {
	// Access token should expire in ~3 minutes
	_, _, err := jwt.Generate("user-123", "admin")
	if err != nil {
		t.Fatal(err)
	}

	if jwt.AccessTokenExpiry != 3*time.Minute {
		t.Errorf("expected 3m access expiry, got %v", jwt.AccessTokenExpiry)
	}
	if jwt.RefreshTokenExpiry != 5*time.Minute {
		t.Errorf("expected 5m refresh expiry, got %v", jwt.RefreshTokenExpiry)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	_, err := jwt.Validate("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestValidateWrongSecret(t *testing.T) {
	// Generate with one secret
	access, _, err := jwt.Generate("user-123", "analyst")
	if err != nil {
		t.Fatal(err)
	}

	// Swap secret
	configs.AppConfig.JWTSecret = "different-secret"
	defer func() { configs.AppConfig.JWTSecret = "test-secret-key" }()

	_, err = jwt.Validate(access)
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}
