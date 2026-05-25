package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
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

type mockAdminUserSvc struct {
	mock.Mock
}

func (m *mockAdminUserSvc) ListUsers(filter repositories.UserFilter, page, limit int) ([]models.User, int64, error) {
	args := m.Called(filter, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Get(1).(int64), args.Error(2)
}
func (m *mockAdminUserSvc) GetUser(id string) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *mockAdminUserSvc) CreateUser(req *dtos.AdminCreateUserRequest, createdBy string) (*models.User, error) {
	return nil, nil
}
func (m *mockAdminUserSvc) UpdateUser(id string, req *dtos.AdminUpdateUserRequest, updatedBy string) (*models.User, error) {
	return nil, nil
}
func (m *mockAdminUserSvc) BlockUser(id string, updatedBy string) error   { return nil }
func (m *mockAdminUserSvc) UnblockUser(id string, updatedBy string) error { return nil }
func (m *mockAdminUserSvc) DeleteUser(id string, deletedBy string) error  { return nil }

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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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
	userSvc := new(mockAdminUserSvc)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
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

// --- helpers ---

func newAdminCtxExt(e *echo.Echo, method, path string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Email: "admin@test.com", Role: "super_admin"})
	return c, rec
}

func newFormCtxExt(e *echo.Echo, method, path string, values url.Values) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Email: "admin@test.com", Role: "super_admin"})
	return c, rec
}

func newMultipartCSVExt(content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "test.csv")
	_, _ = fw.Write([]byte(content))
	w.Close()
	return body, w.FormDataContentType()
}

func setupAdminEcho() *echo.Echo {
	e := newTestEcho()
	e.Renderer = &stubRenderer{}
	return e
}

func defaultMockSetup(svc *mockServiceCatalogSvc, userSvc *mockAdminUserSvc) {
	svc.On("ListDepartments", mock.Anything).Return([]models.Department{}, nil)
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
}

// --- AdminLanding ---

