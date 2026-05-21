package handlers

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type AdminUserService interface {
	ListUsers(filter repositories.UserFilter, page, limit int) ([]models.User, int64, error)
	GetUser(id string) (*models.User, error)
	CreateUser(req *dtos.AdminCreateUserRequest, createdBy string) (*models.User, error)
	UpdateUser(id string, req *dtos.AdminUpdateUserRequest, updatedBy string) (*models.User, error)
	BlockUser(id string, updatedBy string) error
	UnblockUser(id string, updatedBy string) error
	DeleteUser(id string, deletedBy string) error
}

type AdminUserHandler struct {
	svc AdminUserService
}

func NewAdminUserHandler(svc AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{svc: svc}
}


func adminCurrentUser(c *echo.Context) *configs.JwtCustomClaims {
	v := c.Get("user")
	if v == nil {
		return nil
	}
	claims, _ := v.(*configs.JwtCustomClaims)
	return claims
}

func adminFlashURL(flash, msg string) string {
	return "/admin/users?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func actorID(c *echo.Context) string {
	if u := adminCurrentUser(c); u != nil {
		return u.ID
	}
	return ""
}

func extractFieldErrors(c *echo.Context, err error) map[string]string {
	var ve *configs.ValidatorError
	if !errors.As(err, &ve) {
		return nil
	}
	out := make(map[string]string, len(ve.Messages))
	for _, m := range ve.Messages {
		params := make(map[string]string, len(m.Params)+1)
		for k, v := range m.Params {
			params[k] = v
		}
		params["field"] = m.Field
		out[m.Field] = configs.T(c, m.Key, params)
	}
	return out
}

func (h *AdminUserHandler) ListUsers(c *echo.Context) error {
	search := c.QueryParam("search")
	role := c.QueryParam("role")
	page, limit := parsePagination(c)

	filter := repositories.UserFilter{Search: search, Role: role}
	users, total, err := h.svc.ListUsers(filter, page, limit)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.users.title", nil),
		"CurrentPath": "/admin/users",
		"CurrentUser": adminCurrentUser(c),
		"Users":       users,
		"Pagination":  utils.NewPagination(page, limit, total),
		"Search":      search,
		"RoleFilter":  role,
		"Flash":       flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/users/list.html", data)
}

func (h *AdminUserHandler) ShowCreateForm(c *echo.Context) error {
	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.users.form.create_title", nil),
		"CurrentPath": "/admin/users",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      false,
	}
	return c.Render(http.StatusOK, "admin/pages/users/form.html", data)
}

func (h *AdminUserHandler) CreateUser(c *echo.Context) error {
	req := new(dtos.AdminCreateUserRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, extractFieldErrors(c, err), "")
	}

	_, err := h.svc.CreateUser(req, actorID(c))
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			msg = configs.T(c, "auth.email_exists", nil)
		}
		return h.renderFormErrors(c, false, nil, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_created", nil)))
}

func (h *AdminUserHandler) ShowUser(c *echo.Context) error {
	id := c.Param("id")
	user, err := h.svc.GetUser(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "auth.user_not_found", nil)))
	}

	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.users.detail.title", nil),
		"CurrentPath": "/admin/users",
		"CurrentUser": adminCurrentUser(c),
		"User":        user,
	}
	return c.Render(http.StatusOK, "admin/pages/users/detail.html", data)
}

func (h *AdminUserHandler) ShowEditForm(c *echo.Context) error {
	id := c.Param("id")
	user, err := h.svc.GetUser(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "auth.user_not_found", nil)))
	}

	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.users.form.edit_title", nil),
		"CurrentPath": "/admin/users",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      true,
		"User":        user,
	}
	return c.Render(http.StatusOK, "admin/pages/users/form.html", data)
}

func (h *AdminUserHandler) UpdateUser(c *echo.Context) error {
	id := c.Param("id")

	user, err := h.svc.GetUser(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "auth.user_not_found", nil)))
	}

	req := new(dtos.AdminUpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, true, user, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, true, user, req, extractFieldErrors(c, err), "")
	}

	if _, err := h.svc.UpdateUser(id, req, actorID(c)); err != nil {
		return h.renderFormErrors(c, true, user, req, nil, configs.T(c, "common.internal_error", nil))
	}

	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_updated", nil)))
}

func (h *AdminUserHandler) BlockUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.BlockUser(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "ui.msg.block_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_blocked", nil)))
}

func (h *AdminUserHandler) UnblockUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.UnblockUser(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "ui.msg.unblock_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_unblocked", nil)))
}

func (h *AdminUserHandler) DeleteUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteUser(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "ui.msg.delete_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_deleted", nil)))
}

func (h *AdminUserHandler) renderFormErrors(c *echo.Context, isEdit bool, user *models.User, req interface{}, fieldErrors map[string]string, globalError string) error {
	titleKey := "ui.users.form.create_title"
	if isEdit {
		titleKey = "ui.users.form.edit_title"
	}
	data := map[string]interface{}{
		"Title":       configs.T(c, titleKey, nil),
		"CurrentPath": "/admin/users",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      isEdit,
		"User":        user,
		"Req":         req,
		"FieldErrors": fieldErrors,
		"Error":       globalError,
	}
	return c.Render(http.StatusUnprocessableEntity, "admin/pages/users/form.html", data)
}

func flashFromQuery(c *echo.Context) map[string]string {
	t := c.QueryParam("flash")
	msg := c.QueryParam("msg")
	if t == "" || msg == "" {
		return nil
	}
	return map[string]string{"Type": t, "Message": msg}
}
