package handlers

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type applicationSvc interface {
	SubmitApplication(citizenUserID string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error)
	ListMyApplications(userID string, page, limit int) ([]models.Application, int64, error)
	GetMyApplication(userID, appID string) (*dtos.ApplicationResponse, error)
}

type ApplicationHandler struct {
	svc applicationSvc
}

func NewApplicationHandler(svc *services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

func NewApplicationHandlerFromSvc(svc applicationSvc) *ApplicationHandler {
	return &ApplicationHandler{svc: svc}
}

func (h *ApplicationHandler) Submit(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)

	dataField := c.FormValue("data")
	if dataField == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "application.invalid_request")
	}

	req := new(dtos.SubmitApplicationRequest)
	if err := json.Unmarshal([]byte(dataField), req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "application.invalid_request")
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	var files []*multipart.FileHeader
	form, err := c.MultipartForm()
	if err == nil && form != nil {
		files = form.File["attachments[]"]
	}

	result, err := h.svc.SubmitApplication(userID, req, files)
	if err != nil {
		return mapApplicationError(err)
	}

	return c.JSON(http.StatusCreated, utils.Map{"application": result})
}

func (h *ApplicationHandler) ListMine(c *echo.Context) error {
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
		"pagination":   utils.Pagination{Page: page, Limit: limit, Total: total},
	})
}

func (h *ApplicationHandler) GetMine(c *echo.Context) error {
	userID := configs.UserIDFromContext(c)
	appID := c.Param("id")

	result, err := h.svc.GetMyApplication(userID, appID)
	if err != nil {
		if errors.Is(err, services.ErrApplicationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "application.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{"application": result})
}

func mapApplicationError(err error) error {
	switch {
	case errors.Is(err, services.ErrServiceTypeNotFound):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "application.service_type_not_found")
	case errors.Is(err, services.ErrServiceTypeInactive):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "application.service_type_inactive")
	case errors.Is(err, services.ErrMissingRequiredField):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "application.missing_required_field")
	case errors.Is(err, services.ErrTooManyAttachments):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "application.attachment_too_many")
	case errors.Is(err, services.ErrAttachmentTooLarge):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "application.attachment_too_large")
	case errors.Is(err, services.ErrAttachmentInvalidType):
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "application.attachment_invalid_type")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
}
