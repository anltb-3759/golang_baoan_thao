package services

import (
	"errors"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyExists = errors.New("auth.email_exists")
var ErrInvalidCredentials = errors.New("auth.invalid_credentials")
var ErrUserNotFound = errors.New("auth.user_not_found")
var ErrPasswordMismatch = errors.New("auth.password_mismatch")

type AuthService struct {
	userRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(reqData *dtos.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(reqData.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(reqData.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         reqData.Name,
		Email:        reqData.Email,
		PasswordHash: string(passwordHash),
		Role:         models.UserRoleCitizen,
		Status:       models.UserStatusActive,
	}

	return s.userRepo.Create(user)
}

func (s *AuthService) Login(reqData *dtos.LoginRequest) (*models.User, string, string, error) {
	user, err := s.userRepo.FindByEmail(reqData.Email)
	if err != nil {
		return nil, "", "", err
	}
	if user == nil {
		return nil, "", "", ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(reqData.Password)); err != nil {
		return nil, "", "", ErrPasswordMismatch
	}
	token, refreshToken, err := configs.GenerateTokenPair(user)
	return user, token, refreshToken, err
}
