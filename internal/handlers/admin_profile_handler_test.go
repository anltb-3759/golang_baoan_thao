package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

type fakeAdminProfileSvc struct {
	user      *models.User
	getErr    error
	updateErr error
	pwdErr    error
}

func (s *fakeAdminProfileSvc) GetSelf(_ string) (*models.User, error) {
	return s.user, s.getErr
}

func (s *fakeAdminProfileSvc) UpdateSelfContact(_ string, _ *dtos.UpdateAdminProfileRequest) (*models.User, error) {
	return s.user, s.updateErr
}

func (s *fakeAdminProfileSvc) ChangeSelfPassword(_ string, _ *dtos.ChangeAdminPasswordRequest) error {
	return s.pwdErr
}

func TestAdminProfileHandler_ShowProfilePage_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{user: &models.User{ID: "u1", Name: "Admin"}})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/profile", "", "")
	err := h.ShowProfilePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAdminProfileHandler_ShowProfilePage_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{getErr: services.ErrUserNotFound})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/profile", "", "")
	err := h.ShowProfilePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	loc := rec.Header().Get(echo.HeaderLocation)
	assert.Contains(t, loc, "/admin?")
	assert.Contains(t, loc, "flash=error")
}

func TestAdminProfileHandler_ShowProfilePage_InternalError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{getErr: errors.New("db fail")})

	c, rec := newAdminCtx(e, http.MethodGet, "/admin/profile", "", "")
	err := h.ShowProfilePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	loc := rec.Header().Get(echo.HeaderLocation)
	decoded, _ := url.QueryUnescape(loc)
	assert.Contains(t, decoded, configs.T(c, "common.internal_error", nil))
}

func TestAdminProfileHandler_UpdateProfile_NoLogger(t *testing.T) {
	// Covers writeActivityLog nil-logger path (early return)
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{user: &models.User{ID: "admin-1"}})
	// no WithActivityLogger → h.logger == nil

	form := url.Values{"phone": {"0909"}, "address": {"HCM"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/profile", form)
	err := h.UpdateProfile(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminProfileHandler_ChangePassword_NoLogger(t *testing.T) {
	// Covers writeActivityLog nil-logger path (early return)
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{})

	form := url.Values{
		"current_password":     {"oldpass123"},
		"new_password":         {"newpass123"},
		"confirm_new_password": {"newpass123"},
	}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/profile/password", form)
	err := h.ChangePassword(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
}

func TestAdminProfileHandler_UpdateProfile_WritesActivityLog(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	logger := &fakeActivityLogger{}
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{user: &models.User{ID: "admin-1"}}).WithActivityLogger(logger)

	form := url.Values{"phone": {"0909"}, "address": {"HCM"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/profile", form)
	err := h.UpdateProfile(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, 1, logger.calls)
	if assert.NotNil(t, logger.lastLog) {
		assert.Equal(t, "admin.profile.update_contact", logger.lastLog.Action)
		assert.Equal(t, "success", logger.lastLog.Result)
	}
}

func TestAdminProfileHandler_UpdateProfile_ServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{updateErr: services.ErrUserNotFound})

	form := url.Values{"phone": {"0909"}, "address": {"HCM"}}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/profile", form)
	err := h.UpdateProfile(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	loc := rec.Header().Get(echo.HeaderLocation)
	assert.Contains(t, loc, "flash=error")
}

func TestAdminProfileHandler_ChangePassword_ServiceErrors(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	cases := []struct {
		name string
		err  error
	}{
		{"user_not_found", services.ErrUserNotFound},
		{"password_mismatch", services.ErrPasswordMismatch},
		{"confirmation_mismatch", services.ErrPasswordConfirmationMismatch},
		{"must_differ", services.ErrNewPasswordMustDiffer},
		{"internal", errors.New("db fail")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newAdminEcho()
			h := NewAdminProfileHandler(&fakeAdminProfileSvc{pwdErr: tc.err})
			form := url.Values{
				"current_password":     {"oldpass123"},
				"new_password":         {"newpass123"},
				"confirm_new_password": {"newpass123"},
			}
			c, rec := newFormCtx(e, http.MethodPost, "/admin/profile/password", form)
			err := h.ChangePassword(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusSeeOther, rec.Code)
			assert.Contains(t, rec.Header().Get(echo.HeaderLocation), "flash=error")
		})
	}
}

func TestAdminProfileHandler_ChangePassword_WritesActivityLog(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	logger := &fakeActivityLogger{}
	h := NewAdminProfileHandler(&fakeAdminProfileSvc{}).WithActivityLogger(logger)

	form := url.Values{
		"current_password":     {"oldpass123"},
		"new_password":         {"newpass123"},
		"confirm_new_password": {"newpass123"},
	}
	c, rec := newFormCtx(e, http.MethodPost, "/admin/profile/password", form)
	err := h.ChangePassword(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, 1, logger.calls)
	if assert.NotNil(t, logger.lastLog) {
		assert.Equal(t, "admin.profile.change_password", logger.lastLog.Action)
		assert.Equal(t, "success", logger.lastLog.Result)
	}
}
