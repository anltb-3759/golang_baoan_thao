package services

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// --- fakes ---

type fakeAppServiceTypeRepo struct {
	st  *models.ServiceType
	err error
}

func (r *fakeAppServiceTypeRepo) List(_ repositories.ListFilter) (*repositories.ListResult, error) {
	return nil, nil
}
func (r *fakeAppServiceTypeRepo) GetByID(_ string) (*models.ServiceType, error) {
	return r.st, r.err
}

type fakeAppUserRepo struct {
	user *models.User
	err  error
}

func (r *fakeAppUserRepo) FindByEmail(_ string) (*models.User, error)  { return nil, nil }
func (r *fakeAppUserRepo) FindByID(_ string) (*models.User, error)     { return r.user, r.err }
func (r *fakeAppUserRepo) Create(u *models.User) (*models.User, error) { return u, nil }
func (r *fakeAppUserRepo) CreateInTx(_ *gorm.DB, _ *models.User) error { return nil }
func (r *fakeAppUserRepo) Update(_ *models.User) error                 { return nil }
func (r *fakeAppUserRepo) List(_ repositories.UserFilter, _, _ int) ([]models.User, int64, error) {
	return nil, 0, nil
}
func (r *fakeAppUserRepo) UpdateStatus(_ string, _ models.UserStatus, _ string) error { return nil }
func (r *fakeAppUserRepo) SoftDelete(_ string, _ string) error                        { return nil }

type fakeAppRepo struct {
	createErr error
	apps      []models.Application
	total     int64
	listErr   error
	app       *models.Application
	getErr    error
}

func (r *fakeAppRepo) CreateWithAttachments(_ *models.Application, _ []models.ApplicationAttachment, _ *models.Notification, _ func() string) error {
	return r.createErr
}
func (r *fakeAppRepo) ListByCitizen(_ string, _, _ int) ([]models.Application, int64, error) {
	return r.apps, r.total, r.listErr
}
func (r *fakeAppRepo) GetByIDForCitizen(_, _ string) (*models.Application, error) {
	return r.app, r.getErr
}

type fakeStorage struct {
	pubURL  string
	mime    string
	size    int64
	saveErr error
}

func (s *fakeStorage) SaveApplicationFile(_ string, fh *multipart.FileHeader) (string, string, int64, error) {
	if s.saveErr != nil {
		return "", "", 0, s.saveErr
	}
	return s.pubURL, s.mime, s.size, nil
}
func (s *fakeStorage) RemoveApplicationDir(_ string) error { return nil }

type fakeMailer struct{ sent bool }

func (m *fakeMailer) Send(_, _, _ string) error { m.sent = true; return nil }

// compile-time interface checks
var _ repositories.ServiceTypeRepository = (*fakeAppServiceTypeRepo)(nil)
var _ repositories.UserRepository = (*fakeAppUserRepo)(nil)
var _ repositories.ApplicationRepository = (*fakeAppRepo)(nil)
var _ utils.FileStorage = (*fakeStorage)(nil)
var _ Mailer = (*fakeMailer)(nil)

// --- helpers ---

func makeSchema(required []string) json.RawMessage {
	b, _ := json.Marshal(map[string]interface{}{"required": required})
	return b
}

func makeData(fields map[string]string) json.RawMessage {
	b, _ := json.Marshal(fields)
	return b
}

func newSvc(
	appRepo *fakeAppRepo,
	stRepo *fakeAppServiceTypeRepo,
	uRepo *fakeAppUserRepo,
	storage *fakeStorage,
	mailer *fakeMailer,
) *ApplicationService {
	return NewApplicationService(appRepo, stRepo, uRepo, storage, mailer)
}

func activeServiceType() *models.ServiceType {
	return &models.ServiceType{
		ID:         "st-1",
		Name:       "Cấp CCCD lần đầu",
		IsActive:   true,
		FormSchema: makeSchema([]string{"full_name", "date_of_birth"}),
	}
}

func validReq() *dtos.SubmitApplicationRequest {
	return &dtos.SubmitApplicationRequest{
		ServiceTypeID: "st-1",
		SubmittedData: makeData(map[string]string{
			"full_name":    "Nguyen Van A",
			"date_of_birth": "1995-01-01",
		}),
	}
}

// --- SubmitApplication ---

func TestSubmitApplication_Success_NoFiles(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com", Name: "An"}},
		&fakeStorage{pubURL: "/uploads/f.pdf", mime: "application/pdf", size: 100},
		&fakeMailer{},
	)

	resp, err := svc.SubmitApplication("u1", validReq(), nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ApplicationCode)
	assert.Equal(t, "received", resp.Status)
}

func TestSubmitApplication_ServiceTypeNotFound(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{err: errors.New("not found")},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.ErrorIs(t, err, ErrServiceTypeNotFound)
}

func TestSubmitApplication_ServiceTypeInactive(t *testing.T) {
	st := activeServiceType()
	st.IsActive = false
	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: st},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.ErrorIs(t, err, ErrServiceTypeInactive)
}

func TestSubmitApplication_MissingRequiredField(t *testing.T) {
	st := activeServiceType()
	st.FormSchema = makeSchema([]string{"full_name", "date_of_birth"})

	req := &dtos.SubmitApplicationRequest{
		ServiceTypeID: "st-1",
		SubmittedData: makeData(map[string]string{"full_name": "An"}), // missing date_of_birth
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: st},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", req, nil)
	assert.ErrorIs(t, err, ErrMissingRequiredField)
}

