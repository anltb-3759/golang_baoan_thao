package services

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var (
	ErrActivityLogCleanupModeInvalid = errors.New("exactly one cleanup mode must be provided")
	ErrActivityLogRetainDaysInvalid  = errors.New("retain days must be greater than 0")
	ErrActivityLogInvalidInput       = errors.New("activity log input is required")
)

type ActivityLogService struct {
	repo repositories.ActivityLogRepository
}

func NewActivityLogService(repo repositories.ActivityLogRepository) *ActivityLogService {
	return &ActivityLogService{repo: repo}
}

func (s *ActivityLogService) Log(log *models.ActivityLog) error {
	if log == nil {
		return ErrActivityLogInvalidInput
	}
	return s.repo.Create(log)
}

func (s *ActivityLogService) List(filter repositories.ActivityLogFilter, page, limit int) ([]models.ActivityLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.List(filter, offset, limit)
}

func (s *ActivityLogService) Cleanup(retainDays *int, deleteBefore *time.Time) (int64, error) {
	if (retainDays == nil && deleteBefore == nil) || (retainDays != nil && deleteBefore != nil) {
		return 0, ErrActivityLogCleanupModeInvalid
	}

	if retainDays != nil {
		if *retainDays <= 0 {
			return 0, ErrActivityLogRetainDaysInvalid
		}
		return s.repo.DeleteRetainDays(*retainDays)
	}

	return s.repo.DeleteBefore(*deleteBefore)
}
