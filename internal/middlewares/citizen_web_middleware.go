package middlewares

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func CitizenWebMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cookie, err := c.Cookie("refresh_token")
		if err != nil || cookie.Value == "" {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		claims, err := configs.ParseToken(cookie.Value, configs.RefreshTokenType)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		if claims.Role != string(models.UserRoleCitizen) {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		c.Set("user", claims)
		return next(c)
	}
}
