package services

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrDepartmentNotFound = errors.New("department.not_found")
var ErrDepartmentCodeExists = errors.New("department.code_exists")

type DepartmentService struct {
	repo repositories.DepartmentRepository
}

func NewDepartmentService(repo repositories.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

func (s *DepartmentService) ListDepartments(filter repositories.DepartmentFilter, page, limit int) ([]models.Department, int64, error) {
	offset := (page - 1) * limit
	return s.repo.List(filter, offset, limit)
}

func (s *DepartmentService) GetDepartment(id string) (*models.Department, error) {
	dept, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, ErrDepartmentNotFound
	}
	return dept, nil
}

func (s *DepartmentService) CreateDepartment(req *dtos.DepartmentCreateRequest, createdBy string) (*models.Department, error) {
	existing, err := s.repo.FindByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDepartmentCodeExists
	}

	now := time.Now()
	dept := &models.Department{
		Name:      req.Name,
		Code:      req.Code,
		Address:   req.Address,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.LeaderUserID != "" {
		dept.LeaderUserID = &req.LeaderUserID
	}

	return s.repo.Create(dept)
}

func (s *DepartmentService) UpdateDepartment(id string, req *dtos.DepartmentUpdateRequest, updatedBy string) (*models.Department, error) {
	dept, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, ErrDepartmentNotFound
	}

	if req.Code != dept.Code {
		existing, err := s.repo.FindByCode(req.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrDepartmentCodeExists
		}
	}

	dept.Name = req.Name
	dept.Code = req.Code
	dept.Address = req.Address
	dept.UpdatedAt = time.Now()

	if req.LeaderUserID != "" {
		dept.LeaderUserID = &req.LeaderUserID
	} else {
		dept.LeaderUserID = nil
	}

	if err := s.repo.Update(dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *DepartmentService) DeleteDepartment(id string, deletedBy string) error {
	dept, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if dept == nil {
		return ErrDepartmentNotFound
	}
	return s.repo.SoftDelete(id, deletedBy)
}
