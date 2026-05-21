package services

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --- fakeTransactor ---

type fakeTransactor struct {
	err error
}

func (t *fakeTransactor) Transaction(fc func(tx *gorm.DB) error, _ ...*sql.TxOptions) error {
	if t.err != nil {
		return t.err
	}
	return fc(nil)
}

// --- fakeUserRepository ---

type fakeUserRepository struct {
	findByEmailUser *models.User
	findByEmailErr  error
	createUser      *models.User
	createErr       error
	createdUser     *models.User
}

func (r *fakeUserRepository) FindByEmail(email string) (*models.User, error) {
	return r.findByEmailUser, r.findByEmailErr
}

func (r *fakeUserRepository) FindByID(_ string) (*models.User, error) { return nil, nil }

func (r *fakeUserRepository) Create(user *models.User) (*models.User, error) {
	r.createdUser = user
	if r.createErr != nil {
		return nil, r.createErr
	}
	if r.createUser != nil {
		return r.createUser, nil
	}
	return user, nil
}

func (r *fakeUserRepository) CreateInTx(_ *gorm.DB, user *models.User) error {
	r.createdUser = user
	return r.createErr
}

func (r *fakeUserRepository) Update(_ *models.User) error { return nil }
func (r *fakeUserRepository) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeUserRepository) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeUserRepository) SoftDelete(_ string, _ string) error                        { return nil }

// --- fakeCitizenProfileRepository ---

type fakeCitizenProfileRepository struct {
	createErr error
	created   bool
}

func (r *fakeCitizenProfileRepository) GetByUserID(_ string) (*models.CitizenProfile, error) {
	return nil, nil
}

func (r *fakeCitizenProfileRepository) Update(_ *models.CitizenProfile) error { return nil }

func (r *fakeCitizenProfileRepository) CreateInTx(_ *gorm.DB, _ *models.CitizenProfile) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = true
	return nil
}

// --- helpers ---

func newAuthService(userRepo *fakeUserRepository, profileRepo *fakeCitizenProfileRepository) *AuthService {
	return NewAuthService(&fakeTransactor{}, userRepo, profileRepo)
}

// --- Register tests ---

