package middlewares

import (
	"net/http"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func TestAdminSessionMiddlewareRedirectsWithoutCookie(t *testing.T) {
	c, rec := newMiddlewareContext(http.MethodGet, "/admin/service-types")

	handler := AdminSessionMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect status %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/admin/login" {
		t.Fatalf("expected login redirect, got %q", location)
	}
}

func TestAdminSessionMiddlewareAllowsValidRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	refreshToken, err := configs.GenerateRefreshToken(&models.User{
		ID:    "user-id",
		Email: "admin@example.com",
		Role:  models.UserRoleSuperAdmin,
	})
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	c, rec := newMiddlewareContext(http.MethodGet, "/admin/service-types")
	c.Request().AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})

	called := false
	handler := AdminSessionMiddleware(func(c *echo.Context) error {
		called = true
		claims, ok := c.Get("user").(*configs.JwtCustomClaims)
		if !ok {
			t.Fatal("expected JWT claims in context")
		}
		if claims.Role != string(models.UserRoleSuperAdmin) {
			t.Fatalf("expected role %q, got %q", models.UserRoleSuperAdmin, claims.Role)
		}
		return c.NoContent(http.StatusNoContent)
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
