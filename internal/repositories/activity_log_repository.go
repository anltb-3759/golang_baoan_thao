package repositories

import (
	"errors"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

const defaultActivityLogListLimit = 20

type ActivityLogFilter struct {
	Action     string
	EntityType string
	Result     string
	ActorUserID string
}

type ActivityLogRepository interface {
	Create(log *models.ActivityLog) error
	List(filter ActivityLogFilter, offset, limit int) ([]models.ActivityLog, int64, error)
	DeleteBefore(before time.Time) (int64, error)
	DeleteRetainDays(days int) (int64, error)
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) Create(log *models.ActivityLog) error {
	return r.db.Create(log).Error
}

func (r *activityLogRepository) List(filter ActivityLogFilter, offset, limit int) ([]models.ActivityLog, int64, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = defaultActivityLogListLimit
	}

	q := r.db.Model(&models.ActivityLog{})
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.EntityType != "" {
		q = q.Where("entity_type = ?", filter.EntityType)
	}
	if filter.Result != "" {
		q = q.Where("result = ?", filter.Result)
	}
	if filter.ActorUserID != "" {
		q = q.Where("actor_user_id = ?", filter.ActorUserID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []models.ActivityLog
	if err := q.Preload("ActorUser").Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *activityLogRepository) DeleteBefore(before time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", before).Delete(&models.ActivityLog{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (r *activityLogRepository) DeleteRetainDays(days int) (int64, error) {
	if days <= 0 {
		return 0, errors.New("retain days must be greater than 0")
	}

	before := time.Now().AddDate(0, 0, -days)
	return r.DeleteBefore(before)
}
