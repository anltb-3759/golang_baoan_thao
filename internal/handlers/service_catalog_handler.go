package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

type serviceCatalogSvc interface {
	List(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error)
	GetByID(ctx context.Context, id string) (*models.ServiceType, error)
	GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error)
	ListDepartments(ctx context.Context) ([]models.Department, error)
	ListCategories(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, st *models.ServiceType) error
	Update(ctx context.Context, st *models.ServiceType) error
	Delete(ctx context.Context, id string) error
}

type serviceTypeFormData struct {
	ID                      string
	Name                    string
	Code                    string
	CategoryID              string
	Description             string
	RequiredDocuments       string
	FormSchema              string
	ProcessingTime          string
	Fee                     string
	ResponsibleDepartmentID string
	IsActive                bool
}

type ServiceCatalogHandler struct {
	svc serviceCatalogSvc
}

func NewServiceCatalogHandler(svc serviceCatalogSvc) *ServiceCatalogHandler {
	return &ServiceCatalogHandler{svc: svc}
}

func serviceTypeFormFromModel(st *models.ServiceType) serviceTypeFormData {
	formSchema := "{}"
	if len(st.FormSchema) > 0 {
		formSchema = string(st.FormSchema)
	}

	processingTime := ""
	if st.ProcessingTime != nil {
		processingTime = strconv.Itoa(*st.ProcessingTime)
	}

	fee := fmt.Sprintf("%.2f", st.Fee)
	responsibleDepartmentID := ""
	if st.ResponsibleDepartmentID != nil {
		responsibleDepartmentID = *st.ResponsibleDepartmentID
	}

	categoryID := ""
	if st.CategoryID != nil {
		categoryID = *st.CategoryID
	}

	return serviceTypeFormData{
		ID:                      st.ID,
		Name:                    st.Name,
		Code:                    st.Code,
		CategoryID:              categoryID,
		Description:             st.Description,
		RequiredDocuments:       st.RequiredDocuments,
		FormSchema:              formSchema,
		ProcessingTime:          processingTime,
		Fee:                     fee,
		ResponsibleDepartmentID: responsibleDepartmentID,
		IsActive:                st.IsActive,
	}
}

func serviceTypeFormFromRequest(req *dtos.ServiceTypeFormRequest) serviceTypeFormData {
	formData := serviceTypeFormData{
		Name:                    req.Name,
		Code:                    req.Code,
		CategoryID:              req.CategoryID,
		Description:             req.Description,
		RequiredDocuments:       req.RequiredDocuments,
		FormSchema:              req.FormSchema,
		ProcessingTime:          req.ProcessingTime,
		Fee:                     req.Fee,
		ResponsibleDepartmentID: req.ResponsibleDepartmentID,
		IsActive:                req.IsActive,
	}
	if formData.FormSchema == "" {
		formData.FormSchema = "{}"
	}
	return formData
}

func serviceTypeToModel(req *dtos.ServiceTypeFormRequest, existing *models.ServiceType) (*models.ServiceType, error) {
	serviceType := &models.ServiceType{}
	if existing != nil {
		serviceType = existing
	}

	serviceType.Name = req.Name
	serviceType.Code = req.Code
	serviceType.Description = req.Description
	serviceType.RequiredDocuments = req.RequiredDocuments
	serviceType.IsActive = req.IsActive

	if strings.TrimSpace(req.CategoryID) != "" {
		catID := strings.TrimSpace(req.CategoryID)
		serviceType.CategoryID = &catID
	} else {
		serviceType.CategoryID = nil
	}

	if req.FormSchema == "" {
		req.FormSchema = "{}"
	}
	if !json.Valid([]byte(req.FormSchema)) {
		return nil, errors.New("service_type.invalid_form_schema")
	}
	serviceType.FormSchema = json.RawMessage(req.FormSchema)

	if strings.TrimSpace(req.ProcessingTime) != "" {
		processingTime, err := strconv.Atoi(req.ProcessingTime)
		if err != nil || processingTime < 0 {
			return nil, errors.New("validation.invalid")
		}
		serviceType.ProcessingTime = &processingTime
	} else {
		serviceType.ProcessingTime = nil
	}

	if strings.TrimSpace(req.Fee) != "" {
		fee, err := strconv.ParseFloat(req.Fee, 64)
		if err != nil || fee < 0 {
			return nil, errors.New("validation.invalid")
		}
		serviceType.Fee = fee
	} else {
		serviceType.Fee = 0
	}

	if strings.TrimSpace(req.ResponsibleDepartmentID) != "" {
		deptID := strings.TrimSpace(req.ResponsibleDepartmentID)
		serviceType.ResponsibleDepartmentID = &deptID
	} else {
		serviceType.ResponsibleDepartmentID = nil
	}

	return serviceType, nil
}

