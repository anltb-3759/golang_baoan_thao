package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type AdminLogService interface {
	List(filter repositories.ActivityLogFilter, page, limit int) ([]models.ActivityLog, int64, error)
	Cleanup(retainDays *int, deleteBefore *time.Time) (int64, error)
}

type AdminLogHandler struct {
	svc AdminLogService
}

func NewAdminLogHandler(svc AdminLogService) *AdminLogHandler {
	return &AdminLogHandler{svc: svc}
}

func (h *AdminLogHandler) ListLogs(c *echo.Context) error {
	page, limit := parsePagination(c)
	filter := repositories.ActivityLogFilter{
		Action:      strings.TrimSpace(c.QueryParam("action")),
		EntityType:  strings.TrimSpace(c.QueryParam("entity_type")),
		Result:      strings.TrimSpace(c.QueryParam("result")),
		ActorUserID: strings.TrimSpace(c.QueryParam("actor_user_id")),
	}
	logs, total, err := h.svc.List(filter, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	currentUser := adminCurrentUser(c)
	data := map[string]any{
		"Title":        configs.T(c, "ui.logs.title", nil),
		"CurrentPath":  "/admin/logs",
		"CurrentUser":  currentUser,
		"CanCleanup":   currentUser != nil && currentUser.Role == string(models.UserRoleSuperAdmin),
		"Logs":         logs,
		"Pagination":   utils.NewPagination(page, limit, total),
		"Flash":        flashFromQuery(c),
		"FilterAction": filter.Action,
		"FilterEntity": filter.EntityType,
		"FilterResult": filter.Result,
		"FilterActor":  filter.ActorUserID,
	}
	return c.Render(http.StatusOK, "admin/pages/logs/list.html", data)
}

func (h *AdminLogHandler) CleanupLogs(c *echo.Context) error {
	currentUser := adminCurrentUser(c)
	if currentUser == nil || currentUser.Role != string(models.UserRoleSuperAdmin) {
		return c.Redirect(http.StatusSeeOther, adminLogsFlashURL("error", configs.T(c, "ui.error.403.message", nil)))
	}

	retainDaysRaw := strings.TrimSpace(c.FormValue("retain_days"))
	deleteBeforeRaw := strings.TrimSpace(c.FormValue("delete_before"))

	var retainDays *int
	if retainDaysRaw != "" {
		v, err := strconv.Atoi(retainDaysRaw)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, adminLogsFlashURL("error", configs.T(c, "activity_log.cleanup.invalid_mode", nil)))
		}
		retainDays = &v
	}

	var deleteBefore *time.Time
	if deleteBeforeRaw != "" {
		t, err := time.Parse("2006-01-02", deleteBeforeRaw)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, adminLogsFlashURL("error", configs.T(c, "activity_log.cleanup.invalid_mode", nil)))
		}
		deleteBefore = &t
	}

	if _, err := h.svc.Cleanup(retainDays, deleteBefore); err != nil {
		if err == services.ErrActivityLogCleanupModeInvalid || err == services.ErrActivityLogRetainDaysInvalid {
			return c.Redirect(http.StatusSeeOther, adminLogsFlashURL("error", configs.T(c, "activity_log.cleanup.invalid_mode", nil)))
		}
		return c.Redirect(http.StatusSeeOther, adminLogsFlashURL("error", configs.T(c, "common.internal_error", nil)))
	}

	return c.Redirect(http.StatusSeeOther, adminLogsFlashURL("success", configs.T(c, "activity_log.cleanup.success", nil)))
}

func adminLogsFlashURL(flash, msg string) string {
	return "/admin/logs?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}
