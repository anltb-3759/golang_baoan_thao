package handlers

import (
	"net/http"
	"net/url"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type AdminCitizenProfileRepo interface {
	ListAllForExport(offset, limit int) ([]repositories.CitizenExportRow, int64, error)
}

type CitizenImportExportService interface {
	ImportCitizens(rows []services.CitizenImportRow, createdBy string) []string
}

type AdminCitizenHandler struct {
	profileRepo AdminCitizenProfileRepo
	importSvc   CitizenImportExportService
}

func NewAdminCitizenHandler(profileRepo AdminCitizenProfileRepo, importSvc CitizenImportExportService) *AdminCitizenHandler {
	return &AdminCitizenHandler{profileRepo: profileRepo, importSvc: importSvc}
}

// DownloadTemplate handles GET /admin/citizens/template
func (h *AdminCitizenHandler) DownloadTemplate(c *echo.Context) error {
	return writeXLSXTemplate(c, "template_cong_dan.xlsx", []XLSXColumn{
		{Header: "so_cccd", Hint: "12 chữ số. VD: 012345678901"},
		{Header: "ho_ten", Hint: "Họ và tên đầy đủ. VD: Nguyễn Văn A"},
		{Header: "email", Hint: "Email hợp lệ. VD: nguyenvana@email.com"},
		{Header: "so_dien_thoai", Hint: "Số điện thoại. VD: 0901234567"},
		{Header: "dia_chi", Hint: "Địa chỉ thường trú. VD: 123 Nguyễn Huệ, Q.1, TP.HCM"},
		{Header: "ngay_sinh", Hint: "DD/MM/YYYY hoặc YYYY-MM-DD. VD: 01/01/1990"},
	})
}

// ImportCSV handles POST /admin/citizens/import
func (h *AdminCitizenHandler) ImportCSV(c *echo.Context) error {
	dataRows, err := parseUploadedCSV(c)
	if err != nil {
		return err
	}

	importRows := make([]services.CitizenImportRow, 0, len(dataRows))
	for _, cols := range dataRows {
		importRows = append(importRows, services.CitizenImportRow{
			SoCCCD:      safeCol(cols, 0),
			HoTen:       safeCol(cols, 1),
			Email:       safeCol(cols, 2),
			SoDienThoai: safeCol(cols, 3),
			DiaChi:      safeCol(cols, 4),
			NgaySinh:    safeCol(cols, 5),
		})
	}

	errs := h.importSvc.ImportCitizens(importRows, actorID(c))
	if len(errs) > 0 {
		return h.renderList(c, errs)
	}

	return c.Redirect(http.StatusSeeOther, citizenFlashURL("success", configs.T(c, "ui.msg.import_success", nil)))
}

func (h *AdminCitizenHandler) renderList(c *echo.Context, importErrors []string) error {
	rows, total, _ := h.profileRepo.ListAllForExport(0, 20)
	data := map[string]interface{}{
		"Title":          configs.T(c, "ui.citizens.title", nil),
		"CurrentPath":    "/admin/citizens",
		"CurrentUser":    adminCurrentUser(c),
		"Citizens":       rows,
		"Pagination":     utils.NewPagination(1, 20, total),
		"ImportErrors":   importErrors,
		"ImportAction":   "/admin/citizens/import",
		"TemplateAction": "/admin/citizens/template",
		"Flash":          flashFromQuery(c),
	}
	return c.Render(http.StatusUnprocessableEntity, "admin/pages/citizens/list.html", data)
}

func citizenFlashURL(flash, msg string) string {
	return "/admin/citizens?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

// ListCitizens handles GET /admin/citizens
func (h *AdminCitizenHandler) ListCitizens(c *echo.Context) error {
	page, limit := parsePagination(c)
	offset := (page - 1) * limit
	rows, total, err := h.profileRepo.ListAllForExport(offset, limit)
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"Title":          configs.T(c, "ui.citizens.title", nil),
		"CurrentPath":    "/admin/citizens",
		"CurrentUser":    adminCurrentUser(c),
		"Citizens":       rows,
		"Pagination":     utils.NewPagination(page, limit, total),
		"ImportAction":   "/admin/citizens/import",
		"TemplateAction": "/admin/citizens/template",
		"Flash":          flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/citizens/list.html", data)
}
