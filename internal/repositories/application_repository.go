package repositories

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type ApplicationRepository interface {
	ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error)
	GetByIDForCitizen(id, citizenUserID string) (*models.Application, error)
}

type applicationRepo struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepo{db: db}
}

func (r *applicationRepo) ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error) {
	q := r.db.Model(&models.Application{}).
		Where("citizen_user_id = ? AND deleted_at IS NULL", citizenUserID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := make([]models.Application, 0)
	if err := q.Preload("ServiceType").
		Order("submitted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *applicationRepo) GetByIDForCitizen(id, citizenUserID string) (*models.Application, error) {
	var app models.Application
	if err := r.db.Preload("ServiceType").
		Where("id = ? AND citizen_user_id = ? AND deleted_at IS NULL", id, citizenUserID).
		First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}
