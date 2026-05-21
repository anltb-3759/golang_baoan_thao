package services

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFoundAdmin = errors.New("admin.user_not_found")

type AdminUserService struct {
	userRepo repositories.UserRepository
}

func NewAdminUserService(userRepo repositories.UserRepository) *AdminUserService {
	return &AdminUserService{userRepo: userRepo}
}

func (s *AdminUserService) ListUsers(filter repositories.UserFilter, page, limit int) ([]models.User, int64, error) {
	offset := (page - 1) * limit
	return s.userRepo.List(filter, offset, limit)
}

func (s *AdminUserService) GetUser(id string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFoundAdmin
	}
	return user, nil
}

func (s *AdminUserService) CreateUser(req *dtos.AdminCreateUserRequest, createdBy string) (*models.User, error) {
	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         models.UserRole(req.Role),
		Phone:        req.Phone,
		Address:      req.Address,
		Status:       models.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &createdBy,
	}
	return s.userRepo.Create(user)
}

func (s *AdminUserService) UpdateUser(id string, req *dtos.AdminUpdateUserRequest, updatedBy string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFoundAdmin
	}

	user.Name = req.Name
	user.Role = models.UserRole(req.Role)
	user.Phone = req.Phone
	user.Address = req.Address
	user.UpdatedAt = time.Now()
	user.UpdatedBy = &updatedBy

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AdminUserService) BlockUser(id string, updatedBy string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFoundAdmin
	}
	return s.userRepo.UpdateStatus(id, models.UserStatusBlocked, updatedBy)
}

func (s *AdminUserService) UnblockUser(id string, updatedBy string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFoundAdmin
	}
	return s.userRepo.UpdateStatus(id, models.UserStatusActive, updatedBy)
}

func (s *AdminUserService) DeleteUser(id string, deletedBy string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFoundAdmin
	}
	return s.userRepo.SoftDelete(id, deletedBy)
}
