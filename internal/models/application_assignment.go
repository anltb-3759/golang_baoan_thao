package models

import "time"

type ApplicationAssignment struct {
	ID               string           `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ApplicationID    string           `json:"application_id" gorm:"type:uuid;not null;index"`
	Application      Application      `json:"application" gorm:"foreignKey:ApplicationID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	FromStaffUserID  *string          `json:"from_staff_user_id" gorm:"type:uuid;index"`
	FromStaffUser    *User            `json:"from_staff_user" gorm:"foreignKey:FromStaffUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ToStaffUserID    *string          `json:"to_staff_user_id" gorm:"type:uuid;index"`
	ToStaffUser      *User            `json:"to_staff_user" gorm:"foreignKey:ToStaffUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	AssignedByUserID *string          `json:"assigned_by_user_id" gorm:"type:uuid;index"`
	AssignedByUser   *User            `json:"assigned_by_user" gorm:"foreignKey:AssignedByUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Action           AssignmentAction `json:"action" gorm:"type:varchar(20);not null;check:action IN ('assigned','transferred','unassigned')"`
	Note             string           `json:"note" gorm:"type:text"`
	CreatedAt        time.Time        `json:"created_at" gorm:"not null"`
}
