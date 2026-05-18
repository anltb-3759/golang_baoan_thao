package models

import "time"

type User struct {
	ID           string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string     `json:"name" gorm:"type:varchar(255);not null"`
	Email        string     `json:"email" gorm:"type:varchar(255);not null;uniqueIndex"`
	PasswordHash string     `json:"-" gorm:"type:varchar(255);not null"`
	Phone        string     `json:"phone" gorm:"type:varchar(50)"`
	Address      string     `json:"address" gorm:"type:text"`
	Role         UserRole   `json:"role" gorm:"type:varchar(20);not null;default:'citizen';index;check:role IN ('citizen','staff','manager','super_admin')"`
	Status       UserStatus `json:"status" gorm:"type:varchar(20);not null;default:'active';index;check:status IN ('active','blocked')"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt    *time.Time `json:"deleted_at" gorm:"index"`
	CreatedBy    *string    `json:"created_by" gorm:"type:uuid"`
	UpdatedBy    *string    `json:"updated_by" gorm:"type:uuid"`
	DeletedBy    *string    `json:"deleted_by" gorm:"type:uuid"`
}
