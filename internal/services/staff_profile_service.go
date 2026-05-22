package services

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

type StaffProfileService struct {
	repo     repositories.StaffProfileRepository
	userRepo repositories.UserRepository
}

func NewStaffProfileService(repo repositories.StaffProfileRepository, userRepo repositories.UserRepository) *StaffProfileService {
	return &StaffProfileService{repo: repo, userRepo: userRepo}
}

func (s *StaffProfileService) ListStaffByDepartment(deptID string, page, limit int) ([]models.StaffProfile, int64, error) {
	offset := (page - 1) * limit
	return s.repo.ListByDepartment(deptID, offset, limit)
}

func (s *StaffProfileService) AssignStaffToDepartment(userID string, deptID string, updatedBy string) error {
	// validate user exists
	u, err := s.userRepo.FindByID(userID)
	if err != nil || u == nil {
		return err
	}
	return s.repo.UpdateDepartment(userID, &deptID, updatedBy)
}

func (s *StaffProfileService) RemoveStaffFromDepartment(userID string, updatedBy string) error {
	return s.repo.UpdateDepartment(userID, nil, updatedBy)
}
