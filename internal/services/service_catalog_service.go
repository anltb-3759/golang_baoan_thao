package services

import (
	"context"
	"errors"
	"strings"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrServiceTypeHasApplications = errors.New("service_type.applications_exist")
var ErrServiceTypeCodeExists = errors.New("service_type.code_duplicate")

type ServiceCatalogService struct {
	repo repositories.ServiceTypeRepository
}

func NewServiceCatalogService(repo repositories.ServiceTypeRepository) *ServiceCatalogService {
	return &ServiceCatalogService{repo: repo}
}

func (s *ServiceCatalogService) List(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error) {
	return s.repo.List(ctx, filter)
}
func (s *ServiceCatalogService) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ServiceCatalogService) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	return s.repo.GetByIDForAdmin(ctx, id)
}

func (s *ServiceCatalogService) ListDepartments(ctx context.Context) ([]models.Department, error) {
	return s.repo.ListDepartments(ctx)
}

func (s *ServiceCatalogService) Create(ctx context.Context, st *models.ServiceType) error {
	if err := s.repo.Create(ctx, st); err != nil {
		if isDuplicateServiceTypeError(err) {
			return ErrServiceTypeCodeExists
		}
		return err
	}
	return nil
}

func (s *ServiceCatalogService) Update(ctx context.Context, st *models.ServiceType) error {
	if err := s.repo.Update(ctx, st); err != nil {
		if isDuplicateServiceTypeError(err) {
			return ErrServiceTypeCodeExists
		}
		return err
	}
	return nil
}

func (s *ServiceCatalogService) Delete(ctx context.Context, id string) error {
	serviceType, err := s.repo.GetByIDForAdmin(ctx, id)
	if err != nil {
		return err
	}

	count, err := s.repo.CountApplications(ctx, serviceType.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrServiceTypeHasApplications
	}

	return s.repo.Delete(ctx, serviceType.ID)
}

func isDuplicateServiceTypeError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key value violates unique constraint") || strings.Contains(message, "service_types_code_key")
}
