package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("JWT_SECRET", "test-secret-for-middleware-tests")
}

func newMwEcho() *echo.Echo { return echo.New() }

func passHandler(c *echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func makeRefreshToken(user *models.User) string {
	token, _ := configs.GenerateRefreshToken(user)
	return token
}

// --- AdminWebMiddleware ---

func TestAdminWebMiddleware_NoCookie(t *testing.T) {
	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := AdminWebMiddleware(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/login", rec.Header().Get("Location"))
}

func TestAdminWebMiddleware_EmptyCookie(t *testing.T) {
	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: ""})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := AdminWebMiddleware(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminWebMiddleware_InvalidToken(t *testing.T) {
	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "not.a.valid.token"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := AdminWebMiddleware(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminWebMiddleware_ValidToken(t *testing.T) {
	user := &models.User{ID: "u1", Email: "admin@test.com", Role: models.UserRoleSuperAdmin}
	token := makeRefreshToken(user)

	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: token})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := AdminWebMiddleware(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	claims, ok := c.Get("user").(*configs.JwtCustomClaims)
	assert.True(t, ok)
	assert.Equal(t, "u1", claims.ID)
}

func TestAdminWebMiddleware_AccessTokenRejected(t *testing.T) {
	// Access token should be rejected (expects refresh token type)
	user := &models.User{ID: "u1", Email: "admin@test.com", Role: models.UserRoleSuperAdmin}
	token, _ := configs.GenerateAccessToken(user)

	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: token})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := AdminWebMiddleware(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- AdminWebRequireRoles ---

func TestAdminWebRequireRoles_NoClaims(t *testing.T) {
	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := AdminWebRequireRoles(models.UserRoleSuperAdmin)
	err := mw(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/login", rec.Header().Get("Location"))
}

func TestAdminWebRequireRoles_WrongRole(t *testing.T) {
	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "staff"})

	mw := AdminWebRequireRoles(models.UserRoleSuperAdmin)
	err := mw(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminWebRequireRoles_CorrectRole(t *testing.T) {
	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "super_admin"})

	mw := AdminWebRequireRoles(models.UserRoleSuperAdmin)
	err := mw(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminWebRequireRoles_MultipleRoles(t *testing.T) {
	e := newMwEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "manager"})

	mw := AdminWebRequireRoles(models.UserRoleSuperAdmin, models.UserRoleManager)
	err := mw(passHandler)(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
