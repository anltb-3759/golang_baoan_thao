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
	"gorm.io/gorm"
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

func (m *mockServiceTypeRepo) CreateInTx(_ *gorm.DB, _ *models.ServiceType) error { return nil }

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

func TestServiceCatalogService_GetByIDForAdmin_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{Name: "Cấp CCCD"}
	repo.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	result, err := svc.GetByIDForAdmin(context.Background(), "st1")
	assert.NoError(t, err)
	assert.Equal(t, st, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_GetByIDForAdmin_Error(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	repo.On("GetByIDForAdmin", mock.Anything, "bad").Return(nil, errors.New("not found"))
	_, err := svc.GetByIDForAdmin(context.Background(), "bad")
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_ListDepartments_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	depts := []models.Department{{ID: "d1", Name: "IT"}}
	repo.On("ListDepartments", mock.Anything).Return(depts, nil)
	result, err := svc.ListDepartments(context.Background())
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_ListDepartments_Error(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	repo.On("ListDepartments", mock.Anything).Return(nil, errors.New("db error"))
	_, err := svc.ListDepartments(context.Background())
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_ListCategories_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	// ListCategories is not mocked — stub returns nil,nil
	result, err := svc.ListCategories(context.Background())
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestServiceCatalogService_Create_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{Name: "Test"}
	repo.On("Create", mock.Anything, st).Return(nil)
	err := svc.Create(context.Background(), st)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Create_GenericError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{Name: "Test"}
	repo.On("Create", mock.Anything, st).Return(errors.New("some other db error"))
	err := svc.Create(context.Background(), st)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, services.ErrServiceTypeCodeExists)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Update_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{Name: "Updated"}
	repo.On("Update", mock.Anything, st).Return(nil)
	err := svc.Update(context.Background(), st)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Update_MapsDuplicateError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{Name: "Test"}
	repo.On("Update", mock.Anything, st).Return(errors.New("duplicate key value violates unique constraint \"service_types_code_key\""))
	err := svc.Update(context.Background(), st)
	assert.ErrorIs(t, err, services.ErrServiceTypeCodeExists)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Update_GenericError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{Name: "Test"}
	repo.On("Update", mock.Anything, st).Return(errors.New("some other error"))
	err := svc.Update(context.Background(), st)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, services.ErrServiceTypeCodeExists)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Delete_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{ID: "st1"}
	repo.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	repo.On("CountApplications", mock.Anything, "st1").Return(int64(0), nil)
	repo.On("Delete", mock.Anything, "st1").Return(nil)
	err := svc.Delete(context.Background(), "st1")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Delete_GetByIDForAdminError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	repo.On("GetByIDForAdmin", mock.Anything, "bad").Return(nil, errors.New("not found"))
	err := svc.Delete(context.Background(), "bad")
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_Delete_CountApplicationsError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)
	st := &models.ServiceType{ID: "st1"}
	repo.On("GetByIDForAdmin", mock.Anything, "st1").Return(st, nil)
	repo.On("CountApplications", mock.Anything, "st1").Return(int64(0), errors.New("count error"))
	err := svc.Delete(context.Background(), "st1")
	assert.Error(t, err)
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

func TestServiceCatalogService_Update_WithID_PopulatesChanges(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	before := &models.ServiceType{ID: "st1", Name: "Old Name", Code: "OLD", IsActive: true}
	st := &models.ServiceType{ID: "st1", Name: "New Name", Code: "NEW", IsActive: false}

	repo.On("GetByIDForAdmin", mock.Anything, "st1").Return(before, nil)
	repo.On("Update", mock.Anything, st).Return(nil)

	err := svc.Update(context.Background(), st)
	assert.NoError(t, err)
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
