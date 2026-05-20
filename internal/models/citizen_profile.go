package models

import "time"

type CitizenProfile struct {
	ID                       string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID                   string     `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	User                     User       `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CitizenIDNumber          string     `json:"citizen_id_number" gorm:"type:varchar(12);not null;uniqueIndex"`
	DateOfBirth              *time.Time `json:"date_of_birth" gorm:"type:date"`
	Gender                   string     `json:"gender" gorm:"type:varchar(50)"`
	PermanentAddress         string     `json:"permanent_address" gorm:"type:text"`
	EmailNotificationEnabled bool       `json:"email_notification_enabled" gorm:"not null;default:true"`
	CreatedAt                time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt                time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt                *time.Time `json:"deleted_at" gorm:"index"`
}
