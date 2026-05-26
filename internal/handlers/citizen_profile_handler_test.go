package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
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

type mockCitizenProfileSvc struct {
	mock.Mock
}

func (m *mockCitizenProfileSvc) GetProfile(userID string) (*dtos.CitizenProfileResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.CitizenProfileResponse), args.Error(1)
}

func (m *mockCitizenProfileSvc) UpdateProfile(userID string, req *dtos.UpdateCitizenProfileRequest) (*dtos.CitizenProfileResponse, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.CitizenProfileResponse), args.Error(1)
}

func (m *mockCitizenProfileSvc) ListMyApplications(userID string, page, limit int) ([]models.Application, int64, error) {
	args := m.Called(userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Application), args.Get(1).(int64), args.Error(2)
}

func (m *mockCitizenProfileSvc) ChangeMyPassword(userID string, req *dtos.ChangeMyPasswordRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

// --- helper to make request with body ---

func makeRequestWithBody(e *echo.Echo, method, target string, body []byte) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// --- GetMe tests ---

func TestGetMe_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)

	expected := &dtos.CitizenProfileResponse{
		UserID:          "u1",
		Name:            "An",
		Email:           "an@example.com",
		CitizenIDNumber: "123456789012",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	svc.On("GetProfile", "u1").Return(expected, nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/citizens/me")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.GetMe(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	data := body["profile"].(map[string]any)
	assert.Equal(t, "An", data["name"])
	assert.Equal(t, "123456789012", data["citizen_id_number"])
	svc.AssertExpectations(t)
}

func TestGetMe_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("GetProfile", "u1").Return(nil, services.ErrProfileNotFound)

	c, _ := makeRequest(e, http.MethodGet, "/api/citizens/me")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.GetMe(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusNotFound, he.Code)
	svc.AssertExpectations(t)
}

func TestGetMe_InternalError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("GetProfile", "u1").Return(nil, errors.New("db error"))

	c, _ := makeRequest(e, http.MethodGet, "/api/citizens/me")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.GetMe(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

// --- UpdateMe tests ---

func TestUpdateMe_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)

	newName := "New Name"
	expected := &dtos.CitizenProfileResponse{Name: "New Name", Email: "an@example.com"}
	svc.On("UpdateProfile", "u1", mock.MatchedBy(func(r *dtos.UpdateCitizenProfileRequest) bool {
		return r.Name != nil && *r.Name == "New Name"
	})).Return(expected, nil)

	body, _ := json.Marshal(&dtos.UpdateCitizenProfileRequest{Name: &newName})
	c, rec := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.UpdateMe(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestUpdateMe_BindError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)

	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me", []byte(`{invalid json`))
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.UpdateMe(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestUpdateMe_ValidateError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)

	// name="" fails omitempty,min=1
	empty := ""
	body, _ := json.Marshal(&dtos.UpdateCitizenProfileRequest{Name: &empty})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.UpdateMe(c)
	assert.Error(t, err)
}

func TestUpdateMe_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("UpdateProfile", "u1", mock.Anything).Return(nil, services.ErrProfileNotFound)

	body, _ := json.Marshal(&dtos.UpdateCitizenProfileRequest{})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.UpdateMe(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusNotFound, he.Code)
	svc.AssertExpectations(t)
}

func TestUpdateMe_InternalError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("UpdateProfile", "u1", mock.Anything).Return(nil, errors.New("unexpected"))

	body, _ := json.Marshal(&dtos.UpdateCitizenProfileRequest{})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.UpdateMe(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	svc.AssertExpectations(t)
}

// --- ListMyApplications tests ---

func TestListMyApplications_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)

	apps := []models.Application{
		{
			ID:              "app1",
			ApplicationCode: "HCM-2024-001",
			ServiceTypeID:   "svc1",
			Status:          models.ApplicationStatusReceived,
			SubmittedAt:     time.Now(),
		},
	}
	svc.On("ListMyApplications", "u1", 1, 10).Return(apps, int64(1), nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/citizens/me/applications")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ListMyApplications(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	pagination := body["pagination"].(map[string]any)
	assert.Equal(t, float64(1), pagination["total"])
	svc.AssertExpectations(t)
}

func TestListMyApplications_WithServiceTypeName(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)

	apps := []models.Application{
		{
			ID:              "app1",
			ApplicationCode: "APP-20240101-ABCDEF",
			ServiceType:     models.ServiceType{Name: "Cấp CCCD lần đầu"},
		},
	}
	svc.On("ListMyApplications", "u1", 1, 10).Return(apps, int64(1), nil)

	c, rec := makeRequest(e, http.MethodGet, "/api/citizens/me/applications")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ListMyApplications(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	items := body["applications"].([]any)
	assert.Equal(t, "Cấp CCCD lần đầu", items[0].(map[string]any)["service_type_name"])
	svc.AssertExpectations(t)
}

