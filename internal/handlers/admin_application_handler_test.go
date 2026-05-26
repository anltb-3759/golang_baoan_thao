package handlers

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	templates "github.com/awesome-academy/golang_baoan_thao/internal/templates"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

type captureRenderer struct {
	name string
	data map[string]interface{}
}

func (r *captureRenderer) Render(_ *echo.Context, w io.Writer, name string, data any) error {
	r.name = name
	if m, ok := data.(map[string]interface{}); ok {
		r.data = m
	}
	_, _ = w.Write([]byte("ok"))
	return nil
}

func claimsWithRole(role models.UserRole) *configs.JwtCustomClaims {
	return &configs.JwtCustomClaims{ID: "u-1", Email: "u@test.com", Role: string(role)}
}

type fakeAdminAppSvc struct {
	apps       []models.Application
	total      int64
	app        *models.Application
	listErr    error
	getErr     error
	assignErr  error
	processErr error
	lastFilter repositories.ApplicationFilter
	lastStatus models.ApplicationStatus
	lastNote   string
	lastFiles  []*multipart.FileHeader
}

func (s *fakeAdminAppSvc) ListApplications(filter repositories.ApplicationFilter, page, limit int) ([]models.Application, int64, error) {
	s.lastFilter = filter
	return s.apps, s.total, s.listErr
}
func (s *fakeAdminAppSvc) GetApplication(id string) (*models.Application, error) {
	return s.app, s.getErr
}
func (s *fakeAdminAppSvc) AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	return s.assignErr
}
func (s *fakeAdminAppSvc) ProcessApplication(_ string, status models.ApplicationStatus, note string, files []*multipart.FileHeader, _ string) error {
	s.lastStatus = status
	s.lastNote = note
	s.lastFiles = files
	return s.processErr
}

type fakeAssignableStaffSvc struct {
	profiles []models.StaffProfile
	err      error
}

func (s *fakeAssignableStaffSvc) ListStaffByDepartment(_ string, _, _ int) ([]models.StaffProfile, int64, error) {
	return s.profiles, int64(len(s.profiles)), s.err
}

func newAdminAppHandler(appSvc *fakeAdminAppSvc, userSvc *fakeAdminUserSvc, staffSvc *fakeAssignableStaffSvc) *AdminApplicationHandler {
	return NewAdminApplicationHandler(appSvc, userSvc, staffSvc)
}

func TestAdminApplicationHandler_List_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminAppSvc{apps: []models.Application{{ID: "a1", ApplicationCode: "C1"}}, total: 1}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications", "", "")
	err := h.ListApplications(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminApplicationHandler_List_WithFilters(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminAppSvc{apps: []models.Application{{ID: "a1", ApplicationCode: "C1"}}, total: 1}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	req := httptest.NewRequest(http.MethodGet, "/admin/applications?status=processing&service=cccd&submitter=an", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListApplications(c)
	assert.NoError(t, err)
	assert.Equal(t, repositories.ApplicationFilter{Status: "processing", Service: "cccd", Submitter: "an"}, svc.lastFilter)
}

func TestAdminApplicationHandler_ShowAssignForm_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	deptID := "dept-1"
	app := &models.Application{ID: "app1", ApplicationCode: "C1", ServiceType: models.ServiceType{Name: "S1", ResponsibleDepartmentID: &deptID}}
	svc := &fakeAdminAppSvc{app: app}
	staffSvc := &fakeAssignableStaffSvc{profiles: []models.StaffProfile{{User: models.User{ID: "u1", Name: "Staff A", Email: "a@test.com", Role: models.UserRoleStaff}}, {User: models.User{ID: "u2", Name: "Manager B", Email: "b@test.com", Role: models.UserRoleManager}}}}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, staffSvc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications/app1/assign", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "app1"}})
	err := h.ShowAssignForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminApplicationHandler_assignableStaffUsers_FiltersDepartmentStaff(t *testing.T) {
	deptID := "dept-1"
	app := &models.Application{ServiceType: models.ServiceType{ResponsibleDepartmentID: &deptID}}
	staffSvc := &fakeAssignableStaffSvc{profiles: []models.StaffProfile{{User: models.User{ID: "u1", Name: "Staff A", Role: models.UserRoleStaff}}, {User: models.User{ID: "u2", Name: "Manager B", Role: models.UserRoleManager}}}}
	h := newAdminAppHandler(&fakeAdminAppSvc{}, &fakeAdminUserSvc{}, staffSvc)

	users, err := h.assignableStaffUsers(app)
	assert.NoError(t, err)
	if assert.Len(t, users, 1) {
		assert.Equal(t, "u1", users[0].ID)
	}
}

func TestAdminApplicationHandler_AssignToStaff_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminAppSvc{}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	req := httptest.NewRequest(http.MethodPost, "/admin/applications/app1/assign", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "app1"}})
	c.Set("user", superAdminClaims())

	err := h.AssignToStaff(c)
	assert.NoError(t, err)
}

