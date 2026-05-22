package services

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

type AdminApplicationService struct {
	appRepo       repositories.ApplicationRepository
	assignService *ApplicationAssignmentService
}

func NewAdminApplicationService(appRepo repositories.ApplicationRepository, assignService *ApplicationAssignmentService) *AdminApplicationService {
	return &AdminApplicationService{appRepo: appRepo, assignService: assignService}
}

func (s *AdminApplicationService) ListApplications(page, limit int) ([]models.Application, int64, error) {
	return s.appRepo.AdminList(page, limit)
}

func (s *AdminApplicationService) GetApplication(id string) (*models.Application, error) {
	return s.appRepo.GetByID(id)
}

func (s *AdminApplicationService) AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	if err := s.assignService.AssignApplicationToStaff(applicationID, toStaffUserID, assignedBy); err != nil {
		return err
	}
	return nil
}
