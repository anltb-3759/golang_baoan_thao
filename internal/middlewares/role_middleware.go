package middlewares

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func RequireRoles(roles ...models.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims, ok := c.Get("user").(*configs.JwtCustomClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}

			for _, role := range roles {
				if claims.Role == string(role) {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}
	}
}
