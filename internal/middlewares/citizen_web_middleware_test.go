package middlewares

import (
	"net/http"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func TestCitizenWebMiddlewareRedirectsWithoutCookie(t *testing.T) {
	c, rec := newMiddlewareContext(http.MethodGet, "/dashboard")

	handler := CitizenWebMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Fatalf("expected /login, got %q", loc)
	}
}

func TestCitizenWebMiddlewareRedirectsWithInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	c, rec := newMiddlewareContext(http.MethodGet, "/dashboard")
	c.Request().AddCookie(&http.Cookie{Name: "refresh_token", Value: "invalid-token"})

	handler := CitizenWebMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect %d, got %d", http.StatusSeeOther, rec.Code)
	}
}

func TestCitizenWebMiddlewareRedirectsForNonCitizenRole(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := configs.GenerateRefreshToken(&models.User{
		ID:    "admin-id",
		Email: "admin@example.com",
		Role:  models.UserRoleSuperAdmin,
	})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	c, rec := newMiddlewareContext(http.MethodGet, "/dashboard")
	c.Request().AddCookie(&http.Cookie{Name: "refresh_token", Value: token})

	handler := CitizenWebMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Fatalf("expected /login, got %q", loc)
	}
}

func TestCitizenWebMiddlewareAllowsValidCitizenToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	token, err := configs.GenerateRefreshToken(&models.User{
		ID:    "citizen-id",
		Email: "citizen@example.com",
		Role:  models.UserRoleCitizen,
	})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	c, rec := newMiddlewareContext(http.MethodGet, "/dashboard")
	c.Request().AddCookie(&http.Cookie{Name: "refresh_token", Value: token})

	called := false
	handler := CitizenWebMiddleware(func(c *echo.Context) error {
		called = true
		claims, ok := c.Get("user").(*configs.JwtCustomClaims)
		if !ok {
			t.Fatal("expected JWT claims in context")
		}
		if claims.Role != string(models.UserRoleCitizen) {
			t.Fatalf("expected citizen role, got %q", claims.Role)
		}
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
