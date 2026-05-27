package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

type AdminAuthHandler struct {
	authService AuthService
	logger      ActivityLogger
}

func NewAdminAuthHandler(authService AuthService) *AdminAuthHandler {
	return &AdminAuthHandler{authService: authService}
}

func (h *AdminAuthHandler) WithActivityLogger(logger ActivityLogger) *AdminAuthHandler {
	h.logger = logger
	return h
}

func (h *AdminAuthHandler) writeActivityLog(log *models.ActivityLog) {
	if h.logger == nil || log == nil {
		return
	}
	_ = h.logger.Log(log)
}

// ShowLoginPage renders the HTML login page for admin users.
func (h *AdminAuthHandler) ShowLoginPage(c *echo.Context) error {
	if cookie, err := c.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		if claims, err := configs.ParseToken(cookie.Value, configs.RefreshTokenType); err == nil {
			if claims.Role != string(models.UserRoleCitizen) {
				return c.Redirect(http.StatusSeeOther, "/admin")
			}
		}
	}

	data := map[string]interface{}{
		"FullPage": true,
		"Title":    configs.T(c, "ui.form.admin_login", nil),
		"Flash":    flashFromQuery(c),
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

	user, _, refreshToken, err := h.authService.Login(reqData)
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
	if user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	if !isAdminRole(user.Role) {
		data := map[string]interface{}{
			"Error":    configs.T(c, "auth.user_not_found", nil),
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
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: &user.ID,
		Action:      "auth.login",
		EntityType:  "user",
		EntityID:    &user.ID,
		Description: "Admin web login successful",
		Result:      "success",
	})

	return c.Redirect(http.StatusSeeOther, "/admin")
}

func (h *AdminAuthHandler) WebLogout(c *echo.Context) error {
	var actorUserID *string
	if claims := adminCurrentUser(c); claims != nil && claims.ID != "" {
		actorUserID = &claims.ID
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: actorUserID,
		Action:      "auth.logout",
		EntityType:  "user",
		EntityID:    actorUserID,
		Description: "Admin web logout successful",
		Result:      "success",
	})
	return c.Redirect(http.StatusSeeOther, "/admin/login?"+url.Values{"flash": {"success"}, "msg": {configs.T(c, "auth.logout_success", nil)}}.Encode())
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
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	redirectTo := "/admin/login"
	rawReferer := strings.TrimSpace(c.Request().Header.Get("Referer"))
	if rawReferer != "" {
		if parsed, err := url.Parse(rawReferer); err == nil {
			reqHost := c.Request().Host

			isRelative := parsed.IsAbs() == false && strings.HasPrefix(rawReferer, "/")
			isSameHostAbs := parsed.IsAbs() && strings.EqualFold(parsed.Host, reqHost)

			if isRelative || isSameHostAbs {
				path := parsed.EscapedPath()
				if path != "" && strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//") {
					redirectTo = path
					if parsed.RawQuery != "" {
						redirectTo += "?" + parsed.RawQuery
					}
				}
			}
		}
	}
	return c.Redirect(http.StatusSeeOther, redirectTo)
}

func isAdminRole(role models.UserRole) bool {
	return role == models.UserRoleStaff || role == models.UserRoleManager || role == models.UserRoleSuperAdmin
}
