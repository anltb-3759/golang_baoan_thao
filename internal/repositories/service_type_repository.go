package repositories

import (
	"context"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type ListFilter struct {
	Category        string
	Search          string
	Page            int
	Limit           int
	IncludeInactive bool
}

type ListResult struct {
	Items []models.ServiceType
	Total int64
}

type ServiceTypeRepository interface {
	List(ctx context.Context, filter ListFilter) (*ListResult, error)
	GetByID(ctx context.Context, id string) (*models.ServiceType, error)
	GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error)
	ListDepartments(ctx context.Context) ([]models.Department, error)
	ListCategories(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, st *models.ServiceType) error
	Update(ctx context.Context, st *models.ServiceType) error
	Delete(ctx context.Context, id string) error
	CountApplications(ctx context.Context, serviceTypeID string) (int64, error)
}

type serviceTypeRepo struct {
	db *gorm.DB
}

func NewServiceTypeRepository(db *gorm.DB) ServiceTypeRepository {
	return &serviceTypeRepo{db: db}
}

func (r *serviceTypeRepo) List(ctx context.Context, filter ListFilter) (*ListResult, error) {
	q := r.db.WithContext(ctx).Model(&models.ServiceType{}).
		Preload("ResponsibleDepartment").
		Preload("Category")

	if filter.IncludeInactive {
		q = q.Where("deleted_at IS NULL")
	} else {
		q = q.Where("is_active = ? AND deleted_at IS NULL", true)
	}

	if filter.Category != "" {
		q = q.Where("category_id = ?", filter.Category)
	}
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("(name ILIKE ? OR code ILIKE ?)", like, like)
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

func (r *serviceTypeRepo) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	return r.getByID(ctx, id, true)
}

func (r *serviceTypeRepo) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	return r.getByID(ctx, id, false)
}

func (r *serviceTypeRepo) getByID(ctx context.Context, id string, activeOnly bool) (*models.ServiceType, error) {
	var st models.ServiceType
	query := r.db.WithContext(ctx).Preload("ResponsibleDepartment").Preload("Category").
		Where("id = ? AND deleted_at IS NULL", id)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.
		First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *serviceTypeRepo) ListDepartments(ctx context.Context) ([]models.Department, error) {
	departments := make([]models.Department, 0)
	if err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&departments).Error; err != nil {
		return nil, err
	}
	return departments, nil
}

func (r *serviceTypeRepo) ListCategories(ctx context.Context) ([]models.Category, error) {
	cats := make([]models.Category, 0)
	if err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND is_active = ?", true).
		Order("name ASC").
		Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}

func (r *serviceTypeRepo) Create(ctx context.Context, st *models.ServiceType) error {
	return r.db.WithContext(ctx).Create(st).Error
}

func (r *serviceTypeRepo) Update(ctx context.Context, st *models.ServiceType) error {
	return r.db.WithContext(ctx).
		Model(&models.ServiceType{}).
		Where("id = ? AND deleted_at IS NULL", st.ID).
		Updates(map[string]any{
			"name":                      st.Name,
			"code":                      st.Code,
			"category_id":               st.CategoryID,
			"description":               st.Description,
			"required_documents":        st.RequiredDocuments,
			"form_schema":               st.FormSchema,
			"processing_time":           st.ProcessingTime,
			"fee":                       st.Fee,
			"responsible_department_id": st.ResponsibleDepartmentID,
			"is_active":                 st.IsActive,
			"updated_at":                st.UpdatedAt,
		}).Error
}

func (r *serviceTypeRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.ServiceType{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", time.Now()).Error
}

func (r *serviceTypeRepo) CountApplications(ctx context.Context, serviceTypeID string) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("service_type_id = ? AND deleted_at IS NULL", serviceTypeID).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
