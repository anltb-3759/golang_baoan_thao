package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

type AdminAuthHandler struct {
	authService AuthService
}

func NewAdminAuthHandler(authService AuthService) *AdminAuthHandler {
	return &AdminAuthHandler{authService: authService}
}

// ShowLoginPage renders the HTML login page for admin users.
func (h *AdminAuthHandler) ShowLoginPage(c *echo.Context) error {

	data := map[string]interface{}{
		"FullPage": true,
		"Title":    configs.T(c, "ui.form.admin_login", nil),
	}
	return c.Render(http.StatusOK, "admin/pages/auth/login.html", data)
}

// WebLogin handles form POST from the login page and redirects on success.
func (h *AdminAuthHandler) WebLogin(c *echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	reqData := &dtos.LoginRequest{
		Email:    email,
		Password: password,
	}

	_, _, refreshToken, err := h.authService.Login(reqData)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			msg = configs.T(c, "auth.user_not_found", nil)
		case errors.Is(err, services.ErrPasswordMismatch):
			msg = configs.T(c, "auth.password_mismatch", nil)
		case errors.Is(err, services.ErrUserBlocked):
			msg = configs.T(c, "auth.user_blocked", nil)
		default:
			msg = configs.T(c, "common.internal_error", nil)
		}
		data := map[string]interface{}{
			"Error":    msg,
			"Email":    email,
			"FullPage": true,
			"Title":    configs.T(c, "ui.form.admin_login", nil),
		}
		return c.Render(http.StatusUnauthorized, "admin/pages/auth/login.html", data)
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
