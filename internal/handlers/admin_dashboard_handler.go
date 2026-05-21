package handlers

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

type AdminDashboardHandler struct{}

func NewAdminDashboardHandler() *AdminDashboardHandler {
	return &AdminDashboardHandler{}
}

func (h *AdminDashboardHandler) ShowDashboard(c *echo.Context) error {
	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.dashboard.title", nil),
		"CurrentPath": "/admin",
		"CurrentUser": adminCurrentUser(c),
	}
	return c.Render(http.StatusOK, "admin/pages/dashboard.html", data)
}
