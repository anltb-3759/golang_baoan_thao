package middlewares

import (
	"net/http"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func TestJWTMiddlewareRejectsMissingAuthorizationHeader(t *testing.T) {
	c, _ := newMiddlewareContext(http.MethodGet, "/protected")

	handler := JWTMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	err := handler(c)
	assertHTTPErrorCode(t, err, http.StatusUnauthorized)
}

func TestJWTMiddlewareRejectsInvalidAuthorizationFormat(t *testing.T) {
	c, _ := newMiddlewareContext(http.MethodGet, "/protected")
	c.Request().Header.Set("Authorization", "Token invalid")

	handler := JWTMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	err := handler(c)
	assertHTTPErrorCode(t, err, http.StatusUnauthorized)
}

func TestJWTMiddlewareRejectsRefreshToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	refreshToken, err := configs.GenerateRefreshToken(&models.User{
		ID:    "user-id",
		Email: "user@example.com",
		Role:  models.UserRoleCitizen,
	})
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	c, _ := newMiddlewareContext(http.MethodGet, "/protected")
	c.Request().Header.Set("Authorization", "Bearer "+refreshToken)

	handler := JWTMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	err = handler(c)
	assertHTTPErrorCode(t, err, http.StatusUnauthorized)
}

func TestJWTMiddlewareAllowsValidAccessToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	accessToken, err := configs.GenerateAccessToken(&models.User{
		ID:    "user-id",
		Email: "user@example.com",
		Role:  models.UserRoleCitizen,
	})
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	c, rec := newMiddlewareContext(http.MethodGet, "/protected")
	c.Request().Header.Set("Authorization", "Bearer "+accessToken)

	called := false
	handler := JWTMiddleware(func(c *echo.Context) error {
		called = true
		claims, ok := c.Get("user").(*configs.JwtCustomClaims)
		if !ok {
			t.Fatal("expected JWT claims in context")
		}
		if claims.ID != "user-id" {
			t.Fatalf("expected user id %q, got %q", "user-id", claims.ID)
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

func TestJWTMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	c, _ := newMiddlewareContext(http.MethodGet, "/protected")
	c.Request().Header.Set("Authorization", "Bearer invalid-token")

	handler := JWTMiddleware(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	err := handler(c)
	assertHTTPErrorCode(t, err, http.StatusUnauthorized)
}
