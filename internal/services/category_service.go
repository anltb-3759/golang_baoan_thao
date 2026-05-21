package services

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/jackc/pgx/v5/pgconn"
)

func isCategoryUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

var ErrCategoryNotFound = errors.New("category.not_found")
var ErrCategoryCodeExists = errors.New("category.code_exists")

type CategoryService struct {
	repo repositories.CategoryRepository
}

func NewCategoryService(repo repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) ListCategories(filter repositories.CategoryFilter, page, limit int) ([]models.Category, int64, error) {
	offset := (page - 1) * limit
	return s.repo.List(filter, offset, limit)
}

func (s *CategoryService) GetCategory(id string) (*models.Category, error) {
	cat, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}
	return cat, nil
}

func (s *CategoryService) CreateCategory(req *dtos.CategoryCreateRequest, createdBy string) (*models.Category, error) {
	existing, err := s.repo.FindByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrCategoryCodeExists
	}

	now := time.Now()
	cat := &models.Category{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	created, err := s.repo.Create(cat)
	if err != nil {
		if isCategoryUniqueViolation(err) {
			return nil, ErrCategoryCodeExists
		}
		return nil, err
	}
	return created, nil
}

func (s *CategoryService) UpdateCategory(id string, req *dtos.CategoryUpdateRequest, updatedBy string) (*models.Category, error) {
	cat, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, ErrCategoryNotFound
	}

	if req.Code != cat.Code {
		existing, err := s.repo.FindByCode(req.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrCategoryCodeExists
		}
	}

	cat.Name = req.Name
	cat.Code = req.Code
	cat.Description = req.Description
	cat.IsActive = req.IsActive
	cat.UpdatedAt = time.Now()

	if err := s.repo.Update(cat); err != nil {
		if isCategoryUniqueViolation(err) {
			return nil, ErrCategoryCodeExists
		}
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) DeleteCategory(id string, deletedBy string) error {
	cat, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if cat == nil {
		return ErrCategoryNotFound
	}
	return s.repo.SoftDelete(id, deletedBy)
}
