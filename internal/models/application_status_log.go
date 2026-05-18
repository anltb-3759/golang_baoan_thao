package models

import "time"

type ApplicationStatusLog struct {
	ID              string             `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ApplicationID   string             `json:"application_id" gorm:"type:uuid;not null;index"`
	Application     Application        `json:"application" gorm:"foreignKey:ApplicationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	OldStatus       *ApplicationStatus `json:"old_status" gorm:"type:varchar(30);check:old_status IN ('received','processing','need_more_info','approved','rejected')"`
	NewStatus       ApplicationStatus  `json:"new_status" gorm:"type:varchar(30);not null;check:new_status IN ('received','processing','need_more_info','approved','rejected')"`
	ChangedByUserID *string            `json:"changed_by_user_id" gorm:"type:uuid;index"`
	ChangedByUser   *User              `json:"changed_by_user" gorm:"foreignKey:ChangedByUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Note            string             `json:"note" gorm:"type:text"`
	CreatedAt       time.Time          `json:"created_at" gorm:"not null"`
}
