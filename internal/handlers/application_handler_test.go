package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mock ---

type mockApplicationSvc struct{ mock.Mock }

func (m *mockApplicationSvc) SubmitApplication(uid string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error) {
	args := m.Called(uid, req, files)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.ApplicationResponse), args.Error(1)
}

func (m *mockApplicationSvc) ListMyApplications(uid string, page, limit int) ([]models.Application, int64, error) {
	args := m.Called(uid, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Application), args.Get(1).(int64), args.Error(2)
}

func (m *mockApplicationSvc) GetMyApplication(uid, appID string) (*dtos.ApplicationResponse, error) {
	args := m.Called(uid, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.ApplicationResponse), args.Error(1)
}

// newApplicationHandler bypasses the concrete type requirement via the interface.
func newApplicationHandler(svc *mockApplicationSvc) *handlers.ApplicationHandler {
	return handlers.NewApplicationHandlerFromSvc(svc)
}

// --- multipart helper ---

func makeMultipartRequest(e *echo.Echo, method, target, dataJSON string) (*echo.Context, *httptest.ResponseRecorder) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	_ = w.WriteField("data", dataJSON)
	w.Close()

	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// --- Submit ---

func TestSubmit_MissingDataField(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/citizens/me/applications", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.Submit(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestSubmit_InvalidJSON(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)

	c, _ := makeMultipartRequest(e, http.MethodPost, "/api/citizens/me/applications", `{bad json`)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.Submit(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestSubmit_ValidationError_MissingServiceTypeID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)

	data, _ := json.Marshal(map[string]any{
		"submitted_data": map[string]string{"name": "An"},
	})
	c, _ := makeMultipartRequest(e, http.MethodPost, "/api/citizens/me/applications", string(data))
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.Submit(c)
	assert.Error(t, err)
}

func TestSubmit_ServiceTypeNotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)

	svc.On("SubmitApplication", "u1", mock.Anything, mock.Anything).
		Return(nil, services.ErrServiceTypeNotFound)

	data, _ := json.Marshal(map[string]any{
		"service_type_id": "550e8400-e29b-41d4-a716-446655440000",
		"submitted_data":  map[string]string{"name": "An"},
	})
	c, _ := makeMultipartRequest(e, http.MethodPost, "/api/citizens/me/applications", string(data))
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.Submit(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusUnprocessableEntity, he.Code)
	svc.AssertExpectations(t)
}

func TestSubmit_ServiceTypeInactive(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)
	svc.On("SubmitApplication", "u1", mock.Anything, mock.Anything).
		Return(nil, services.ErrServiceTypeInactive)

	data, _ := json.Marshal(map[string]any{
		"service_type_id": "550e8400-e29b-41d4-a716-446655440000",
		"submitted_data":  map[string]string{"name": "An"},
	})
	c, _ := makeMultipartRequest(e, http.MethodPost, "/api/citizens/me/applications", string(data))
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.Submit(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusUnprocessableEntity, he.Code)
	svc.AssertExpectations(t)
}

func TestSubmit_TooManyAttachments(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)
	svc.On("SubmitApplication", "u1", mock.Anything, mock.Anything).
		Return(nil, services.ErrTooManyAttachments)

	data, _ := json.Marshal(map[string]any{
		"service_type_id": "550e8400-e29b-41d4-a716-446655440000",
		"submitted_data":  map[string]string{"name": "An"},
	})
	c, _ := makeMultipartRequest(e, http.MethodPost, "/api/citizens/me/applications", string(data))
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.Submit(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusUnprocessableEntity, he.Code)
	svc.AssertExpectations(t)
}

func TestSubmit_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)

	expected := &dtos.ApplicationResponse{
		ID:              "app-1",
		ApplicationCode: "APP-20260520-ABCDEF",
		ServiceTypeName: "Cấp CCCD",
		Status:          "received",
		SubmittedAt:     time.Now(),
	}
	svc.On("SubmitApplication", "u1", mock.Anything, mock.Anything).Return(expected, nil)

	data, _ := json.Marshal(map[string]any{
		"service_type_id": "550e8400-e29b-41d4-a716-446655440000",
		"submitted_data":  map[string]string{"full_name": "An"},
	})
	c, rec := makeMultipartRequest(e, http.MethodPost, "/api/citizens/me/applications", string(data))
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.Submit(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	app := body["application"].(map[string]any)
	assert.Equal(t, "APP-20260520-ABCDEF", app["application_code"])
	assert.Equal(t, "received", app["status"])
	svc.AssertExpectations(t)
}

// --- ListMine ---

func TestListMine_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)

	apps := []models.Application{
		{
			ID:              "app1",
			ApplicationCode: "APP-20260520-ABCDEF",
			ServiceType:     models.ServiceType{Name: "Cấp CCCD"},
			Status:          models.ApplicationStatusReceived,
			SubmittedAt:     time.Now(),
		},
	}
	svc.On("ListMyApplications", "u1", 1, 10).Return(apps, int64(1), nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/citizens/me/applications")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ListMine(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	items := body["applications"].([]any)
	assert.Equal(t, "Cấp CCCD", items[0].(map[string]any)["service_type_name"])
	pagination := body["pagination"].(map[string]any)
	assert.Equal(t, float64(1), pagination["total"])
	svc.AssertExpectations(t)
}

func TestListMine_Empty(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)
	svc.On("ListMyApplications", "u1", 1, 10).Return([]models.Application{}, int64(0), nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/citizens/me/applications")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ListMine(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestListMine_InternalError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)
	svc.On("ListMyApplications", "u1", 1, 10).Return(nil, int64(0), errors.New("db error"))

	c, _ := makeRequest(e, http.MethodGet, "/api/citizens/me/applications")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ListMine(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	svc.AssertExpectations(t)
}

// --- GetMine ---

func TestGetMine_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)

	expected := &dtos.ApplicationResponse{
		ID:              "app-1",
		ApplicationCode: "APP-20260520-ABCDEF",
		ServiceTypeName: "Cấp CCCD",
		Status:          "received",
	}
	svc.On("GetMyApplication", "u1", "app-1").Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/applications/app-1", nil)
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "app-1"}})
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.GetMine(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	app := body["application"].(map[string]any)
	assert.Equal(t, "APP-20260520-ABCDEF", app["application_code"])
	svc.AssertExpectations(t)
}

func TestGetMine_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)
	svc.On("GetMyApplication", "u1", "missing").Return(nil, services.ErrApplicationNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/applications/missing", nil)
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "missing"}})
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.GetMine(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusNotFound, he.Code)
	svc.AssertExpectations(t)
}

func TestGetMine_InternalError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockApplicationSvc)
	h := newApplicationHandler(svc)
	svc.On("GetMyApplication", "u1", "app-1").Return(nil, errors.New("unexpected db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/applications/app-1", nil)
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "app-1"}})
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.GetMine(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	svc.AssertExpectations(t)
}