func TestListMyApplications_PagingClamping(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("ListMyApplications", "u1", 1, 10).Return([]models.Application{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/applications?page=0&limit=200", nil)
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ListMyApplications(c)
	assert.NoError(t, err)
	svc.AssertExpectations(t)
}

func TestListMyApplications_InternalError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("ListMyApplications", "u1", 1, 10).Return(nil, int64(0), errors.New("db error"))

	c, _ := makeRequest(e, http.MethodGet, "/api/citizens/me/applications")
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ListMyApplications(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
}

func TestChangeMyPassword_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()

	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)

	req := &dtos.ChangeMyPasswordRequest{
		CurrentPassword:    "oldpass123",
		NewPassword:        "newpass123",
		ConfirmNewPassword: "newpass123",
	}
	svc.On("ChangeMyPassword", "u1", mock.MatchedBy(func(r *dtos.ChangeMyPasswordRequest) bool {
		return r.CurrentPassword == req.CurrentPassword && r.NewPassword == req.NewPassword && r.ConfirmNewPassword == req.ConfirmNewPassword
	})).Return(nil)

	body, _ := json.Marshal(req)
	c, rec := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestChangeMyPassword_ValidateError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	h := handlers.NewCitizenProfileHandler(new(mockCitizenProfileSvc))

	body, _ := json.Marshal(&dtos.ChangeMyPasswordRequest{CurrentPassword: "123", NewPassword: "123", ConfirmNewPassword: "123"})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	assert.Error(t, err)
}

func TestChangeMyPassword_BindError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	h := handlers.NewCitizenProfileHandler(new(mockCitizenProfileSvc))

	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", []byte(`{invalid json`))
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "auth.invalid_request", he.Message)
}

func TestChangeMyPassword_CurrentPasswordMismatch(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("ChangeMyPassword", "u1", mock.Anything).Return(services.ErrPasswordMismatch)

	body, _ := json.Marshal(&dtos.ChangeMyPasswordRequest{CurrentPassword: "oldpass123", NewPassword: "newpass123", ConfirmNewPassword: "newpass123"})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, "auth.password_mismatch", he.Message)
}

func TestChangeMyPassword_ConfirmationMismatch(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("ChangeMyPassword", "u1", mock.Anything).Return(services.ErrPasswordConfirmationMismatch)

	body, _ := json.Marshal(&dtos.ChangeMyPasswordRequest{CurrentPassword: "oldpass123", NewPassword: "newpass123", ConfirmNewPassword: "differentpass"})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusUnprocessableEntity, he.Code)
	assert.Equal(t, "auth.password_confirmation_mismatch", he.Message)
}

func TestChangeMyPassword_NewEqualsCurrent(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("ChangeMyPassword", "u1", mock.Anything).Return(services.ErrNewPasswordMustDiffer)

	body, _ := json.Marshal(&dtos.ChangeMyPasswordRequest{CurrentPassword: "oldpass123", NewPassword: "oldpass123", ConfirmNewPassword: "oldpass123"})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusUnprocessableEntity, he.Code)
	assert.Equal(t, "auth.new_password_must_differ", he.Message)
}

func TestChangeMyPassword_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("ChangeMyPassword", "u1", mock.Anything).Return(services.ErrProfileNotFound)

	body, _ := json.Marshal(&dtos.ChangeMyPasswordRequest{CurrentPassword: "oldpass123", NewPassword: "newpass123", ConfirmNewPassword: "newpass123"})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusNotFound, he.Code)
	assert.Equal(t, "profile.not_found", he.Message)
}

func TestChangeMyPassword_InternalError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	svc := new(mockCitizenProfileSvc)
	h := handlers.NewCitizenProfileHandler(svc)
	svc.On("ChangeMyPassword", "u1", mock.Anything).Return(errors.New("db error"))

	body, _ := json.Marshal(&dtos.ChangeMyPasswordRequest{CurrentPassword: "oldpass123", NewPassword: "newpass123", ConfirmNewPassword: "newpass123"})
	c, _ := makeRequestWithBody(e, http.MethodPut, "/api/citizens/me/password", body)
	c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: "citizen"})

	err := h.ChangeMyPassword(c)
	var he *echo.HTTPError
	assert.True(t, errors.As(err, &he))
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	assert.Equal(t, "common.internal_error", he.Message)
}
