package repositories

import (
	"errors"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type CategoryFilter struct {
	Search string
}

type CategoryRepository interface {
	FindByID(id string) (*models.Category, error)
	FindByCode(code string) (*models.Category, error)
	Create(cat *models.Category) (*models.Category, error)
	Update(cat *models.Category) error
	List(filter CategoryFilter, offset, limit int) ([]models.Category, int64, error)
	SoftDelete(id string, deletedBy string) error
}

type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) CategoryRepository {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) FindByID(id string) (*models.Category, error) {
	var cat models.Category
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&cat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) FindByCode(code string) (*models.Category, error) {
	var cat models.Category
	err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&cat).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) Create(cat *models.Category) (*models.Category, error) {
	if err := r.db.Create(cat).Error; err != nil {
		return nil, err
	}
	return cat, nil
}

func (r *CategoryRepo) Update(cat *models.Category) error {
	return r.db.Save(cat).Error
}

func (r *CategoryRepo) List(filter CategoryFilter, offset, limit int) ([]models.Category, int64, error) {
	q := r.db.Model(&models.Category{}).Where("deleted_at IS NULL")
	if filter.Search != "" {
		like := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(code) LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var cats []models.Category
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&cats).Error; err != nil {
		return nil, 0, err
	}
	return cats, total, nil
}

func (r *CategoryRepo) SoftDelete(id string, deletedBy string) error {
	now := time.Now()
	return r.db.Model(&models.Category{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
			"deleted_by": deletedBy,
		}).Error
}