func TestAdminLanding_Redirect(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin")
	err := h.AdminLanding(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/admin/service-types", rec.Header().Get("Location"))
}

// --- ListServiceTypesAdmin ---

func TestListServiceTypesAdmin_OK(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	svc.On("List", mock.Anything, mock.MatchedBy(func(f repositories.ListFilter) bool {
		return f.IncludeInactive
	})).Return(&repositories.ListResult{Items: []models.ServiceType{{Name: "ST1"}}, Total: 1}, nil)

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types")
	err := h.ListServiceTypesAdmin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListServiceTypesAdmin_Error(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	svc.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	c, _ := newAdminCtxExt(e, http.MethodGet, "/admin/service-types")
	err := h.ListServiceTypesAdmin(c)
	assert.Error(t, err)
}

func TestListServiceTypesAdmin_FlashFromSuccess(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	svc.On("List", mock.Anything, mock.Anything).Return(&repositories.ListResult{Items: []models.ServiceType{}, Total: 0}, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/service-types?success=created", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "a1"})

	err := h.ListServiceTypesAdmin(c)
	assert.NoError(t, err)
}

// --- CreateServiceTypeForm ---

func TestCreateServiceTypeForm_OK(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/new")
	err := h.CreateServiceTypeForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateServiceTypeForm_DepsError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("ListDepartments", mock.Anything).Return(nil, errors.New("db error"))
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, _ := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/new")
	err := h.CreateServiceTypeForm(c)
	assert.Error(t, err)
}

// --- CreateServiceType ---

func TestCreateServiceType_OK(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	svc.On("Create", mock.Anything, mock.Anything).Return(nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Dịch vụ 1"}, "code": {"SVC001"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types", form)
	err := h.CreateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestCreateServiceType_ValidationError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {""}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types", form)
	err := h.CreateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateServiceType_InvalidProcessingTime(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Svc"}, "code": {"S1"}, "processing_time": {"not-a-number"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types", form)
	err := h.CreateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateServiceType_InvalidFee(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Svc"}, "code": {"S1"}, "fee": {"not-a-fee"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types", form)
	err := h.CreateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateServiceType_InvalidFormSchema(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Svc"}, "code": {"S1"}, "form_schema": {"not-json"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types", form)
	err := h.CreateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateServiceType_ServiceError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	svc.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Svc"}, "code": {"S1"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types", form)
	err := h.CreateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateServiceType_CodeExists(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	svc.On("Create", mock.Anything, mock.Anything).Return(services.ErrServiceTypeCodeExists)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Svc"}, "code": {"S1"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types", form)
	err := h.CreateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- EditServiceTypeForm ---

func TestEditServiceTypeForm_OK(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	st := &models.ServiceType{ID: "st1", Name: "Svc", Code: "S1"}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/st1/edit")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.EditServiceTypeForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEditServiceTypeForm_NotFound(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("GetByIDForAdmin", mock.Anything, "bad").Return(nil, gorm.ErrRecordNotFound)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, _ := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/bad/edit")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.EditServiceTypeForm(c)
	assert.Error(t, err)
}

func TestEditServiceTypeForm_DepsError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	st := &models.ServiceType{ID: "st1", Name: "Svc", Code: "S1"}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	svc.On("ListDepartments", mock.Anything).Return(nil, errors.New("db error"))
	userSvc.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return([]models.User{}, int64(0), nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, _ := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/st1/edit")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.EditServiceTypeForm(c)
	assert.Error(t, err)
}

// --- UpdateServiceType ---

func TestUpdateServiceType_OK(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	st := &models.ServiceType{ID: "st1", Name: "Old", Code: "O1"}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	svc.On("Update", mock.Anything, mock.Anything).Return(nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"New Svc"}, "code": {"NEW1"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types/st1", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.UpdateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestUpdateServiceType_NotFound(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("GetByIDForAdmin", mock.Anything, "bad").Return(nil, gorm.ErrRecordNotFound)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Svc"}, "code": {"S1"}}
	c, _ := newFormCtxExt(e, http.MethodPost, "/admin/service-types/bad", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.UpdateServiceType(c)
	assert.Error(t, err)
}

func TestUpdateServiceType_ValidationError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	st := &models.ServiceType{ID: "st1", Name: "Old", Code: "O1"}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {""}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types/st1", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.UpdateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateServiceType_ModelError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	st := &models.ServiceType{ID: "st1", Name: "Old", Code: "O1"}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"Svc"}, "code": {"S1"}, "processing_time": {"bad"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types/st1", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.UpdateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateServiceType_ServiceError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	st := &models.ServiceType{ID: "st1", Name: "Old", Code: "O1"}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	svc.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"New Svc"}, "code": {"NEW1"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types/st1", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.UpdateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateServiceType_CodeExists(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	st := &models.ServiceType{ID: "st1", Name: "Old", Code: "O1"}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	svc.On("Update", mock.Anything, mock.Anything).Return(services.ErrServiceTypeCodeExists)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	form := url.Values{"name": {"New Svc"}, "code": {"NEW1"}}
	c, rec := newFormCtxExt(e, http.MethodPost, "/admin/service-types/st1", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.UpdateServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- ExportCSV ---

type mockServiceTypeImportSvc struct {
	errs []string
}

func (m *mockServiceTypeImportSvc) ImportServiceTypes(_ []services.ServiceTypeImportRow, _ string) []string {
	return m.errs
}

func TestServiceCatalogHandler_ExportCSV_OK(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	pt := 5
	svc.On("List", mock.Anything, mock.Anything).Return(&repositories.ListResult{
		Items: []models.ServiceType{{Name: "Svc1", Description: "Desc", ProcessingTime: &pt, Fee: 100.0}},
		Total: 1,
	}, nil)

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/export")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, "text/csv; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "Svc1")
}

func TestServiceCatalogHandler_ExportCSV_Empty(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	svc.On("List", mock.Anything, mock.Anything).Return(&repositories.ListResult{Items: []models.ServiceType{}, Total: 0}, nil)

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/export")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "ten")
}

func TestServiceCatalogHandler_ExportCSV_Error(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	svc.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/export")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "ten")
}

func TestServiceCatalogHandler_ExportCSV_WithDept(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	dept := &models.Department{Code: "IT001"}
	svc.On("List", mock.Anything, mock.Anything).Return(&repositories.ListResult{
		Items: []models.ServiceType{{Name: "Svc", ResponsibleDepartment: dept}},
		Total: 1,
	}, nil)

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/export")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "IT001")
}

// --- ImportCSV ---

func TestServiceCatalogHandler_ImportCSV_NoFile(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc).WithImportExport(&mockServiceTypeImportSvc{})
	e := setupAdminEcho()

	req := httptest.NewRequest(http.MethodPost, "/admin/service-types/import", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "a1"})

	err := h.ImportCSV(c)
	assert.Error(t, err)
}

