package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

type fakeAdminLogService struct {
	logs       []models.ActivityLog
	total      int64
	listErr    error
	deleted    int64
	cleanupErr error
}

func (s *fakeAdminLogService) List(_ repositories.ActivityLogFilter, _, _ int) ([]models.ActivityLog, int64, error) {
	return s.logs, s.total, s.listErr
}

func (s *fakeAdminLogService) Cleanup(_, _ interface{}) (int64, error) {
	return s.deleted, s.cleanupErr
}

// Implement the interface properly
type adminLogSvcImpl struct {
	fakeAdminLogService
}

func (s *adminLogSvcImpl) Cleanup(retainDays *int, deleteBefore *time.Time) (int64, error) {
	return s.deleted, s.cleanupErr
}

func newFakeAdminLogSvc(logs []models.ActivityLog, total int64, listErr error, deleted int64, cleanupErr error) AdminLogService {
	return &adminLogSvcImpl{
		fakeAdminLogService: fakeAdminLogService{
			logs:       logs,
			total:      total,
			listErr:    listErr,
			deleted:    deleted,
			cleanupErr: cleanupErr,
		},
	}
}

func TestAdminLogHandler_ListLogs_Success(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc([]models.ActivityLog{{Action: "auth.login"}}, 1, nil, 0, nil)
	handler := NewAdminLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.ListLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminLogHandler_ListLogs_ServiceError(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, errors.New("db error"), 0, nil)
	handler := NewAdminLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListLogs(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 http error, got %v", err)
	}
}

func TestAdminLogHandler_ListLogs_WithFilters(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 0, nil)
	handler := NewAdminLogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/logs?action=auth.login&entity_type=user&result=success", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.ListLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_NonSuperAdmin(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 5, nil)
	handler := NewAdminLogHandler(svc)

	form := "retain_days=30"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No session user set → redirect to error
	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_SuperAdmin_RetainDays(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 5, nil)
	handler := NewAdminLogHandler(svc)

	form := "retain_days=30"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set super admin session
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_InvalidRetainDays(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 0, nil)
	handler := NewAdminLogHandler(svc)

	form := "retain_days=abc"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_InvalidDeleteBefore(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 0, nil)
	handler := NewAdminLogHandler(svc)

	form := "delete_before=not-a-date"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_ServiceModeError(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 0, services.ErrActivityLogCleanupModeInvalid)
	handler := NewAdminLogHandler(svc)

	form := "retain_days=30"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_ServiceRetainDaysError(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 0, services.ErrActivityLogRetainDaysInvalid)
	handler := NewAdminLogHandler(svc)

	form := "retain_days=30"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_InternalError(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 0, errors.New("internal error"))
	handler := NewAdminLogHandler(svc)

	form := "retain_days=30"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogHandler_CleanupLogs_ValidDeleteBefore(t *testing.T) {
	e := newAdminEcho()
	svc := newFakeAdminLogSvc(nil, 0, nil, 3, nil)
	handler := NewAdminLogHandler(svc)

	form := "delete_before=2024-01-01"
	req := httptest.NewRequest(http.MethodPost, "/admin/logs/cleanup", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "admin-1", Role: string(models.UserRoleSuperAdmin)})

	if err := handler.CleanupLogs(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestAdminLogsFlashURL(t *testing.T) {
	url := adminLogsFlashURL("success", "Done")
	if !strings.Contains(url, "/admin/logs") {
		t.Fatalf("expected /admin/logs in url, got %q", url)
	}
	if !strings.Contains(url, "flash=success") {
		t.Fatalf("expected flash=success in url, got %q", url)
	}
}
