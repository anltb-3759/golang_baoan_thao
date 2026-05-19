package configs

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func CustomLogger(e *echo.Echo) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		HandleError: true,
		LogMethod:   true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {

			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("method", v.Method),
				)
			} else {
				errLog := v.Error.Error()
				if he, ok := v.Error.(*echo.HTTPError); ok {
					if internalErr := errors.Unwrap(he); internalErr != nil {
						errLog = internalErr.Error()
					}
				}

				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", errLog),
				)
			}
			return nil
		},
	}))
}
