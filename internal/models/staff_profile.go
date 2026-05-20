package models

import "time"

type StaffProfile struct {
	ID           string      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       string      `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	User         User        `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	DepartmentID *string     `json:"department_id" gorm:"type:uuid;index"`
	Department   *Department `json:"department" gorm:"foreignKey:DepartmentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Position     string      `json:"position" gorm:"type:varchar(255)"`
	CreatedAt    time.Time   `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time   `json:"updated_at" gorm:"not null"`
	DeletedAt    *time.Time  `json:"deleted_at" gorm:"index"`
}
