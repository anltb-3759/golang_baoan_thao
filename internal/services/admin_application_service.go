package services

import (
	"errors"
	"mime/multipart"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
)

var ErrAdminApplicationNotFound = errors.New("application.not_found")
var ErrAdminApplicationInvalidTransition = errors.New("application.invalid_transition")
var ErrAdminApplicationRejectReasonRequired = errors.New("application.reject_reason_required")

type AdminApplicationService struct {
	appRepo       repositories.ApplicationRepository
	assignService *ApplicationAssignmentService
	storage       utils.FileStorage
}

func NewAdminApplicationService(appRepo repositories.ApplicationRepository, assignService *ApplicationAssignmentService, storage utils.FileStorage) *AdminApplicationService {
	return &AdminApplicationService{appRepo: appRepo, assignService: assignService, storage: storage}
}

func (s *AdminApplicationService) ListApplications(filter repositories.ApplicationFilter, page, limit int) ([]models.Application, int64, error) {
	return s.appRepo.AdminList(filter, page, limit)
}

func (s *AdminApplicationService) GetApplication(id string) (*models.Application, error) {
	return s.appRepo.GetByID(id)
}

func (s *AdminApplicationService) AssignToStaff(applicationID string, toStaffUserID *string, assignedBy string) error {
	if err := s.assignService.AssignApplicationToStaff(applicationID, toStaffUserID, assignedBy); err != nil {
		return err
	}
	return nil
}

func (s *AdminApplicationService) ProcessApplication(applicationID string, newStatus models.ApplicationStatus, note string, files []*multipart.FileHeader, processedBy string) error {
	app, err := s.appRepo.GetByID(applicationID)
	if err != nil || app == nil {
		return ErrAdminApplicationNotFound
	}

	if !isAllowedAdminTransition(app.Status, newStatus) {
		return ErrAdminApplicationInvalidTransition
	}
	if newStatus == models.ApplicationStatusRejected && strings.TrimSpace(note) == "" {
		return ErrAdminApplicationRejectReasonRequired
	}

	now := time.Now()
	processingStartedAt := app.ProcessingStartedAt
	completedAt := app.CompletedAt
	if newStatus == models.ApplicationStatusProcessing && processingStartedAt == nil {
		processingStartedAt = &now
	}
	if newStatus == models.ApplicationStatusApproved || newStatus == models.ApplicationStatusRejected {
		completedAt = &now
	}

	resultNote := note
	rejectedReason := ""
	if newStatus == models.ApplicationStatusRejected {
		rejectedReason = note
		resultNote = ""
	}

	attachments := make([]models.ApplicationAttachment, 0, len(files))
	savedURLs := make([]string, 0, len(files))
	if len(files) > 0 {
		if s.storage == nil {
			return errors.New("storage not configured")
		}
		for _, fh := range files {
			pubURL, mimeType, size, saveErr := s.storage.SaveApplicationFile(app.ID, fh)
			if saveErr != nil {
				for _, u := range savedURLs {
					_ = s.storage.RemoveFile(u)
				}
				if err := s.storage.RemoveApplicationDir(app.ID); err != nil {
					_ = err
				}
				return saveErr
			}
			savedURLs = append(savedURLs, pubURL)
			sz := size
			attachments = append(attachments, models.ApplicationAttachment{
				UploadedByUserID: processedBy,
				FileName:         fh.Filename,
				FileURL:          pubURL,
				FileType:         mimeType,
				FileSize:         &sz,
				AttachmentType:   models.AttachmentTypeResult,
				CreatedAt:        now,
			})
		}
	}

	if err := s.appRepo.ProcessStatusUpdate(app.ID, &app.Status, newStatus, resultNote, rejectedReason, processingStartedAt, completedAt, processedBy, attachments); err != nil {
		if s.storage != nil {
			for _, a := range attachments {
				_ = s.storage.RemoveFile(a.FileURL)
			}
		}
		return err
	}

	return nil
}

func isAllowedAdminTransition(current, next models.ApplicationStatus) bool {
	switch current {
	case models.ApplicationStatusReceived:
		return next == models.ApplicationStatusProcessing
	case models.ApplicationStatusProcessing:
		return next == models.ApplicationStatusApproved || next == models.ApplicationStatusRejected
	default:
		return false
	}
}
