package repositories

import (
	"errors"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type DepartmentFilter struct {
	Search string
}

type DepartmentRepository interface {
	FindByID(id string) (*models.Department, error)
	FindByCode(code string) (*models.Department, error)
	Create(dept *models.Department) (*models.Department, error)
	CreateInTx(tx *gorm.DB, dept *models.Department) error
	Update(dept *models.Department) error
	List(filter DepartmentFilter, offset, limit int) ([]models.Department, int64, error)
	SoftDelete(id string, deletedBy string) error
}

type DepartmentRepo struct {
	db *gorm.DB
}

func NewDepartmentRepo(db *gorm.DB) DepartmentRepository {
	return &DepartmentRepo{db: db}
}

func (r *DepartmentRepo) FindByID(id string) (*models.Department, error) {
	var dept models.Department
	err := r.db.Preload("LeaderUser").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&dept).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dept, nil
}

func (r *DepartmentRepo) FindByCode(code string) (*models.Department, error) {
	var dept models.Department
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&dept).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &dept, nil
}

func (r *DepartmentRepo) Create(dept *models.Department) (*models.Department, error) {
	if err := r.db.Create(dept).Error; err != nil {
		return nil, err
	}
	return dept, nil
}

func (r *DepartmentRepo) Update(dept *models.Department) error {
	return r.db.Save(dept).Error
}

func (r *DepartmentRepo) List(filter DepartmentFilter, offset, limit int) ([]models.Department, int64, error) {
	q := r.db.Model(&models.Department{}).Preload("LeaderUser").Where("deleted_at IS NULL")
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var depts []models.Department
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&depts).Error; err != nil {
		return nil, 0, err
	}
	return depts, total, nil
}

func (r *DepartmentRepo) CreateInTx(tx *gorm.DB, dept *models.Department) error {
	return tx.Create(dept).Error
}

func (r *DepartmentRepo) SoftDelete(id string, deletedBy string) error {
	now := time.Now()
	return r.db.Model(&models.Department{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
			"deleted_by": deletedBy,
		}).Error
}
