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
		} else if header := c.Request().Header.Get("Accept-Language"); header != "" {
			locale = configs.NormalizeLocale(header)
		}

		c.Set(configs.LocaleKey, locale)
		return next(c)
	}
}
