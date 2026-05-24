package handlers

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/stretchr/testify/assert"
)

// --- fakes ---

type fakeCitizenProfileRepo struct {
	rows     []repositories.CitizenExportRow
	total    int64
	listErr  error
}

func (r *fakeCitizenProfileRepo) ListAllForExport(_, _ int) ([]repositories.CitizenExportRow, int64, error) {
	return r.rows, r.total, r.listErr
}

type fakeCitizenImportSvc struct {
	errs []string
	err  error
}

func (s *fakeCitizenImportSvc) ImportCitizens(_ []services.CitizenImportRow, _ string) []string {
	return s.errs
}

func newCitizenHandler(repo *fakeCitizenProfileRepo, svc *fakeCitizenImportSvc) *AdminCitizenHandler {
	return NewAdminCitizenHandler(repo, svc)
}

// --- ImportCSV ---

func newMultipartCSV(content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "test.csv")
	_, _ = fw.Write([]byte(content))
	w.Close()
	return body, w.FormDataContentType()
}

func TestAdminCitizenHandler_ImportCSV_NoFile(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newCitizenHandler(&fakeCitizenProfileRepo{}, &fakeCitizenImportSvc{})

	req := httptest.NewRequest(http.MethodPost, "/admin/citizens/import", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ImportCSV(c)
	assert.Error(t, err)
}

func TestAdminCitizenHandler_ImportCSV_EmptyCSV(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newCitizenHandler(&fakeCitizenProfileRepo{}, &fakeCitizenImportSvc{})

	body, ct := newMultipartCSV("so_cccd,ho_ten,email\n") // header only, no data
	req := httptest.NewRequest(http.MethodPost, "/admin/citizens/import", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ImportCSV(c)
	assert.Error(t, err)
}

func TestAdminCitizenHandler_ImportCSV_WithImportErrors(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeCitizenImportSvc{errs: []string{"Dòng 2: Số CCCD phải có đúng 12 chữ số"}}
	h := newCitizenHandler(&fakeCitizenProfileRepo{}, svc)

	csvData := "so_cccd,ho_ten,email\n123,Test,t@t.com\n"
	body, ct := newMultipartCSV(csvData)
	req := httptest.NewRequest(http.MethodPost, "/admin/citizens/import", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ImportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminCitizenHandler_ImportCSV_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeCitizenImportSvc{errs: nil}
	h := newCitizenHandler(&fakeCitizenProfileRepo{}, svc)

	csvData := "so_cccd,ho_ten,email\n123456789012,Nguyễn Văn A,a@a.com\n"
	body, ct := newMultipartCSV(csvData)
	req := httptest.NewRequest(http.MethodPost, "/admin/citizens/import", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ImportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- ListCitizens ---

func TestAdminCitizenHandler_ListCitizens_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	rows := []repositories.CitizenExportRow{{SoCCCD: "111111111111", HoTen: "Test"}}
	h := newCitizenHandler(&fakeCitizenProfileRepo{rows: rows, total: 1}, &fakeCitizenImportSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/citizens", "", "")
	err := h.ListCitizens(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminCitizenHandler_ListCitizens_RepoError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newCitizenHandler(&fakeCitizenProfileRepo{listErr: errors.New("db error")}, &fakeCitizenImportSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/citizens", "", "")
	err := h.ListCitizens(c)
	assert.Error(t, err)
}
