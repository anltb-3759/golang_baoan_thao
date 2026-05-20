package models

import "time"

type Department struct {
	ID           string     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name         string     `json:"name" gorm:"type:varchar(255);not null"`
	Code         string     `json:"code" gorm:"type:varchar(100);not null;uniqueIndex"`
	Address      string     `json:"address" gorm:"type:text"`
	LeaderUserID *string    `json:"leader_user_id" gorm:"type:uuid"`
	LeaderUser   *User      `json:"leader_user" gorm:"foreignKey:LeaderUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"not null"`
	DeletedAt    *time.Time `json:"deleted_at" gorm:"index"`
}
