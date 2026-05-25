package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

type ServiceTypeImportExportService interface {
	ImportServiceTypes(rows []services.ServiceTypeImportRow, createdBy string) []string
}

type ServiceCatalogHandler struct {
	svc       serviceCatalogSvc
	userSvc   AdminUserService
	importSvc ServiceTypeImportExportService
}

func NewServiceCatalogHandler(svc serviceCatalogSvc, userSvc AdminUserService) *ServiceCatalogHandler {
	return &ServiceCatalogHandler{svc: svc, userSvc: userSvc}
}

func (h *ServiceCatalogHandler) WithImportExport(svc ServiceTypeImportExportService) *ServiceCatalogHandler {
	h.importSvc = svc
	return h
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

func serviceTypeFlashURL(flash, msg string) string {
	return "/admin/service-types?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
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

	result, err := h.svc.List(c.Request().Context(), repositories.ListFilter{
		Search:          search,
		Page:            page,
		Limit:           limit,
		IncludeInactive: true,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	flash := flashFromQuery(c)
	if flash == nil {
		switch c.QueryParam("success") {
		case "created":
			flash = map[string]string{"Type": "success", "Message": configs.T(c, "service_type.created", nil)}
		case "updated":
			flash = map[string]string{"Type": "success", "Message": configs.T(c, "service_type.updated", nil)}
		case "deleted":
			flash = map[string]string{"Type": "success", "Message": configs.T(c, "service_type.deleted", nil)}
		}
	}

	return c.Render(http.StatusOK, "admin/pages/service-types/list.html", map[string]any{
		"ServiceTypes":   result.Items,
		"Pagination":     utils.NewPagination(page, limit, result.Total),
		"Search":         search,
		"Flash":          flash,
		"CurrentPath":    "/admin/service-types",
		"CurrentUser":    adminCurrentUser(c),
		"ImportAction":   "/admin/service-types/import",
		"TemplateAction": "/admin/service-types/template",
	})
}

// CreateServiceTypeForm handles GET /admin/service-types/new.
func (h *ServiceCatalogHandler) CreateServiceTypeForm(c *echo.Context) error {
	depts, cats, err := h.loadFormDeps(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.Render(http.StatusOK, "admin/pages/service-types/form.html", map[string]any{
		"Mode":         "create",
		"Title":        configs.T(c, "service_type.create_title", nil),
		"Action":       "/admin/service-types",
		"ServiceType":  serviceTypeFormData{},
		"Departments":  depts,
		"Categories":   cats,
		"CurrentPath":  "/admin/service-types",
		"CurrentUser":  adminCurrentUser(c),
	})
}

// CreateServiceType handles POST /admin/service-types.
func (h *ServiceCatalogHandler) CreateServiceType(c *echo.Context) error {
	req := new(dtos.ServiceTypeFormRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "service_type.invalid_request")
	}

	if err := c.Validate(req); err != nil {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":         "create",
			"Title":        configs.T(c, "service_type.create_title", nil),
			"Action":       "/admin/service-types",
			"Error":        validationErrorMessage(c),
			"ServiceType":  serviceTypeFormFromRequest(req),
			"Departments":  depts,
			"Categories":   cats,
			"CurrentPath":  "/admin/service-types",
			"CurrentUser":  adminCurrentUser(c),
		})
	}

	serviceType, err := serviceTypeToModel(req, nil)
	if err != nil {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		message := configs.T(c, "common.internal_error", nil)
		switch err.Error() {
		case "validation.invalid":
			message = configs.T(c, "validation.invalid", nil)
		case "service_type.invalid_form_schema":
			message = configs.T(c, "service_type.invalid_form_schema", nil)
		}
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":         "create",
			"Title":        configs.T(c, "service_type.create_title", nil),
			"Action":       "/admin/service-types",
			"Error":        message,
			"ServiceType":  serviceTypeFormFromRequest(req),
			"Departments":  depts,
			"Categories":   cats,
			"CurrentPath":  "/admin/service-types",
			"CurrentUser":  adminCurrentUser(c),
		})
	}

	serviceType.CreatedAt = time.Now()
	serviceType.UpdatedAt = time.Now()

	if err := h.svc.Create(c.Request().Context(), serviceType); err != nil {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		message := configs.T(c, "common.internal_error", nil)
		switch {
		case errors.Is(err, services.ErrServiceTypeCodeExists):
			message = configs.T(c, "service_type.code_duplicate", nil)
		case strings.Contains(err.Error(), "service_type.invalid_form_schema"):
			message = configs.T(c, "service_type.invalid_form_schema", nil)
		}
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":         "create",
			"Title":        configs.T(c, "service_type.create_title", nil),
			"Action":       "/admin/service-types",
			"Error":        message,
			"ServiceType":  serviceTypeFormFromRequest(req),
			"Departments":  depts,
			"Categories":   cats,
			"CurrentPath":  "/admin/service-types",
			"CurrentUser":  adminCurrentUser(c),
		})
	}

	return c.Redirect(http.StatusSeeOther, serviceTypeFlashURL("success", configs.T(c, "service_type.created", nil)))
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
		"Mode":         "edit",
		"Title":        configs.T(c, "service_type.edit_title", nil),
		"Action":       "/admin/service-types/" + serviceType.ID,
		"ServiceType":  serviceTypeFormFromModel(serviceType),
		"Departments":  depts,
		"Categories":   cats,
		"CurrentPath":  "/admin/service-types",
		"CurrentUser":  adminCurrentUser(c),
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

	if err := c.Validate(req); err != nil {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		data := serviceTypeFormFromRequest(req)
		data.ID = serviceType.ID
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":         "edit",
			"Title":        configs.T(c, "service_type.edit_title", nil),
			"Action":       "/admin/service-types/" + serviceType.ID,
			"Error":        validationErrorMessage(c),
			"ServiceType":  data,
			"Departments":  depts,
			"Categories":   cats,
			"CurrentPath":  "/admin/service-types",
			"CurrentUser":  adminCurrentUser(c),
		})
	}

	updated, err := serviceTypeToModel(req, serviceType)
	if err != nil {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		msg := configs.T(c, "common.internal_error", nil)
		switch err.Error() {
		case "validation.invalid":
			msg = configs.T(c, "validation.invalid", nil)
		case "service_type.invalid_form_schema":
			msg = configs.T(c, "service_type.invalid_form_schema", nil)
		}
		data := serviceTypeFormFromRequest(req)
		data.ID = serviceType.ID
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":         "edit",
			"Title":        configs.T(c, "service_type.edit_title", nil),
			"Action":       "/admin/service-types/" + serviceType.ID,
			"Error":        msg,
			"ServiceType":  data,
			"Departments":  depts,
			"Categories":   cats,
			"CurrentPath":  "/admin/service-types",
			"CurrentUser":  adminCurrentUser(c),
		})
	}

	updated.ID = serviceType.ID
	updated.CreatedAt = serviceType.CreatedAt
	updated.UpdatedAt = time.Now()

	if err := h.svc.Update(c.Request().Context(), updated); err != nil {
		depts, cats, depErr := h.loadFormDeps(c.Request().Context())
		if depErr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
		}
		message := configs.T(c, "common.internal_error", nil)
		switch {
		case errors.Is(err, services.ErrServiceTypeCodeExists):
			message = configs.T(c, "service_type.code_duplicate", nil)
		}
		data := serviceTypeFormFromRequest(req)
		data.ID = serviceType.ID
		return c.Render(http.StatusBadRequest, "admin/pages/service-types/form.html", map[string]any{
			"Mode":         "edit",
			"Title":        configs.T(c, "service_type.edit_title", nil),
			"Action":       "/admin/service-types/" + serviceType.ID,
			"Error":        message,
			"ServiceType":  data,
			"Departments":  depts,
			"Categories":   cats,
			"CurrentPath":  "/admin/service-types",
			"CurrentUser":  adminCurrentUser(c),
		})
	}

	return c.Redirect(http.StatusSeeOther, serviceTypeFlashURL("success", configs.T(c, "service_type.updated", nil)))
}

