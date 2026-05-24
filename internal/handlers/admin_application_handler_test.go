package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

type fakeAdminAppSvc struct {
	apps      []models.Application
	total     int64
	app       *models.Application
	listErr   error
	getErr    error
	assignErr error
}

func (s *fakeAdminAppSvc) ListApplications(page, limit int) ([]models.Application, int64, error) {
	return s.apps, s.total, s.listErr
}
func (s *fakeAdminAppSvc) GetApplication(id string) (*models.Application, error) {
	return s.app, s.getErr
}
func (s *fakeAdminAppSvc) AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	return s.assignErr
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
		ApplicationCode:  "APP-2024-002",
		CitizenUser:      models.User{Name: "Cit"},
		ServiceType:      models.ServiceType{Name: "Svc", ResponsibleDepartment: dept},
		CompletedAt:      &now,
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

func TestAdminApplicationHandler_ShowApplication_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newAdminAppHandler(&fakeAdminAppSvc{getErr: errors.New("not found")}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/applications/bad", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.ShowApplication(c)
	assert.Error(t, err)
}

// --- assignableStaffUsers edge cases ---

func TestAdminApplicationHandler_assignableStaffUsers_NilApp(t *testing.T) {
	h := newAdminAppHandler(&fakeAdminAppSvc{}, &fakeAdminUserSvc{}, &fakeAssignableStaffSvc{})
	users, err := h.assignableStaffUsers(nil)
	assert.NoError(t, err)
	assert.Nil(t, users)
}

func TestAdminApplicationHandler_assignableStaffUsers_ResponsibleStaffUser(t *testing.T) {
	staffID := "u1"
	app := &models.Application{ServiceType: models.ServiceType{ResponsibleStaffUserID: &staffID}}
	userSvc := &fakeAdminUserSvc{user: &models.User{ID: "u1", Name: "Staff"}}
	h := newAdminAppHandler(&fakeAdminAppSvc{}, userSvc, &fakeAssignableStaffSvc{})

	users, err := h.assignableStaffUsers(app)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "u1", users[0].ID)
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
