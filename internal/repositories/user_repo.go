package repositories

import (
	"errors"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	Create(user *models.User) (*models.User, error)
}

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepository {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(user *models.User) (*models.User, error) {
	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}
