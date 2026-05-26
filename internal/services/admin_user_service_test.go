package services

import (
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// --- fakeAdminUserRepo ---

type fakeAdminUserRepo struct {
	user        *models.User
	users       []models.User
	total       int64
	findErr     error
	findByEmail *models.User
	createErr   error
	updateErr   error
	statusErr   error
	deleteErr   error
}

func (r *fakeAdminUserRepo) FindByEmail(_ string) (*models.User, error) {
	return r.findByEmail, r.findErr
}
func (r *fakeAdminUserRepo) FindByID(_ string) (*models.User, error) {
	return r.user, r.findErr
}
func (r *fakeAdminUserRepo) Create(u *models.User) (*models.User, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	return u, nil
}
func (r *fakeAdminUserRepo) CreateInTx(_ *gorm.DB, _ *models.User) error { return nil }
func (r *fakeAdminUserRepo) Update(u *models.User) error {
	return r.updateErr
}
func (r *fakeAdminUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	if r.findErr != nil {
		return nil, 0, r.findErr
	}
	return r.users, r.total, nil
}
func (r *fakeAdminUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error {
	return r.statusErr
}
func (r *fakeAdminUserRepo) SoftDelete(_ string, _ string) error {
	return r.deleteErr
}

var _ repositories.UserRepository = (*fakeAdminUserRepo)(nil)

func newAdminSvc(repo *fakeAdminUserRepo) *AdminUserService {
	return NewAdminUserService(repo)
}

// --- ListUsers ---

func TestAdminUserService_ListUsers_OK(t *testing.T) {
	users := []models.User{{ID: "1", Name: "Test"}}
	svc := newAdminSvc(&fakeAdminUserRepo{users: users, total: 1})
	result, total, err := svc.ListUsers(repositories.UserFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestAdminUserService_ListUsers_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newAdminSvc(&fakeAdminUserRepo{findErr: repoErr})
	_, _, err := svc.ListUsers(repositories.UserFilter{}, 1, 10)
	assert.ErrorIs(t, err, repoErr)
}

// --- GetUser ---

func TestAdminUserService_GetUser_Found(t *testing.T) {
	u := &models.User{ID: "abc", Name: "An"}
	svc := newAdminSvc(&fakeAdminUserRepo{user: u})
	result, err := svc.GetUser("abc")
	assert.NoError(t, err)
	assert.Equal(t, "abc", result.ID)
}

func TestAdminUserService_GetUser_NotFound(t *testing.T) {
	svc := newAdminSvc(&fakeAdminUserRepo{user: nil})
	_, err := svc.GetUser("missing")
	assert.ErrorIs(t, err, ErrUserNotFoundAdmin)
}

func TestAdminUserService_GetUser_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newAdminSvc(&fakeAdminUserRepo{findErr: repoErr})
	_, err := svc.GetUser("x")
	assert.ErrorIs(t, err, repoErr)
}

// --- CreateUser ---

func TestAdminUserService_CreateUser_OK(t *testing.T) {
	svc := newAdminSvc(&fakeAdminUserRepo{findByEmail: nil})
	req := &dtos.AdminCreateUserRequest{
		Name:     "New User",
		Email:    "new@example.com",
		Password: "password123",
		Role:     "citizen",
	}
	user, err := svc.CreateUser(req, "actor-id")
	assert.NoError(t, err)
	assert.Equal(t, "New User", user.Name)
	assert.Equal(t, models.UserRoleCitizen, user.Role)
	assert.Equal(t, models.UserStatusActive, user.Status)
}

func TestAdminUserService_CreateUser_EmailExists(t *testing.T) {
	existing := &models.User{Email: "new@example.com"}
	svc := newAdminSvc(&fakeAdminUserRepo{findByEmail: existing})
	req := &dtos.AdminCreateUserRequest{
		Name: "User", Email: "new@example.com", Password: "pass123", Role: "citizen",
	}
	_, err := svc.CreateUser(req, "actor")
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
}

func TestAdminUserService_CreateUser_FindByEmailError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newAdminSvc(&fakeAdminUserRepo{findErr: repoErr})
	req := &dtos.AdminCreateUserRequest{
		Name: "User", Email: "new@example.com", Password: "pass123", Role: "citizen",
	}
	_, err := svc.CreateUser(req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestAdminUserService_CreateUser_CreateError(t *testing.T) {
	createErr := errors.New("create failed")
	svc := newAdminSvc(&fakeAdminUserRepo{createErr: createErr})
	req := &dtos.AdminCreateUserRequest{
		Name: "User", Email: "new@example.com", Password: "pass123", Role: "citizen",
	}
	_, err := svc.CreateUser(req, "actor")
	assert.ErrorIs(t, err, createErr)
}

// --- UpdateUser ---

func TestAdminUserService_UpdateUser_OK(t *testing.T) {
	u := &models.User{ID: "u1", Name: "Old", Role: "citizen"}
	svc := newAdminSvc(&fakeAdminUserRepo{user: u})
	req := &dtos.AdminUpdateUserRequest{Name: "New Name", Role: "staff"}
	result, err := svc.UpdateUser("u1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, models.UserRole("staff"), result.Role)
}

func TestAdminUserService_UpdateUser_NotFound(t *testing.T) {
	svc := newAdminSvc(&fakeAdminUserRepo{user: nil})
	req := &dtos.AdminUpdateUserRequest{Name: "Name", Role: "staff"}
	_, err := svc.UpdateUser("missing", req, "actor")
	assert.ErrorIs(t, err, ErrUserNotFoundAdmin)
}

func TestAdminUserService_UpdateUser_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newAdminSvc(&fakeAdminUserRepo{findErr: repoErr})
	req := &dtos.AdminUpdateUserRequest{Name: "Name", Role: "staff"}
	_, err := svc.UpdateUser("u1", req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestAdminUserService_UpdateUser_SaveError(t *testing.T) {
	u := &models.User{ID: "u1", Name: "Old"}
	saveErr := errors.New("save failed")
	svc := newAdminSvc(&fakeAdminUserRepo{user: u, updateErr: saveErr})
	req := &dtos.AdminUpdateUserRequest{Name: "New", Role: "citizen"}
	_, err := svc.UpdateUser("u1", req, "actor")
	assert.ErrorIs(t, err, saveErr)
}

// --- BlockUser ---

func TestAdminUserService_BlockUser_OK(t *testing.T) {
	u := &models.User{ID: "u1", Status: models.UserStatusActive}
	svc := newAdminSvc(&fakeAdminUserRepo{user: u})
	err := svc.BlockUser("u1", "actor")
	assert.NoError(t, err)
}

func TestAdminUserService_BlockUser_NotFound(t *testing.T) {
	svc := newAdminSvc(&fakeAdminUserRepo{user: nil})
	err := svc.BlockUser("missing", "actor")
	assert.ErrorIs(t, err, ErrUserNotFoundAdmin)
}

func TestAdminUserService_BlockUser_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newAdminSvc(&fakeAdminUserRepo{findErr: repoErr})
	err := svc.BlockUser("u1", "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestAdminUserService_BlockUser_StatusError(t *testing.T) {
	u := &models.User{ID: "u1"}
	statusErr := errors.New("status error")
	svc := newAdminSvc(&fakeAdminUserRepo{user: u, statusErr: statusErr})
	err := svc.BlockUser("u1", "actor")
	assert.ErrorIs(t, err, statusErr)
}

// --- UnblockUser ---

func TestAdminUserService_UnblockUser_OK(t *testing.T) {
	u := &models.User{ID: "u1", Status: models.UserStatusBlocked}
	svc := newAdminSvc(&fakeAdminUserRepo{user: u})
	err := svc.UnblockUser("u1", "actor")
	assert.NoError(t, err)
}

func TestAdminUserService_UnblockUser_NotFound(t *testing.T) {
	svc := newAdminSvc(&fakeAdminUserRepo{user: nil})
	err := svc.UnblockUser("missing", "actor")
	assert.ErrorIs(t, err, ErrUserNotFoundAdmin)
}

func TestAdminUserService_UnblockUser_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newAdminSvc(&fakeAdminUserRepo{findErr: repoErr})
	err := svc.UnblockUser("u1", "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestAdminUserService_UnblockUser_StatusError(t *testing.T) {
	u := &models.User{ID: "u1"}
	statusErr := errors.New("status error")
	svc := newAdminSvc(&fakeAdminUserRepo{user: u, statusErr: statusErr})
	err := svc.UnblockUser("u1", "actor")
	assert.ErrorIs(t, err, statusErr)
}

// --- DeleteUser ---

func TestAdminUserService_DeleteUser_OK(t *testing.T) {
	u := &models.User{ID: "u1"}
	svc := newAdminSvc(&fakeAdminUserRepo{user: u})
	err := svc.DeleteUser("u1", "actor")
	assert.NoError(t, err)
}

func TestAdminUserService_DeleteUser_NotFound(t *testing.T) {
	svc := newAdminSvc(&fakeAdminUserRepo{user: nil})
	err := svc.DeleteUser("missing", "actor")
	assert.ErrorIs(t, err, ErrUserNotFoundAdmin)
}

func TestAdminUserService_DeleteUser_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newAdminSvc(&fakeAdminUserRepo{findErr: repoErr})
	err := svc.DeleteUser("u1", "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestAdminUserService_DeleteUser_DeleteError(t *testing.T) {
	u := &models.User{ID: "u1"}
	deleteErr := errors.New("delete error")
	svc := newAdminSvc(&fakeAdminUserRepo{user: u, deleteErr: deleteErr})
	err := svc.DeleteUser("u1", "actor")
	assert.ErrorIs(t, err, deleteErr)
}

func TestAdminUserService_LogActivity_WithLogger(t *testing.T) {
	u := &models.User{ID: "u1", Email: "new@test.com", Name: "New", Role: models.UserRoleStaff}
	repo := &fakeAdminUserRepo{findByEmail: nil, user: u}
	logger := &fakeActivityLogger{}
	svc := NewAdminUserService(repo, logger)

	req := &dtos.AdminCreateUserRequest{Email: "new@test.com", Name: "New", Role: "staff"}
	_, err := svc.CreateUser(req, "actor1")
	assert.NoError(t, err)
	assert.Len(t, logger.entries, 1)
	assert.Equal(t, "user.create", logger.entries[0].Action)
}

func TestAdminUserService_LogActivity_LoggerError(t *testing.T) {
	u := &models.User{ID: "u1", Email: "new@test.com", Name: "New", Role: models.UserRoleStaff}
	repo := &fakeAdminUserRepo{findByEmail: nil, user: u}
	logger := &fakeActivityLogger{err: assert.AnError}
	svc := NewAdminUserService(repo, logger)

	req := &dtos.AdminCreateUserRequest{Email: "new@test.com", Name: "New", Role: "staff"}
	// Should not return error even if logger fails
	_, err := svc.CreateUser(req, "actor1")
	assert.NoError(t, err)
}

func TestAdminUserService_DefaultAdminPassword(t *testing.T) {
	pwd := defaultAdminPassword()
	assert.NotEmpty(t, pwd)
}
