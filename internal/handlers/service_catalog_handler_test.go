package handlers_test

import (
	"encoding/json"
	"errors"
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

func (m *mockServiceCatalogSvc) List(filter repositories.ListFilter) (*repositories.ListResult, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ListResult), args.Error(1)
}

func (m *mockServiceCatalogSvc) GetByID(id string) (*models.ServiceType, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceType), args.Error(1)
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
	svc.On("List", filter).Return(&repositories.ListResult{
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
	svc.On("List", filter).Return(&repositories.ListResult{Items: []models.ServiceType{}, Total: 0}, nil)

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
	svc.On("List", filter).Return(&repositories.ListResult{Items: []models.ServiceType{}, Total: 0}, nil)

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
	svc.On("List", filter).Return(nil, errors.New("db error"))

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
	svc.On("GetByID", id).Return(&models.ServiceType{Name: "Cấp CCCD lần đầu"}, nil)

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

	svc.On("GetByID", "bad-id").Return(nil, gorm.ErrRecordNotFound)

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

	svc.On("GetByID", "some-id").Return(nil, errors.New("unexpected db error"))

	c, _ := makeRequest(e, http.MethodGet, "/api/v1/services/some-id")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "some-id"}})

	err := h.GetService(c)

	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}
