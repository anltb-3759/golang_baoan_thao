package models

import (
	"encoding/json"
	"time"
)

type ActivityLog struct {
	ID          string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ActorUserID *string   `json:"actor_user_id" gorm:"type:uuid;index"`
	ActorUser   *User     `json:"actor_user" gorm:"foreignKey:ActorUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Action      string    `json:"action" gorm:"type:varchar(255);not null"`
	EntityType  string    `json:"entity_type" gorm:"type:varchar(255);index"`
	EntityID    *string   `json:"entity_id" gorm:"type:uuid;index"`
	Description string    `json:"description" gorm:"type:text"`
	MetadataJSON json.RawMessage `json:"metadata_json" gorm:"type:jsonb"`
	Result       string          `json:"result" gorm:"type:varchar(20);index"`
	IPAddress   string    `json:"ip_address" gorm:"type:varchar(100)"`
	UserAgent   string    `json:"user_agent" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null;index"`
}
