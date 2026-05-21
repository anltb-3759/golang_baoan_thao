package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// --- fake department service ---

type fakeDeptSvc struct {
	depts     []models.Department
	total     int64
	dept      *models.Department
	listErr   error
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (s *fakeDeptSvc) ListDepartments(_ repositories.DepartmentFilter, _, _ int) ([]models.Department, int64, error) {
	return s.depts, s.total, s.listErr
}
func (s *fakeDeptSvc) GetDepartment(_ string) (*models.Department, error) {
	return s.dept, s.getErr
}
func (s *fakeDeptSvc) CreateDepartment(_ *dtos.DepartmentCreateRequest, _ string) (*models.Department, error) {
	return s.dept, s.createErr
}
func (s *fakeDeptSvc) UpdateDepartment(_ string, _ *dtos.DepartmentUpdateRequest, _ string) (*models.Department, error) {
	return s.dept, s.updateErr
}
func (s *fakeDeptSvc) DeleteDepartment(_ string, _ string) error { return s.deleteErr }

var _ DepartmentService = (*fakeDeptSvc)(nil)

func newDeptHandler(deptSvc *fakeDeptSvc, userSvc *fakeAdminUserSvc) *AdminDepartmentHandler {
	return NewAdminDepartmentHandler(deptSvc, userSvc)
}

// --- ListDepartments ---

func TestAdminDeptHandler_ListDepartments_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeDeptSvc{depts: []models.Department{{ID: "1", Name: "IT"}}, total: 1}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments", "", "")
	err := h.ListDepartments(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminDeptHandler_ListDepartments_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeDeptSvc{listErr: errors.New("db error")}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments", "", "")
	err := h.ListDepartments(c)
	assert.Error(t, err)
}

// --- ShowCreateForm ---

func TestAdminDeptHandler_ShowCreateForm_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/new", "", "")
	err := h.ShowCreateForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminDeptHandler_ShowCreateForm_UserSvcError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{listErr: errors.New("user error")})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments/new", "", "")
	err := h.ShowCreateForm(c)
	assert.Error(t, err)
}

// --- CreateDepartment ---

func TestAdminDeptHandler_CreateDepartment_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	created := &models.Department{ID: "d1", Name: "IT"}
	svc := &fakeDeptSvc{dept: created}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	form := url.Values{"name": {"IT Dept"}, "code": {"IT001"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments", form)
	err := h.CreateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/admin/departments")
}

func TestAdminDeptHandler_CreateDepartment_ValidationError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{})

	form := url.Values{"name": {""}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments", form)
	err := h.CreateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminDeptHandler_CreateDepartment_CodeExists(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeDeptSvc{createErr: services.ErrDepartmentCodeExists}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	form := url.Values{"name": {"IT"}, "code": {"IT001"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments", form)
	err := h.CreateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminDeptHandler_CreateDepartment_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeDeptSvc{createErr: errors.New("internal")}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	form := url.Values{"name": {"IT"}, "code": {"IT001"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments", form)
	err := h.CreateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// --- ShowEditForm ---

func TestAdminDeptHandler_ShowEditForm_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	d := &models.Department{ID: "d1", Name: "IT"}
	svc := &fakeDeptSvc{dept: d}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/edit", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ShowEditForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminDeptHandler_ShowEditForm_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeDeptSvc{getErr: errors.New("not found")}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/bad/edit", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.ShowEditForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- UpdateDepartment ---

func TestAdminDeptHandler_UpdateDepartment_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	d := &models.Department{ID: "d1", Name: "IT"}
	svc := &fakeDeptSvc{dept: d}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	form := url.Values{"name": {"Updated"}, "code": {"IT001"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments/d1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.UpdateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminDeptHandler_UpdateDepartment_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeDeptSvc{getErr: errors.New("not found")}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	form := url.Values{"name": {"X"}, "code": {"X"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments/bad/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.UpdateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminDeptHandler_UpdateDepartment_ValidationError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	d := &models.Department{ID: "d1"}
	svc := &fakeDeptSvc{dept: d}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	form := url.Values{"name": {""}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments/d1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.UpdateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminDeptHandler_UpdateDepartment_CodeExists(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	d := &models.Department{ID: "d1"}
	svc := &fakeDeptSvc{dept: d, updateErr: services.ErrDepartmentCodeExists}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	form := url.Values{"name": {"IT"}, "code": {"IT001"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments/d1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.UpdateDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// --- DeleteDepartment ---

func TestAdminDeptHandler_DeleteDepartment_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/departments/d1/delete", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.DeleteDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=success")
}

func TestAdminDeptHandler_DeleteDepartment_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeDeptSvc{deleteErr: errors.New("delete failed")}
	h := newDeptHandler(svc, &fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/departments/d1/delete", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.DeleteDepartment(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=error")
}