func TestServiceCatalogHandler_ImportCSV_WithErrors(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("List", mock.Anything, mock.Anything).Return(&repositories.ListResult{Items: []models.ServiceType{}, Total: 0}, nil)
	importSvc := &mockServiceTypeImportSvc{errs: []string{"Dòng 2: lỗi"}}
	h := handlers.NewServiceCatalogHandler(svc, userSvc).WithImportExport(importSvc)
	e := setupAdminEcho()

	body, ct := newMultipartCSVExt("ten,mo_ta\nSvc,Desc\n")
	req := httptest.NewRequest(http.MethodPost, "/admin/service-types/import", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "a1"})

	err := h.ImportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestServiceCatalogHandler_ImportCSV_Success(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc).WithImportExport(&mockServiceTypeImportSvc{})
	e := setupAdminEcho()

	body, ct := newMultipartCSVExt("ten,mo_ta\nCấp hộ khẩu,Mô tả\n")
	req := httptest.NewRequest(http.MethodPost, "/admin/service-types/import", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "a1"})

	err := h.ImportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- DeleteServiceType ---

func TestDeleteServiceType_OK(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("Delete", mock.Anything, "st1").Return(nil)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, rec := newAdminCtxExt(e, http.MethodPost, "/admin/service-types/st1/delete")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.DeleteServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestDeleteServiceType_NotFound(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("Delete", mock.Anything, "bad").Return(gorm.ErrRecordNotFound)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, rec := newAdminCtxExt(e, http.MethodPost, "/admin/service-types/bad/delete")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.DeleteServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestDeleteServiceType_HasApplications(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("Delete", mock.Anything, "st1").Return(services.ErrServiceTypeHasApplications)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, rec := newAdminCtxExt(e, http.MethodPost, "/admin/service-types/st1/delete")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.DeleteServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestDeleteServiceType_InternalError(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	svc.On("Delete", mock.Anything, "st1").Return(errors.New("db error"))
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	c, rec := newAdminCtxExt(e, http.MethodPost, "/admin/service-types/st1/delete")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.DeleteServiceType(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- serviceTypeFormFromModel coverage via EditServiceTypeForm ---

func TestEditServiceTypeForm_WithAllOptionalFields(t *testing.T) {
	svc := new(mockServiceCatalogSvc)
	userSvc := new(mockAdminUserSvc)
	defaultMockSetup(svc, userSvc)
	h := handlers.NewServiceCatalogHandler(svc, userSvc)
	e := setupAdminEcho()

	pt := 5
	deptID := "dept-1"
	catID := "cat-1"
	schema := []byte(`{"fields":[]}`)
	st := &models.ServiceType{
		ID: "st1", Name: "Svc", Code: "S1",
		ProcessingTime:          &pt,
		ResponsibleDepartmentID: &deptID,
		CategoryID:              &catID,
		FormSchema:              schema,
	}
	svc.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)

	c, rec := newAdminCtxExt(e, http.MethodGet, "/admin/service-types/st1/edit")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "st1"}})
	err := h.EditServiceTypeForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
