package services_test

import (
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type fakeAdminProfileUserRepo struct {
	user      *models.User
	err       error
	updateErr error
	updated   *models.User
}

func (r *fakeAdminProfileUserRepo) FindByEmail(_ string) (*models.User, error)  { return nil, nil }
func (r *fakeAdminProfileUserRepo) FindByID(_ string) (*models.User, error)     { return r.user, r.err }
func (r *fakeAdminProfileUserRepo) Create(u *models.User) (*models.User, error) { return u, nil }
func (r *fakeAdminProfileUserRepo) CreateInTx(_ *gorm.DB, u *models.User) error { return nil }
func (r *fakeAdminProfileUserRepo) Update(u *models.User) error                 { r.updated = u; return r.updateErr }
func (r *fakeAdminProfileUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeAdminProfileUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeAdminProfileUserRepo) SoftDelete(_ string, _ string) error                        { return nil }

var _ repositories.UserRepository = (*fakeAdminProfileUserRepo)(nil)

func TestAdminProfileService_GetSelf_Success(t *testing.T) {
	user := &models.User{ID: "admin1", Name: "Admin User", Role: models.UserRoleSuperAdmin}
	repo := &fakeAdminProfileUserRepo{user: user}
	svc := services.NewAdminProfileService(repo)

	res, err := svc.GetSelf("admin1")
	assert.NoError(t, err)
	assert.Equal(t, user, res)
}

func TestAdminProfileService_GetSelf_DBError(t *testing.T) {
	dbErr := assert.AnError
	repo := &fakeAdminProfileUserRepo{err: dbErr}
	svc := services.NewAdminProfileService(repo)

	res, err := svc.GetSelf("admin1")
	assert.ErrorIs(t, err, dbErr)
	assert.Nil(t, res)
}

func TestAdminProfileService_GetSelf_UserNotFound(t *testing.T) {
	repo := &fakeAdminProfileUserRepo{user: nil}
	svc := services.NewAdminProfileService(repo)

	res, err := svc.GetSelf("admin1")
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	assert.Nil(t, res)
}

func TestAdminProfileService_UpdateSelfContact_Success(t *testing.T) {
	user := &models.User{ID: "admin1", Name: "Admin User", Phone: "111", Address: "Old"}
	repo := &fakeAdminProfileUserRepo{user: user}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.UpdateAdminProfileRequest{
		Phone:   "222",
		Address: "New",
	}

	res, err := svc.UpdateSelfContact("admin1", req)
	assert.NoError(t, err)
	assert.Equal(t, "222", res.Phone)
	assert.Equal(t, "New", res.Address)
	assert.Equal(t, "222", repo.updated.Phone)
}

func TestAdminProfileService_UpdateSelfContact_UserNotFound(t *testing.T) {
	repo := &fakeAdminProfileUserRepo{user: nil}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.UpdateAdminProfileRequest{Phone: "222", Address: "New"}
	res, err := svc.UpdateSelfContact("admin1", req)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
	assert.Nil(t, res)
}

func TestAdminProfileService_UpdateSelfContact_DBError(t *testing.T) {
	dbErr := assert.AnError
	repo := &fakeAdminProfileUserRepo{err: dbErr}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.UpdateAdminProfileRequest{Phone: "222", Address: "New"}
	res, err := svc.UpdateSelfContact("admin1", req)
	assert.ErrorIs(t, err, dbErr)
	assert.Nil(t, res)
}

func TestAdminProfileService_UpdateSelfContact_UpdateError(t *testing.T) {
	user := &models.User{ID: "admin1", Phone: "111"}
	updateErr := assert.AnError
	repo := &fakeAdminProfileUserRepo{user: user, updateErr: updateErr}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.UpdateAdminProfileRequest{Phone: "222", Address: "New"}
	res, err := svc.UpdateSelfContact("admin1", req)
	assert.ErrorIs(t, err, updateErr)
	assert.Nil(t, res)
}

func TestAdminProfileService_ChangeSelfPassword_UserNotFound(t *testing.T) {
	repo := &fakeAdminProfileUserRepo{user: nil}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.ChangeAdminPasswordRequest{
		CurrentPassword:    "old",
		NewPassword:        "new",
		ConfirmNewPassword: "new",
	}
	err := svc.ChangeSelfPassword("admin1", req)
	assert.ErrorIs(t, err, services.ErrUserNotFound)
}

func TestAdminProfileService_ChangeSelfPassword_DBError(t *testing.T) {
	dbErr := assert.AnError
	repo := &fakeAdminProfileUserRepo{err: dbErr}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.ChangeAdminPasswordRequest{
		CurrentPassword:    "old",
		NewPassword:        "new",
		ConfirmNewPassword: "new",
	}
	err := svc.ChangeSelfPassword("admin1", req)
	assert.ErrorIs(t, err, dbErr)
}

func TestAdminProfileService_ChangeSelfPassword_UpdateError(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("realold"), bcrypt.DefaultCost)
	updateErr := assert.AnError
	user := &models.User{ID: "admin1", PasswordHash: string(hash)}
	repo := &fakeAdminProfileUserRepo{user: user, updateErr: updateErr}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.ChangeAdminPasswordRequest{
		CurrentPassword:    "realold",
		NewPassword:        "newpass",
		ConfirmNewPassword: "newpass",
	}
	err := svc.ChangeSelfPassword("admin1", req)
	assert.ErrorIs(t, err, updateErr)
}

