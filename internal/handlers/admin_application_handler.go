package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type AdminApplicationService interface {
	ListApplications(page, limit int) ([]models.Application, int64, error)
	GetApplication(id string) (*models.Application, error)
	AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error
}

type AssignableStaffService interface {
	ListStaffByDepartment(deptID string, page, limit int) ([]models.StaffProfile, int64, error)
}

type AdminApplicationHandler struct {
	svc             AdminApplicationService
	userSvc         AdminUserService
	staffProfileSvc AssignableStaffService
}

func NewAdminApplicationHandler(svc AdminApplicationService, userSvc AdminUserService, staffProfileSvc AssignableStaffService) *AdminApplicationHandler {
	return &AdminApplicationHandler{svc: svc, userSvc: userSvc, staffProfileSvc: staffProfileSvc}
}

func adminAppFlashURL(flash, msg string) string {
	return "/admin/applications?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func (h *AdminApplicationHandler) ListApplications(c *echo.Context) error {
	page, limit := parsePagination(c)
	apps, total, err := h.svc.ListApplications(page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	data := map[string]interface{}{
		"Title":        configs.T(c, "ui.applications.title", nil),
		"CurrentPath":  "/admin/applications",
		"CurrentUser":  adminCurrentUser(c),
		"Applications": apps,
		"Pagination":   utils.NewPagination(page, limit, total),
		"Flash":        flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/applications/list.html", data)
}

func (h *AdminApplicationHandler) ShowApplication(c *echo.Context) error {
	id := c.Param("id")
	app, err := h.svc.GetApplication(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "application.not_found")
	}
	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.applications.detail_title", nil),
		"CurrentPath": "/admin/applications",
		"CurrentUser": adminCurrentUser(c),
		"Application": app,
	}
	return c.Render(http.StatusOK, "admin/pages/applications/detail.html", data)
}

func (h *AdminApplicationHandler) ShowAssignForm(c *echo.Context) error {
	id := c.Param("id")
	app, err := h.svc.GetApplication(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "application.not_found")
	}
	staffUsers, err := h.assignableStaffUsers(app)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	data := map[string]interface{}{
		"Title":                      configs.T(c, "ui.applications.assign_title", nil),
		"CurrentPath":                "/admin/applications",
		"CurrentUser":                adminCurrentUser(c),
		"Application":                app,
		"StaffUsers":                 staffUsers,
		"CurrentAssignedStaffUserID": derefStr(app.AssignedStaffUserID),
	}
	return c.Render(http.StatusOK, "admin/pages/applications/assign_form.html", data)
}

func (h *AdminApplicationHandler) assignableStaffUsers(app *models.Application) ([]models.User, error) {
	if app == nil {
		return nil, nil
	}

	users := make([]models.User, 0)
	seen := make(map[string]struct{})

	addUser := func(user models.User) {
		if _, ok := seen[user.ID]; ok {
			return
		}
		seen[user.ID] = struct{}{}
		users = append(users, user)
	}

	if app.ServiceType.ResponsibleStaffUserID != nil && strings.TrimSpace(*app.ServiceType.ResponsibleStaffUserID) != "" {
		staff, err := h.userSvc.GetUser(*app.ServiceType.ResponsibleStaffUserID)
		if err != nil {
			return nil, err
		}
		if staff != nil {
			addUser(*staff)
		}
	} else if app.ServiceType.ResponsibleDepartmentID != nil && strings.TrimSpace(*app.ServiceType.ResponsibleDepartmentID) != "" && h.staffProfileSvc != nil {
		profiles, _, err := h.staffProfileSvc.ListStaffByDepartment(*app.ServiceType.ResponsibleDepartmentID, 1, 1000)
		if err != nil {
			return nil, err
		}
		for _, profile := range profiles {
			if profile.User.Role != models.UserRoleStaff {
				continue
			}
			addUser(profile.User)
		}
	}

	if app.AssignedStaffUser != nil {
		addUser(*app.AssignedStaffUser)
	}

	return users, nil
}

func (h *AdminApplicationHandler) AssignToStaff(c *echo.Context) error {
	id := c.Param("id")
	userID := c.FormValue("user_id")
	var toID *string
	if userID != "" {
		toID = &userID
	}
	if err := h.svc.AssignToStaff(id, toID, actorID(c)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.Redirect(http.StatusSeeOther, adminAppFlashURL("success", configs.T(c, "ui.msg.application_assigned", nil)))
}
