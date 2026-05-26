package handlers

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

type adminProfileService interface {
	GetSelf(userID string) (*models.User, error)
	UpdateSelfContact(userID string, req *dtos.UpdateAdminProfileRequest) (*models.User, error)
	ChangeSelfPassword(userID string, req *dtos.ChangeAdminPasswordRequest) error
}

type AdminProfileHandler struct {
	svc    adminProfileService
	logger ActivityLogger
}

func NewAdminProfileHandler(svc adminProfileService) *AdminProfileHandler {
	return &AdminProfileHandler{svc: svc}
}

func (h *AdminProfileHandler) WithActivityLogger(logger ActivityLogger) *AdminProfileHandler {
	h.logger = logger
	return h
}

func (h *AdminProfileHandler) writeActivityLog(log *models.ActivityLog) {
	if h.logger == nil || log == nil {
		return
	}
	_ = h.logger.Log(log)
}

func adminProfileFlashURL(c *echo.Context, flash, msg string) string {
	return "/admin/profile?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func (h *AdminProfileHandler) ShowProfilePage(c *echo.Context) error {
	uid := actorID(c)
	user, err := h.svc.GetSelf(uid)
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrUserNotFound) {
			msg = configs.T(c, "auth.user_not_found", nil)
		}
		return c.Redirect(http.StatusSeeOther, "/admin?"+url.Values{"flash": {"error"}, "msg": {msg}}.Encode())
	}
	return c.Render(http.StatusOK, "admin/pages/profile/index.html", map[string]interface{}{
		"Title":       configs.T(c, "ui.profile.title", nil),
		"CurrentPath": "/admin/profile",
		"CurrentUser": adminCurrentUser(c),
		"User":        user,
		"Flash":       flashFromQuery(c),
	})
}

func (h *AdminProfileHandler) UpdateProfile(c *echo.Context) error {
	req := new(dtos.UpdateAdminProfileRequest)
	if err := c.Bind(req); err != nil {
		return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "error", configs.T(c, "common.invalid_data", nil)))
	}
	if err := c.Validate(req); err != nil {
		return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "error", configs.T(c, "common.invalid_data", nil)))
	}
	if _, err := h.svc.UpdateSelfContact(actorID(c), req); err != nil {
		return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "error", mapAdminProfileError(c, err)))
	}
	uid := actorID(c)
	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: &uid,
		Action:      "admin.profile.update_contact",
		EntityType:  "user",
		EntityID:    &uid,
		Description: "Admin updated own contact information",
		Result:      "success",
	})
	return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "success", configs.T(c, "ui.msg.profile_updated", nil)))
}

func (h *AdminProfileHandler) ChangePassword(c *echo.Context) error {
	req := new(dtos.ChangeAdminPasswordRequest)
	if err := c.Bind(req); err != nil {
		return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "error", configs.T(c, "common.invalid_data", nil)))
	}
	if err := c.Validate(req); err != nil {
		return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "error", configs.T(c, "common.invalid_data", nil)))
	}
	if err := h.svc.ChangeSelfPassword(actorID(c), req); err != nil {
		return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "error", mapAdminProfileError(c, err)))
	}
	uid := actorID(c)
	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: &uid,
		Action:      "admin.profile.change_password",
		EntityType:  "user",
		EntityID:    &uid,
		Description: "Admin changed own password",
		Result:      "success",
	})
	return c.Redirect(http.StatusSeeOther, adminProfileFlashURL(c, "success", configs.T(c, "profile.password_changed", nil)))
}

func mapAdminProfileError(c *echo.Context, err error) string {
	switch {
	case errors.Is(err, services.ErrUserNotFound):
		return configs.T(c, "auth.user_not_found", nil)
	case errors.Is(err, services.ErrPasswordMismatch):
		return configs.T(c, "auth.password_mismatch", nil)
	case errors.Is(err, services.ErrPasswordConfirmationMismatch):
		return configs.T(c, "auth.password_confirmation_mismatch", nil)
	case errors.Is(err, services.ErrNewPasswordMustDiffer):
		return configs.T(c, "auth.new_password_must_differ", nil)
	default:
		return configs.T(c, "common.internal_error", nil)
	}
}