func TestAdminProfileService_ChangeSelfPassword_ConfirmMismatch(t *testing.T) {
	user := &models.User{ID: "admin1", PasswordHash: "oldhash"}
	repo := &fakeAdminProfileUserRepo{user: user}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.ChangeAdminPasswordRequest{
		CurrentPassword:    "oldpass",
		NewPassword:        "newpass",
		ConfirmNewPassword: "different",
	}

	err := svc.ChangeSelfPassword("admin1", req)
	assert.ErrorIs(t, err, services.ErrPasswordConfirmationMismatch)
}

func TestAdminProfileService_ChangeSelfPassword_NewEqualsCurrent(t *testing.T) {
	user := &models.User{ID: "admin1", PasswordHash: "oldhash"}
	repo := &fakeAdminProfileUserRepo{user: user}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.ChangeAdminPasswordRequest{
		CurrentPassword:    "oldpass",
		NewPassword:        "oldpass",
		ConfirmNewPassword: "oldpass",
	}

	err := svc.ChangeSelfPassword("admin1", req)
	assert.ErrorIs(t, err, services.ErrNewPasswordMustDiffer)
}

func TestAdminProfileService_ChangeSelfPassword_CurrentMismatch(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("realold"), bcrypt.DefaultCost)
	user := &models.User{ID: "admin1", PasswordHash: string(hash)}
	repo := &fakeAdminProfileUserRepo{user: user}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.ChangeAdminPasswordRequest{
		CurrentPassword:    "wrongold",
		NewPassword:        "newpass",
		ConfirmNewPassword: "newpass",
	}

	err := svc.ChangeSelfPassword("admin1", req)
	assert.ErrorIs(t, err, services.ErrPasswordMismatch)
}

func TestAdminProfileService_ChangeSelfPassword_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("realold"), bcrypt.DefaultCost)
	user := &models.User{ID: "admin1", PasswordHash: string(hash)}
	repo := &fakeAdminProfileUserRepo{user: user}
	svc := services.NewAdminProfileService(repo)

	req := &dtos.ChangeAdminPasswordRequest{
		CurrentPassword:    "realold",
		NewPassword:        "newpass",
		ConfirmNewPassword: "newpass",
	}

	err := svc.ChangeSelfPassword("admin1", req)
	assert.NoError(t, err)
	assert.NotNil(t, repo.updated)

	err = bcrypt.CompareHashAndPassword([]byte(repo.updated.PasswordHash), []byte("newpass"))
	assert.NoError(t, err)
}
