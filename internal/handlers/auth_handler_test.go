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
	c, rec := newJSONContext(e, http.MethodPost, "/api/auth/register", `{"name":"User","email":"user@example.com","password":"123456"}`)

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
	c, _ := newJSONContext(e, http.MethodPost, "/api/auth/register", `{"name":"User","email":"user@example.com","password":"123456"}`)

	err := handler.Register(c)
	assertHTTPError(t, err, http.StatusConflict, "auth.email_exists")
}

func TestAuthHandlerLoginReturnsTokenAndRefreshCookie(t *testing.T) {
	e := newTestEcho()
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
	})
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
	handler := NewAuthHandler(&fakeAuthService{})
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