func TestSubmitApplication_TooManyAttachments(t *testing.T) {
	files := make([]*multipart.FileHeader, maxAttachments+1)
	for i := range files {
		files[i] = &multipart.FileHeader{Size: 1024}
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrTooManyAttachments)
}

func TestSubmitApplication_FileTooLarge(t *testing.T) {
	files := []*multipart.FileHeader{
		{Size: maxFileSizeBytes + 1},
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrAttachmentTooLarge)
}

func TestSubmitApplication_TotalSizeTooLarge(t *testing.T) {
	// 3 files × 11 MiB each = 33 MiB > 30 MiB limit
	const elevenMiB = 11 * 1024 * 1024
	files := []*multipart.FileHeader{
		{Size: elevenMiB},
		{Size: elevenMiB},
		{Size: elevenMiB},
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrAttachmentTooLarge)
}

func TestSubmitApplication_InvalidMimeType(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "malware.exe", Size: 1024}}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{saveErr: utils.ErrDisallowedMime},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.ErrorIs(t, err, ErrAttachmentInvalidType)
}

func TestSubmitApplication_StorageError(t *testing.T) {
	files := []*multipart.FileHeader{{Filename: "doc.pdf", Size: 1024}}
	storageErr := errors.New("disk full")

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{saveErr: storageErr},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), files)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrAttachmentInvalidType))
}

func TestSubmitApplication_RepoCreateError(t *testing.T) {
	repoErr := errors.New("db constraint violation")
	svc := newSvc(
		&fakeAppRepo{createErr: repoErr},
		&fakeAppServiceTypeRepo{st: activeServiceType()},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.SubmitApplication("u1", validReq(), nil)
	assert.Error(t, err)
}

func TestSubmitApplication_NoRequiredFieldsInSchema(t *testing.T) {
	st := activeServiceType()
	st.FormSchema = makeSchema(nil) // no required fields

	req := &dtos.SubmitApplicationRequest{
		ServiceTypeID: "st-1",
		SubmittedData: makeData(map[string]string{}),
	}

	svc := newSvc(
		&fakeAppRepo{},
		&fakeAppServiceTypeRepo{st: st},
		&fakeAppUserRepo{user: &models.User{ID: "u1", Email: "a@b.com"}},
		&fakeStorage{},
		&fakeMailer{},
	)

	resp, err := svc.SubmitApplication("u1", req, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// --- ListMyApplications ---

func TestAppService_ListMyApplications_Success(t *testing.T) {
	apps := []models.Application{
		{ID: "app1", ApplicationCode: "APP-20240101-ABCDEF", SubmittedAt: time.Now()},
	}
	svc := newSvc(
		&fakeAppRepo{apps: apps, total: 1},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	result, total, err := svc.ListMyApplications("u1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
}

func TestAppService_ListMyApplications_Error(t *testing.T) {
	repoErr := errors.New("db error")
	svc := newSvc(
		&fakeAppRepo{listErr: repoErr},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, _, err := svc.ListMyApplications("u1", 1, 10)
	assert.ErrorIs(t, err, repoErr)
}

// --- GetMyApplication ---

func TestAppService_GetMyApplication_Success(t *testing.T) {
	app := &models.Application{
		ID:              "app1",
		ApplicationCode: "APP-20240101-ABCDEF",
		ServiceType:     models.ServiceType{Name: "Cấp CCCD"},
		Status:          models.ApplicationStatusReceived,
		SubmittedAt:     time.Now(),
	}
	svc := newSvc(
		&fakeAppRepo{app: app},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	resp, err := svc.GetMyApplication("u1", "app1")
	assert.NoError(t, err)
	assert.Equal(t, "app1", resp.ID)
	assert.Equal(t, "APP-20240101-ABCDEF", resp.ApplicationCode)
	assert.Equal(t, "Cấp CCCD", resp.ServiceTypeName)
}

func TestAppService_GetMyApplication_NotFound(t *testing.T) {
	svc := newSvc(
		&fakeAppRepo{getErr: gorm.ErrRecordNotFound},
		&fakeAppServiceTypeRepo{},
		&fakeAppUserRepo{},
		&fakeStorage{},
		&fakeMailer{},
	)

	_, err := svc.GetMyApplication("u1", "missing")
	assert.ErrorIs(t, err, ErrApplicationNotFound)
}

// --- validateSubmittedData ---

func TestValidateSubmittedData_EmptySchema(t *testing.T) {
	err := validateSubmittedData(makeData(nil), json.RawMessage{})
	assert.NoError(t, err)
}

func TestValidateSubmittedData_AllPresent(t *testing.T) {
	schema := makeSchema([]string{"name", "dob"})
	data := makeData(map[string]string{"name": "An", "dob": "1990-01-01"})
	assert.NoError(t, validateSubmittedData(data, schema))
}

func TestValidateSubmittedData_MissingField(t *testing.T) {
	schema := makeSchema([]string{"name", "dob"})
	data := makeData(map[string]string{"name": "An"})
	err := validateSubmittedData(data, schema)
	assert.ErrorIs(t, err, ErrMissingRequiredField)
}

func TestValidateSubmittedData_EmptyStringField(t *testing.T) {
	schema := makeSchema([]string{"name"})
	data := makeData(map[string]string{"name": ""})
	err := validateSubmittedData(data, schema)
	assert.ErrorIs(t, err, ErrMissingRequiredField)
}

func TestValidateSubmittedData_InvalidJSON(t *testing.T) {
	schema := makeSchema([]string{"name"})
	err := validateSubmittedData(json.RawMessage(`{bad json`), schema)
	assert.ErrorIs(t, err, ErrMissingRequiredField)
}