// DownloadTemplate handles GET /admin/service-types/template.
func (h *ServiceCatalogHandler) DownloadTemplate(c *echo.Context) error {
	return writeXLSXTemplate(c, "template_loai_dich_vu.xlsx", []XLSXColumn{
		{Header: "ten", Hint: "Tên loại dịch vụ. VD: Cấp giấy phép xây dựng"},
		{Header: "mo_ta", Hint: "Mô tả dịch vụ (tùy chọn)"},
		{Header: "thoi_gian_xu_ly_ngay", Hint: "Thời gian xử lý (số nguyên, ngày). VD: 15"},
		{Header: "phi", Hint: "Lệ phí (số thực, đồng). VD: 50000"},
		{Header: "ma_phong_ban", Hint: "Mã phòng ban phụ trách (tùy chọn). VD: QLDT"},
	})
}

// ExportCSV handles GET /admin/service-types/export.
func (h *ServiceCatalogHandler) ExportCSV(c *echo.Context) error {
	setCSVHeaders(c, "loai_dich_vu.csv")
	writeCSVBOM(c)

	w := csv.NewWriter(c.Response())
	_ = w.Write([]string{"ten", "mo_ta", "thoi_gian_xu_ly_ngay", "phi", "ma_phong_ban"})

	batchSize := 1000
	for page := 1; ; page++ {
		result, err := h.svc.List(c.Request().Context(), repositories.ListFilter{
			Page: page, Limit: batchSize, IncludeInactive: true,
		})
		if err != nil {
			break
		}
		for _, st := range result.Items {
			pt := ""
			if st.ProcessingTime != nil {
				pt = strconv.Itoa(*st.ProcessingTime)
			}
			deptCode := ""
			if st.ResponsibleDepartment != nil {
				deptCode = st.ResponsibleDepartment.Code
			}
			_ = w.Write([]string{st.Name, st.Description, pt, fmt.Sprintf("%.2f", st.Fee), deptCode})
		}
		if len(result.Items) < batchSize {
			break
		}
	}
	w.Flush()
	return nil
}

