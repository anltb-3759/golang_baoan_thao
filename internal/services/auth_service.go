package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Transactor abstracts gorm.DB.Transaction so AuthService can be unit-tested without a real DB.
type Transactor interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}

var ErrEmailAlreadyExists = errors.New("auth.email_exists")
var ErrInvalidCredentials = errors.New("auth.invalid_credentials")
var ErrUserNotFound = errors.New("auth.user_not_found")
var ErrPasswordMismatch = errors.New("auth.password_mismatch")
var ErrUserBlocked = errors.New("auth.user_blocked")

type AuthService struct {
	db          Transactor
	userRepo    repositories.UserRepository
	profileRepo repositories.CitizenProfileRepository
}

func NewAuthService(db Transactor, userRepo repositories.UserRepository, profileRepo repositories.CitizenProfileRepository) *AuthService {
	return &AuthService{db: db, userRepo: userRepo, profileRepo: profileRepo}
}

func (s *AuthService) Register(reqData *dtos.RegisterRequest) (*models.User, error) {
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

	now := time.Now()
	user := &models.User{
		Name:         reqData.Name,
		Email:        reqData.Email,
		PasswordHash: string(passwordHash),
		Role:         models.UserRoleCitizen,
		Status:       models.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.CreateInTx(tx, user); err != nil {
			return err
		}
		profile := &models.CitizenProfile{
			UserID:                   user.ID,
			CitizenIDNumber:          reqData.CitizenIDNumber,
			EmailNotificationEnabled: true,
			CreatedAt:                now,
			UpdatedAt:                now,
		}
		return s.profileRepo.CreateInTx(tx, profile)
	})
	if txErr != nil {
		return nil, txErr
	}

	return user, nil
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
	if user.Status == models.UserStatusBlocked {
		return nil, "", "", ErrUserBlocked
	}
	token, refreshToken, err := configs.GenerateTokenPair(user)
	return user, token, refreshToken, err
}
