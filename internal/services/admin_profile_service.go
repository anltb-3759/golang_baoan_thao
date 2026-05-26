package services

import (
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type AdminProfileService struct {
	userRepo repositories.UserRepository
}

func NewAdminProfileService(userRepo repositories.UserRepository) *AdminProfileService {
	return &AdminProfileService{userRepo: userRepo}
}

func (s *AdminProfileService) GetSelf(userID string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *AdminProfileService) UpdateSelfContact(userID string, req *dtos.UpdateAdminProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	user.Phone = req.Phone
	user.Address = req.Address
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AdminProfileService) ChangeSelfPassword(userID string, req *dtos.ChangeAdminPasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if req.NewPassword != req.ConfirmNewPassword {
		return ErrPasswordConfirmationMismatch
	}
	if req.NewPassword == req.CurrentPassword {
		return ErrNewPasswordMustDiffer
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrPasswordMismatch
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(user)
}
