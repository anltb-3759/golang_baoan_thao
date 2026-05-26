package services

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

type AdminDashboardService struct {
	appRepo repositories.DashboardRepository
}

func NewAdminDashboardService(appRepo repositories.DashboardRepository) *AdminDashboardService {
	return &AdminDashboardService{appRepo: appRepo}
}

type DashboardData struct {
	Stats          repositories.DashboardStats
	RecentApplications []models.Application
}

func (s *AdminDashboardService) GetDashboardData(staffID string) (DashboardData, error) {
	var (
		stats  repositories.DashboardStats
		recent []models.Application
		err    error
	)
	if staffID != "" {
		stats, err = s.appRepo.GetDashboardStatsForStaff(staffID)
		if err != nil {
			return DashboardData{}, err
		}
		recent, err = s.appRepo.ListRecentForStaff(staffID, 5)
	} else {
		stats, err = s.appRepo.GetDashboardStats()
		if err != nil {
			return DashboardData{}, err
		}
		recent, err = s.appRepo.ListRecent(5)
	}
	if err != nil {
		return DashboardData{}, err
	}
	return DashboardData{Stats: stats, RecentApplications: recent}, nil
}