func TestAuthServiceRegisterCreatesCitizenUser(t *testing.T) {
	userRepo := &fakeUserRepository{}
	profileRepo := &fakeCitizenProfileRepository{}
	service := newAuthService(userRepo, profileRepo)

	user, err := service.Register(&dtos.RegisterRequest{
		Name:     "User",
		Email:    "user@example.com",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if userRepo.createdUser == nil {
		t.Fatal("expected user to be persisted")
	}
	if userRepo.createdUser.Role != models.UserRoleCitizen {
		t.Fatalf("expected citizen role, got %q", userRepo.createdUser.Role)
	}
	if userRepo.createdUser.Status != models.UserStatusActive {
		t.Fatalf("expected active status, got %q", userRepo.createdUser.Status)
	}
	if userRepo.createdUser.PasswordHash == "123456" {
		t.Fatal("expected password to be hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(userRepo.createdUser.PasswordHash), []byte("123456")); err != nil {
		t.Fatalf("expected password hash to match original password: %v", err)
	}
	if !profileRepo.created {
		t.Fatal("expected citizen profile to be created alongside user")
	}
}

func TestAuthServiceRegisterReturnsEmailExists(t *testing.T) {
	service := newAuthService(
		&fakeUserRepository{findByEmailUser: &models.User{Email: "user@example.com"}},
		&fakeCitizenProfileRepository{},
	)

	user, err := service.Register(&dtos.RegisterRequest{
		Name:     "User",
		Email:    "user@example.com",
		Password: "123456",
	})
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}
}

func TestAuthServiceRegisterReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("repository error")
	service := newAuthService(
		&fakeUserRepository{findByEmailErr: repoErr},
		&fakeCitizenProfileRepository{},
	)

	user, err := service.Register(&dtos.RegisterRequest{
		Name:     "User",
		Email:    "user@example.com",
		Password: "123456",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}
}

// --- Login tests ---

func TestAuthServiceLoginReturnsTokens(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	service := newAuthService(
		&fakeUserRepository{
			findByEmailUser: &models.User{
				ID:           "user-id",
				Email:        "user@example.com",
				PasswordHash: string(passwordHash),
				Role:         models.UserRoleCitizen,
			},
		},
		&fakeCitizenProfileRepository{},
	)

	user, token, refreshToken, err := service.Login(&dtos.LoginRequest{
		Email:    "user@example.com",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if token == "" {
		t.Fatal("expected access token")
	}
	if refreshToken == "" {
		t.Fatal("expected refresh token")
	}
}

func TestAuthServiceLoginReturnsUserNotFound(t *testing.T) {
	service := newAuthService(&fakeUserRepository{}, &fakeCitizenProfileRepository{})

	user, token, refreshToken, err := service.Login(&dtos.LoginRequest{
		Email:    "missing@example.com",
		Password: "123456",
	})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if user != nil || token != "" || refreshToken != "" {
		t.Fatalf("expected empty login result, got user=%#v token=%q refresh=%q", user, token, refreshToken)
	}
}

func TestAuthServiceLoginReturnsRepositoryError(t *testing.T) {
	repoErr := errors.New("db error")
	service := newAuthService(
		&fakeUserRepository{findByEmailErr: repoErr},
		&fakeCitizenProfileRepository{},
	)

	user, token, refreshToken, err := service.Login(&dtos.LoginRequest{
		Email:    "user@example.com",
		Password: "123456",
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if user != nil || token != "" || refreshToken != "" {
		t.Fatalf("expected empty result, got user=%#v token=%q refresh=%q", user, token, refreshToken)
	}
}

func TestAuthServiceRegisterPasswordTooLong(t *testing.T) {
	service := newAuthService(&fakeUserRepository{}, &fakeCitizenProfileRepository{})

	// bcrypt rejects passwords longer than 72 bytes
	longPassword := string(make([]byte, 73))
	user, err := service.Register(&dtos.RegisterRequest{
		Name:     "User",
		Email:    "user@example.com",
		Password: longPassword,
	})
	if err == nil {
		t.Fatal("expected error for too-long password, got nil")
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}
}

func TestAuthServiceRegisterUserCreateInTxFails(t *testing.T) {
	createErr := errors.New("insert error")
	service := newAuthService(
		&fakeUserRepository{createErr: createErr},
		&fakeCitizenProfileRepository{},
	)

	user, err := service.Register(&dtos.RegisterRequest{
		Name:     "User",
		Email:    "user@example.com",
		Password: "123456",
	})
	if !errors.Is(err, createErr) {
		t.Fatalf("expected createErr, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}
}

func TestAuthServiceRegisterProfileCreateInTxFails(t *testing.T) {
	profileErr := errors.New("profile insert error")
	service := newAuthService(
		&fakeUserRepository{},
		&fakeCitizenProfileRepository{createErr: profileErr},
	)

	user, err := service.Register(&dtos.RegisterRequest{
		Name:     "User",
		Email:    "user@example.com",
		Password: "123456",
	})
	if !errors.Is(err, profileErr) {
		t.Fatalf("expected profileErr, got %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}
}

func TestAuthServiceLoginReturnsPasswordMismatch(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	service := newAuthService(
		&fakeUserRepository{
			findByEmailUser: &models.User{
				Email:        "user@example.com",
				PasswordHash: string(passwordHash),
			},
		},
		&fakeCitizenProfileRepository{},
	)

	user, token, refreshToken, err := service.Login(&dtos.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("expected ErrPasswordMismatch, got %v", err)
	}
	if user != nil || token != "" || refreshToken != "" {
		t.Fatalf("expected empty login result, got user=%#v token=%q refresh=%q", user, token, refreshToken)
	}
}
