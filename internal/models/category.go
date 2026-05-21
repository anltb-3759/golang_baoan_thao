package models

import "time"

type Category struct {
	ID          string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Code        string     `json:"code" gorm:"type:varchar(100);not null;uniqueIndex"`
	Description string     `json:"description" gorm:"type:text"`
	IsActive    bool       `json:"is_active" gorm:"not null;default:true"`
	CreatedAt   time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt   *time.Time `json:"deleted_at" gorm:"index"`
	DeletedBy   *string    `json:"deleted_by" gorm:"type:uuid"`
}
