package handlers

import (
	"errors"
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type citizenProfileSvc interface {
	GetProfile(userID string) (*dtos.CitizenProfileResponse, error)
	UpdateProfile(userID string, req *dtos.UpdateCitizenProfileRequest) (*dtos.CitizenProfileResponse, error)
	ChangeMyPassword(userID string, req *dtos.ChangeMyPasswordRequest) error
	ListMyApplications(userID string, page, limit int) ([]models.Application, int64, error)
}

type CitizenProfileHandler struct {
	svc citizenProfileSvc
}

func NewCitizenProfileHandler(svc citizenProfileSvc) *CitizenProfileHandler {
	return &CitizenProfileHandler{svc: svc}
}

// GetMe handles GET /api/citizens/me
func (h *CitizenProfileHandler) GetMe(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)

	profile, err := h.svc.GetProfile(userID)
	if err != nil {
		if errors.Is(err, services.ErrProfileNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "profile.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{"profile": profile})
}

// UpdateMe handles PUT /api/citizens/me
func (h *CitizenProfileHandler) UpdateMe(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)

	req := new(dtos.UpdateCitizenProfileRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "profile.invalid_request")
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	profile, err := h.svc.UpdateProfile(userID, req)
	if err != nil {
		if errors.Is(err, services.ErrProfileNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "profile.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{"profile": profile})
}

// ChangeMyPassword handles PUT /api/citizens/me/password
func (h *CitizenProfileHandler) ChangeMyPassword(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)
	req := new(dtos.ChangeMyPasswordRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "auth.invalid_request")
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.svc.ChangeMyPassword(userID, req); err != nil {
		switch {
		case errors.Is(err, services.ErrPasswordMismatch):
			return echo.NewHTTPError(http.StatusUnauthorized, "auth.password_mismatch")
		case errors.Is(err, services.ErrPasswordConfirmationMismatch):
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "auth.password_confirmation_mismatch")
		case errors.Is(err, services.ErrNewPasswordMustDiffer):
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "auth.new_password_must_differ")
		case errors.Is(err, services.ErrProfileNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "profile.not_found")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
	}

	return c.JSON(http.StatusOK, utils.Map{"message": configs.T(c, "profile.password_changed", nil)})
}

// ListMyApplications handles GET /api/citizens/me/applications
func (h *CitizenProfileHandler) ListMyApplications(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)

	page, limit := parsePagination(c)

	apps, total, err := h.svc.ListMyApplications(userID, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	items := make([]utils.Map, 0, len(apps))
	for _, a := range apps {
		items = append(items, utils.Map{
			"id":                a.ID,
			"application_code":  a.ApplicationCode,
			"service_type_id":   a.ServiceTypeID,
			"service_type_name": a.ServiceType.Name,
			"status":            string(a.Status),
			"submitted_at":      a.SubmittedAt,
		})
	}

	return c.JSON(http.StatusOK, utils.Map{
		"applications": items,
		"pagination":   utils.NewPagination(page, limit, total),
	})
}
