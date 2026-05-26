package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/labstack/echo/v5"
)

type fakeNotificationSvc struct {
	listItems   []dtos.NotificationResponse
	listTotal   int64
	listErr     error
	markReadErr error
	markAllErr  error
}

func (s *fakeNotificationSvc) List(_ string, _ repositories.NotificationFilter, _, _ int) ([]dtos.NotificationResponse, int64, error) {
	return s.listItems, s.listTotal, s.listErr
}

func (s *fakeNotificationSvc) MarkAsRead(_, _ string) error {
	return s.markReadErr
}

func (s *fakeNotificationSvc) MarkAllAsRead(_ string) error {
	return s.markAllErr
}

func TestNotificationHandler_List_Success(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{
		listItems: []dtos.NotificationResponse{{ID: "n1", Title: "Hello"}},
		listTotal: 1,
	}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/notifications", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})

	if err := handler.List(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNotificationHandler_List_WithIsReadTrue(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{listItems: []dtos.NotificationResponse{}, listTotal: 0}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/notifications?is_read=true", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})

	if err := handler.List(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNotificationHandler_List_WithIsReadFalse(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{listItems: []dtos.NotificationResponse{}, listTotal: 0}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/notifications?is_read=false", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})

	if err := handler.List(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNotificationHandler_List_ServiceError(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{listErr: errors.New("db error")}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/citizens/me/notifications", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})

	err := handler.List(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNotificationHandler_MarkAsRead_Success(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/citizens/me/notifications/n1/read", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "n1"}})

	if err := handler.MarkAsRead(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNotificationHandler_MarkAsRead_Error(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{markReadErr: errors.New("mark error")}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/citizens/me/notifications/n1/read", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "n1"}})

	err := handler.MarkAsRead(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNotificationHandler_MarkAllAsRead_Success(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/citizens/me/notifications/read-all", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})

	if err := handler.MarkAllAsRead(c); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNotificationHandler_MarkAllAsRead_Error(t *testing.T) {
	e := newTestEcho()
	svc := &fakeNotificationSvc{markAllErr: errors.New("mark all error")}
	handler := NewNotificationHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/citizens/me/notifications/read-all", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &configs.JwtCustomClaims{ID: "user-1"})

	err := handler.MarkAllAsRead(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
