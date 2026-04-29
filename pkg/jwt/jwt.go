package jwt

import (
	"errors"
	"time"

	"insighta/configs"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// Token expiry per spec: access=3min, refresh=5min
const (
	AccessTokenExpiry  = 3 * time.Minute
	RefreshTokenExpiry = 5 * time.Minute
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Type   string `json:"type"` // "access" or "refresh"
	gojwt.RegisteredClaims
}

func Generate(userID, role string) (string, string, error) {
	secret := []byte(configs.AppConfig.JWTSecret)

	now := time.Now()

	accessClaims := Claims{
		UserID: userID,
		Role:   role,
		Type:   "access",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(AccessTokenExpiry)),
			IssuedAt:  gojwt.NewNumericDate(now),
		},
	}

	refreshClaims := Claims{
		UserID: userID,
		Role:   role,
		Type:   "refresh",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(RefreshTokenExpiry)),
			IssuedAt:  gojwt.NewNumericDate(now),
		},
	}

	at := gojwt.NewWithClaims(gojwt.SigningMethodHS256, accessClaims)
	rt := gojwt.NewWithClaims(gojwt.SigningMethodHS256, refreshClaims)

	accessToken, err := at.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := rt.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func Validate(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(configs.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
