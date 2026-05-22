package repositories

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type StaffProfileRepository interface {
	FindByUserID(userID string) (*models.StaffProfile, error)
	ListByDepartment(deptID string, offset, limit int) ([]models.StaffProfile, int64, error)
	UpdateDepartment(userID string, deptID *string, updatedBy string) error
	Create(profile *models.StaffProfile) (*models.StaffProfile, error)
}

type staffProfileRepo struct {
	db *gorm.DB
}

func NewStaffProfileRepo(db *gorm.DB) StaffProfileRepository {
	return &staffProfileRepo{db: db}
}

func (r *staffProfileRepo) FindByUserID(userID string) (*models.StaffProfile, error) {
	var p models.StaffProfile
	if err := r.db.Preload("User").Preload("Department").Where("user_id = ? AND deleted_at IS NULL", userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *staffProfileRepo) ListByDepartment(deptID string, offset, limit int) ([]models.StaffProfile, int64, error) {
	q := r.db.Model(&models.StaffProfile{}).Preload("User").Preload("Department").Where("deleted_at IS NULL")
	if deptID != "" {
		q = q.Where("department_id = ?", deptID)
	} else {
		q = q.Where("department_id IS NULL")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.StaffProfile
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *staffProfileRepo) UpdateDepartment(userID string, deptID *string, updatedBy string) error {
	return r.db.Model(&models.StaffProfile{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]interface{}{
			"department_id": deptID,
			"updated_at":    time.Now(),
		}).Error
}

func (r *staffProfileRepo) Create(profile *models.StaffProfile) (*models.StaffProfile, error) {
	if err := r.db.Create(profile).Error; err != nil {
		return nil, err
	}
	return profile, nil
}
