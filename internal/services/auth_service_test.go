package services

import (
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"golang.org/x/crypto/bcrypt"
)

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

func TestAuthServiceRegisterCreatesCitizenUser(t *testing.T) {
	repo := &fakeUserRepository{}
	service := NewAuthService(repo)

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
	if repo.createdUser == nil {
		t.Fatal("expected user to be persisted")
	}
	if repo.createdUser.Role != models.UserRoleCitizen {
		t.Fatalf("expected citizen role, got %q", repo.createdUser.Role)
	}
	if repo.createdUser.Status != models.UserStatusActive {
		t.Fatalf("expected active status, got %q", repo.createdUser.Status)
	}
	if repo.createdUser.PasswordHash == "123456" {
		t.Fatal("expected password to be hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.createdUser.PasswordHash), []byte("123456")); err != nil {
		t.Fatalf("expected password hash to match original password: %v", err)
	}
}

func TestAuthServiceRegisterReturnsEmailExists(t *testing.T) {
	service := NewAuthService(&fakeUserRepository{
		findByEmailUser: &models.User{Email: "user@example.com"},
	})

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
	service := NewAuthService(&fakeUserRepository{
		findByEmailErr: repoErr,
	})

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

func TestAuthServiceLoginReturnsTokens(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	service := NewAuthService(&fakeUserRepository{
		findByEmailUser: &models.User{
			ID:           "user-id",
			Email:        "user@example.com",
			PasswordHash: string(passwordHash),
			Role:         models.UserRoleCitizen,
		},
	})

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
	service := NewAuthService(&fakeUserRepository{})

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

func TestAuthServiceLoginReturnsPasswordMismatch(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	service := NewAuthService(&fakeUserRepository{
		findByEmailUser: &models.User{
			Email:        "user@example.com",
			PasswordHash: string(passwordHash),
		},
	})

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
