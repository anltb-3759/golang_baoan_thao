package repositories

import (
	"errors"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"gorm.io/gorm"
)

const codeRetryAttempts = 3

type ApplicationRepository interface {
	CreateWithAttachments(app *models.Application, atts []models.ApplicationAttachment, notif *models.Notification, codeGen func() string) error
	GetByID(id string) (*models.Application, error)
	AdminList(page, limit int) ([]models.Application, int64, error)
	ListByCitizen(citizenUserID string, page, limit int) ([]models.Application, int64, error)
	GetByIDForCitizen(id, citizenUserID string) (*models.Application, error)
	ListStatusLogsByCitizen(appID, citizenUserID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error)
	CreateAttachments(appID string, atts []models.ApplicationAttachment) error
	UpdateAssignedStaff(applicationID string, assignedStaffUserID *string, updatedBy string) error
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

func (r *applicationRepo) GetByID(id string) (*models.Application, error) {
	var app models.Application
	if err := r.db.Preload("CitizenUser").Preload("ServiceType").Preload("AssignedStaffUser").Preload("ApplicationAttachments").Where("id = ? AND deleted_at IS NULL", id).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *applicationRepo) AdminList(page, limit int) ([]models.Application, int64, error) {
	q := r.db.Model(&models.Application{}).Where("deleted_at IS NULL")

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := make([]models.Application, 0)
	if err := q.Preload("ServiceType").Preload("CitizenUser").Preload("AssignedStaffUser").Order("submitted_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
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

func (r *applicationRepo) ListStatusLogsByCitizen(appID, citizenUserID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	q := r.db.Model(&models.ApplicationStatusLog{}).
		Joins("JOIN applications ON applications.id = application_status_logs.application_id").
		Where("application_status_logs.application_id = ? AND applications.citizen_user_id = ? AND applications.deleted_at IS NULL", appID, citizenUserID)

	if since != nil {
		q = q.Where("application_status_logs.created_at > ?", *since)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	items := make([]models.ApplicationStatusLog, 0)
	if err := q.Preload("ChangedByUser").
		Order("application_status_logs.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *applicationRepo) CreateAttachments(appID string, atts []models.ApplicationAttachment) error {
	if len(atts) == 0 {
		return nil
	}
	for i := range atts {
		atts[i].ApplicationID = appID
	}
	return r.db.Create(&atts).Error
}

func (r *applicationRepo) UpdateAssignedStaff(applicationID string, assignedStaffUserID *string, updatedBy string) error {
	return r.db.Model(&models.Application{}).
		Where("id = ? AND deleted_at IS NULL", applicationID).
		Updates(map[string]interface{}{
			"assigned_staff_user_id": assignedStaffUserID,
			"updated_at":             time.Now(),
		}).Error
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