func (h *ServiceCatalogHandler) loadFormDeps(ctx context.Context) ([]models.Department, []models.Category, error) {
	depts, err := h.svc.ListDepartments(ctx)
	if err != nil {
		return nil, nil, err
	}
	cats, err := h.svc.ListCategories(ctx)
	if err != nil {
		return nil, nil, err
	}
	return depts, cats, nil
}

func validationErrorMessage(c *echo.Context) string {
	return configs.T(c, "validation.invalid", nil)
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

	result, err := h.svc.List(c.Request().Context(), repositories.ListFilter{
		Category:        c.QueryParam("category"),
		Search:          c.QueryParam("search"),
		Page:            page,
		Limit:           limit,
		IncludeInactive: false,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"services":   result.Items,
		"pagination": utils.NewPagination(page, limit, result.Total),
	})
}

// GetService handles GET /api/services/:id
func (h *ServiceCatalogHandler) GetService(c *echo.Context) error {
	service, err := h.svc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "service.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, map[string]any{
		"service": service,
	})
}

// AdminLanding handles GET /admin.
func (h *ServiceCatalogHandler) AdminLanding(c *echo.Context) error {
	return c.Redirect(http.StatusSeeOther, "/admin/service-types")
}

// ListServiceTypesAdmin handles GET /admin/service-types.
func (h *ServiceCatalogHandler) ListServiceTypesAdmin(c *echo.Context) error {
	page, limit := parsePagination(c)
	search := c.QueryParam("search")
	success := c.QueryParam("success")

	result, err := h.svc.List(c.Request().Context(), repositories.ListFilter{
		Search:          search,
		Page:            page,
		Limit:           limit,
		IncludeInactive: true,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	message := ""
	switch success {
	case "created":
		message = configs.T(c, "service_type.created", nil)
	case "updated":
		message = configs.T(c, "service_type.updated", nil)
	case "deleted":
		message = configs.T(c, "service_type.deleted", nil)
	}

	return c.Render(http.StatusOK, "admin/pages/service-types/list.html", map[string]any{
		"ServiceTypes": result.Items,
		"Pagination":   utils.NewPagination(page, limit, result.Total),
		"Search":       search,
		"Message":      message,
		"CurrentPath":  "/admin/service-types",
		"CurrentUser":  adminCurrentUser(c),
	})
}

// CreateServiceTypeForm handles GET /admin/service-types/new.
func (h *ServiceCatalogHandler) CreateServiceTypeForm(c *echo.Context) error {
	depts, cats, err := h.loadFormDeps(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.Render(http.StatusOK, "admin/pages/service-types/form.html", map[string]any{
		"Mode":        "create",
		"Title":       configs.T(c, "service_type.create_title", nil),
		"Action":      "/admin/service-types",
		"ServiceType": serviceTypeFormData{FormSchema: "{}", IsActive: true},
		"Departments": depts,
		"Categories":  cats,
	})
}

// CreateServiceType handles POST /admin/service-types.
func (h *ServiceCatalogHandler) CreateServiceType(c *echo.Context) error {
	req := new(dtos.ServiceTypeFormRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "service_type.invalid_request")
	}

	renderCreateErr := func(msg string) error {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":        "create",
			"Title":       configs.T(c, "service_type.create_title", nil),
			"Action":      "/admin/service-types",
			"Error":       msg,
			"ServiceType": serviceTypeFormFromRequest(req),
			"Departments": depts,
			"Categories":  cats,
			"CurrentPath": "/admin/service-types",
			"CurrentUser": adminCurrentUser(c),
		})
	}

	if err := c.Validate(req); err != nil {
		return renderCreateErr(validationErrorMessage(c))
	}

	serviceType, err := serviceTypeToModel(req, nil)
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		switch err.Error() {
		case "validation.invalid":
			msg = configs.T(c, "validation.invalid", nil)
		case "service_type.invalid_form_schema":
			msg = configs.T(c, "service_type.invalid_form_schema", nil)
		}
		return renderCreateErr(msg)
	}

	serviceType.CreatedAt = time.Now()
	serviceType.UpdatedAt = time.Now()

	if err := h.svc.Create(c.Request().Context(), serviceType); err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		switch {
		case errors.Is(err, services.ErrServiceTypeCodeExists):
			msg = configs.T(c, "service_type.code_duplicate", nil)
		case strings.Contains(err.Error(), "service_type.invalid_form_schema"):
			msg = configs.T(c, "service_type.invalid_form_schema", nil)
		}
		return renderCreateErr(msg)
	}

	return c.Redirect(http.StatusSeeOther, "/admin/service-types?success=created")
}

