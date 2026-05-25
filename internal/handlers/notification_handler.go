package handlers

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type notificationSvc interface {
	List(userID string, filter repositories.NotificationFilter, page, limit int) ([]dtos.NotificationResponse, int64, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
}

type NotificationHandler struct {
	svc notificationSvc
}

func NewNotificationHandler(svc notificationSvc) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// List handles GET /api/citizens/me/notifications
func (h *NotificationHandler) List(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)
	page, limit := parsePagination(c)

	filter := repositories.NotificationFilter{
		Type: c.QueryParam("type"),
	}
	if c.QueryParam("is_read") == "true" {
		v := true
		filter.IsRead = &v
	} else if c.QueryParam("is_read") == "false" {
		v := false
		filter.IsRead = &v
	}

	items, total, err := h.svc.List(userID, filter, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{
		"notifications": items,
		"pagination":    utils.NewPagination(page, limit, total),
	})
}

// MarkAsRead handles PUT /api/citizens/me/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)
	id := c.Param("id")

	if err := h.svc.MarkAsRead(id, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{
		"message": configs.T(c, "notification.marked_read", nil),
	})
}

// MarkAllAsRead handles PUT /api/citizens/me/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)

	if err := h.svc.MarkAllAsRead(userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{
		"message": configs.T(c, "notification.all_marked_read", nil),
	})
}
