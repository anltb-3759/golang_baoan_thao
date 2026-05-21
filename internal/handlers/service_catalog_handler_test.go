package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- mock ---

type mockServiceCatalogSvc struct {
	mock.Mock
}

func (m *mockServiceCatalogSvc) List(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ListResult), args.Error(1)
}

func (m *mockServiceCatalogSvc) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceType), args.Error(1)
}

func (m *mockServiceCatalogSvc) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceType), args.Error(1)
}

func (m *mockServiceCatalogSvc) ListDepartments(ctx context.Context) ([]models.Department, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Department), args.Error(1)
}

func (m *mockServiceCatalogSvc) ListCategories(ctx context.Context) ([]models.Category, error) {
	return nil, nil
}

func (m *mockServiceCatalogSvc) Create(ctx context.Context, st *models.ServiceType) error {
	args := m.Called(ctx, st)
	return args.Error(0)
}

func (m *mockServiceCatalogSvc) Update(ctx context.Context, st *models.ServiceType) error {
	args := m.Called(ctx, st)
	return args.Error(0)
}

func (m *mockServiceCatalogSvc) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// --- helpers ---

func newTestEcho() *echo.Echo {
	_ = configs.LoadI18nMessages("../../locales")
	e := echo.New()
	e.Validator = &configs.CustomValidator{Validator: validator.New()}
	e.HTTPErrorHandler = configs.CustomHTTPErrorHandler
	return e
}

func makeRequest(e *echo.Echo, method, target string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// --- tests ---

func TestListServices_Success(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()

	filter := repositories.ListFilter{Category: "", Search: "", Page: 1, Limit: 20}
	svc.On("List", mock.Anything, filter).Return(&repositories.ListResult{
		Items: []models.ServiceType{{Name: "Cấp CCCD lần đầu", Code: "CCCD_NEW"}},
		Total: 1,
	}, nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/v1/services")
	err := h.ListServices(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	pagination := body["pagination"].(map[string]any)
	assert.Equal(t, float64(1), pagination["total"])
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(20), pagination["limit"])
	svc.AssertExpectations(t)
}

func TestListServices_WithQueryParams(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()

	filter := repositories.ListFilter{Category: "hanh_chinh_cong", Search: "CCCD", Page: 2, Limit: 5}
	svc.On("List", mock.Anything, filter).Return(&repositories.ListResult{Items: []models.ServiceType{}, Total: 0}, nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/v1/services?page=2&limit=5&category=hanh_chinh_cong&search=CCCD")
	err := h.ListServices(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestListServices_InvalidPageFallsBackToDefault(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()

	// page=-5 → clamped to 1; limit=999 → clamped to 20
	filter := repositories.ListFilter{Page: 1, Limit: 20}
	svc.On("List", mock.Anything, filter).Return(&repositories.ListResult{Items: []models.ServiceType{}, Total: 0}, nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/v1/services?page=-5&limit=999")
	err := h.ListServices(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestListServices_ServiceError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()

	filter := repositories.ListFilter{Page: 1, Limit: 20}
	svc.On("List", mock.Anything, filter).Return(nil, errors.New("db error"))

	c, _ := makeRequest(e, http.MethodGet, "/api/v1/services")
	err := h.ListServices(c)

	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

func TestGetService_Success(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()

	id := "some-uuid"
	svc.On("GetByID", mock.Anything, id).Return(&models.ServiceType{Name: "Cấp CCCD lần đầu"}, nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/v1/services/"+id)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: id}})

	err := h.GetService(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestGetService_NotFound(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()

	svc.On("GetByID", mock.Anything, "bad-id").Return(nil, gorm.ErrRecordNotFound)

	c, _ := makeRequest(e, http.MethodGet, "/api/v1/services/bad-id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad-id"}})

	err := h.GetService(c)

	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusNotFound, he.Code)
	assert.Equal(t, "service.not_found", he.Message)
}

func TestGetService_InternalError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()

	svc.On("GetByID", mock.Anything, "some-id").Return(nil, errors.New("unexpected db error"))

	c, _ := makeRequest(e, http.MethodGet, "/api/v1/services/some-id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "some-id"}})

	err := h.GetService(c)

	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

// --- tests for admin ShowServiceType ---

type stubRenderer struct{}

func (r *stubRenderer) Render(_ *echo.Context, w io.Writer, _ string, _ any) error {
	_, _ = w.Write([]byte("ok"))
	return nil
}

func TestShowServiceType_Success(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()
	e.Renderer = &stubRenderer{}

	id := "st-1"
	svc.On("GetByIDForAdmin", mock.Anything, id).Return(&models.ServiceType{ID: id, Name: "Test Service", Code: "TST"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-types/"+id, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: id}})
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Email: "admin@test.com", Role: "super_admin"})

	err := h.ShowServiceType(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestShowServiceType_NotFound(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()
	e.Renderer = &stubRenderer{}

	svc.On("GetByIDForAdmin", mock.Anything, "bad-id").Return(nil, gorm.ErrRecordNotFound)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-types/bad-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad-id"}})

	err := h.ShowServiceType(c)

	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusNotFound, he.Code)
	assert.Equal(t, "service_type.not_found", he.Message)
}

func TestShowServiceType_InternalError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	h := handlers.NewServiceCatalogHandler(svc)
	e := newTestEcho()
	e.Renderer = &stubRenderer{}

	svc.On("GetByIDForAdmin", mock.Anything, "err-id").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/admin/service-types/err-id", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "err-id"}})

	err := h.ShowServiceType(c)

	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}
