package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// --- fake category service ---

type fakeCatSvc struct {
	cats      []models.Category
	total     int64
	cat       *models.Category
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (s *fakeCatSvc) ListCategories(_ repositories.CategoryFilter, _, _ int) ([]models.Category, int64, error) {
	return s.cats, s.total, s.listErr
}
func (s *fakeCatSvc) GetCategory(_ string) (*models.Category, error) {
	return s.cat, s.getErr
}
func (s *fakeCatSvc) CreateCategory(_ *dtos.CategoryCreateRequest, _ string) (*models.Category, error) {
	return s.cat, s.createErr
}
func (s *fakeCatSvc) UpdateCategory(_ string, _ *dtos.CategoryUpdateRequest, _ string) (*models.Category, error) {
	return s.cat, s.updateErr
}
func (s *fakeCatSvc) DeleteCategory(_ string, _ string) error { return s.deleteErr }

var _ CategoryService = (*fakeCatSvc)(nil)

func newCatHandler(svc *fakeCatSvc) *AdminCategoryHandler {
	return NewAdminCategoryHandler(svc)
}

// --- ListCategories ---

func TestAdminCatHandler_ListCategories_OK(t *testing.T) {
	e := newAdminEcho()
	svc := &fakeCatSvc{cats: []models.Category{{ID: "1", Name: "Y tế"}}, total: 1}
	h := newCatHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/categories", "", "")
	err := h.ListCategories(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminCatHandler_ListCategories_ServiceError(t *testing.T) {
	e := newAdminEcho()
	svc := &fakeCatSvc{listErr: errors.New("db error")}
	h := newCatHandler(svc)

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/categories", "", "")
	err := h.ListCategories(c)
	assert.Error(t, err)
}

// --- ShowCreateForm ---

func TestAdminCatHandler_ShowCreateForm_OK(t *testing.T) {
	e := newAdminEcho()
	h := newCatHandler(&fakeCatSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/categories/new", "", "")
	err := h.ShowCreateForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- CreateCategory ---

func TestAdminCatHandler_CreateCategory_OK(t *testing.T) {
	e := newAdminEcho()
	created := &models.Category{ID: "c1", Name: "Y tế"}
	svc := &fakeCatSvc{cat: created}
	h := newCatHandler(svc)

	form := url.Values{"name": {"Y tế"}, "code": {"y_te"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories", form)
	err := h.CreateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/admin/categories")
}

func TestAdminCatHandler_CreateCategory_ValidationError(t *testing.T) {
	e := newAdminEcho()
	h := newCatHandler(&fakeCatSvc{})

	form := url.Values{"name": {""}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories", form)
	err := h.CreateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminCatHandler_CreateCategory_CodeExists(t *testing.T) {
	e := newAdminEcho()
	svc := &fakeCatSvc{createErr: services.ErrCategoryCodeExists}
	h := newCatHandler(svc)

	form := url.Values{"name": {"Y tế"}, "code": {"y_te"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories", form)
	err := h.CreateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminCatHandler_CreateCategory_ServiceError(t *testing.T) {
	e := newAdminEcho()
	svc := &fakeCatSvc{createErr: errors.New("internal")}
	h := newCatHandler(svc)

	form := url.Values{"name": {"Y tế"}, "code": {"y_te"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories", form)
	err := h.CreateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// --- ShowEditForm ---

func TestAdminCatHandler_ShowEditForm_OK(t *testing.T) {
	e := newAdminEcho()
	cat := &models.Category{ID: "c1", Name: "Y tế"}
	svc := &fakeCatSvc{cat: cat}
	h := newCatHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/categories/c1/edit", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "c1"}})
	err := h.ShowEditForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminCatHandler_ShowEditForm_NotFound(t *testing.T) {
	e := newAdminEcho()
	svc := &fakeCatSvc{getErr: errors.New("not found")}
	h := newCatHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/categories/bad/edit", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.ShowEditForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- UpdateCategory ---

func TestAdminCatHandler_UpdateCategory_OK(t *testing.T) {
	e := newAdminEcho()
	cat := &models.Category{ID: "c1", Name: "Y tế"}
	svc := &fakeCatSvc{cat: cat}
	h := newCatHandler(svc)

	form := url.Values{"name": {"Updated"}, "code": {"y_te"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories/c1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "c1"}})
	err := h.UpdateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminCatHandler_UpdateCategory_NotFound(t *testing.T) {
	e := newAdminEcho()
	svc := &fakeCatSvc{getErr: errors.New("not found")}
	h := newCatHandler(svc)

	form := url.Values{"name": {"X"}, "code": {"x"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories/bad/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.UpdateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminCatHandler_UpdateCategory_ValidationError(t *testing.T) {
	e := newAdminEcho()
	cat := &models.Category{ID: "c1"}
	svc := &fakeCatSvc{cat: cat}
	h := newCatHandler(svc)

	form := url.Values{"name": {""}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories/c1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "c1"}})
	err := h.UpdateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminCatHandler_UpdateCategory_CodeExists(t *testing.T) {
	e := newAdminEcho()
	cat := &models.Category{ID: "c1"}
	svc := &fakeCatSvc{cat: cat, updateErr: services.ErrCategoryCodeExists}
	h := newCatHandler(svc)

	form := url.Values{"name": {"Y tế"}, "code": {"y_te"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/categories/c1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "c1"}})
	err := h.UpdateCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// --- DeleteCategory ---

func TestAdminCatHandler_DeleteCategory_OK(t *testing.T) {
	e := newAdminEcho()
	h := newCatHandler(&fakeCatSvc{})

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/categories/c1/delete", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "c1"}})
	err := h.DeleteCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=success")
}

func TestAdminCatHandler_DeleteCategory_Error(t *testing.T) {
	e := newAdminEcho()
	svc := &fakeCatSvc{deleteErr: errors.New("delete failed")}
	h := newCatHandler(svc)

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/categories/c1/delete", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "c1"}})
	err := h.DeleteCategory(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=error")
}
