package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"gorm.io/gorm"
)

const (
	MaxTotalAttachmentBytes = 30 * 1024 * 1024 // 30 MiB — also used as server body limit
	maxAttachments          = 10
	maxFileSizeBytes        = 10 * 1024 * 1024 // 10 MiB per file
)

var (
	ErrApplicationNotFound   = errors.New("application.not_found")
	ErrServiceTypeNotFound   = errors.New("application.service_type_not_found")
	ErrServiceTypeInactive   = errors.New("application.service_type_inactive")
	ErrMissingRequiredField  = errors.New("application.missing_required_field")
	ErrInvalidSubmittedData  = errors.New("application.invalid_submitted_data")
	ErrInvalidFormSchema     = errors.New("application.invalid_form_schema")
	ErrAttachmentRequired    = errors.New("application.attachment_required")
	ErrSupplementNotAllowed  = errors.New("application.supplement_not_allowed")
	ErrTooManyAttachments    = errors.New("application.attachment_too_many")
	ErrAttachmentTooLarge    = errors.New("application.attachment_too_large")
	ErrAttachmentInvalidType = errors.New("application.attachment_invalid_type")
)

type ApplicationService struct {
	appRepo         repositories.ApplicationRepository
	serviceTypeRepo repositories.ServiceTypeRepository
	userRepo        repositories.UserRepository
	storage         utils.FileStorage
	mailer          Mailer
}

func NewApplicationService(
	appRepo repositories.ApplicationRepository,
	serviceTypeRepo repositories.ServiceTypeRepository,
	userRepo repositories.UserRepository,
	storage utils.FileStorage,
	mailer Mailer,
) *ApplicationService {
	return &ApplicationService{
		appRepo:         appRepo,
		serviceTypeRepo: serviceTypeRepo,
		userRepo:        userRepo,
		storage:         storage,
		mailer:          mailer,
	}
}

func (s *ApplicationService) SubmitApplication(
	citizenUserID string,
	req *dtos.SubmitApplicationRequest,
	files []*multipart.FileHeader,
) (*dtos.ApplicationResponse, error) {
	st, err := s.serviceTypeRepo.GetByID(context.Background(), req.ServiceTypeID)
	if err != nil {
		return nil, ErrServiceTypeNotFound
	}
	if !st.IsActive {
		return nil, ErrServiceTypeInactive
	}

	if err := validateSubmittedData(req.SubmittedData, st.FormSchema); err != nil {
		return nil, err
	}

	if err := validateAttachmentLimits(files); err != nil {
		return nil, err
	}

	now := time.Now()
	app := &models.Application{
		ApplicationCode: utils.GenerateApplicationCode(),
		CitizenUserID:   citizenUserID,
		ServiceTypeID:   req.ServiceTypeID,
		Status:          models.ApplicationStatusReceived,
		SubmittedData:   req.SubmittedData,
		SubmittedAt:     now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	tmpID := utils.GenerateUUID()
	app.ID = tmpID

	atts := make([]models.ApplicationAttachment, 0, len(files))

	savedURLs := make([]string, 0, len(files))
	for _, fh := range files {
		pubURL, mimeType, size, err := s.storage.SaveApplicationFile(tmpID, fh)
		if err != nil {
			_ = s.storage.RemoveApplicationDir(tmpID)
			if errors.Is(err, utils.ErrDisallowedMime) {
				return nil, ErrAttachmentInvalidType
			}
			return nil, fmt.Errorf("save attachment: %w", err)
		}
		savedURLs = append(savedURLs, pubURL)
		sz := size
		atts = append(atts, models.ApplicationAttachment{
			UploadedByUserID: citizenUserID,
			FileName:         fh.Filename,
			FileURL:          pubURL,
			FileType:         mimeType,
			FileSize:         &sz,
			AttachmentType:   models.AttachmentTypeSubmitted,
			CreatedAt:        now,
		})
	}

	loc := configs.DefaultLocale
	notifParams := map[string]string{"code": app.ApplicationCode, "service": st.Name}
	notif := &models.Notification{
		UserID:    citizenUserID,
		Title:     configs.TLang(loc, "notification.received.title", notifParams),
		Message:   configs.TLang(loc, "notification.received.message", notifParams),
		Type:      models.NotificationTypeReceived,
		CreatedAt: now,
	}

	if err := s.appRepo.CreateWithAttachments(app, atts, notif, utils.GenerateApplicationCode); err != nil {
		_ = s.storage.RemoveApplicationDir(tmpID)
		return nil, fmt.Errorf("create application: %w", err)
	}

	go func() {
		user, err := s.userRepo.FindByID(citizenUserID)
		if err != nil || user == nil {
			return
		}
		citizenName := user.Name
		if citizenName == "" {
			citizenName = configs.TLang("vi", "common.default_citizen_name", nil)
		}
		params := map[string]string{
			"name":         citizenName,
			"service":      st.Name,
			"code":         app.ApplicationCode,
			"submitted_at": app.SubmittedAt.Format("15:04 02/01/2006"),
			"from_name":    utils.EnvOr("SMTP_FROM_NAME", configs.TLang("vi", "common.app_name", nil)),
		}
		subject := configs.TLang("vi", "mail.application_received.subject", params)
		body := configs.TLang("vi", "mail.application_received.body", params)
		if err := s.mailer.Send(user.Email, subject, body); err != nil {
			log.Printf("send confirmation email failed: %v", err)
		}
	}()

	return toApplicationResponse(app, st, atts), nil
}

func (s *ApplicationService) ListMyApplications(userID string, page, limit int) ([]models.Application, int64, error) {
	return s.appRepo.ListByCitizen(userID, page, limit)
}

func (s *ApplicationService) GetMyApplication(userID, appID string) (*dtos.ApplicationResponse, error) {
	app, err := s.appRepo.GetByIDForCitizen(appID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("get application: %w", err)
	}
	return toApplicationResponseFromModel(app), nil
}

func (s *ApplicationService) ListMyApplicationStatusHistory(userID, appID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	if _, err := s.appRepo.GetByIDForCitizen(appID, userID); err != nil {
		return nil, 0, ErrApplicationNotFound
	}
	return s.appRepo.ListStatusLogsByCitizen(appID, userID, page, limit, since)
}

func (s *ApplicationService) UploadMyApplicationSupplements(userID, appID string, files []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error) {
	if len(files) == 0 {
		return nil, ErrAttachmentRequired
	}

	if err := validateAttachmentLimits(files); err != nil {
		return nil, err
	}

	app, err := s.appRepo.GetByIDForCitizen(appID, userID)
	if err != nil {
		return nil, ErrApplicationNotFound
	}

	if app.Status != models.ApplicationStatusProcessing && app.Status != models.ApplicationStatusNeedMoreInfo {
		return nil, ErrSupplementNotAllowed
	}

	now := time.Now()
	atts := make([]models.ApplicationAttachment, 0, len(files))
	savedURLs := make([]string, 0, len(files))
	for _, fh := range files {
		pubURL, mimeType, size, saveErr := s.storage.SaveApplicationFile(app.ID, fh)
		if saveErr != nil {
			for _, u := range savedURLs {
				_ = s.storage.RemoveFile(u)
			}
			if errors.Is(saveErr, utils.ErrDisallowedMime) {
				return nil, ErrAttachmentInvalidType
			}
			return nil, fmt.Errorf("save supplement attachment: %w", saveErr)
		}
		savedURLs = append(savedURLs, pubURL)
		sz := size
		atts = append(atts, models.ApplicationAttachment{
			UploadedByUserID: userID,
			FileName:         fh.Filename,
			FileURL:          pubURL,
			FileType:         mimeType,
			FileSize:         &sz,
			AttachmentType:   models.AttachmentTypeSupplement,
			CreatedAt:        now,
		})
	}

	if err := s.appRepo.CreateAttachments(app.ID, atts); err != nil {
		for _, a := range atts {
			_ = s.storage.RemoveFile(a.FileURL)
		}
		return nil, fmt.Errorf("create supplement attachments: %w", err)
	}

	resp := make([]dtos.ApplicationAttachmentResponse, 0, len(atts))
	for _, a := range atts {
		resp = append(resp, dtos.ApplicationAttachmentResponse{
			ID:       a.ID,
			FileName: a.FileName,
			FileURL:  a.FileURL,
			FileType: a.FileType,
			FileSize: a.FileSize,
		})
	}

	return resp, nil
}

func validateSubmittedData(data json.RawMessage, schema json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}

	var s struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(schema, &s); err != nil {
		return ErrInvalidFormSchema
	}
	if len(s.Required) == 0 {
		return nil
	}

	var submitted map[string]interface{}
	if err := json.Unmarshal(data, &submitted); err != nil {
		return ErrInvalidSubmittedData
	}
	for _, field := range s.Required {
		v, ok := submitted[field]
		if !ok || v == nil {
			return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
		}
		if str, ok := v.(string); ok {
			if strings.TrimSpace(str) == "" {
				return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
			}
		}
	}
	return nil
}

