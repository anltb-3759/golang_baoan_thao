package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func TestRequireRolesAllowsMatchingRole(t *testing.T) {
	c, rec := newMiddlewareContext(http.MethodGet, "/")
	c.Set("user", &configs.JwtCustomClaims{
		Role: string(models.UserRoleStaff),
	})

	called := false
	handler := RequireRoles(models.UserRoleStaff)(func(c *echo.Context) error {
		called = true
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

func TestRequireRolesAllowsOneOfMultipleRoles(t *testing.T) {
	c, _ := newMiddlewareContext(http.MethodGet, "/")
	c.Set("user", &configs.JwtCustomClaims{
		Role: string(models.UserRoleManager),
	})

	called := false
	handler := RequireRoles(models.UserRoleStaff, models.UserRoleManager)(func(c *echo.Context) error {
		called = true
		return nil
	})

	if err := handler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestRequireRolesRejectsMissingUserClaims(t *testing.T) {
	c, _ := newMiddlewareContext(http.MethodGet, "/")

	handler := RequireRoles(models.UserRoleStaff)(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	err := handler(c)
	assertHTTPErrorCode(t, err, http.StatusUnauthorized)
}

func TestRequireRolesRejectsWrongRole(t *testing.T) {
	c, _ := newMiddlewareContext(http.MethodGet, "/")
	c.Set("user", &configs.JwtCustomClaims{
		Role: string(models.UserRoleCitizen),
	})

	handler := RequireRoles(models.UserRoleStaff)(func(c *echo.Context) error {
		t.Fatal("next handler should not be called")
		return nil
	})

	err := handler(c)
	assertHTTPErrorCode(t, err, http.StatusForbidden)
}

func newMiddlewareContext(method string, target string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()

	return e.NewContext(req, rec), rec
}

func assertHTTPErrorCode(t *testing.T, err error, code int) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected HTTP error with status %d, got nil", code)
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != code {
		t.Fatalf("expected status %d, got %d", code, httpErr.Code)
	}
}
