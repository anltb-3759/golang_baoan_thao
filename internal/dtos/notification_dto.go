package dtos

import (
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
)

type NotificationResponse struct {
	ID            string     `json:"id"`
	ApplicationID *string    `json:"application_id,omitempty"`
	Title         string     `json:"title"`
	Message       string     `json:"message"`
	Type          string     `json:"type"`
	IsRead        bool       `json:"is_read"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func NewNotificationResponse(n *models.Notification) NotificationResponse {
	return NotificationResponse{
		ID:            n.ID,
		ApplicationID: n.ApplicationID,
		Title:         n.Title,
		Message:       n.Message,
		Type:          string(n.Type),
		IsRead:        n.IsRead,
		ReadAt:        n.ReadAt,
		CreatedAt:     n.CreatedAt,
	}
}
