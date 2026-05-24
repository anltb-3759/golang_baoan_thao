package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
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

type fakeStaffProfileSvc struct {
	profiles  []models.StaffProfile
	listErr   error
	assignErr error
	removeErr error
}

func (s *fakeStaffProfileSvc) ListStaffByDepartment(_ string, _, _ int) ([]models.StaffProfile, int64, error) {
	return s.profiles, int64(len(s.profiles)), s.listErr
}

func (s *fakeStaffProfileSvc) AssignStaffToDepartment(_ string, _ string, _ string) error {
	return s.assignErr
}

func (s *fakeStaffProfileSvc) RemoveStaffFromDepartment(_ string, _ string) error {
	return s.removeErr
}

var _ StaffProfileService = (*fakeStaffProfileSvc)(nil)

type fakeDeptImportSvc struct {
	errs []string
}

func (s *fakeDeptImportSvc) ImportDepartments(_ []services.DepartmentImportRow, _ string) []string {
	return s.errs
}

func newDeptHandler(deptSvc *fakeDeptSvc, userSvc *fakeAdminUserSvc) *AdminDepartmentHandler {
	return NewAdminDepartmentHandler(deptSvc, userSvc, nil)
}

func newDeptHandlerWithImport(deptSvc *fakeDeptSvc, importSvc *fakeDeptImportSvc) *AdminDepartmentHandler {
	return NewAdminDepartmentHandler(deptSvc, &fakeAdminUserSvc{}, nil).WithImportExport(importSvc)
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

func TestAdminDeptHandler_ListDepartmentStaff_RejectsPlaceholderID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments/:id/staff", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: ":id"}})

	err := h.ListDepartmentStaff(c)
	assert.Error(t, err)
	if httpErr, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusNotFound, httpErr.Code)
	}
}

// --- ExportCSV ---

func TestAdminDeptHandler_ExportCSV_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	depts := []models.Department{{ID: "d1", Name: "IT", Code: "IT001"}}
	h := newDeptHandlerWithImport(&fakeDeptSvc{depts: depts, total: 1}, &fakeDeptImportSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/export", "", "")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, "text/csv; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), "IT001")
}

