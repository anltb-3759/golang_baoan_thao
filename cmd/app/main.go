package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/middlewares"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	// Load file .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := configs.LoadI18nMessages("locales"); err != nil {
		log.Fatalf("failed to load i18n messages: %v", err)
	}

	//  Database Connection & Migration
	configs.InitDB()

	e := echo.New()

	// Register Validator
	e.Validator = &configs.CustomValidator{Validator: validator.New()}

	// middleware
	configs.CustomLogger(e)
	e.Use(middleware.Recover())
	e.Use(middlewares.LocaleMiddleware)

	// error handler
	e.HTTPErrorHandler = configs.CustomHTTPErrorHandler

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sc := echo.StartConfig{
		Address:         ":8080",
		GracefulTimeout: 5 * time.Second,
	}

	if err := sc.Start(ctx, e); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
