package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

// stubRenderer is already declared in admin_user_handler_test.go (same package).
// newAdminEcho() (admin_user_handler_test.go) sets e.Renderer = &stubRenderer{}.

// --- fake serviceCatalogSvc ---

type fakeCatalogSvc struct {
	listFn     func(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error)
	getByIDFn  func(ctx context.Context, id string) (*models.ServiceType, error)
}

func (f *fakeCatalogSvc) List(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error) {
	if f.listFn != nil {
		return f.listFn(ctx, filter)
	}
	return &repositories.ListResult{Items: nil, Total: 0}, nil
}
func (f *fakeCatalogSvc) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return &models.ServiceType{
		ID:         id,
		Name:       "Test Service",
		FormSchema: json.RawMessage(`{"fields":["name"],"required_fields":["name"]}`),
	}, nil
}
func (f *fakeCatalogSvc) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	return f.GetByID(ctx, id)
}
func (f *fakeCatalogSvc) ListDepartments(ctx context.Context) ([]models.Department, error) { return nil, nil }
func (f *fakeCatalogSvc) ListCategories(ctx context.Context) ([]models.Category, error)    { return nil, nil }
func (f *fakeCatalogSvc) Create(ctx context.Context, st *models.ServiceType) error         { return nil }
func (f *fakeCatalogSvc) Update(ctx context.Context, st *models.ServiceType) error         { return nil }
func (f *fakeCatalogSvc) Delete(ctx context.Context, id string) error                      { return nil }

// --- fake citizenAppSvc ---

type fakeAppWebSvc struct {
	submitFn     func(uid string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error)
	listFn       func(uid string, page, limit int) ([]models.Application, int64, error)
	getFn        func(uid, appID string) (*dtos.ApplicationResponse, error)
	supplementFn func(uid, appID string, files []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error)
}

func (f *fakeAppWebSvc) SubmitApplication(uid string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error) {
	if f.submitFn != nil {
		return f.submitFn(uid, req, files)
	}
	return &dtos.ApplicationResponse{ID: "app-id", ApplicationCode: "CODE-001"}, nil
}
func (f *fakeAppWebSvc) ListMyApplications(uid string, page, limit int) ([]models.Application, int64, error) {
	if f.listFn != nil {
		return f.listFn(uid, page, limit)
	}
	return nil, 0, nil
}
func (f *fakeAppWebSvc) GetMyApplication(uid, appID string) (*dtos.ApplicationResponse, error) {
	if f.getFn != nil {
		return f.getFn(uid, appID)
	}
	return &dtos.ApplicationResponse{ID: appID, ApplicationCode: "CODE-001", Status: "received"}, nil
}
func (f *fakeAppWebSvc) ListMyApplicationStatusHistory(uid, appID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	return nil, 0, nil
}
func (f *fakeAppWebSvc) UploadMyApplicationSupplements(uid, appID string, files []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error) {
	if f.supplementFn != nil {
		return f.supplementFn(uid, appID, files)
	}
	return nil, nil
}

// --- helpers ---

