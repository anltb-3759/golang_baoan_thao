package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type fakeAuthService struct {
	registerFn func(reqData *dtos.RegisterRequest) (*models.User, error)
	loginFn    func(reqData *dtos.LoginRequest) (*models.User, string, string, error)
}

type fakeActivityLogger struct {
	logFn      func(log *models.ActivityLog) error
	calls      int
	lastLog    *models.ActivityLog
	returnErr  error
}

func (l *fakeActivityLogger) Log(log *models.ActivityLog) error {
	l.calls++
	l.lastLog = log
	if l.logFn != nil {
		return l.logFn(log)
	}
	return l.returnErr
}

func (s *fakeAuthService) Register(reqData *dtos.RegisterRequest) (*models.User, error) {
	return s.registerFn(reqData)
}

func (s *fakeAuthService) Login(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
	return s.loginFn(reqData)
}

func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &configs.CustomValidator{Validator: validator.New()}
	return e
}

func newJSONContext(e *echo.Echo, method string, path string, body string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestAuthHandlerRegisterReturnsCreatedUser(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		registerFn: func(reqData *dtos.RegisterRequest) (*models.User, error) {
			if reqData.Email != "user@example.com" {
				t.Fatalf("expected email to be bound, got %q", reqData.Email)
			}
			return &models.User{
				ID:    "user-id",
				Name:  reqData.Name,
				Email: reqData.Email,
				Role:  models.UserRoleCitizen,
			}, nil
		},
	})
	c, rec := newJSONContext(e, http.MethodPost, "/api/auth/register", `{"name":"User","email":"user@example.com","password":"123456","citizen_id_number":"123456789012"}`)

	if err := handler.Register(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response map[string]models.User
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["user"].Email != "user@example.com" {
		t.Fatalf("expected user email in response, got %q", response["user"].Email)
	}
}

func TestAuthHandlerRegisterMapsEmailExistsError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		registerFn: func(reqData *dtos.RegisterRequest) (*models.User, error) {
			return nil, services.ErrEmailAlreadyExists
		},
	})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/register", `{"name":"User","email":"user@example.com","password":"123456","citizen_id_number":"123456789012"}`)

	err := handler.Register(c)
	assertHTTPError(t, err, http.StatusConflict, "auth.email_exists")
}

func TestAuthHandlerLoginReturnsTokenAndRefreshCookie(t *testing.T) {
	e := newTestEcho()
	logger := &fakeActivityLogger{}
	handler := NewAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			if reqData.Email != "user@example.com" {
				t.Fatalf("expected email to be bound, got %q", reqData.Email)
			}
			return &models.User{
				ID:    "user-id",
				Email: reqData.Email,
				Role:  models.UserRoleCitizen,
			}, "access-token", "refresh-token", nil
		},
	}).WithActivityLogger(logger)
	c, rec := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"123456"}`)

	if err := handler.Login(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if setCookie := rec.Header().Get(echo.HeaderSetCookie); !strings.Contains(setCookie, "refresh_token=refresh-token") || !strings.Contains(setCookie, "HttpOnly") {
		t.Fatalf("expected refresh token HttpOnly cookie, got %q", setCookie)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["token"] != "access-token" {
		t.Fatalf("expected access token in response, got %#v", response["token"])
	}
	if logger.calls != 1 {
		t.Fatalf("expected 1 activity log call, got %d", logger.calls)
	}
	if logger.lastLog == nil || logger.lastLog.Action != "auth.login" || logger.lastLog.Result != "success" {
		t.Fatalf("expected auth.login success log, got %#v", logger.lastLog)
	}
}

func TestAuthHandlerRegisterValidateError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{})
	// missing required "name" field → validation fails
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/register", `{"email":"user@example.com","password":"123456"}`)

	err := handler.Register(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthHandlerLoginValidateError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{})
	// missing "password" field → validation fails
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"user@example.com"}`)

	err := handler.Login(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuthHandlerRegisterBindError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/register", `{bad json`)

	err := handler.Register(c)
	assertHTTPError(t, err, http.StatusBadRequest, "auth.invalid_request")
}

func TestAuthHandlerRegisterInternalError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		registerFn: func(reqData *dtos.RegisterRequest) (*models.User, error) {
			return nil, errors.New("unexpected error")
		},
	})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/register", `{"name":"User","email":"user@example.com","password":"123456","citizen_id_number":"123456789012"}`)

	err := handler.Register(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// non-HTTPError should propagate as-is (Echo maps it to 500)
	var he *echo.HTTPError
	if errors.As(err, &he) {
		t.Fatalf("expected raw error, got HTTPError %d", he.Code)
	}
}

func TestAuthHandlerLoginBindError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/login", `{bad json`)

	err := handler.Login(c)
	assertHTTPError(t, err, http.StatusBadRequest, "auth.invalid_request")
}

