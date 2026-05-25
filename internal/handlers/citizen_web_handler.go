package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type citizenNotificationSvc interface {
	List(userID string, filter repositories.NotificationFilter, page, limit int) ([]dtos.NotificationResponse, int64, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
	CountUnread(userID string) (int64, error)
}

type CitizenWebHandler struct {
	authService     AuthService
	logger          ActivityLogger
	notificationSvc citizenNotificationSvc
}

func NewCitizenWebHandler(authService AuthService) *CitizenWebHandler {
	return &CitizenWebHandler{authService: authService}
}

func (h *CitizenWebHandler) WithActivityLogger(logger ActivityLogger) *CitizenWebHandler {
	h.logger = logger
	return h
}

func (h *CitizenWebHandler) WithNotificationService(svc citizenNotificationSvc) *CitizenWebHandler {
	h.notificationSvc = svc
	return h
}

func (h *CitizenWebHandler) unreadCount(userID string) int64 {
	if h.notificationSvc == nil || userID == "" {
		return 0
	}
	n, err := h.notificationSvc.CountUnread(userID)
	if err != nil {
		return 0
	}
	return n
}

func (h *CitizenWebHandler) writeActivityLog(log *models.ActivityLog) {
	if h.logger == nil || log == nil {
		return
	}
	_ = h.logger.Log(log)
}

func citizenCurrentUser(c *echo.Context) *configs.JwtCustomClaims {
	v := c.Get("user")
	if v == nil {
		return nil
	}
	claims, _ := v.(*configs.JwtCustomClaims)
	return claims
}

func (h *CitizenWebHandler) ShowLoginPage(c *echo.Context) error {
	data := map[string]interface{}{
		"FullPage": true,
		"Title":    configs.T(c, "ui.form.citizen_login", nil),
		"Flash":    flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "citizen/pages/auth/login.html", data)
}

func (h *CitizenWebHandler) WebLogin(c *echo.Context) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := c.FormValue("password")

	reqData := &dtos.LoginRequest{Email: email, Password: password}

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
		return c.Render(http.StatusUnauthorized, "citizen/pages/auth/login.html", map[string]interface{}{
			"Error":    msg,
			"Email":    email,
			"FullPage": true,
			"Title":    configs.T(c, "ui.form.citizen_login", nil),
		})
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	if user.Role != models.UserRoleCitizen {
		return c.Render(http.StatusUnauthorized, "citizen/pages/auth/login.html", map[string]interface{}{
			"Error":    configs.T(c, "auth.user_not_found", nil),
			"Email":    email,
			"FullPage": true,
			"Title":    configs.T(c, "ui.form.citizen_login", nil),
		})
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
		Description: "Citizen web login successful",
		Result:      "success",
	})

	return c.Redirect(http.StatusSeeOther, "/citizen")
}

func (h *CitizenWebHandler) ShowRegisterPage(c *echo.Context) error {
	return c.Render(http.StatusOK, "citizen/pages/auth/register.html", map[string]interface{}{
		"FullPage": true,
		"Title":    configs.T(c, "ui.form.citizen_register", nil),
	})
}

