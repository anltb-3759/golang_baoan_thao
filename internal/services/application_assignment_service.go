package services

import (
	"errors"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrApplicationNotFoundAssign = errors.New("application.not_found")
var ErrUserNotFoundAssign = errors.New("user.not_found")

type ApplicationAssignmentService struct {
	appRepo    repositories.ApplicationRepository
	assignRepo repositories.ApplicationAssignmentRepository
	userRepo   repositories.UserRepository
}

func NewApplicationAssignmentService(appRepo repositories.ApplicationRepository, assignRepo repositories.ApplicationAssignmentRepository, userRepo repositories.UserRepository) *ApplicationAssignmentService {
	return &ApplicationAssignmentService{appRepo: appRepo, assignRepo: assignRepo, userRepo: userRepo}
}

func (s *ApplicationAssignmentService) AssignApplicationToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	app, err := s.appRepo.GetByID(applicationID)
	if err != nil {
		return ErrApplicationNotFoundAssign
	}

	var toUser *models.User
	if toStaffUserID != nil && *toStaffUserID != "" {
		toUser, err = s.userRepo.FindByID(*toStaffUserID)
		if err != nil || toUser == nil {
			return ErrUserNotFoundAssign
		}
	}

	var action models.AssignmentAction
	if app.AssignedStaffUserID == nil && toStaffUserID != nil {
		action = models.AssignmentActionAssigned
	} else if app.AssignedStaffUserID != nil && toStaffUserID != nil && *app.AssignedStaffUserID != *toStaffUserID {
		action = models.AssignmentActionTransferred
	} else if toStaffUserID == nil {
		action = models.AssignmentActionUnassigned
	} else {
		// no-op (assigning to same user)
		return nil
	}

	assignment := &models.ApplicationAssignment{
		ApplicationID:    applicationID,
		FromStaffUserID:  app.AssignedStaffUserID,
		ToStaffUserID:    toStaffUserID,
		AssignedByUserID: &assignedBy,
		Action:           action,
		Note:             "",
	}

	if _, err := s.assignRepo.Create(assignment); err != nil {
		return err
	}

	if err := s.appRepo.UpdateAssignedStaff(applicationID, toStaffUserID, assignedBy); err != nil {
		return err
	}

	return nil
}
