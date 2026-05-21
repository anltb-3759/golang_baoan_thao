package middlewares

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

func LocaleMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		locale := configs.DefaultLocale

		// Prefer cookie if set (user choice), otherwise fall back to Accept-Language header
		if cookieLang, err := c.Cookie("lang"); err == nil && cookieLang.Value != "" {
			locale = configs.NormalizeLocale(cookieLang.Value)
		} else if al := c.Request().Header.Get("Accept-Language"); al != "" {
			locale = configs.NormalizeLocale(al)
		}

		c.Set(configs.LocaleKey, locale)
		return next(c)
	}
}
