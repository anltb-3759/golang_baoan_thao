package middlewares

import (
	"net/http"
	"strings"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization format")
		}

		tokenString := parts[1]
		claims, err := configs.ParseToken(tokenString, configs.AccessTokenType)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
		}

		c.Set("user", claims)
		return next(c)
	}
}
