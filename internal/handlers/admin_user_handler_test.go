package handlers

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// --- stub renderer ---

type stubRenderer struct{}

func (r *stubRenderer) Render(_ *echo.Context, w io.Writer, _ string, _ any) error {
	_, _ = w.Write([]byte("ok"))
	return nil
}

func newAdminEcho() *echo.Echo {
	e := newTestEcho()
	e.Renderer = &stubRenderer{}
	return e
}

// --- fake admin user service ---

type fakeAdminUserSvc struct {
	users    []models.User
	total    int64
	user     *models.User
	listErr  error
	getErr   error
	createErr error
	updateErr error
	blockErr  error
	deleteErr error
}

func (s *fakeAdminUserSvc) ListUsers(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return s.users, s.total, s.listErr
}
func (s *fakeAdminUserSvc) GetUser(_ string) (*models.User, error) {
	return s.user, s.getErr
}
func (s *fakeAdminUserSvc) CreateUser(_ *dtos.AdminCreateUserRequest, _ string) (*models.User, error) {
	return s.user, s.createErr
}
func (s *fakeAdminUserSvc) UpdateUser(_ string, _ *dtos.AdminUpdateUserRequest, _ string) (*models.User, error) {
	return s.user, s.updateErr
}
func (s *fakeAdminUserSvc) BlockUser(_ string, _ string) error   { return s.blockErr }
func (s *fakeAdminUserSvc) UnblockUser(_ string, _ string) error { return s.blockErr }
func (s *fakeAdminUserSvc) DeleteUser(_ string, _ string) error  { return s.deleteErr }

var _ AdminUserService = (*fakeAdminUserSvc)(nil)

func superAdminClaims() *configs.JwtCustomClaims {
	return &configs.JwtCustomClaims{ID: "admin-1", Email: "admin@test.com", Role: "super_admin"}
}

func newAdminCtx(e *echo.Echo, method, path string, body string, contentType string) (*echo.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())
	return c, rec
}

func newFormCtx(e *echo.Echo, method, path string, values url.Values) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())
	return c, rec
}

// --- ListUsers ---

func TestAdminUserHandler_ListUsers_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{users: []models.User{{ID: "1", Name: "Test"}}, total: 1}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/users", "", "")
	err := h.ListUsers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUserHandler_ListUsers_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{listErr: errors.New("db error")}
	h := NewAdminUserHandler(svc)

	c, _ := newAdminCtx(e, http.MethodGet, "/admin/users", "", "")
	err := h.ListUsers(c)
	assert.Error(t, err)
}

func TestAdminUserHandler_ListUsers_WithFlash(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{}
	h := NewAdminUserHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/admin/users?flash=success&msg=done", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.ListUsers(c)
	assert.NoError(t, err)
}

// --- ShowCreateForm ---

