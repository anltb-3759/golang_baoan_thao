package repositories

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type ListFilter struct {
	Category string
	Search   string
	Page     int
	Limit    int
}

type ListResult struct {
	Items []models.ServiceType
	Total int64
}

type ServiceTypeRepository interface {
	List(filter ListFilter) (*ListResult, error)
	GetByID(id string) (*models.ServiceType, error)
}

type serviceTypeRepo struct {
	db *gorm.DB
}

func NewServiceTypeRepository(db *gorm.DB) ServiceTypeRepository {
	return &serviceTypeRepo{db: db}
}

func (r *serviceTypeRepo) List(filter ListFilter) (*ListResult, error) {
	q := r.db.Model(&models.ServiceType{}).
		Preload("ResponsibleDepartment").
		Where("is_active = ? AND deleted_at IS NULL", true)

	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Search != "" {
		q = q.Where("name ILIKE ?", "%"+filter.Search+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (filter.Page - 1) * filter.Limit
	items := make([]models.ServiceType, 0)
	if err := q.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, err
	}

	return &ListResult{Items: items, Total: total}, nil
}

func (r *serviceTypeRepo) GetByID(id string) (*models.ServiceType, error) {
	var st models.ServiceType
	err := r.db.Preload("ResponsibleDepartment").
		Where("id = ? AND is_active = ? AND deleted_at IS NULL", id, true).
		First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}
