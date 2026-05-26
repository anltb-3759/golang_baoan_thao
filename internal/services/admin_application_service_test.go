package services

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
)

type fakeAdminAppRepo struct {
	apps          []models.Application
	total         int64
	app           *models.Application
	err           error
	lastFilter    repositories.ApplicationFilter
	processStatus models.ApplicationStatus
	processNote   string
}

func (r *fakeAdminAppRepo) AdminList(filter repositories.ApplicationFilter, _, _ int) ([]models.Application, int64, error) {
	r.lastFilter = filter
	return r.apps, r.total, r.err
}
func (r *fakeAdminAppRepo) GetByID(_ string) (*models.Application, error) {
	return r.app, r.err
}
func (r *fakeAdminAppRepo) UpdateAssignedStaff(_ string, _ *string, _ string) error { return r.err }
func (r *fakeAdminAppRepo) ProcessStatusUpdate(_ string, _ *models.ApplicationStatus, newStatus models.ApplicationStatus, resultNote string, _ string, _, _ *time.Time, _ string, _ []models.ApplicationAttachment) error {
	r.processStatus = newStatus
	r.processNote = resultNote
	return r.err
}
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
func (r *fakeAdminAppRepo) GetDashboardStats() (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAdminAppRepo) ListRecent(_ int) ([]models.Application, error) { return nil, nil }
func (r *fakeAdminAppRepo) GetDashboardStatsForStaff(_ string) (repositories.DashboardStats, error) {
	return repositories.DashboardStats{}, nil
}
func (r *fakeAdminAppRepo) ListRecentForStaff(_ string, _ int) ([]models.Application, error) {
	return nil, nil
}

func newAdminAppSvc(repo *fakeAdminAppRepo) *AdminApplicationService {
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	return NewAdminApplicationService(repo, assignSvc, nil)
}

func newAdminAppSvcWithLogger(repo *fakeAdminAppRepo, logger activityLogger) *AdminApplicationService {
	assignSvc := NewApplicationAssignmentService(repo, &fakeAssignRepo{}, &fakeUserRepoAssign{})
	return NewAdminApplicationService(repo, assignSvc, nil, logger)
}

func TestAdminApplicationService_ListApplications_OK(t *testing.T) {
	apps := []models.Application{{ID: "a1", ApplicationCode: "APP-2024-001"}}
	svc := newAdminAppSvc(&fakeAdminAppRepo{apps: apps, total: 1})
	result, total, err := svc.ListApplications(repositories.ApplicationFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestAdminApplicationService_ListApplications_Error(t *testing.T) {
	svc := newAdminAppSvc(&fakeAdminAppRepo{err: errors.New("db error")})
	_, _, err := svc.ListApplications(repositories.ApplicationFilter{}, 1, 10)
	assert.Error(t, err)
}

func TestAdminApplicationService_ListApplications_WithFilter(t *testing.T) {
	repo := &fakeAdminAppRepo{apps: []models.Application{{ID: "a1"}}, total: 1}
	svc := newAdminAppSvc(repo)
	filter := repositories.ApplicationFilter{Status: string(models.ApplicationStatusProcessing), Service: "CCCD", Submitter: "An"}
	_, _, err := svc.ListApplications(filter, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, filter, repo.lastFilter)
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

func TestAdminApplicationService_ProcessApplication_OK(t *testing.T) {
	prevStatus := models.ApplicationStatusReceived
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: prevStatus}}
	logger := &fakeActivityLogger{}
	svc := newAdminAppSvcWithLogger(repo, logger)

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "processing", nil, "admin-1")
	assert.NoError(t, err)
	assert.Equal(t, models.ApplicationStatusProcessing, repo.processStatus)
	assert.Equal(t, "processing", repo.processNote)
	if assert.Len(t, logger.entries, 1) {
		assert.Equal(t, "application.status_update", logger.entries[0].Action)
		var metadata map[string]any
		assert.NoError(t, json.Unmarshal(logger.entries[0].MetadataJSON, &metadata))
		changes, ok := metadata["changes"].(map[string]any)
		assert.True(t, ok)
		status, ok := changes["status"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, string(models.ApplicationStatusReceived), status["before"])
		assert.Equal(t, string(models.ApplicationStatusProcessing), status["after"])
	}
}

func TestAdminApplicationService_ProcessApplication_LogFailureDoesNotBreakMainFlow(t *testing.T) {
	prevStatus := models.ApplicationStatusReceived
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: prevStatus}}
	svc := newAdminAppSvcWithLogger(repo, &fakeActivityLogger{err: errors.New("log failed")})

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "ok", nil, "admin-1")
	assert.NoError(t, err)
}

func TestAdminApplicationService_ProcessApplication_InvalidTransition(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusApproved}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusProcessing, "", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationInvalidTransition)
}

func TestAdminApplicationService_ProcessApplication_RejectedRequiresReason(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusProcessing}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusRejected, "   ", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationRejectReasonRequired)
}

func TestAdminApplicationService_ProcessApplication_AllowNeedMoreInfoTransitions(t *testing.T) {
	cases := []struct {
		name string
		from models.ApplicationStatus
		to   models.ApplicationStatus
		note string
	}{
		{
			name: "received to need_more_info",
			from: models.ApplicationStatusReceived,
			to:   models.ApplicationStatusNeedMoreInfo,
			note: "need additional papers",
		},
		{
			name: "processing to need_more_info",
			from: models.ApplicationStatusProcessing,
			to:   models.ApplicationStatusNeedMoreInfo,
			note: "missing photo",
		},
		{
			name: "need_more_info to processing",
			from: models.ApplicationStatusNeedMoreInfo,
			to:   models.ApplicationStatusProcessing,
			note: "citizen submitted extra docs",
		},
		{
			name: "need_more_info to approved",
			from: models.ApplicationStatusNeedMoreInfo,
			to:   models.ApplicationStatusApproved,
			note: "all docs are valid",
		},
		{
			name: "need_more_info to rejected",
			from: models.ApplicationStatusNeedMoreInfo,
			to:   models.ApplicationStatusRejected,
			note: "documents are inconsistent",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: tc.from}}
			svc := newAdminAppSvc(repo)

			err := svc.ProcessApplication("app1", tc.to, tc.note, nil, "admin-1")
			assert.NoError(t, err)
			assert.Equal(t, tc.to, repo.processStatus)
		})
	}
}

func TestAdminApplicationService_ProcessApplication_RequireNoteForNeedMoreInfo(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusProcessing}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusNeedMoreInfo, "   ", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationNeedMoreInfoNoteRequired)
	assert.Equal(t, models.ApplicationStatus(""), repo.processStatus)
}

func TestAdminApplicationService_ProcessApplication_RequireReasonForRejected(t *testing.T) {
	repo := &fakeAdminAppRepo{app: &models.Application{ID: "app1", Status: models.ApplicationStatusNeedMoreInfo}}
	svc := newAdminAppSvc(repo)

	err := svc.ProcessApplication("app1", models.ApplicationStatusRejected, "   ", nil, "admin-1")
	assert.ErrorIs(t, err, ErrAdminApplicationRejectReasonRequired)
	assert.Equal(t, models.ApplicationStatus(""), repo.processStatus)
}
