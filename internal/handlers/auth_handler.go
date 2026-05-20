package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

const AGE_REFRESH_TOKEN = 7 * 24 * 60 * 60
const AGE_LANGUAGE_COOKIE = 365 * 24 * 60 * 60

type AuthHandler struct {
	authService AuthService
}

type AdminAuthHandler struct {
	authService AuthService
}

type AuthService interface {
	Register(reqData *dtos.RegisterRequest) (*models.User, error)
	Login(reqData *dtos.LoginRequest) (*models.User, string, string, error)
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func NewAdminAuthHandler(authService AuthService) *AdminAuthHandler {
	return &AdminAuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *echo.Context) error {
	reqData := new(dtos.RegisterRequest)
	if err := c.Bind(reqData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "auth.invalid_request").Wrap(err)
	}

	if err := c.Validate(reqData); err != nil {
		return err
	}

	user, err := h.authService.Register(reqData)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return err
	}

	return c.JSON(http.StatusCreated, utils.Map{
		"user": user,
	})
}

func (h *AuthHandler) Login(c *echo.Context) error {
	reqData := new(dtos.LoginRequest)
	if err := c.Bind(reqData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "auth.invalid_request").Wrap(err)
	}

	if err := c.Validate(reqData); err != nil {
		return err
	}

	user, token, refreshToken, err := h.authService.Login(reqData)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return echo.NewHTTPError(http.StatusUnauthorized, "auth.user_not_found")
		case errors.Is(err, services.ErrPasswordMismatch):
			return echo.NewHTTPError(http.StatusUnauthorized, "auth.password_mismatch")
		default:
			return err
		}
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   AGE_REFRESH_TOKEN,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return c.JSON(http.StatusOK, utils.Map{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) RefreshTokenHandler(c *echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "auth.missing_refresh_token")
	}

	claims, err := configs.ParseToken(cookie.Value, configs.RefreshTokenType)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "auth.invalid_refresh_token")
	}

	token, err := configs.GenerateAccessToken(&models.User{
		ID:    claims.ID,
		Email: claims.Email,
		Role:  models.UserRole(claims.Role),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{
		"token": token,
	})
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return c.JSON(http.StatusOK, utils.Map{
		"message": configs.T(c, "auth.logout_success", nil),
	})
}

// ShowLoginPage renders the HTML login page for admin users.
func (h *AdminAuthHandler) ShowLoginPage(c *echo.Context) error {
	data := map[string]interface{}{
		"Roles": []map[string]string{
			{"value": "super_admin", "label": "Super Admin"},
			{"value": "manager", "label": "Manager"},
			{"value": "staff", "label": "Staff"},
		},
	}
	return c.Render(http.StatusOK, "admin/pages/auth/login.html", data)
}

// WebLogin handles form POST from the login page and redirects on success.
func (h *AdminAuthHandler) WebLogin(c *echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	selectedRole := c.FormValue("role")

	reqData := &dtos.LoginRequest{
		Email:    email,
		Password: password,
	}

	user, _, refreshToken, err := h.authService.Login(reqData)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			msg = configs.T(c, "auth.user_not_found", nil)
		case errors.Is(err, services.ErrPasswordMismatch):
			msg = configs.T(c, "auth.password_mismatch", nil)
		default:
			msg = configs.T(c, "common.internal_error", nil)
		}
		data := map[string]interface{}{
			"Error": msg,
			"Email": email,
			"Role":  selectedRole,
		}
		return c.Render(http.StatusUnauthorized, "admin/pages/auth/login.html", data)
	}

	if selectedRole != "" && selectedRole != string(user.Role) {
		data := map[string]interface{}{
			"Error": configs.T(c, "auth.invalid_role", nil),
			"Email": email,
			"Role":  selectedRole,
		}
		return c.Render(http.StatusForbidden, "admin/pages/auth/login.html", data)
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   AGE_REFRESH_TOKEN,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusSeeOther, "/admin")
}

func (h *AdminAuthHandler) WebLogout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	return c.Redirect(http.StatusSeeOther, "/admin/login")
}

// SetLocale handles language switching by setting a cookie and redirecting back.
func (h *AdminAuthHandler) SetLocale(c *echo.Context) error {
	lang := c.QueryParam("lang")
	lang = configs.NormalizeLocale(lang)

	c.SetCookie(&http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		MaxAge:   AGE_LANGUAGE_COOKIE,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	referer := c.Request().Header.Get("Referer")
	if referer == "" {
		referer = "/admin/login"
	}
	return c.Redirect(http.StatusSeeOther, referer)
}
