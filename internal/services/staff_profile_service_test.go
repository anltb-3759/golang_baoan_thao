package services

import (
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type fakeStaffProfileRepo struct {
	profiles  []models.StaffProfile
	total     int64
	listErr   error
	updateErr error
}

func (r *fakeStaffProfileRepo) FindByUserID(_ string) (*models.StaffProfile, error) {
	return nil, nil
}
func (r *fakeStaffProfileRepo) ListByDepartment(_ string, _, _ int) ([]models.StaffProfile, int64, error) {
	return r.profiles, r.total, r.listErr
}
func (r *fakeStaffProfileRepo) UpdateDepartment(_ string, _ *string, _ string) error {
	return r.updateErr
}
func (r *fakeStaffProfileRepo) Create(p *models.StaffProfile) (*models.StaffProfile, error) {
	return p, nil
}
func (r *fakeStaffProfileRepo) CreateInTx(_ *gorm.DB, _ *models.StaffProfile) error { return nil }

var _ repositories.StaffProfileRepository = (*fakeStaffProfileRepo)(nil)

type fakeStaffUserRepo struct {
	user *models.User
}

func (r *fakeStaffUserRepo) FindByID(_ string) (*models.User, error) { return r.user, nil }
func (r *fakeStaffUserRepo) FindByEmail(_ string) (*models.User, error) { return nil, nil }
func (r *fakeStaffUserRepo) Create(u *models.User) (*models.User, error) { return u, nil }
func (r *fakeStaffUserRepo) CreateInTx(_ *gorm.DB, _ *models.User) error { return nil }
func (r *fakeStaffUserRepo) Update(_ *models.User) error { return nil }
func (r *fakeStaffUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeStaffUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeStaffUserRepo) SoftDelete(_ string, _ string) error                        { return nil }

func TestStaffProfileService_ListStaffByDepartment_OK(t *testing.T) {
	profiles := []models.StaffProfile{{User: models.User{ID: "u1"}}}
	repo := &fakeStaffProfileRepo{profiles: profiles, total: 1}
	svc := NewStaffProfileService(repo, nil)
	result, total, err := svc.ListStaffByDepartment("dept-1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestStaffProfileService_AssignStaffToDepartment_UserFound(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	userRepo := &fakeIEUserRepo{} // reuse existing fake that returns nil, nil for FindByID
	svc := NewStaffProfileService(repo, userRepo)
	// FindByID returns nil user → should return nil without error
	err := svc.AssignStaffToDepartment("u1", "dept-1", "admin-1")
	assert.NoError(t, err)
}

func TestStaffProfileService_AssignStaffToDepartment_UpdateDepartment(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	// Use a fake user repo that returns a valid user
	svc := NewStaffProfileService(repo, &fakeStaffUserRepo{user: &models.User{ID: "u1"}})
	err := svc.AssignStaffToDepartment("u1", "dept-1", "admin-1")
	assert.NoError(t, err)
}

func TestStaffProfileService_RemoveStaffFromDepartment_OK(t *testing.T) {
	repo := &fakeStaffProfileRepo{}
	svc := NewStaffProfileService(repo, nil)
	err := svc.RemoveStaffFromDepartment("u1", "admin-1")
	assert.NoError(t, err)
}
