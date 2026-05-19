package models

import "time"

type ImportExportLog struct {
	ID           string             `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       *string            `json:"user_id" gorm:"type:uuid;index"`
	User         *User              `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	TableName    string             `json:"table_name" gorm:"type:varchar(255);not null;index"`
	Type         ImportExportType   `json:"type" gorm:"type:varchar(20);not null;check:type IN ('import','export')"`
	Status       ImportExportStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index;check:status IN ('pending','processing','success','failed')"`
	FileName     string             `json:"file_name" gorm:"type:varchar(255)"`
	FileURL      string             `json:"file_url" gorm:"type:text"`
	TotalRows    int                `json:"total_rows" gorm:"default:0"`
	SuccessRows  int                `json:"success_rows" gorm:"default:0"`
	FailedRows   int                `json:"failed_rows" gorm:"default:0"`
	ErrorMessage string             `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time          `json:"created_at" gorm:"not null;index"`
	CompletedAt  *time.Time         `json:"completed_at"`
}
