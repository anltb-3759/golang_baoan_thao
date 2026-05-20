package middlewares

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

func LocaleMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		locale := configs.DefaultLocale

		if cookieLang, err := c.Cookie("lang"); err == nil && cookieLang.Value != "" {
			locale = configs.NormalizeLocale(cookieLang.Value)
		}

		c.Set(configs.LocaleKey, locale)
		return next(c)
	}
}