func validateAttachmentLimits(files []*multipart.FileHeader) error {
	if len(files) > maxAttachments {
		return ErrTooManyAttachments
	}
	var total int64
	for _, f := range files {
		if f.Size > maxFileSizeBytes {
			return ErrAttachmentTooLarge
		}
		total += f.Size
	}
	if total > MaxTotalAttachmentBytes {
		return ErrAttachmentTooLarge
	}
	return nil
}

func toApplicationResponse(app *models.Application, st *models.ServiceType, atts []models.ApplicationAttachment) *dtos.ApplicationResponse {
	attResponses := make([]dtos.ApplicationAttachmentResponse, 0, len(atts))
	for _, a := range atts {
		attResponses = append(attResponses, dtos.ApplicationAttachmentResponse{
			ID:       a.ID,
			FileName: a.FileName,
			FileURL:  a.FileURL,
			FileType: a.FileType,
			FileSize: a.FileSize,
		})
	}
	return &dtos.ApplicationResponse{
		ID:              app.ID,
		ApplicationCode: app.ApplicationCode,
		ServiceTypeID:   app.ServiceTypeID,
		ServiceTypeName: st.Name,
		Status:          string(app.Status),
		SubmittedData:   app.SubmittedData,
		SubmittedAt:     app.SubmittedAt,
		Attachments:     attResponses,
	}
}

func toApplicationResponseFromModel(app *models.Application) *dtos.ApplicationResponse {
	attResponses := make([]dtos.ApplicationAttachmentResponse, 0)
	for _, a := range app.ApplicationAttachments {
		attResponses = append(attResponses, dtos.ApplicationAttachmentResponse{
			ID:       a.ID,
			FileName: a.FileName,
			FileURL:  a.FileURL,
			FileType: a.FileType,
			FileSize: a.FileSize,
		})
	}
	return &dtos.ApplicationResponse{
		ID:              app.ID,
		ApplicationCode: app.ApplicationCode,
		ServiceTypeID:   app.ServiceTypeID,
		ServiceTypeName: app.ServiceType.Name,
		Status:          string(app.Status),
		SubmittedData:   app.SubmittedData,
		SubmittedAt:     app.SubmittedAt,
		Attachments:     attResponses,
	}
}
