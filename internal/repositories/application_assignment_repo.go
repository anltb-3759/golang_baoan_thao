package repositories

import (
	"errors"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type ApplicationAssignmentRepository interface {
	Create(a *models.ApplicationAssignment) (*models.ApplicationAssignment, error)
	ListByApplication(applicationID string) ([]models.ApplicationAssignment, error)
}

type applicationAssignmentRepo struct {
	db *gorm.DB
}

func NewApplicationAssignmentRepo(db *gorm.DB) ApplicationAssignmentRepository {
	return &applicationAssignmentRepo{db: db}
}

func (r *applicationAssignmentRepo) Create(a *models.ApplicationAssignment) (*models.ApplicationAssignment, error) {
	if err := r.db.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (r *applicationAssignmentRepo) ListByApplication(applicationID string) ([]models.ApplicationAssignment, error) {
	var items []models.ApplicationAssignment
	if err := r.db.Preload("FromStaffUser").Preload("ToStaffUser").Preload("AssignedByUser").Where("application_id = ?", applicationID).Order("created_at DESC").Find(&items).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return items, nil
}
