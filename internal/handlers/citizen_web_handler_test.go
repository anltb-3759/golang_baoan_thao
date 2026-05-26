package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