// EditServiceTypeForm handles GET /admin/service-types/:id/edit.
func (h *ServiceCatalogHandler) EditServiceTypeForm(c *echo.Context) error {
	serviceType, err := h.svc.GetByIDForAdmin(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "service_type.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	depts, cats, err := h.loadFormDeps(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.Render(http.StatusOK, "admin/pages/service-types/form.html", map[string]any{
		"Mode":        "edit",
		"Title":       configs.T(c, "service_type.edit_title", nil),
		"Action":      "/admin/service-types/" + serviceType.ID,
		"ServiceType": serviceTypeFormFromModel(serviceType),
		"Departments": depts,
		"Categories":  cats,
	})
}

// ShowServiceType handles GET /admin/service-types/:id.
func (h *ServiceCatalogHandler) ShowServiceType(c *echo.Context) error {
	serviceType, err := h.svc.GetByIDForAdmin(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "service_type.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.Render(http.StatusOK, "admin/pages/service-types/detail.html", map[string]any{
		"ServiceType": serviceType,
		"CurrentPath": "/admin/service-types",
		"CurrentUser": adminCurrentUser(c),
	})
}

// UpdateServiceType handles POST /admin/service-types/:id.
func (h *ServiceCatalogHandler) UpdateServiceType(c *echo.Context) error {
	serviceType, err := h.svc.GetByIDForAdmin(c.Request().Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "service_type.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	req := new(dtos.ServiceTypeFormRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "service_type.invalid_request")
	}

	renderEditErr := func(msg string) error {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		data := serviceTypeFormFromRequest(req)
		data.ID = serviceType.ID
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":        "edit",
			"Title":       configs.T(c, "service_type.edit_title", nil),
			"Action":      "/admin/service-types/" + serviceType.ID,
			"Error":       msg,
			"ServiceType": data,
			"Departments": depts,
			"Categories":  cats,
			"CurrentPath": "/admin/service-types",
			"CurrentUser": adminCurrentUser(c),
		})
	}

	if err := c.Validate(req); err != nil {
		return renderEditErr(validationErrorMessage(c))
	}

	updated, err := serviceTypeToModel(req, serviceType)
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		switch err.Error() {
		case "validation.invalid":
			msg = configs.T(c, "validation.invalid", nil)
		case "service_type.invalid_form_schema":
			msg = configs.T(c, "service_type.invalid_form_schema", nil)
		}
		return renderEditErr(msg)
	}

	updated.ID = serviceType.ID
	updated.CreatedAt = serviceType.CreatedAt
	updated.UpdatedAt = time.Now()

	if err := h.svc.Update(c.Request().Context(), updated); err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrServiceTypeCodeExists) {
			msg = configs.T(c, "service_type.code_duplicate", nil)
		}
		return renderEditErr(msg)
	}

	return c.Redirect(http.StatusSeeOther, "/admin/service-types?success=updated")
}

// DeleteServiceType handles POST /admin/service-types/:id/delete.
func (h *ServiceCatalogHandler) DeleteServiceType(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "service_type.not_found")
		case errors.Is(err, services.ErrServiceTypeHasApplications):
			return echo.NewHTTPError(http.StatusConflict, "service_type.applications_exist")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
	}

	return c.Redirect(http.StatusSeeOther, "/admin/service-types?success=deleted")
}
