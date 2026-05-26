package middlewares

import (
	"net/http"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

func TestLocaleMiddlewareSetsCookieLocale(t *testing.T) {
	if err := configs.LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	c, _ := newMiddlewareContext(http.MethodGet, "/")
	c.Request().AddCookie(&http.Cookie{Name: "lang", Value: "en"})

	handler := LocaleMiddleware(func(c *echo.Context) error {
		locale, ok := c.Get(configs.LocaleKey).(string)
		if !ok {
			t.Fatal("expected locale in context")
		}
		if locale != "en" {
			t.Fatalf("expected locale %q from cookie, got %q", "en", locale)
		}
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestLocaleMiddlewareSetsRequestedLocale(t *testing.T) {
	if err := configs.LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	c, _ := newMiddlewareContext(http.MethodGet, "/")
	c.Request().Header.Set("Accept-Language", "en-US,en;q=0.9")

	handler := LocaleMiddleware(func(c *echo.Context) error {
		locale, ok := c.Get(configs.LocaleKey).(string)
		if !ok {
			t.Fatal("expected locale in context")
		}
		if locale != "en" {
			t.Fatalf("expected locale %q, got %q", "en", locale)
		}

		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestLocaleMiddlewareFallsBackToDefaultLocale(t *testing.T) {
	if err := configs.LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	c, _ := newMiddlewareContext(http.MethodGet, "/")
	c.Request().Header.Set("Accept-Language", "fr")

	handler := LocaleMiddleware(func(c *echo.Context) error {
		locale, ok := c.Get(configs.LocaleKey).(string)
		if !ok {
			t.Fatal("expected locale in context")
		}
		if locale != configs.DefaultLocale {
			t.Fatalf("expected locale %q, got %q", configs.DefaultLocale, locale)
		}

		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
