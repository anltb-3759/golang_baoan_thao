package middlewares

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func AdminWebMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cookie, err := c.Cookie("refresh_token")
		if err != nil || cookie.Value == "" {
			return c.Redirect(http.StatusSeeOther, "/admin/login")
		}

		claims, err := configs.ParseToken(cookie.Value, configs.RefreshTokenType)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/admin/login")
		}

		c.Set("user", claims)
		return next(c)
	}
}

func AdminWebRequireRoles(roles ...models.UserRole) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims, ok := c.Get("user").(*configs.JwtCustomClaims)
			if !ok {
				return c.Redirect(http.StatusSeeOther, "/admin/login")
			}
			for _, role := range roles {
				if claims.Role == string(role) {
					return next(c)
				}
			}
			return c.Redirect(http.StatusSeeOther, "/admin/login")
		}
	}
}
