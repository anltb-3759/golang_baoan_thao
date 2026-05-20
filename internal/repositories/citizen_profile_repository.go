package repositories

import (
	"errors"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type CitizenProfileRepository interface {
	GetByUserID(userID string) (*models.CitizenProfile, error)
	Update(profile *models.CitizenProfile) error
	CreateInTx(tx *gorm.DB, profile *models.CitizenProfile) error
}

type citizenProfileRepo struct {
	db *gorm.DB
}

func NewCitizenProfileRepository(db *gorm.DB) CitizenProfileRepository {
	return &citizenProfileRepo{db: db}
}

func (r *citizenProfileRepo) GetByUserID(userID string) (*models.CitizenProfile, error) {
	var p models.CitizenProfile
	if err := r.db.Where("user_id = ?", userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *citizenProfileRepo) Update(profile *models.CitizenProfile) error {
	return r.db.Save(profile).Error
}

func (r *citizenProfileRepo) CreateInTx(tx *gorm.DB, profile *models.CitizenProfile) error {
	return tx.Create(profile).Error
}
