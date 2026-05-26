package services

import (
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
)

type fakeDashboardRepo struct {
	stats           repositories.DashboardStats
	statsErr        error
	staffStats      repositories.DashboardStats
	staffStatsErr   error
	recent          []models.Application
	recentErr       error
	staffRecent     []models.Application
	staffRecentErr  error
}

func (r *fakeDashboardRepo) GetDashboardStats() (repositories.DashboardStats, error) {
	return r.stats, r.statsErr
}

func (r *fakeDashboardRepo) ListRecent(_ int) ([]models.Application, error) {
	return r.recent, r.recentErr
}

func (r *fakeDashboardRepo) GetDashboardStatsForStaff(_ string) (repositories.DashboardStats, error) {
	return r.staffStats, r.staffStatsErr
}

func (r *fakeDashboardRepo) ListRecentForStaff(_ string, _ int) ([]models.Application, error) {
	return r.staffRecent, r.staffRecentErr
}

var _ repositories.DashboardRepository = (*fakeDashboardRepo)(nil)

func TestAdminDashboardService_GetDashboardData_Admin_Success(t *testing.T) {
	stats := repositories.DashboardStats{Total: 10, Pending: 3, Approved: 5, Rejected: 2}
	recent := []models.Application{{ID: "app-1"}, {ID: "app-2"}}
	repo := &fakeDashboardRepo{stats: stats, recent: recent}
	svc := NewAdminDashboardService(repo)

	data, err := svc.GetDashboardData("")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), data.Stats.Total)
	assert.Len(t, data.RecentApplications, 2)
}

func TestAdminDashboardService_GetDashboardData_Admin_StatsError(t *testing.T) {
	repoErr := errors.New("stats error")
	repo := &fakeDashboardRepo{statsErr: repoErr}
	svc := NewAdminDashboardService(repo)

	data, err := svc.GetDashboardData("")
	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, DashboardData{}, data)
}

func TestAdminDashboardService_GetDashboardData_Admin_RecentError(t *testing.T) {
	repoErr := errors.New("recent error")
	repo := &fakeDashboardRepo{
		stats:     repositories.DashboardStats{Total: 5},
		recentErr: repoErr,
	}
	svc := NewAdminDashboardService(repo)

	data, err := svc.GetDashboardData("")
	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, DashboardData{}, data)
}

func TestAdminDashboardService_GetDashboardData_Staff_Success(t *testing.T) {
	staffStats := repositories.DashboardStats{Total: 4, Pending: 2, Approved: 2}
	staffRecent := []models.Application{{ID: "app-s1"}}
	repo := &fakeDashboardRepo{staffStats: staffStats, staffRecent: staffRecent}
	svc := NewAdminDashboardService(repo)

	data, err := svc.GetDashboardData("staff-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(4), data.Stats.Total)
	assert.Len(t, data.RecentApplications, 1)
	assert.Equal(t, "app-s1", data.RecentApplications[0].ID)
}

func TestAdminDashboardService_GetDashboardData_Staff_StatsError(t *testing.T) {
	repoErr := errors.New("staff stats error")
	repo := &fakeDashboardRepo{staffStatsErr: repoErr}
	svc := NewAdminDashboardService(repo)

	data, err := svc.GetDashboardData("staff-1")
	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, DashboardData{}, data)
}

func TestAdminDashboardService_GetDashboardData_Staff_RecentError(t *testing.T) {
	repoErr := errors.New("staff recent error")
	repo := &fakeDashboardRepo{
		staffStats:     repositories.DashboardStats{Total: 2},
		staffRecentErr: repoErr,
	}
	svc := NewAdminDashboardService(repo)

	data, err := svc.GetDashboardData("staff-1")
	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, DashboardData{}, data)
}