// --- ExportCSV ---

func TestAdminApplicationHandler_ExportCSV_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	apps := []models.Application{{
		ApplicationCode: "APP-2024-001",
		CitizenUser:     models.User{Name: "Test Citizen"},
		ServiceType:     models.ServiceType{Name: "Cấp CCCD"},
	}}
	h := newAdminAppHandler(&fakeAdminAppSvc{apps: apps, total: 1}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications/export", "", "")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, "text/csv; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "APP-2024-001")
}

func TestAdminApplicationHandler_ExportCSV_Empty(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newAdminAppHandler(&fakeAdminAppSvc{}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications/export", "", "")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "ma_ho_so")
}

func TestAdminApplicationHandler_ExportCSV_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newAdminAppHandler(&fakeAdminAppSvc{listErr: errors.New("db error")}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications/export", "", "")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "ma_ho_so")
}

func TestAdminApplicationHandler_ExportCSV_OptionalFields(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	now := time.Now()
	dept := &models.Department{Name: "Phòng IT"}
	staff := &models.User{Name: "Nguyễn Staff"}
	apps := []models.Application{{
		ApplicationCode:   "APP-2024-002",
		CitizenUser:       models.User{Name: "Cit"},
		ServiceType:       models.ServiceType{Name: "Svc", ResponsibleDepartment: dept},
		CompletedAt:       &now,
		AssignedStaffUser: staff,
	}}
	h := newAdminAppHandler(&fakeAdminAppSvc{apps: apps, total: 1}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications/export", "", "")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "Phòng IT")
	assert.Contains(t, rec.Body.String(), "Nguyễn Staff")
}

// --- ShowApplication ---

func TestAdminApplicationHandler_ShowApplication_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	app := &models.Application{ID: "a1", ApplicationCode: "C1"}
	h := newAdminAppHandler(&fakeAdminAppSvc{app: app}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications/a1", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "a1"}})
	err := h.ShowApplication(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminApplicationHandler_ShowApplication_RenderPermissionsByRole(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	renderer := &captureRenderer{}
	e.Renderer = renderer

	app := &models.Application{
		ID:             "a1",
		ApplicationCode: "C1",
		Status:         models.ApplicationStatusReceived,
	}
	h := newAdminAppHandler(&fakeAdminAppSvc{app: app}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	tests := []struct {
		name       string
		role       models.UserRole
		canAssign  bool
		canProcess bool
	}{
		{name: "manager can assign only", role: models.UserRoleManager, canAssign: true, canProcess: false},
		{name: "staff can process only", role: models.UserRoleStaff, canAssign: false, canProcess: true},
		{name: "super admin cannot assign or process", role: models.UserRoleSuperAdmin, canAssign: false, canProcess: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/applications/a1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", claimsWithRole(tc.role))
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "a1"}})

			err := h.ShowApplication(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "admin/pages/applications/detail.html", renderer.name)
			assert.Equal(t, tc.canAssign, renderer.data["CanAssign"])
			assert.Equal(t, tc.canProcess, renderer.data["CanProcess"])
		})
	}
}

func TestAdminApplicationHandler_ShowApplication_ProcessOptionsIncludeNeedMoreInfo(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	renderer := &captureRenderer{}
	e.Renderer = renderer

	tests := []struct {
		name        string
		current     models.ApplicationStatus
		expectValue []string
	}{
		{
			name:        "received options",
			current:     models.ApplicationStatusReceived,
			expectValue: []string{string(models.ApplicationStatusProcessing), string(models.ApplicationStatusNeedMoreInfo)},
		},
		{
			name:        "processing options",
			current:     models.ApplicationStatusProcessing,
			expectValue: []string{string(models.ApplicationStatusNeedMoreInfo), string(models.ApplicationStatusApproved), string(models.ApplicationStatusRejected)},
		},
		{
			name:        "need more info options",
			current:     models.ApplicationStatusNeedMoreInfo,
			expectValue: []string{string(models.ApplicationStatusProcessing), string(models.ApplicationStatusApproved), string(models.ApplicationStatusRejected)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := &models.Application{ID: "a1", ApplicationCode: "C1", Status: tc.current}
			h := newAdminAppHandler(&fakeAdminAppSvc{app: app}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

			req := httptest.NewRequest(http.MethodGet, "/admin/applications/a1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user", claimsWithRole(models.UserRoleStaff))
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "a1"}})

			err := h.ShowApplication(c)
			assert.NoError(t, err)

			rawOptions, ok := renderer.data["ProcessOptions"].([]applicationStatusOption)
			if assert.True(t, ok) {
				got := make([]string, 0, len(rawOptions))
				for _, opt := range rawOptions {
					got = append(got, opt.Value)
				}
				assert.Equal(t, tc.expectValue, got)
			}
		})
	}
}

