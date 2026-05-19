package models

import (
	"encoding/json"
	"time"
)

type ServiceType struct {
	ID                      string          `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name                    string          `json:"name" gorm:"type:varchar(255);not null;index"`
	Code                    string          `json:"code" gorm:"type:varchar(100);not null;uniqueIndex"`
	Category                ServiceCategory `json:"category" gorm:"type:varchar(50);not null;default:'hanh_chinh_cong';index"`
	Description             string          `json:"description" gorm:"type:text"`
	RequiredDocuments       string          `json:"required_documents" gorm:"type:text"`
	FormSchema              json.RawMessage `json:"form_schema" gorm:"type:jsonb;not null"`
	ProcessingTime          *int            `json:"processing_time"`
	Fee                     float64         `json:"fee" gorm:"type:numeric(12,2);not null;default:0"`
	ResponsibleDepartmentID *string         `json:"responsible_department_id" gorm:"type:uuid;index"`
	ResponsibleDepartment   *Department     `json:"responsible_department" gorm:"foreignKey:ResponsibleDepartmentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	IsActive                bool            `json:"is_active" gorm:"not null;default:true"`
	CreatedAt               time.Time       `json:"created_at" gorm:"not null"`
	UpdatedAt               time.Time       `json:"updated_at" gorm:"not null"`
	DeletedAt               *time.Time      `json:"deleted_at" gorm:"index"`
}
