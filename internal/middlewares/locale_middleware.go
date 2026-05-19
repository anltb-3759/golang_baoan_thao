package middlewares

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

func LocaleMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		locale := c.Request().Header.Get("Accept-Language")
		c.Set(configs.LocaleKey, configs.NormalizeLocale(locale))

		return next(c)
	}
}
