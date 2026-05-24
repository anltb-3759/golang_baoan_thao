package services

import (
	"errors"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/stretchr/testify/assert"
)

type fakeAdminAppRepo struct {
	apps  []models.Application
	total int64
	app   *models.Application
	err   error
}

func (r *fakeAdminAppRepo) AdminList(_, _ int) ([]models.Application, int64, error) {
	return r.apps, r.total, r.err
}
func (r *fakeAdminAppRepo) GetByID(_ string) (*models.Application, error) {
	return r.app, r.err
}
func (r *fakeAdminAppRepo) UpdateAssignedStaff(_ string, _ *string, _ string) error { return r.err }
func (r *fakeAdminAppRepo) CreateWithAttachments(_ *models.Application, _ []models.ApplicationAttachment, _ *models.Notification, _ func() string) error {
	return nil
}
func (r *fakeAdminAppRepo) ListByCitizen(_ string, _, _ int) ([]models.Application, int64, error) {
	return nil, 0, nil
}
func (r *fakeAdminAppRepo) ListStatusLogsByCitizen(_ string, _ string, _, _ int, _ *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	return nil, 0, nil
}
func (r *fakeAdminAppRepo) CreateAttachments(_ string, _ []models.ApplicationAttachment) error {
	return nil
}
func (r *fakeAdminAppRepo) GetByIDForCitizen(_, _ string) (*models.Application, error) {
	return r.app, r.err
}

func newAdminAppSvc(repo *fakeAdminAppRepo) *AdminApplicationService {
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	return NewAdminApplicationService(repo, assignSvc)
}

func TestAdminApplicationService_ListApplications_OK(t *testing.T) {
	apps := []models.Application{{ID: "a1", ApplicationCode: "APP-2024-001"}}
	svc := newAdminAppSvc(&fakeAdminAppRepo{apps: apps, total: 1})
	result, total, err := svc.ListApplications(1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestAdminApplicationService_ListApplications_Error(t *testing.T) {
	svc := newAdminAppSvc(&fakeAdminAppRepo{err: errors.New("db error")})
	_, _, err := svc.ListApplications(1, 10)
	assert.Error(t, err)
}

func TestAdminApplicationService_GetApplication_OK(t *testing.T) {
	app := &models.Application{ID: "a1"}
	svc := newAdminAppSvc(&fakeAdminAppRepo{app: app})
	result, err := svc.GetApplication("a1")
	assert.NoError(t, err)
	assert.Equal(t, "a1", result.ID)
}

func TestAdminApplicationService_GetApplication_Error(t *testing.T) {
	svc := newAdminAppSvc(&fakeAdminAppRepo{err: errors.New("not found")})
	_, err := svc.GetApplication("missing")
	assert.Error(t, err)
}

func TestAdminApplicationService_AssignToStaff_AppNotFound(t *testing.T) {
	svc := newAdminAppSvc(&fakeAdminAppRepo{err: errors.New("not found")})
	err := svc.AssignToStaff("appX", nil, "admin-1")
	assert.Error(t, err)
}

func TestAdminApplicationService_AssignToStaff_Unassign(t *testing.T) {
	prevID := "u0"
	app := &models.Application{ID: "app1", AssignedStaffUserID: &prevID}
	svc := newAdminAppSvc(&fakeAdminAppRepo{app: app})
	err := svc.AssignToStaff("app1", nil, "admin-1")
	assert.NoError(t, err)
}
