package middlewares

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

func AdminSessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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