func (h *CitizenWebHandler) WebRegister(c *echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	cccd := strings.TrimSpace(c.FormValue("citizen_id_number"))
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	renderError := func(msg string) error {
		return c.Render(http.StatusUnprocessableEntity, "citizen/pages/auth/register.html", map[string]interface{}{
			"FullPage": true,
			"Title":    configs.T(c, "ui.form.citizen_register", nil),
			"Error":    msg,
			"Name":     name,
			"Email":    email,
			"CCCD":     cccd,
		})
	}

	if name == "" || email == "" || cccd == "" || password == "" {
		return renderError(configs.T(c, "auth.invalid_request", nil))
	}
	if password != confirmPassword {
		return renderError(configs.T(c, "auth.password_mismatch", nil))
	}
	if len(cccd) != 12 || !isAllDigits(cccd) {
		return renderError(configs.T(c, "validation.invalid", map[string]string{"field": configs.T(c, "ui.form.cccd", nil)}))
	}
	if len(password) < 6 {
		return renderError(configs.T(c, "validation.min", map[string]string{"field": configs.T(c, "ui.form.password", nil), "param": "6"}))
	}

	reqData := &dtos.RegisterRequest{
		Name:            name,
		Email:           email,
		Password:        password,
		CitizenIDNumber: cccd,
	}

	user, err := h.authService.Register(reqData)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			msg = configs.T(c, "auth.email_exists", nil)
		default:
			msg = configs.T(c, "common.internal_error", nil)
		}
		return renderError(msg)
	}

	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: &user.ID,
		Action:      "auth.register",
		EntityType:  "user",
		EntityID:    &user.ID,
		Description: "Citizen registered via web",
		Result:      "success",
	})

	return c.Redirect(http.StatusSeeOther, "/login?"+url.Values{
		"flash": {"success"},
		"msg":   {configs.T(c, "auth.register_success", nil)},
	}.Encode())
}

func (h *CitizenWebHandler) WebLogout(c *echo.Context) error {
	var actorUserID *string
	if claims := citizenCurrentUser(c); claims != nil && claims.ID != "" {
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
		Description: "Citizen web logout successful",
		Result:      "success",
	})

	return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *CitizenWebHandler) ShowDashboard(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	var userID string
	if claims != nil {
		userID = claims.ID
	}
	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.citizen.dashboard.title", nil),
		"CurrentPath": "/citizen",
		"CurrentUser": claims,
		"UnreadCount": h.unreadCount(userID),
	}
	return c.Render(http.StatusOK, "citizen/pages/dashboard.html", data)
}

func (h *CitizenWebHandler) ListNotifications(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil || h.notificationSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	page, limit := parsePagination(c)

	filterRead := strings.TrimSpace(c.QueryParam("is_read"))
	filterType := strings.TrimSpace(c.QueryParam("type"))
	filter := repositories.NotificationFilter{Type: filterType}
	switch filterRead {
	case "true":
		v := true
		filter.IsRead = &v
	case "false":
		v := false
		filter.IsRead = &v
	}

	items, total, err := h.notificationSvc.List(claims.ID, filter, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	data := map[string]interface{}{
		"Title":         configs.T(c, "ui.nav.citizen.notifications", nil),
		"CurrentPath":   "/citizen/notifications",
		"CurrentUser":   claims,
		"UnreadCount":   h.unreadCount(claims.ID),
		"Notifications": items,
		"Pagination":    utils.NewPagination(page, limit, total),
		"FilterIsRead":  filterRead,
		"FilterType":    filterType,
		"Flash":         flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "citizen/pages/notifications/list.html", data)
}

func (h *CitizenWebHandler) MarkNotificationRead(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil || h.notificationSvc == nil {
		return c.Redirect(http.StatusSeeOther, "/citizen/notifications")
	}
	id := c.Param("id")
	if err := h.notificationSvc.MarkAsRead(id, claims.ID); err != nil {
		log.Printf("mark notification %s as read failed for user %s: %v", id, claims.ID, err)
	}
	return c.Redirect(http.StatusSeeOther, "/citizen/notifications")
}

func (h *CitizenWebHandler) MarkAllNotificationsRead(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil || h.notificationSvc == nil {
		return c.Redirect(http.StatusSeeOther, "/citizen/notifications")
	}
	if err := h.notificationSvc.MarkAllAsRead(claims.ID); err != nil {
		return c.Redirect(http.StatusSeeOther, "/citizen/notifications?"+url.Values{
			"flash": {"error"},
			"msg":   {configs.T(c, "common.internal_error", nil)},
		}.Encode())
	}
	return c.Redirect(http.StatusSeeOther, "/citizen/notifications?"+url.Values{
		"flash": {"success"},
		"msg":   {configs.T(c, "notification.all_marked_read", nil)},
	}.Encode())
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
