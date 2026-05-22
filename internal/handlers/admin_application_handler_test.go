package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

type fakeAdminAppSvc struct {
	apps  []models.Application
	total int64
	app   *models.Application
}

func (s *fakeAdminAppSvc) ListApplications(page, limit int) ([]models.Application, int64, error) {
	return s.apps, s.total, nil
}
func (s *fakeAdminAppSvc) GetApplication(id string) (*models.Application, error) { return s.app, nil }
func (s *fakeAdminAppSvc) AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	return nil
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
