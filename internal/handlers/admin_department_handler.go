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

type DepartmentService interface {
	ListDepartments(filter repositories.DepartmentFilter, page, limit int) ([]models.Department, int64, error)
	GetDepartment(id string) (*models.Department, error)
	CreateDepartment(req *dtos.DepartmentCreateRequest, createdBy string) (*models.Department, error)
	UpdateDepartment(id string, req *dtos.DepartmentUpdateRequest, updatedBy string) (*models.Department, error)
	DeleteDepartment(id string, deletedBy string) error
}

type AdminDepartmentHandler struct {
	svc     DepartmentService
	userSvc AdminUserService
}

func NewAdminDepartmentHandler(svc DepartmentService, userSvc AdminUserService) *AdminDepartmentHandler {
	return &AdminDepartmentHandler{svc: svc, userSvc: userSvc}
}

func deptFlashURL(flash, msg string) string {
	return "/admin/departments?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func (h *AdminDepartmentHandler) ListDepartments(c *echo.Context) error {
	search := c.QueryParam("search")
	page, limit := parsePagination(c)

	filter := repositories.DepartmentFilter{Search: search}
	depts, total, err := h.svc.ListDepartments(filter, page, limit)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.departments.title", nil),
		"CurrentPath": "/admin/departments",
		"CurrentUser": adminCurrentUser(c),
		"Departments": depts,
		"Pagination":  utils.NewPagination(page, limit, total),
		"Search":      search,
		"Flash":       flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/departments/list.html", data)
}

func (h *AdminDepartmentHandler) ShowCreateForm(c *echo.Context) error {
	staffUsers, err := h.listStaffUsers()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"Title":           configs.T(c, "ui.departments.form.create_title", nil),
		"CurrentPath":     "/admin/departments",
		"CurrentUser":     adminCurrentUser(c),
		"IsEdit":          false,
		"StaffUsers":      staffUsers,
		"CurrentLeaderID": "",
	}
	return c.Render(http.StatusOK, "admin/pages/departments/form.html", data)
}

func (h *AdminDepartmentHandler) CreateDepartment(c *echo.Context) error {
	req := new(dtos.DepartmentCreateRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, extractFieldErrors(c, err), "")
	}

	_, err := h.svc.CreateDepartment(req, actorID(c))
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrDepartmentCodeExists) {
			msg = configs.T(c, "department.code_exists", nil)
		}
		return h.renderFormErrors(c, false, nil, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, deptFlashURL("success", configs.T(c, "ui.msg.department_created", nil)))
}

func (h *AdminDepartmentHandler) ShowEditForm(c *echo.Context) error {
	id := c.Param("id")
	dept, err := h.svc.GetDepartment(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, deptFlashURL("error", configs.T(c, "department.not_found", nil)))
	}

	staffUsers, err := h.listStaffUsers()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"Title":           configs.T(c, "ui.departments.form.edit_title", nil),
		"CurrentPath":     "/admin/departments",
		"CurrentUser":     adminCurrentUser(c),
		"IsEdit":          true,
		"Department":      dept,
		"StaffUsers":      staffUsers,
		"CurrentLeaderID": derefStr(dept.LeaderUserID),
	}
	return c.Render(http.StatusOK, "admin/pages/departments/form.html", data)
}

func (h *AdminDepartmentHandler) UpdateDepartment(c *echo.Context) error {
	id := c.Param("id")

	dept, err := h.svc.GetDepartment(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, deptFlashURL("error", configs.T(c, "department.not_found", nil)))
	}

	req := new(dtos.DepartmentUpdateRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, true, dept, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, true, dept, req, extractFieldErrors(c, err), "")
	}

	if _, err := h.svc.UpdateDepartment(id, req, actorID(c)); err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrDepartmentCodeExists) {
			msg = configs.T(c, "department.code_exists", nil)
		}
		return h.renderFormErrors(c, true, dept, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, deptFlashURL("success", configs.T(c, "ui.msg.department_updated", nil)))
}

func (h *AdminDepartmentHandler) DeleteDepartment(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteDepartment(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, deptFlashURL("error", configs.T(c, "ui.msg.department_delete_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, deptFlashURL("success", configs.T(c, "ui.msg.department_deleted", nil)))
}

func (h *AdminDepartmentHandler) renderFormErrors(c *echo.Context, isEdit bool, dept *models.Department, req interface{}, fieldErrors map[string]string, globalError string) error {
	titleKey := "ui.departments.form.create_title"
	if isEdit {
		titleKey = "ui.departments.form.edit_title"
	}
	staffUsers, _ := h.listStaffUsers()
	curLeaderID := ""
	if dept != nil {
		curLeaderID = derefStr(dept.LeaderUserID)
	}
	data := map[string]interface{}{
		"Title":           configs.T(c, titleKey, nil),
		"CurrentPath":     "/admin/departments",
		"CurrentUser":     adminCurrentUser(c),
		"IsEdit":          isEdit,
		"Department":      dept,
		"Req":             req,
		"FieldErrors":     fieldErrors,
		"Error":           globalError,
		"StaffUsers":      staffUsers,
		"CurrentLeaderID": curLeaderID,
	}
	return c.Render(http.StatusUnprocessableEntity, "admin/pages/departments/form.html", data)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h *AdminDepartmentHandler) listStaffUsers() ([]models.User, error) {
	users, _, err := h.userSvc.ListUsers(repositories.UserFilter{}, 1, 1000)
	if err != nil {
		return nil, err
	}
	var staff []models.User
	for _, u := range users {
		if u.Role == models.UserRoleStaff || u.Role == models.UserRoleManager || u.Role == models.UserRoleSuperAdmin {
			staff = append(staff, u)
		}
	}
	return staff, nil
}
