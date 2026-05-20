package services_test

import (
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

func (m *mockServiceTypeRepo) List(filter repositories.ListFilter) (*repositories.ListResult, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ListResult), args.Error(1)
}

func (m *mockServiceTypeRepo) GetByID(id string) (*models.ServiceType, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ServiceType), args.Error(1)
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
	repo.On("List", filter).Return(expected, nil)

	result, err := svc.List(filter)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_List_RepoError(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	filter := repositories.ListFilter{Page: 1, Limit: 10}
	repo.On("List", filter).Return(nil, errors.New("db error"))

	result, err := svc.List(filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_GetByID_Success(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	id := "some-uuid"
	expected := &models.ServiceType{Name: "Cấp CCCD lần đầu"}
	repo.On("GetByID", id).Return(expected, nil)

	result, err := svc.GetByID(id)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestServiceCatalogService_GetByID_NotFound(t *testing.T) {
	repo := new(mockServiceTypeRepo)
	svc := services.NewServiceCatalogService(repo)

	repo.On("GetByID", "bad-id").Return(nil, errors.New("record not found"))

	result, err := svc.GetByID("bad-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}
