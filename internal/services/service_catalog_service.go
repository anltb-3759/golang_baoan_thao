package services

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

type ServiceCatalogService struct {
	repo repositories.ServiceTypeRepository
}

func NewServiceCatalogService(repo repositories.ServiceTypeRepository) *ServiceCatalogService {
	return &ServiceCatalogService{repo: repo}
}

func (s *ServiceCatalogService) List(filter repositories.ListFilter) (*repositories.ListResult, error) {
	return s.repo.List(filter)
}

func (s *ServiceCatalogService) GetByID(id string) (*models.ServiceType, error) {
	return s.repo.GetByID(id)
}
