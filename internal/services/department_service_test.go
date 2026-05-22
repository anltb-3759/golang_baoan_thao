package services

import (
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/stretchr/testify/assert"
)

// --- fakeDepartmentRepo ---

type fakeDepartmentRepo struct {
	dept      *models.Department
	depts     []models.Department
	total     int64
	findErr   error
	codeErr   error
	codeDept  *models.Department
	createErr error
	updateErr error
	deleteErr error
}

func (r *fakeDepartmentRepo) FindByID(_ string) (*models.Department, error) {
	return r.dept, r.findErr
}
func (r *fakeDepartmentRepo) FindByCode(_ string) (*models.Department, error) {
	return r.codeDept, r.codeErr
}
func (r *fakeDepartmentRepo) Create(d *models.Department) (*models.Department, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	return d, nil
}
func (r *fakeDepartmentRepo) Update(_ *models.Department) error { return r.updateErr }
func (r *fakeDepartmentRepo) List(_ repositories.DepartmentFilter, _, _ int) ([]models.Department, int64, error) {
	if r.findErr != nil {
		return nil, 0, r.findErr
	}
	return r.depts, r.total, nil
}
func (r *fakeDepartmentRepo) SoftDelete(_ string, _ string) error { return r.deleteErr }

var _ repositories.DepartmentRepository = (*fakeDepartmentRepo)(nil)

func newDeptSvc(repo *fakeDepartmentRepo) *DepartmentService {
	return NewDepartmentService(repo, nil)
}

// --- ListDepartments ---

func TestDepartmentService_ListDepartments_OK(t *testing.T) {
	depts := []models.Department{{ID: "1", Name: "IT"}}
	svc := newDeptSvc(&fakeDepartmentRepo{depts: depts, total: 1})
	result, total, err := svc.ListDepartments(repositories.DepartmentFilter{}, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestDepartmentService_ListDepartments_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	_, _, err := svc.ListDepartments(repositories.DepartmentFilter{}, 1, 10)
	assert.ErrorIs(t, err, repoErr)
}

// --- GetDepartment ---

func TestDepartmentService_GetDepartment_Found(t *testing.T) {
	d := &models.Department{ID: "d1", Name: "IT"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d})
	result, err := svc.GetDepartment("d1")
	assert.NoError(t, err)
	assert.Equal(t, "d1", result.ID)
}

func TestDepartmentService_GetDepartment_NotFound(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{dept: nil})
	_, err := svc.GetDepartment("missing")
	assert.ErrorIs(t, err, ErrDepartmentNotFound)
}

func TestDepartmentService_GetDepartment_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	_, err := svc.GetDepartment("x")
	assert.ErrorIs(t, err, repoErr)
}

// --- CreateDepartment ---

func TestDepartmentService_CreateDepartment_OK(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{codeDept: nil})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001", Address: "123 St"}
	dept, err := svc.CreateDepartment(req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "IT", dept.Name)
	assert.Equal(t, "IT001", dept.Code)
}

func TestDepartmentService_CreateDepartment_WithLeader(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{codeDept: nil})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001", LeaderUserID: "user-1"}
	dept, err := svc.CreateDepartment(req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "user-1", *dept.LeaderUserID)
}

func TestDepartmentService_CreateDepartment_CodeExists(t *testing.T) {
	existing := &models.Department{Code: "IT001"}
	svc := newDeptSvc(&fakeDepartmentRepo{codeDept: existing})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001"}
	_, err := svc.CreateDepartment(req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentCodeExists)
}

func TestDepartmentService_CreateDepartment_FindCodeError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{codeErr: repoErr})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001"}
	_, err := svc.CreateDepartment(req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestDepartmentService_CreateDepartment_CreateError(t *testing.T) {
	createErr := errors.New("create failed")
	svc := newDeptSvc(&fakeDepartmentRepo{createErr: createErr})
	req := &dtos.DepartmentCreateRequest{Name: "IT", Code: "IT001"}
	_, err := svc.CreateDepartment(req, "actor")
	assert.ErrorIs(t, err, createErr)
}

// --- UpdateDepartment ---

func TestDepartmentService_UpdateDepartment_OK(t *testing.T) {
	d := &models.Department{ID: "d1", Name: "Old", Code: "OLD"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d})
	req := &dtos.DepartmentUpdateRequest{Name: "New", Code: "OLD", Address: "New St"}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "New", result.Name)
}

func TestDepartmentService_UpdateDepartment_CodeChange_Unique(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, codeDept: nil})
	req := &dtos.DepartmentUpdateRequest{Name: "Dept", Code: "NEW"}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Equal(t, "NEW", result.Code)
}

func TestDepartmentService_UpdateDepartment_CodeChange_Conflict(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	existing := &models.Department{ID: "d2", Code: "NEW"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, codeDept: existing})
	req := &dtos.DepartmentUpdateRequest{Name: "Dept", Code: "NEW"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentCodeExists)
}

func TestDepartmentService_UpdateDepartment_NotFound(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{dept: nil})
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "X"}
	_, err := svc.UpdateDepartment("missing", req, "actor")
	assert.ErrorIs(t, err, ErrDepartmentNotFound)
}

func TestDepartmentService_UpdateDepartment_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "X"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestDepartmentService_UpdateDepartment_SaveError(t *testing.T) {
	d := &models.Department{ID: "d1", Code: "OLD"}
	saveErr := errors.New("save failed")
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, updateErr: saveErr})
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "OLD"}
	_, err := svc.UpdateDepartment("d1", req, "actor")
	assert.ErrorIs(t, err, saveErr)
}

func TestDepartmentService_UpdateDepartment_ClearLeader(t *testing.T) {
	leaderID := "user-1"
	d := &models.Department{ID: "d1", Code: "OLD", LeaderUserID: &leaderID}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d})
	req := &dtos.DepartmentUpdateRequest{Name: "X", Code: "OLD", LeaderUserID: ""}
	result, err := svc.UpdateDepartment("d1", req, "actor")
	assert.NoError(t, err)
	assert.Nil(t, result.LeaderUserID)
}

// --- DeleteDepartment ---

func TestDepartmentService_DeleteDepartment_OK(t *testing.T) {
	d := &models.Department{ID: "d1"}
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d})
	err := svc.DeleteDepartment("d1", "actor")
	assert.NoError(t, err)
}

func TestDepartmentService_DeleteDepartment_NotFound(t *testing.T) {
	svc := newDeptSvc(&fakeDepartmentRepo{dept: nil})
	err := svc.DeleteDepartment("missing", "actor")
	assert.ErrorIs(t, err, ErrDepartmentNotFound)
}

func TestDepartmentService_DeleteDepartment_FindError(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newDeptSvc(&fakeDepartmentRepo{findErr: repoErr})
	err := svc.DeleteDepartment("d1", "actor")
	assert.ErrorIs(t, err, repoErr)
}

func TestDepartmentService_DeleteDepartment_DeleteError(t *testing.T) {
	d := &models.Department{ID: "d1"}
	deleteErr := errors.New("delete error")
	svc := newDeptSvc(&fakeDepartmentRepo{dept: d, deleteErr: deleteErr})
	err := svc.DeleteDepartment("d1", "actor")
	assert.ErrorIs(t, err, deleteErr)
}