func newCitizenCtx(e *echo.Echo, method, target string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func setCitizenUser(c *echo.Context, id string) {
	c.Set("user", &configs.JwtCustomClaims{ID: id, Email: "citizen@test.com", Role: "citizen"})
}

func setPathParam(c *echo.Context, name, value string) {
	c.SetPathValues(echo.PathValues{{Name: name, Value: value}})
}

// --- ShowServiceCatalog tests ---

func TestShowServiceCatalog_ReturnsErrorWhenCatalogNil(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/services")
	setCitizenUser(c, "u1")

	err := h.ShowServiceCatalog(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 HTTPError, got %v", err)
	}
}

func TestShowServiceCatalog_RendersListPage(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(catalog)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/services")
	setCitizenUser(c, "u1")

	if err := h.ShowServiceCatalog(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- ShowApplicationsList tests ---

func TestShowApplicationsList_ReturnsErrorWhenNoUser(t *testing.T) {
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications")
	// no user set in context

	err := h.ShowApplicationsList(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 HTTPError, got %v", err)
	}
}

func TestShowApplicationsList_RendersListPage(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/applications")
	setCitizenUser(c, "u1")

	if err := h.ShowApplicationsList(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- SubmitApplication tests ---

func TestSubmitApplication_RedirectsOnSuccess(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(catalog).
		WithApplicationService(appSvc)

	req := httptest.NewRequest(http.MethodPost, "/citizen/applications", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Language", "vi")
	req.ParseForm()
	req.Form.Set("service_type_id", "svc-id")
	req.Form.Set("name", "Nguyen Van A")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.SubmitApplication(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", rec.Code)
	}
	if rec.Header().Get("Location") == "" {
		t.Fatal("expected Location header on redirect")
	}
}

func TestSubmitApplication_RendersFormOnServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	appSvc := &fakeAppWebSvc{
		submitFn: func(uid string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error) {
			return nil, services.ErrMissingRequiredField
		},
	}
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(catalog).
		WithApplicationService(appSvc)

	req := httptest.NewRequest(http.MethodPost, "/citizen/applications", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Language", "vi")
	req.ParseForm()
	req.Form.Set("service_type_id", "svc-id")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.SubmitApplication(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

// --- ShowApplicationDetail tests ---

func TestShowApplicationDetail_Returns404WhenNotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{
		getFn: func(uid, appID string) (*dtos.ApplicationResponse, error) {
			return nil, services.ErrApplicationNotFound
		},
	}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications/unknown-id")
	setPathParam(c, "id", "unknown-id")
	setCitizenUser(c, "u1")

	err := h.ShowApplicationDetail(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusNotFound {
		t.Fatalf("expected 404 HTTPError, got %v", err)
	}
}

func TestShowApplicationDetail_RendersDetailPage(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/applications/app-id")
	setPathParam(c, "id", "app-id")
	setCitizenUser(c, "u1")

	if err := h.ShowApplicationDetail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- ShowLoginPage ---

func TestShowLoginPage_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/login")
	if err := h.ShowLoginPage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- WebLogin ---

func TestWebLogin_Success_CitizenRedirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "u1", Role: models.UserRoleCitizen}, "access", "refresh", nil
		},
	})
	form := "email=citizen%40example.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebLogin(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestWebLogin_NonCitizen_ReturnsUnauthorized(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "a1", Role: models.UserRoleSuperAdmin}, "access", "refresh", nil
		},
	})
	form := "email=admin%40example.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebLogin(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWebLogin_NilUser_Returns500(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", nil
		},
	})
	form := "email=x%40x.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.WebLogin(c)
	if err == nil {
		t.Fatal("expected error for nil user")
	}
}

func TestWebLogin_LoginError_RendersForm(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrUserNotFound
		},
	})
	form := "email=x%40x.com&password=wrong"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebLogin(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWebLogin_PasswordMismatch(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrPasswordMismatch
		},
	})
	form := "email=x%40x.com&password=wrong"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.WebLogin(c)
}

func TestWebLogin_UserBlocked(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrUserBlocked
		},
	})
	form := "email=x%40x.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.WebLogin(c)
}

func TestWebLogin_GenericError(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", errors.New("generic error")
		},
	})
	form := "email=x%40x.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.WebLogin(c)
}

// --- ShowRegisterPage ---

func TestShowRegisterPage_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/register")
	if err := h.ShowRegisterPage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- WebRegister ---

func TestWebRegister_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		registerFn: func(req *dtos.RegisterRequest) (*models.User, error) {
			return &models.User{ID: "u1", Role: models.UserRoleCitizen}, nil
		},
	})
	// All required fields: name, email, citizen_id_number (12 digits), password, confirm_password
	form := "email=new%40example.com&password=pass123&confirm_password=pass123&name=User&citizen_id_number=123456789012"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestWebRegister_Error_RendersForm(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		registerFn: func(req *dtos.RegisterRequest) (*models.User, error) {
			return nil, errors.New("email already exists")
		},
	})
	form := "email=existing%40example.com&password=pass123&name=User"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

// --- WebLogout ---

func TestWebLogout_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/logout")
	if err := h.WebLogout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestWebLogout_WithUser_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/logout")
	setCitizenUser(c, "u1")
	if err := h.WebLogout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

// --- ShowDashboard ---