func TestAdminDeptHandler_ExportCSV_Empty(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandlerWithImport(&fakeDeptSvc{}, &fakeDeptImportSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/export", "", "")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "ten")
}

// --- ImportCSV ---

func TestAdminDeptHandler_ImportCSV_NoFile(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandlerWithImport(&fakeDeptSvc{}, &fakeDeptImportSvc{})

	req := httptest.NewRequest(http.MethodPost, "/admin/departments/import", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ImportCSV(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_ImportCSV_WithErrors(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	importSvc := &fakeDeptImportSvc{errs: []string{"Dòng 2: Tên không được để trống"}}
	h := newDeptHandlerWithImport(&fakeDeptSvc{}, importSvc)

	body, ct := newMultipartCSV("ten,mo_ta,ma_code\n,desc,CODE\n")
	req := httptest.NewRequest(http.MethodPost, "/admin/departments/import", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ImportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminDeptHandler_ImportCSV_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandlerWithImport(&fakeDeptSvc{}, &fakeDeptImportSvc{})

	body, ct := newMultipartCSV("ten,mo_ta,ma_code\nPhòng IT,Phòng IT,IT001\n")
	req := httptest.NewRequest(http.MethodPost, "/admin/departments/import", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ImportCSV(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- ExportCSV error break ---

func TestAdminDeptHandler_ExportCSV_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandlerWithImport(&fakeDeptSvc{listErr: errors.New("db error")}, &fakeDeptImportSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/export", "", "")
	err := h.ExportCSV(c)
	assert.NoError(t, err)
	assert.Contains(t, rec.Body.String(), "ten")
}

// --- derefStr ---

func TestDerefStr_Nil(t *testing.T) {
	assert.Equal(t, "", derefStr(nil))
}

func TestDerefStr_Value(t *testing.T) {
	s := "hello"
	assert.Equal(t, "hello", derefStr(&s))
}

// --- departmentIDParam ---

func TestDepartmentIDParam_Empty(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/departments//staff", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	c.SetPathValues(echo.PathValues{{Name: "id", Value: ""}})

	_, err := departmentIDParam(c)
	assert.Error(t, err)
}

// --- ListDepartmentStaff ---

func TestAdminDeptHandler_ListDepartmentStaff_NoProfileSvc(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/staff", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ListDepartmentStaff(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_ListDepartmentStaff_DeptError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{getErr: errors.New("not found")}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/staff", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ListDepartmentStaff(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_ListDepartmentStaff_ProfileError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	dept := &models.Department{ID: "d1"}
	profileSvc := &fakeStaffProfileSvc{listErr: errors.New("profile error")}
	h := NewAdminDepartmentHandler(&fakeDeptSvc{dept: dept}, &fakeAdminUserSvc{}, profileSvc)

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/staff", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ListDepartmentStaff(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_ListDepartmentStaff_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	dept := &models.Department{ID: "d1", Name: "IT"}
	profiles := []models.StaffProfile{{UserID: "u1", User: models.User{ID: "u1", Name: "Staff A"}}}
	profileSvc := &fakeStaffProfileSvc{profiles: profiles}
	h := NewAdminDepartmentHandler(&fakeDeptSvc{dept: dept}, &fakeAdminUserSvc{}, profileSvc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/staff", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ListDepartmentStaff(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminDeptHandler_ListDepartmentStaff_WithLeader(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	leaderID := "u-leader"
	dept := &models.Department{ID: "d1", Name: "IT", LeaderUserID: &leaderID}
	profiles := []models.StaffProfile{
		{UserID: "u-leader", User: models.User{ID: "u-leader", Name: "Leader"}},
		{UserID: "u2", User: models.User{ID: "u2", Name: "Staff"}},
	}
	profileSvc := &fakeStaffProfileSvc{profiles: profiles}
	h := NewAdminDepartmentHandler(&fakeDeptSvc{dept: dept}, &fakeAdminUserSvc{}, profileSvc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/staff", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ListDepartmentStaff(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- ShowAssignStaffForm ---

func TestAdminDeptHandler_ShowAssignStaffForm_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/staff/assign", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ShowAssignStaffForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminDeptHandler_ShowAssignStaffForm_UserSvcError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{listErr: errors.New("user err")}, &fakeStaffProfileSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments/d1/staff/assign", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.ShowAssignStaffForm(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_ShowAssignStaffForm_BadID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/departments/:id/staff/assign", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: ":id"}})
	err := h.ShowAssignStaffForm(c)
	assert.Error(t, err)
}

// --- AssignStaffToDept ---

func TestAdminDeptHandler_AssignStaffToDept_NoProfileSvc(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{})

	form := url.Values{"user_id": {"u1"}}
	c, _ := newFormCtx(e, http.MethodPost, "/admin/departments/d1/staff/assign", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.AssignStaffToDept(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_AssignStaffToDept_BadID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	form := url.Values{"user_id": {"u1"}}
	c, _ := newFormCtx(e, http.MethodPost, "/admin/departments/:id/staff/assign", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: ":id"}})
	err := h.AssignStaffToDept(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_AssignStaffToDept_EmptyUserID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	form := url.Values{"user_id": {""}}
	c, _ := newFormCtx(e, http.MethodPost, "/admin/departments/d1/staff/assign", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.AssignStaffToDept(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_AssignStaffToDept_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	form := url.Values{"user_id": {"u1"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/departments/d1/staff/assign", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.AssignStaffToDept(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminDeptHandler_AssignStaffToDept_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	profileSvc := &fakeStaffProfileSvc{assignErr: errors.New("assign error")}
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, profileSvc)

	form := url.Values{"user_id": {"u1"}}
	c, _ := newFormCtx(e, http.MethodPost, "/admin/departments/d1/staff/assign", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}})
	err := h.AssignStaffToDept(c)
	assert.Error(t, err)
}

// --- RemoveStaffFromDept ---

func TestAdminDeptHandler_RemoveStaffFromDept_NoProfileSvc(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{})

	c, _ := newAdminCtx(e, http.MethodPost, "/admin/departments/d1/staff/u1/remove", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}, {Name: "user_id", Value: "u1"}})
	err := h.RemoveStaffFromDept(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_RemoveStaffFromDept_BadID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	c, _ := newAdminCtx(e, http.MethodPost, "/admin/departments/:id/staff/u1/remove", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: ":id"}, {Name: "user_id", Value: "u1"}})
	err := h.RemoveStaffFromDept(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_RemoveStaffFromDept_EmptyUserID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	c, _ := newAdminCtx(e, http.MethodPost, "/admin/departments/d1/staff//remove", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}, {Name: "user_id", Value: ""}})
	err := h.RemoveStaffFromDept(c)
	assert.Error(t, err)
}

func TestAdminDeptHandler_RemoveStaffFromDept_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, &fakeStaffProfileSvc{})

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/departments/d1/staff/u1/remove", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}, {Name: "user_id", Value: "u1"}})
	err := h.RemoveStaffFromDept(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminDeptHandler_RemoveStaffFromDept_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	profileSvc := &fakeStaffProfileSvc{removeErr: errors.New("remove error")}
	h := NewAdminDepartmentHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{}, profileSvc)

	c, _ := newAdminCtx(e, http.MethodPost, "/admin/departments/d1/staff/u1/remove", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "d1"}, {Name: "user_id", Value: "u1"}})
	err := h.RemoveStaffFromDept(c)
	assert.Error(t, err)
}

// --- listStaffUsers role filter coverage ---

func TestAdminDeptHandler_ShowCreateForm_WithStaffUsers(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	staffUsers := []models.User{
		{ID: "u1", Role: models.UserRoleStaff},
		{ID: "u2", Role: models.UserRoleManager},
		{ID: "u3", Role: models.UserRoleSuperAdmin},
		{ID: "u4", Role: models.UserRoleCitizen},
	}
	h := newDeptHandler(&fakeDeptSvc{}, &fakeAdminUserSvc{users: staffUsers, total: 4})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/departments/new", "", "")
	err := h.ShowCreateForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
