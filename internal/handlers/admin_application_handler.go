package handlers

import (
	"encoding/csv"
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type AdminApplicationService interface {
	ListApplications(filter repositories.ApplicationFilter, page, limit int) ([]models.Application, int64, error)
	ListApplicationsForActor(filter repositories.ApplicationFilter, page, limit int, role models.UserRole, actorID string) ([]models.Application, int64, error)
	GetApplication(id string) (*models.Application, error)
	AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error
	ProcessApplication(applicationID string, newStatus models.ApplicationStatus, note string, files []*multipart.FileHeader, processedBy string) error
}

type AssignableStaffService interface {
	ListStaffByDepartment(deptID string, page, limit int) ([]models.StaffProfile, int64, error)
}

type AdminApplicationHandler struct {
	svc             AdminApplicationService
	userSvc         AdminUserService
	staffProfileSvc AssignableStaffService
}

type applicationStatusOption struct {
	Value string
	Label string
}

func NewAdminApplicationHandler(svc AdminApplicationService, userSvc AdminUserService, staffProfileSvc AssignableStaffService) *AdminApplicationHandler {
	return &AdminApplicationHandler{svc: svc, userSvc: userSvc, staffProfileSvc: staffProfileSvc}
}

func adminAppFlashURL(flash, msg string) string {
	return "/admin/applications?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func adminAppDetailFlashURL(id, flash, msg string) string {
	return "/admin/applications/" + id + "?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func adminAppExportURL(filter repositories.ApplicationFilter) string {
	query := adminAppQueryString(filter)
	if query == "" {
		return "/admin/applications/export"
	}
	return "/admin/applications/export?" + query
}

func adminAppQueryString(filter repositories.ApplicationFilter) string {
	values := url.Values{}
	if filter.Status != "" {
		values.Set("status", filter.Status)
	}
	if filter.Service != "" {
		values.Set("service", filter.Service)
	}
	if filter.Submitter != "" {
		values.Set("submitter", filter.Submitter)
	}
	return values.Encode()
}

func adminAppFilterFromQuery(c *echo.Context) repositories.ApplicationFilter {
	return repositories.ApplicationFilter{
		Status:    strings.TrimSpace(c.QueryParam("status")),
		Service:   strings.TrimSpace(c.QueryParam("service")),
		Submitter: strings.TrimSpace(c.QueryParam("submitter")),
	}
}

func applicationStatusOptionsForFilter() []applicationStatusOption {
	return []applicationStatusOption{
		{Value: string(models.ApplicationStatusReceived), Label: "received"},
		{Value: string(models.ApplicationStatusProcessing), Label: "processing"},
		{Value: string(models.ApplicationStatusNeedMoreInfo), Label: "need_more_info"},
		{Value: string(models.ApplicationStatusApproved), Label: "approved"},
		{Value: string(models.ApplicationStatusRejected), Label: "rejected"},
	}
}

func applicationStatusOptionsForProcess(current models.ApplicationStatus) []applicationStatusOption {
	switch current {
	case models.ApplicationStatusReceived:
		return []applicationStatusOption{
			{Value: string(models.ApplicationStatusProcessing), Label: "processing"},
			{Value: string(models.ApplicationStatusNeedMoreInfo), Label: "need_more_info"},
		}
	case models.ApplicationStatusProcessing:
		return []applicationStatusOption{
			{Value: string(models.ApplicationStatusNeedMoreInfo), Label: "need_more_info"},
			{Value: string(models.ApplicationStatusApproved), Label: "approved"},
			{Value: string(models.ApplicationStatusRejected), Label: "rejected"},
		}
	case models.ApplicationStatusNeedMoreInfo:
		return []applicationStatusOption{
			{Value: string(models.ApplicationStatusProcessing), Label: "processing"},
			{Value: string(models.ApplicationStatusApproved), Label: "approved"},
			{Value: string(models.ApplicationStatusRejected), Label: "rejected"},
		}
	default:
		return nil
	}
}

func (h *AdminApplicationHandler) ListApplications(c *echo.Context) error {
	filter := adminAppFilterFromQuery(c)
	page, limit := parsePagination(c)
	currentUser := adminCurrentUser(c)
	role := models.UserRole("")
	actor := ""
	if currentUser != nil {
		role = models.UserRole(currentUser.Role)
		actor = currentUser.ID
	}
	apps, total, err := h.svc.ListApplicationsForActor(filter, page, limit, role, actor)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	data := map[string]interface{}{
		"Title":           configs.T(c, "ui.applications.title", nil),
		"CurrentPath":     "/admin/applications",
		"CurrentUser":     currentUser,
		"CanManage":       currentUser != nil && currentUser.Role == string(models.UserRoleManager),
		"Applications":    apps,
		"Pagination":      utils.NewPagination(page, limit, total),
		"Flash":           flashFromQuery(c),
		"FilterStatus":    filter.Status,
		"FilterService":   filter.Service,
		"FilterSubmitter": filter.Submitter,
		"ExportAction":    adminAppExportURL(filter),
		"PaginationQuery": adminAppQueryString(filter),
		"StatusOptions":   applicationStatusOptionsForFilter(),
	}
	return c.Render(http.StatusOK, "admin/pages/applications/list.html", data)
}

func (h *AdminApplicationHandler) ShowApplication(c *echo.Context) error {
	id := c.Param("id")
	app, err := h.svc.GetApplication(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "application.not_found")
	}
	currentUser := adminCurrentUser(c)
	canAssign := currentUser != nil && currentUser.Role == string(models.UserRoleManager)
	canProcess := currentUser != nil && currentUser.Role == string(models.UserRoleStaff) && len(applicationStatusOptionsForProcess(app.Status)) > 0
	citizenAttachments, responseAttachments := splitAdminAppAttachmentsByPhase(app.ApplicationAttachments)
	data := map[string]interface{}{
		"Title":               configs.T(c, "ui.applications.detail_title", nil),
		"CurrentPath":         "/admin/applications",
		"CurrentUser":         currentUser,
		"Application":         app,
		"CitizenAttachments":  citizenAttachments,
		"ResponseAttachments": responseAttachments,
		"ProcessOptions":      applicationStatusOptionsForProcess(app.Status),
		"CanProcess":          canProcess,
		"CanAssign":           canAssign,
		"Flash":               flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/applications/detail.html", data)
}

func splitAdminAppAttachmentsByPhase(atts []models.ApplicationAttachment) ([]models.ApplicationAttachment, []models.ApplicationAttachment) {
	citizen := make([]models.ApplicationAttachment, 0)
	response := make([]models.ApplicationAttachment, 0)
	for _, att := range atts {
		if att.AttachmentType == models.AttachmentTypeResult {
			response = append(response, att)
			continue
		}
		citizen = append(citizen, att)
	}
	return citizen, response
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

	if app.ServiceType.ResponsibleDepartmentID != nil && strings.TrimSpace(*app.ServiceType.ResponsibleDepartmentID) != "" && h.staffProfileSvc != nil {
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

// ExportCSV handles GET /admin/applications/export
func (h *AdminApplicationHandler) ExportCSV(c *echo.Context) error {
	setCSVHeaders(c, "ho_so.csv")
	writeCSVBOM(c)

	w := csv.NewWriter(c.Response())
	_ = w.Write([]string{
		"ma_ho_so", "ten_cong_dan", "loai_dich_vu", "phong_ban",
		"trang_thai", "ngay_nop", "ngay_hoan_thanh", "can_bo_xu_ly",
	})

	page := 1
	limit := 1000
	filter := adminAppFilterFromQuery(c)
	for {
		apps, _, err := h.svc.ListApplications(filter, page, limit)
		if err != nil {
			break
		}
		for _, a := range apps {
			deptName := ""
			if a.ServiceType.ResponsibleDepartment != nil {
				deptName = a.ServiceType.ResponsibleDepartment.Name
			}
			completedAt := ""
			if a.CompletedAt != nil {
				completedAt = a.CompletedAt.Format("02/01/2006")
			}
			staffName := ""
			if a.AssignedStaffUser != nil {
				staffName = a.AssignedStaffUser.Name
			}
			_ = w.Write([]string{
				a.ApplicationCode,
				a.CitizenUser.Name,
				a.ServiceType.Name,
				deptName,
				string(a.Status),
				a.SubmittedAt.Format("02/01/2006"),
				completedAt,
				staffName,
			})
		}
		if len(apps) < limit {
			break
		}
		page++
	}
	w.Flush()
	return nil
}

func (h *AdminApplicationHandler) ProcessApplication(c *echo.Context) error {
	id := c.Param("id")
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "application.invalid_request")
	}

	statusValue := strings.TrimSpace(c.FormValue("status"))
	if statusValue == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "application.invalid_request")
	}
	newStatus := models.ApplicationStatus(statusValue)
	if len(applicationStatusOptionsForProcess(models.ApplicationStatusReceived)) > 0 {
		allowed := map[string]struct{}{
			string(models.ApplicationStatusProcessing):   {},
			string(models.ApplicationStatusNeedMoreInfo): {},
			string(models.ApplicationStatusApproved):     {},
			string(models.ApplicationStatusRejected):     {},
		}
		if _, ok := allowed[statusValue]; !ok {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "application.invalid_transition")
		}
	}

	note := strings.TrimSpace(c.FormValue("note"))
	files := form.File["attachments[]"]
	if err := h.svc.ProcessApplication(id, newStatus, note, files, actorID(c)); err != nil {
		if errors.Is(err, services.ErrAdminApplicationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "application.not_found")
		}
		errKey := mapAdminApplicationProcessErrorKey(err)
		if errKey == "" {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		msg := configs.T(c, errKey, nil)
		return c.Redirect(http.StatusSeeOther, adminAppDetailFlashURL(id, "error", msg))
	}

	return c.Redirect(http.StatusSeeOther, adminAppDetailFlashURL(id, "success", configs.T(c, "ui.msg.application_processed", nil)))
}

func mapAdminApplicationProcessErrorKey(err error) string {
	switch {
	case errors.Is(err, services.ErrAdminApplicationInvalidTransition):
		return "application.invalid_transition"
	case errors.Is(err, services.ErrAdminApplicationRejectReasonRequired):
		return "application.reject_reason_required"
	case errors.Is(err, services.ErrAdminApplicationNeedMoreInfoNoteRequired):
		return "application.need_more_info_note_required"
	case errors.Is(err, utils.ErrDisallowedMime),
		errors.Is(err, utils.ErrEmptyFileName),
		errors.Is(err, utils.ErrUnsafeFileName),
		errors.Is(err, utils.ErrPathEscape):
		return "application.attachment_invalid_type"
	default:
		return ""
	}
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
