package services

import (
	"errors"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
)

type fakeActivityLogRepo struct {
	createErr         error
	deleteRetainErr   error
	deleteBeforeErr   error
	createdLog        *models.ActivityLog
	deleteRetainDays  int
	deleteBeforeValue time.Time
	deleteRetainCalls int
	deleteBeforeCalls int
}

func (r *fakeActivityLogRepo) Create(log *models.ActivityLog) error {
	r.createdLog = log
	return r.createErr
}

func (r *fakeActivityLogRepo) List(_ repositories.ActivityLogFilter, _, _ int) ([]models.ActivityLog, int64, error) {
	return nil, 0, nil
}

func (r *fakeActivityLogRepo) DeleteBefore(before time.Time) (int64, error) {
	r.deleteBeforeCalls++
	r.deleteBeforeValue = before
	if r.deleteBeforeErr != nil {
		return 0, r.deleteBeforeErr
	}
	return 2, nil
}

func (r *fakeActivityLogRepo) DeleteRetainDays(days int) (int64, error) {
	r.deleteRetainCalls++
	r.deleteRetainDays = days
	if r.deleteRetainErr != nil {
		return 0, r.deleteRetainErr
	}
	return 5, nil
}

var _ repositories.ActivityLogRepository = (*fakeActivityLogRepo)(nil)

func TestActivityLogService_LogSuccess(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	log := &models.ActivityLog{Action: "auth.login"}
	err := svc.Log(log)

	assert.NoError(t, err)
	assert.Same(t, log, repo.createdLog)
}

func TestActivityLogService_Cleanup_InvalidBothModes(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	days := 7
	before := time.Now().AddDate(0, 0, -30)

	deleted, err := svc.Cleanup(&days, &before)

	assert.ErrorIs(t, err, ErrActivityLogCleanupModeInvalid)
	assert.Equal(t, int64(0), deleted)
	assert.Equal(t, 0, repo.deleteRetainCalls)
	assert.Equal(t, 0, repo.deleteBeforeCalls)
}

func TestActivityLogService_Cleanup_ByRetainDays(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	days := 14
	deleted, err := svc.Cleanup(&days, nil)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), deleted)
	assert.Equal(t, 14, repo.deleteRetainDays)
	assert.Equal(t, 1, repo.deleteRetainCalls)
	assert.Equal(t, 0, repo.deleteBeforeCalls)
}

func TestActivityLogService_Cleanup_ByRetainDays_RepoError(t *testing.T) {
	repoErr := errors.New("delete retain days failed")
	repo := &fakeActivityLogRepo{deleteRetainErr: repoErr}
	svc := NewActivityLogService(repo)

	days := 14
	deleted, err := svc.Cleanup(&days, nil)

	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, int64(0), deleted)
	assert.Equal(t, 1, repo.deleteRetainCalls)
	assert.Equal(t, 0, repo.deleteBeforeCalls)
}

func TestActivityLogService_Cleanup_ByDeleteBefore(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	before := time.Now().AddDate(0, 0, -10).UTC()
	deleted, err := svc.Cleanup(nil, &before)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), deleted)
	assert.Equal(t, 1, repo.deleteBeforeCalls)
	assert.True(t, before.Equal(repo.deleteBeforeValue))
	assert.Equal(t, 0, repo.deleteRetainCalls)
}

func TestActivityLogService_Cleanup_ByDeleteBefore_RepoError(t *testing.T) {
	repoErr := errors.New("delete before failed")
	repo := &fakeActivityLogRepo{deleteBeforeErr: repoErr}
	svc := NewActivityLogService(repo)

	before := time.Now().AddDate(0, 0, -10).UTC()
	deleted, err := svc.Cleanup(nil, &before)

	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, int64(0), deleted)
	assert.Equal(t, 1, repo.deleteBeforeCalls)
	assert.Equal(t, 0, repo.deleteRetainCalls)
}

func TestActivityLogService_Cleanup_InvalidRetainDays(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	days := 0
	deleted, err := svc.Cleanup(&days, nil)

	assert.ErrorIs(t, err, ErrActivityLogRetainDaysInvalid)
	assert.Equal(t, int64(0), deleted)
	assert.Equal(t, 0, repo.deleteRetainCalls)
}

func TestActivityLogService_Cleanup_InvalidNoMode(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	deleted, err := svc.Cleanup(nil, nil)

	assert.ErrorIs(t, err, ErrActivityLogCleanupModeInvalid)
	assert.Equal(t, int64(0), deleted)
}

func TestActivityLogService_LogRepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &fakeActivityLogRepo{createErr: repoErr}
	svc := NewActivityLogService(repo)

	err := svc.Log(&models.ActivityLog{Action: "auth.login"})

	assert.ErrorIs(t, err, repoErr)
}

func TestActivityLogService_LogNilInput(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	err := svc.Log(nil)

	assert.ErrorIs(t, err, ErrActivityLogInvalidInput)
	assert.Nil(t, repo.createdLog)
}

func TestActivityLogService_List_DefaultPagination(t *testing.T) {
	repo := &fakeActivityLogRepo{}
	svc := NewActivityLogService(repo)

	// page=0, limit=0 should be normalized to 1, 20
	logs, total, err := svc.List(repositories.ActivityLogFilter{}, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, logs)
}

func TestActivityLogService_List_WithFilter(t *testing.T) {
	repoWithList := &fakeActivityLogRepoWithList{
		logs:  []models.ActivityLog{{Action: "auth.login"}},
		total: 1,
	}
	svc := NewActivityLogService(repoWithList)

	logs, total, err := svc.List(repositories.ActivityLogFilter{Action: "auth.login"}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
	assert.Equal(t, "auth.login", logs[0].Action)
}

func TestActivityLogService_List_Error(t *testing.T) {
	listErr := errors.New("list error")
	repoWithList := &fakeActivityLogRepoWithList{err: listErr}
	svc := NewActivityLogService(repoWithList)

	_, _, err := svc.List(repositories.ActivityLogFilter{}, 1, 10)
	assert.ErrorIs(t, err, listErr)
}

type fakeActivityLogRepoWithList struct {
	fakeActivityLogRepo
	logs  []models.ActivityLog
	total int64
	err   error
}

func (r *fakeActivityLogRepoWithList) List(_ repositories.ActivityLogFilter, _, _ int) ([]models.ActivityLog, int64, error) {
	return r.logs, r.total, r.err
}
