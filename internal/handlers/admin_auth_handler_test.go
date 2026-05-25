package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

func TestAdminAuthHandlerWebLoginWritesActivityLog(t *testing.T) {
	e := newTestEcho()
	logger := &fakeActivityLogger{}
	handler := NewAdminAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "admin-id", Email: reqData.Email, Role: models.UserRoleSuperAdmin}, "access-token", "refresh-token", nil
		},
	}).WithActivityLogger(logger)

	form := "email=admin%40example.com&password=123456"
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.WebLogin(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if logger.calls != 1 {
		t.Fatalf("expected 1 activity log call, got %d", logger.calls)
	}
	if logger.lastLog == nil || logger.lastLog.Action != "auth.login" || logger.lastLog.Result != "success" {
		t.Fatalf("expected auth.login success log, got %#v", logger.lastLog)
	}
}

func TestAdminAuthHandlerWebLoginSucceedsWhenActivityLogFails(t *testing.T) {
	e := newTestEcho()
	logger := &fakeActivityLogger{returnErr: errors.New("log failed")}
	handler := NewAdminAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "admin-id", Email: reqData.Email, Role: models.UserRoleSuperAdmin}, "access-token", "refresh-token", nil
		},
	}).WithActivityLogger(logger)

	form := "email=admin%40example.com&password=123456"
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.WebLogin(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}
}

func TestAdminAuthHandlerWebLoginRejectsNonAdminRole(t *testing.T) {
	e := newAdminEcho()
	handler := NewAdminAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "citizen-id", Email: reqData.Email, Role: models.UserRoleCitizen}, "access-token", "refresh-token", nil
		},
	})

	form := "email=user%40example.com&password=123456"
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.WebLogin(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if setCookie := rec.Header().Get(echo.HeaderSetCookie); setCookie != "" {
		t.Fatalf("expected no refresh cookie for non-admin login, got %q", setCookie)
	}
}

func TestAdminAuthHandlerWebLoginReturnsSafeErrorWhenUserIsNil(t *testing.T) {
	e := newTestEcho()
	handler := NewAdminAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "access-token", "refresh-token", nil
		},
	})

	form := "email=admin%40example.com&password=123456"
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.WebLogin(c)
	assertHTTPError(t, err, http.StatusInternalServerError, "common.internal_error")
}

func TestAdminAuthHandlerWebLogoutWritesActivityLog(t *testing.T) {
	if err := configs.LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	e := newTestEcho()
	logger := &fakeActivityLogger{}
	handler := NewAdminAuthHandler(&fakeAuthService{}).WithActivityLogger(logger)

	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(configs.LocaleKey, "en")
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-id", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.WebLogout(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if logger.calls != 1 {
		t.Fatalf("expected 1 activity log call, got %d", logger.calls)
	}
	if logger.lastLog == nil || logger.lastLog.Action != "auth.logout" || logger.lastLog.Result != "success" {
		t.Fatalf("expected auth.logout success log, got %#v", logger.lastLog)
	}
}

func TestAdminAuthHandlerWebLogoutSucceedsWhenActivityLogFails(t *testing.T) {
	if err := configs.LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	e := newTestEcho()
	logger := &fakeActivityLogger{returnErr: errors.New("log failed")}
	handler := NewAdminAuthHandler(&fakeAuthService{}).WithActivityLogger(logger)

	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(configs.LocaleKey, "en")
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-id", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.WebLogout(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}
}

func TestAdminAuthHandlerSetLocaleRejectsOpenRedirect(t *testing.T) {
	e := newTestEcho()
	handler := NewAdminAuthHandler(&fakeAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/locale?lang=en", nil)
	req.Header.Set("Referer", "https://evil.example.com/phish")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.SetLocale(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}
	if loc := rec.Header().Get(echo.HeaderLocation); loc != "/admin/login" {
		t.Fatalf("expected safe fallback redirect, got %q", loc)
	}
}