func TestAdminUserHandler_ShowCreateForm(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/users/new", "", "")
	err := h.ShowCreateForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// --- CreateUser ---

func TestAdminUserHandler_CreateUser_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	created := &models.User{ID: "new-id", Name: "New User"}
	svc := &fakeAdminUserSvc{user: created}
	h := NewAdminUserHandler(svc)

	form := url.Values{"name": {"New User"}, "email": {"new@test.com"}, "password": {"pass1234"}, "role": {"citizen"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/users", form)
	err := h.CreateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/admin/users")
}

func TestAdminUserHandler_CreateUser_BindError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	// Send malformed form
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())

	err := h.CreateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminUserHandler_CreateUser_ValidationError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	// Missing required fields
	form := url.Values{"name": {""}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/users", form)
	err := h.CreateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminUserHandler_CreateUser_EmailExists(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{createErr: services.ErrEmailAlreadyExists}
	h := NewAdminUserHandler(svc)

	form := url.Values{"name": {"User"}, "email": {"dup@test.com"}, "password": {"pass1234"}, "role": {"citizen"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/users", form)
	err := h.CreateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminUserHandler_CreateUser_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{createErr: errors.New("internal error")}
	h := NewAdminUserHandler(svc)

	form := url.Values{"name": {"User"}, "email": {"x@test.com"}, "password": {"pass1234"}, "role": {"citizen"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/users", form)
	err := h.CreateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminUserHandler_CreateUser_NoUser(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	form := url.Values{"name": {"User"}, "email": {"x@test.com"}, "password": {"pass1234"}, "role": {"citizen"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No user in context → actorID = ""
	err := h.CreateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- ShowUser ---

func TestAdminUserHandler_ShowUser_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	u := &models.User{ID: "u1", Name: "Test"}
	svc := &fakeAdminUserSvc{user: u}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/users/u1", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.ShowUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUserHandler_ShowUser_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{getErr: errors.New("not found")}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/users/bad", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.ShowUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- ShowEditForm ---

func TestAdminUserHandler_ShowEditForm_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	u := &models.User{ID: "u1", Name: "Test"}
	svc := &fakeAdminUserSvc{user: u}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/users/u1/edit", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.ShowEditForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminUserHandler_ShowEditForm_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{getErr: errors.New("not found")}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/users/bad/edit", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "bad"}})
	err := h.ShowEditForm(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- UpdateUser ---

func TestAdminUserHandler_UpdateUser_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	u := &models.User{ID: "u1", Name: "Updated"}
	svc := &fakeAdminUserSvc{user: u}
	h := NewAdminUserHandler(svc)

	form := url.Values{"name": {"Updated"}, "role": {"staff"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/users/u1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.UpdateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminUserHandler_UpdateUser_BindError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	u := &models.User{ID: "u1", Name: "Existing"}
	h := NewAdminUserHandler(&fakeAdminUserSvc{user: u})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/u1/edit", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", superAdminClaims())
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})

	err := h.UpdateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminUserHandler_UpdateUser_ValidationError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	u := &models.User{ID: "u1", Name: "Existing"}
	h := NewAdminUserHandler(&fakeAdminUserSvc{user: u})

	form := url.Values{"name": {""}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/users/u1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.UpdateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminUserHandler_UpdateUser_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	u := &models.User{ID: "u1", Name: "Existing"}
	svc := &fakeAdminUserSvc{user: u, updateErr: errors.New("update error")}
	h := NewAdminUserHandler(svc)

	form := url.Values{"name": {"Name"}, "role": {"citizen"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/users/u1/edit", form)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.UpdateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestAdminUserHandler_UpdateUser_NoUser(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	u := &models.User{ID: "u1", Name: "Updated"}
	svc := &fakeAdminUserSvc{user: u}
	h := NewAdminUserHandler(svc)

	form := url.Values{"name": {"Updated"}, "role": {"staff"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/users/u1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No user in context
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.UpdateUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- BlockUser ---

func TestAdminUserHandler_BlockUser_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/users/u1/block", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.BlockUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=success")
}

func TestAdminUserHandler_BlockUser_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{blockErr: errors.New("block failed")}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/users/u1/block", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.BlockUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=error")
}

func TestAdminUserHandler_BlockUser_NoUser(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/u1/block", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// No claims in context
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.BlockUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- UnblockUser ---

func TestAdminUserHandler_UnblockUser_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/users/u1/unblock", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.UnblockUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=success")
}

func TestAdminUserHandler_UnblockUser_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{blockErr: errors.New("unblock failed")}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/users/u1/unblock", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.UnblockUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=error")
}

func TestAdminUserHandler_UnblockUser_NoUser(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/u1/unblock", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.UnblockUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- DeleteUser ---

func TestAdminUserHandler_DeleteUser_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/users/u1/delete", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.DeleteUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=success")
}

func TestAdminUserHandler_DeleteUser_Error(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	svc := &fakeAdminUserSvc{deleteErr: errors.New("delete failed")}
	h := NewAdminUserHandler(svc)

	c, rec := newAdminCtx(e, http.MethodPost, "/admin/users/u1/delete", "", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.DeleteUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "flash=error")
}

func TestAdminUserHandler_DeleteUser_NoUser(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminUserHandler(&fakeAdminUserSvc{})

	req := httptest.NewRequest(http.MethodPost, "/admin/users/u1/delete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "u1"}})
	err := h.DeleteUser(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

// --- flashFromQuery ---

func TestFlashFromQuery_Empty(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	assert.Nil(t, flashFromQuery(c))
}

func TestFlashFromQuery_OnlyFlash(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/users?flash=success", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	assert.Nil(t, flashFromQuery(c))
}

func TestFlashFromQuery_BothParams(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/admin/users?flash=success&msg=done", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	result := flashFromQuery(c)
	assert.NotNil(t, result)
	assert.Equal(t, "success", result["Type"])
	assert.Equal(t, "done", result["Message"])
}

// --- newPagination ---

func TestNewPagination_Normal(t *testing.T) {
	p := newPagination(2, 10, 25)
	assert.Equal(t, 2, p.Page)
	assert.Equal(t, 11, p.From)
	assert.Equal(t, 20, p.To)
	assert.True(t, p.HasPrev)
	assert.True(t, p.HasNext)
	assert.Equal(t, 1, p.PrevPage)
	assert.Equal(t, 3, p.NextPage)
}

func TestNewPagination_LastPage(t *testing.T) {
	p := newPagination(3, 10, 25)
	assert.Equal(t, 25, p.To)
	assert.False(t, p.HasNext)
}

func TestNewPagination_Empty(t *testing.T) {
	p := newPagination(1, 10, 0)
	assert.Equal(t, 0, p.From)
	assert.Equal(t, 0, p.To)
	assert.False(t, p.HasPrev)
	assert.False(t, p.HasNext)
}

// --- adminCurrentUser ---

func TestAdminCurrentUser_NoClaims(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	assert.Nil(t, adminCurrentUser(c))
}

func TestAdminCurrentUser_WithClaims(t *testing.T) {
	e := newTestEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c := e.NewContext(req, httptest.NewRecorder())
	claims := &configs.JwtCustomClaims{ID: "1", Email: "x@y.com"}
	c.Set("user", claims)
	result := adminCurrentUser(c)
	assert.Equal(t, "1", result.ID)
}

// --- AdminDashboardHandler ---

func TestAdminDashboardHandler_ShowDashboard(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminDashboardHandler()

	c, rec := newAdminCtx(e, http.MethodGet, "/admin", "", "")
	err := h.ShowDashboard(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