func TestAdminApplicationHandler_ProcessApplication_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminAppSvc{app: &models.Application{ID: "a1", Status: models.ApplicationStatusReceived}}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	_ = w.WriteField("status", string(models.ApplicationStatusProcessing))
	_ = w.WriteField("note", "đang xử lý")
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/admin/applications/a1/process", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "a1"}})
	c.Set("user", superAdminClaims())

	err := h.ProcessApplication(c)
	assert.NoError(t, err)
	assert.Equal(t, string(models.ApplicationStatusProcessing), string(svc.lastStatus))
	assert.Equal(t, "đang xử lý", svc.lastNote)
}

func TestAdminApplicationHandler_ShowApplication_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newAdminAppHandler(&fakeAdminAppSvc{getErr: errors.New("not found")}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/applications/bad", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.ShowApplication(c)
	assert.Error(t, err)
}

func TestMapAdminApplicationProcessError_RejectReasonRequired(t *testing.T) {
	err := mapAdminApplicationProcessError(services.ErrAdminApplicationRejectReasonRequired)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, httpErr.Code)
	assert.Equal(t, "application.reject_reason_required", httpErr.Message)
}

func TestMapAdminApplicationProcessError_NeedMoreInfoNoteRequired(t *testing.T) {
	err := mapAdminApplicationProcessError(services.ErrAdminApplicationNeedMoreInfoNoteRequired)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, httpErr.Code)
	assert.Equal(t, "application.need_more_info_note_required", httpErr.Message)
}

// --- assignableStaffUsers edge cases ---

func TestAdminApplicationHandler_assignableStaffUsers_NilApp(t *testing.T) {
	h := newAdminAppHandler(&fakeAdminAppSvc{}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})
	users, err := h.assignableStaffUsers(nil)
	assert.NoError(t, err)
	assert.Nil(t, users)
}

func TestAdminApplicationHandler_assignableStaffUsers_WithAssignedStaff(t *testing.T) {
	assignedUser := &models.User{ID: "u-assigned", Name: "Assigned"}
	app := &models.Application{AssignedStaffUser: assignedUser}
	h := newAdminAppHandler(&fakeAdminAppSvc{}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	users, err := h.assignableStaffUsers(app)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "u-assigned", users[0].ID)
}

// --- AssignToStaff error path ---

func TestAdminApplicationHandler_AssignToStaff_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminAppSvc{assignErr: errors.New("assign failed")}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	req := httptest.NewRequest(http.MethodPost, "/admin/applications/app1/assign", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "app1"}})
	c.Set("user", superAdminClaims())

	err := h.AssignToStaff(c)
	assert.Error(t, err)
}

// --- ListApplications error path ---

func TestAdminApplicationHandler_List_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminAppSvc{listErr: errors.New("db error")}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/applications", "", "")
	err := h.ListApplications(c)
	assert.Error(t, err)
}

func TestAdminApplicationsList_RendersTable(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newTestEcho()
	// use real renderer to render templates and assert HTML output
	r, err := templates.NewRenderer("../../templates")
	if err != nil {
		t.Fatalf("failed to init templates: %v", err)
	}
	e.Renderer = r

	svc := &fakeAdminAppSvc{apps: []models.Application{{ID: "a1", ApplicationCode: "C1"}}, total: 1}
	h := newAdminAppHandler(svc, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/applications", "", "")
	err = h.ListApplications(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.Contains(t, body, "<table")
	// header text from template default
	assert.Contains(t, body, "Mã hồ sơ")
}
