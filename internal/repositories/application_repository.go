package repositories

import (
	"errors"
	"strings"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"gorm.io/gorm"
)

const codeRetryAttempts = 3

type ApplicationRepository interface {
	CreateWithAttachments(app *models.Application, atts []models.ApplicationAttachment, notif *models.Notification, codeGen func() string) error
	ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error)
	GetByIDForCitizen(id, citizenUserID string) (*models.Application, error)
}

type applicationRepo struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepo{db: db}
}

func (r *applicationRepo) CreateWithAttachments(
	app *models.Application,
	atts []models.ApplicationAttachment,
	notif *models.Notification,
	codeGen func() string,
) error {
	var lastErr error
	for attempt := 0; attempt < codeRetryAttempts; attempt++ {
		err := r.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(app).Error; err != nil {
				return err
			}
			for i := range atts {
				atts[i].ApplicationID = app.ID
			}
			if len(atts) > 0 {
				if err := tx.Create(&atts).Error; err != nil {
					return err
				}
			}
			notif.ApplicationID = &app.ID
			return tx.Create(notif).Error
		})
		if err == nil {
			return nil
		}
		lastErr = err
		if !isApplicationCodeConflict(err) {
			return err
		}
		app.ID = ""
		app.ApplicationCode = codeGen()
		for i := range atts {
			atts[i].ID = ""
			atts[i].ApplicationID = ""
		}
		notif.ID = ""
		notif.ApplicationID = nil
	}
	return lastErr
}

func (r *applicationRepo) ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error) {
	q := r.db.Model(&models.Application{}).
		Where("citizen_user_id = ? AND deleted_at IS NULL", citizenUserID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := make([]models.Application, 0)
	if err := q.Preload("ServiceType").
		Order("submitted_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *applicationRepo) GetByIDForCitizen(id, citizenUserID string) (*models.Application, error) {
	var app models.Application
	if err := r.db.Preload("ServiceType").Preload("ApplicationAttachments").
		Where("id = ? AND citizen_user_id = ? AND deleted_at IS NULL", id, citizenUserID).
		First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func isApplicationCodeConflict(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "applications_application_code") || strings.Contains(msg, "application_code")
}

var _ func() string = utils.GenerateApplicationCode