// ImportCSV handles POST /admin/service-types/import.
func (h *ServiceCatalogHandler) ImportCSV(c *echo.Context) error {
	if h.importSvc == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "Chức năng import chưa được kích hoạt")
	}
	dataRows, err := parseUploadedCSV(c)
	if err != nil {
		return err
	}

	importRows := make([]services.ServiceTypeImportRow, 0, len(dataRows))
	for _, cols := range dataRows {
		importRows = append(importRows, services.ServiceTypeImportRow{
			Ten:              safeCol(cols, 0),
			MoTa:             safeCol(cols, 1),
			ThoiGianXuLyNgay: safeCol(cols, 2),
			Phi:              safeCol(cols, 3),
			MaPhongBan:       safeCol(cols, 4),
		})
	}

	errs := h.importSvc.ImportServiceTypes(importRows, actorID(c))
	if len(errs) > 0 {
		page, limit := parsePagination(c)
		search := c.QueryParam("search")
		result, _ := h.svc.List(c.Request().Context(), repositories.ListFilter{
			Search: search, Page: page, Limit: limit, IncludeInactive: true,
		})
		flash := flashFromQuery(c)
		return c.Render(http.StatusUnprocessableEntity, "admin/pages/service-types/list.html", map[string]any{
			"ServiceTypes": result.Items,
			"Pagination":   utils.NewPagination(page, limit, result.Total),
			"Search":       search,
			"Flash":          flash,
			"CurrentPath":    "/admin/service-types",
			"CurrentUser":    adminCurrentUser(c),
			"ImportErrors":   errs,
			"ImportAction":   "/admin/service-types/import",
			"TemplateAction": "/admin/service-types/template",
		})
	}

	return c.Redirect(http.StatusSeeOther, serviceTypeFlashURL("success", configs.T(c, "ui.msg.import_success", nil)))
}

// DeleteServiceType handles POST /admin/service-types/:id/delete.
func (h *ServiceCatalogHandler) DeleteServiceType(c *echo.Context) error {
	if err := h.svc.Delete(c.Request().Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return c.Redirect(http.StatusSeeOther, serviceTypeFlashURL("error", configs.T(c, "service_type.not_found", nil)))
		case errors.Is(err, services.ErrServiceTypeHasApplications):
			return c.Redirect(http.StatusSeeOther, serviceTypeFlashURL("error", configs.T(c, "service_type.applications_exist", nil)))
		default:
			return c.Redirect(http.StatusSeeOther, serviceTypeFlashURL("error", configs.T(c, "common.internal_error", nil)))
		}
	}

	return c.Redirect(http.StatusSeeOther, serviceTypeFlashURL("success", configs.T(c, "service_type.deleted", nil)))
}
