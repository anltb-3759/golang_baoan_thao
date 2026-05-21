package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- mock ---

type mockServiceTypeRepo struct {
	mock.Mock
}

func (m *mockServiceTypeRepo) List(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ListResult), args.Error(1)
}

func (m *mockServiceTypeRepo) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceType), args.Error(1)
}

func (m *mockServiceTypeRepo) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceType), args.Error(1)
}

func (m *mockServiceTypeRepo) ListCategories(ctx context.Context) ([]models.Category, error) {
	return nil, nil
}

func (m *mockServiceTypeRepo) ListDepartments(ctx context.Context) ([]models.Department, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Department), args.Error(1)
}

func (m *mockServiceTypeRepo) Create(ctx context.Context, st *models.ServiceType) error {
	args := m.Called(ctx, st)
	return args.Error(0)
}

func (m *mockServiceTypeRepo) Update(ctx context.Context, st *models.ServiceType) error {
	args := m.Called(ctx, st)
	return args.Error(0)
}

func (m *mockServiceTypeRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockServiceTypeRepo) CountApplications(ctx context.Context, serviceTypeID string) (int64, error) {
	args := m.Called(ctx, serviceTypeID)
	return args.Get(0).(int64), args.Error(1)
}

// --- tests ---

func TestServiceCatalogService_List_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	filter := repositories.ListFilter{Page: 1, Limit: 10}
	expected := &repositories.ListResult{
		Items: []models.ServiceType{{Name: "Cấp CCCD lần đầu"}},
		Total: 1,
	}
	repo.On("List", mock.Anything, filter).Return(expected, nil)

	result, err := svc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_List_RepoError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	filter := repositories.ListFilter{Page: 1, Limit: 10}
	repo.On("List", mock.Anything, filter).Return(nil, errors.New("db error"))

	result, err := svc.List(context.Background(), filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_GetByID_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	id := "some-uuid"
	expected := &models.ServiceType{Name: "Cấp CCCD lần đầu"}
	repo.On("GetByID", mock.Anything, id).Return(expected, nil)

	result, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_GetByID_NotFound(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	repo.On("GetByID", mock.Anything, "bad-id").Return(nil, errors.New("record not found"))

	result, err := svc.GetByID(context.Background(), "bad-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Create_MapsDuplicateError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	serviceType := &models.ServiceType{Name: "Test"}
	repo.On("Create", mock.Anything, serviceType).Return(errors.New("duplicate key value violates unique constraint \"service_types_code_key\""))

	err := svc.Create(context.Background(), serviceType)

	assert.ErrorIs(t, err, services.ErrServiceTypeCodeExists)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Delete_BlocksWhenApplicationsExist(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	serviceType := &models.ServiceType{ID: "st1"}
	repo.On("GetByIDForAdmin", mock.Anything, "st1").Return(serviceType, nil)
	repo.On("CountApplications", mock.Anything, "st1").Return(int64(2), nil)

	err := svc.Delete(context.Background(), "st1")

	assert.ErrorIs(t, err, services.ErrServiceTypeHasApplications)
	repo.AssertExpectations(t)
}
