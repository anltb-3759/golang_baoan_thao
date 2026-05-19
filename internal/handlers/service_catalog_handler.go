package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/types"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type serviceCatalogSvc interface {
	List(filter repositories.ListFilter) (*repositories.ListResult, error)
	GetByID(id string) (*models.ServiceType, error)
}

type ServiceCatalogHandler struct {
	svc serviceCatalogSvc
}

func NewServiceCatalogHandler(svc serviceCatalogSvc) *ServiceCatalogHandler {
	return &ServiceCatalogHandler{svc: svc}
}

// ListServices handles GET /api/services
func (h *ServiceCatalogHandler) ListServices(c *echo.Context) error {
	page := parseIntParam(c, "page", 1)
	limit := parseIntParam(c, "limit", 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := h.svc.List(repositories.ListFilter{
		Category: c.QueryParam("category"),
		Search:   c.QueryParam("search"),
		Page:     page,
		Limit:    limit,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data":       result.Items,
		"pagination": types.Pagination{Page: page, Limit: limit, Total: result.Total},
	})
}

// GetService handles GET /api/services/:id
func (h *ServiceCatalogHandler) GetService(c *echo.Context) error {
	service, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "service.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": service,
	})
}

func parseIntParam(c *echo.Context, key string, fallback int) int {
	v, err := strconv.Atoi(c.QueryParam(key))
	if err != nil {
		return fallback
	}
	return v
}
