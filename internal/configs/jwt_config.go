package configs

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

type JwtCustomClaims struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

const (
	AccessTokenType  = "access"
	RefreshTokenType = "refresh"

	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

func LoadJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Fatal("JWT_SECRET is not set")
		}
		jwtSecret = []byte(secret)
	})
	return jwtSecret
}

func GenerateToken(user *models.User) (string, error) {
	return GenerateAccessToken(user)
}

func GenerateAccessToken(user *models.User) (string, error) {
	return generateToken(user, AccessTokenType, AccessTokenDuration)
}

func GenerateRefreshToken(user *models.User) (string, error) {
	return generateToken(user, RefreshTokenType, RefreshTokenDuration)
}

func GenerateTokenPair(user *models.User) (accessToken string, refreshToken string, err error) {
	accessToken, err = GenerateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = GenerateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func generateToken(user *models.User, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &JwtCustomClaims{
		ID:        user.ID,
		Email:     user.Email,
		TokenType: tokenType,
		Role:      string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(LoadJWTSecret())
}

func UserIDFromContext(c *echo.Context) string {
	claims, ok := c.Get("user").(*JwtCustomClaims)
	if !ok || claims == nil {
		return ""
	}
	return claims.ID
}

func ClaimsFromContext(c *echo.Context) *JwtCustomClaims {
	claims, _ := c.Get("user").(*JwtCustomClaims)
	return claims
}

func ParseToken(tokenString string, expectedTokenType string) (*JwtCustomClaims, error) {
	claims := &JwtCustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return LoadJWTSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid || claims.TokenType != expectedTokenType {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
