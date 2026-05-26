package handlers

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

type AdminDashboardHandler struct {
	svc *services.AdminDashboardService
}

func NewAdminDashboardHandler(svc *services.AdminDashboardService) *AdminDashboardHandler {
	return &AdminDashboardHandler{svc: svc}
}

func (h *AdminDashboardHandler) ShowDashboard(c *echo.Context) error {
	data, err := h.svc.GetDashboardData()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.Render(http.StatusOK, "admin/pages/dashboard.html", map[string]interface{}{
		"Title":              configs.T(c, "ui.dashboard.title", nil),
		"CurrentPath":        "/admin",
		"CurrentUser":        adminCurrentUser(c),
		"Stats":              data.Stats,
		"RecentApplications": data.RecentApplications,
	})
}
