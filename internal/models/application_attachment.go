package models

import "time"

type ApplicationAttachment struct {
	ID               string         `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ApplicationID    string         `json:"application_id" gorm:"type:uuid;not null;index"`
	Application      Application    `json:"application" gorm:"foreignKey:ApplicationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UploadedByUserID string         `json:"uploaded_by_user_id" gorm:"type:uuid;not null;index"`
	UploadedByUser   User           `json:"uploaded_by_user" gorm:"foreignKey:UploadedByUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	FileName         string         `json:"file_name" gorm:"type:varchar(255);not null"`
	FileURL          string         `json:"file_url" gorm:"type:text;not null"`
	FileType         string         `json:"file_type" gorm:"type:varchar(100)"`
	FileSize         *int64         `json:"file_size"`
	AttachmentType   AttachmentType `json:"attachment_type" gorm:"type:varchar(20);not null;default:'submitted';check:attachment_type IN ('submitted','supplement','result')"`
	CreatedAt        time.Time      `json:"created_at" gorm:"not null"`
	DeletedAt        *time.Time     `json:"deleted_at" gorm:"index"`
}
