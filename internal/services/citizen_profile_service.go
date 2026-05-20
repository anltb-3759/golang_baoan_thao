package services

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrProfileNotFound = errors.New("profile.not_found")

type CitizenProfileService struct {
	userRepo    repositories.UserRepository
	profileRepo repositories.CitizenProfileRepository
	appRepo     repositories.ApplicationRepository
}

func NewCitizenProfileService(
	userRepo repositories.UserRepository,
	profileRepo repositories.CitizenProfileRepository,
	appRepo repositories.ApplicationRepository,
) *CitizenProfileService {
	return &CitizenProfileService{
		userRepo:    userRepo,
		profileRepo: profileRepo,
		appRepo:     appRepo,
	}
}

func (s *CitizenProfileService) GetProfile(userID string) (*dtos.CitizenProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrProfileNotFound
	}

	profile, err := s.profileRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		profile = &models.CitizenProfile{
			UserID:    userID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	return dtos.NewCitizenProfileResponse(user, profile), nil
}

func (s *CitizenProfileService) UpdateProfile(userID string, req *dtos.UpdateCitizenProfileRequest) (*dtos.CitizenProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrProfileNotFound
	}

	profile, err := s.profileRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Address != nil {
		user.Address = *req.Address
	}
	if req.Gender != nil {
		profile.Gender = *req.Gender
	}
	if req.PermanentAddress != nil {
		profile.PermanentAddress = *req.PermanentAddress
	}
	if req.DateOfBirth != nil {
		profile.DateOfBirth = req.DateOfBirth
	}
	if req.EmailNotificationEnabled != nil {
		profile.EmailNotificationEnabled = *req.EmailNotificationEnabled
	}

	now := time.Now()
	user.UpdatedAt = now
	profile.UpdatedAt = now

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	if err := s.profileRepo.Update(profile); err != nil {
		return nil, err
	}

	return dtos.NewCitizenProfileResponse(user, profile), nil
}

func (s *CitizenProfileService) ListMyApplications(userID string, page, limit int) ([]models.Application, int64, error) {
	return s.appRepo.ListByCitizen(userID, page, limit)
}