func TestShowDashboard_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen")
	setCitizenUser(c, "u1")
	if err := h.ShowDashboard(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- ListNotifications ---

func TestListNotifications_NoUser_Returns500(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/notifications")
	err := h.ListNotifications(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type fakeNotifSvc struct {
	items      []dtos.NotificationResponse
	total      int64
	listErr    error
	markErr    error
	markAllErr error
	countVal   int64
}

func (s *fakeNotifSvc) List(_ string, _ repositories.NotificationFilter, _, _ int) ([]dtos.NotificationResponse, int64, error) {
	return s.items, s.total, s.listErr
}
func (s *fakeNotifSvc) MarkAsRead(_, _ string) error { return s.markErr }
func (s *fakeNotifSvc) MarkAllAsRead(_ string) error { return s.markAllErr }
func (s *fakeNotifSvc) CountUnread(_ string) (int64, error) {
	return s.countVal, nil
}

func TestListNotifications_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/notifications")
	setCitizenUser(c, "u1")
	if err := h.ListNotifications(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListNotifications_WithIsReadFilter(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	req := httptest.NewRequest(http.MethodGet, "/citizen/notifications?is_read=true&type=system", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")
	if err := h.ListNotifications(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListNotifications_WithIsReadFalse(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	req := httptest.NewRequest(http.MethodGet, "/citizen/notifications?is_read=false", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")
	_ = h.ListNotifications(c)
}

func TestListNotifications_ServiceError(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{listErr: errors.New("db error")})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/notifications")
	setCitizenUser(c, "u1")
	err := h.ListNotifications(c)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- MarkNotificationRead ---

func TestMarkNotificationRead_NoUser_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/n1/read")
	_ = h.MarkNotificationRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestMarkNotificationRead_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/n1/read")
	setCitizenUser(c, "u1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "n1"}})
	_ = h.MarkNotificationRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

// --- MarkAllNotificationsRead ---

func TestMarkAllNotificationsRead_NoUser_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/read-all")
	_ = h.MarkAllNotificationsRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestMarkAllNotificationsRead_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/read-all")
	setCitizenUser(c, "u1")
	_ = h.MarkAllNotificationsRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestMarkAllNotificationsRead_Error_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{markAllErr: errors.New("fail")})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/read-all")
	setCitizenUser(c, "u1")
	_ = h.MarkAllNotificationsRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

// --- ShowServiceDetail ---

func TestShowServiceDetail_NoService_Returns500(t *testing.T) {
	e := newAdminEcho()
	// catalogSvc is nil → should return 500
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/services/svc-1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "svc-1"}})
	err := h.ShowServiceDetail(c)
	if err == nil {
		t.Fatal("expected error for nil catalogSvc")
	}
}

func TestShowServiceDetail_GetByIDError_Returns404(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(&fakeCatalogSvc{
		getByIDFn: func(_ context.Context, _ string) (*models.ServiceType, error) {
			return nil, errors.New("not found")
		},
	})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/services/svc-1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "svc-1"}})
	err := h.ShowServiceDetail(c)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestShowServiceDetail_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(&fakeCatalogSvc{
		getByIDFn: func(_ context.Context, _ string) (*models.ServiceType, error) {
			return &models.ServiceType{ID: "svc-1", Name: "Service"}, nil
		},
	})
	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/services/svc-1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "svc-1"}})
	setCitizenUser(c, "u1")
	if err := h.ShowServiceDetail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- WithActivityLogger / WithNotificationService ---

func TestCitizenWebHandler_WithActivityLogger_Returns(t *testing.T) {
	h := NewCitizenWebHandler(&fakeAuthService{})
	logger := &fakeActivityLogger{}
	result := h.WithActivityLogger(logger)
	if result == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCitizenWebHandler_WithNotificationService_Returns(t *testing.T) {
	h := NewCitizenWebHandler(&fakeAuthService{})
	result := h.WithNotificationService(&fakeNotifSvc{})
	if result == nil {
		t.Fatal("expected non-nil handler")
	}
}

// --- isAllDigits ---

func TestIsAllDigits_True(t *testing.T) {
	if !isAllDigits("12345") {
		t.Fatal("expected true for all digits")
	}
}

func TestIsAllDigits_False(t *testing.T) {
	if isAllDigits("123abc") {
		t.Fatal("expected false for non-digits")
	}
}

func TestIsAllDigits_Empty(t *testing.T) {
	if isAllDigits("") {
		t.Fatal("expected false for empty string")
	}
}
