package configs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func init() {
	os.Setenv("JWT_SECRET", "test-secret-key-for-configs")
}

func newEchoCtx(t *testing.T) *echo.Context {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestGenerateAndParseAccessToken(t *testing.T) {
	user := &models.User{ID: "u1", Email: "user@example.com", Role: models.UserRoleCitizen}
	token, err := GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := ParseToken(token, AccessTokenType)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.ID != "u1" {
		t.Fatalf("expected ID u1, got %q", claims.ID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email, got %q", claims.Email)
	}
}

func TestGenerateToken_IsSameAsAccessToken(t *testing.T) {
	user := &models.User{ID: "u2", Email: "u2@example.com", Role: models.UserRoleStaff}
	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	claims, err := ParseToken(token, AccessTokenType)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.ID != "u2" {
		t.Fatalf("expected ID u2, got %q", claims.ID)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	user := &models.User{ID: "u3", Email: "u3@example.com", Role: models.UserRoleCitizen}
	token, err := GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("GenerateRefreshToken error: %v", err)
	}

	claims, err := ParseToken(token, RefreshTokenType)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.TokenType != RefreshTokenType {
		t.Fatalf("expected token type %q, got %q", RefreshTokenType, claims.TokenType)
	}
}

func TestGenerateTokenPair(t *testing.T) {
	user := &models.User{ID: "u4", Email: "u4@example.com", Role: models.UserRoleSuperAdmin}
	access, refresh, err := GenerateTokenPair(user)
	if err != nil {
		t.Fatalf("GenerateTokenPair error: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}

	aClaims, err := ParseToken(access, AccessTokenType)
	if err != nil || aClaims.ID != "u4" {
		t.Fatalf("access token invalid: %v", err)
	}

	rClaims, err := ParseToken(refresh, RefreshTokenType)
	if err != nil || rClaims.ID != "u4" {
		t.Fatalf("refresh token invalid: %v", err)
	}
}

func TestParseToken_WrongType(t *testing.T) {
	user := &models.User{ID: "u5", Email: "u5@example.com", Role: models.UserRoleCitizen}
	access, _ := GenerateAccessToken(user)

	_, err := ParseToken(access, RefreshTokenType)
	if err == nil {
		t.Fatal("expected error when token type mismatch")
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	_, err := ParseToken("not-a-valid-token", AccessTokenType)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestUserIDFromContext_WithClaims(t *testing.T) {
	c := newEchoCtx(t)
	c.Set("user", &JwtCustomClaims{ID: "user-id-123"})

	id := UserIDFromContext(c)
	if id != "user-id-123" {
		t.Fatalf("expected user-id-123, got %q", id)
	}
}

func TestUserIDFromContext_NoClaims(t *testing.T) {
	c := newEchoCtx(t)

	id := UserIDFromContext(c)
	if id != "" {
		t.Fatalf("expected empty ID, got %q", id)
	}
}

func TestClaimsFromContext_WithClaims(t *testing.T) {
	c := newEchoCtx(t)
	expected := &JwtCustomClaims{ID: "u1", Role: "citizen"}
	c.Set("user", expected)

	got := ClaimsFromContext(c)
	if got == nil || got.ID != "u1" {
		t.Fatalf("expected claims with ID u1, got %#v", got)
	}
}

func TestClaimsFromContext_NoClaims(t *testing.T) {
	c := newEchoCtx(t)

	got := ClaimsFromContext(c)
	if got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestCustomLogger_DoesNotPanic(t *testing.T) {
	e := echo.New()
	CustomLogger(e)
	// CustomLogger registers request logging middleware — just verify it doesn't panic
}

func TestCustomLogger_RequestLogging(t *testing.T) {
	e := echo.New()
	CustomLogger(e)

	// Success path: v.Error == nil
	e.GET("/ok", func(c *echo.Context) error {
		return c.String(200, "ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCustomLogger_RequestLogging_Error(t *testing.T) {
	e := echo.New()
	CustomLogger(e)

	// Error path: v.Error != nil (generic error)
	e.GET("/fail", func(c *echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "server error")
	})
	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestCustomLogger_RequestLogging_HTTPErrorWithWrapped(t *testing.T) {
	e := echo.New()
	CustomLogger(e)

	// HTTPError with wrapped inner error → Unwrap branch
	e.GET("/wrapped", func(c *echo.Context) error {
		inner := errors.New("inner cause")
		return echo.NewHTTPError(http.StatusBadRequest, "wrapped").Wrap(inner)
	})
	req := httptest.NewRequest(http.MethodGet, "/wrapped", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