func TestAuthHandlerLoginUserNotFound(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrUserNotFound
		},
	})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"missing@example.com","password":"123456"}`)

	err := handler.Login(c)
	assertHTTPError(t, err, http.StatusUnauthorized, "auth.user_not_found")
}

func TestAuthHandlerLoginInternalError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", errors.New("unexpected")
		},
	})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"123456"}`)

	err := handler.Login(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var he *echo.HTTPError
	if errors.As(err, &he) {
		t.Fatalf("expected raw error, got HTTPError %d", he.Code)
	}
}

func TestAuthHandlerLoginReturnsSafeErrorWhenUserIsNil(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "access-token", "refresh-token", nil
		},
	})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"123456"}`)

	err := handler.Login(c)
	assertHTTPError(t, err, http.StatusInternalServerError, "common.internal_error")
}

func TestAuthHandlerRefreshTokenInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "not-a-valid-token"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.RefreshTokenHandler(c)
	assertHTTPError(t, err, http.StatusUnauthorized, "auth.invalid_refresh_token")
}

func TestAuthHandlerLoginMapsPasswordMismatchError(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrPasswordMismatch
		},
	})
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"wrong-password"}`)

	err := handler.Login(c)
	assertHTTPError(t, err, http.StatusUnauthorized, "auth.password_mismatch")
}

func TestAuthHandlerRefreshTokenReturnsUnauthorizedWhenCookieMissing(t *testing.T) {
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.RefreshTokenHandler(c)
	assertHTTPError(t, err, http.StatusUnauthorized, "auth.missing_refresh_token")
}

func TestAuthHandlerRefreshTokenReturnsAccessToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{})
	refreshToken, err := configs.GenerateRefreshToken(&models.User{
		ID:    "user-id",
		Email: "user@example.com",
		Role:  models.UserRoleCitizen,
	})
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.RefreshTokenHandler(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["token"] == "" {
		t.Fatal("expected access token in response")
	}
}

func TestAuthHandlerLogoutClearsRefreshCookie(t *testing.T) {
	if err := configs.LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	e := newTestEcho()
	logger := &fakeActivityLogger{}
	handler := NewAuthHandler(&fakeAuthService{}).WithActivityLogger(logger)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "refresh-token",
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(configs.LocaleKey, "en")

	if err := handler.Logout(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	setCookie := rec.Header().Get(echo.HeaderSetCookie)
	if !strings.Contains(setCookie, "refresh_token=") || !strings.Contains(setCookie, "Max-Age=0") || !strings.Contains(setCookie, "HttpOnly") {
		t.Fatalf("expected cleared refresh token HttpOnly cookie, got %q", setCookie)
	}

	var response map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["message"] != "Logout successfully" {
		t.Fatalf("expected logout success message, got %q", response["message"])
	}
	if logger.calls != 1 {
		t.Fatalf("expected 1 activity log call, got %d", logger.calls)
	}
	if logger.lastLog == nil || logger.lastLog.Action != "auth.logout" || logger.lastLog.Result != "success" {
		t.Fatalf("expected auth.logout success log, got %#v", logger.lastLog)
	}
}

func TestAuthHandlerLoginSucceedsWhenActivityLogFails(t *testing.T) {
	e := newTestEcho()
	logger := &fakeActivityLogger{returnErr: errors.New("log failed")}
	handler := NewAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "user-id", Email: reqData.Email, Role: models.UserRoleCitizen}, "access-token", "refresh-token", nil
		},
	}).WithActivityLogger(logger)
	c, rec := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"123456"}`)

	if err := handler.Login(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuthHandlerLogoutSucceedsWhenActivityLogFails(t *testing.T) {
	if err := configs.LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	t.Setenv("JWT_SECRET", "test-secret")
	e := newTestEcho()
	logger := &fakeActivityLogger{returnErr: errors.New("log failed")}
	handler := NewAuthHandler(&fakeAuthService{}).WithActivityLogger(logger)
	refreshToken, err := configs.GenerateRefreshToken(&models.User{
		ID:    "user-id",
		Email: "user@example.com",
		Role:  models.UserRoleCitizen,
	})
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(configs.LocaleKey, "en")

	if err := handler.Logout(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAuthHandlerLogin_NoLogger_WriteActivityLogNilPath(t *testing.T) {
	// Tests that writeActivityLog with nil logger returns safely (no panic)
	e := newTestEcho()
	handler := NewAuthHandler(&fakeAuthService{
		loginFn: func(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "user-id", Email: reqData.Email, Role: models.UserRoleCitizen}, "tok", "ref", nil
		},
	}) // no WithActivityLogger → h.logger == nil
	c, rec := newJSONContext(e, http.MethodPost, "/api/auth/login", `{"email":"user@example.com","password":"123456"}`)

	if err := handler.Login(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func assertHTTPError(t *testing.T, err error, code int, message string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected HTTP error with status %d, got nil", code)
	}

	var httpErr *echo.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != code {
		t.Fatalf("expected status %d, got %d", code, httpErr.Code)
	}
	if httpErr.Message != message {
		t.Fatalf("expected message %q, got %#v", message, httpErr.Message)
	}
}
