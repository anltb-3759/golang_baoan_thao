package dtos

import (
	"encoding/json"
	"time"
)

type SubmitApplicationRequest struct {
	ServiceTypeID string          `json:"service_type_id" validate:"required,uuid"`
	SubmittedData json.RawMessage `json:"submitted_data" validate:"required"`
}

type ApplicationAttachmentResponse struct {
	ID             string `json:"id"`
	FileName       string `json:"file_name"`
	FileURL        string `json:"file_url"`
	FileType       string `json:"file_type"`
	FileSize       *int64 `json:"file_size"`
	AttachmentType string `json:"attachment_type"`
	CreatedAt      time.Time `json:"created_at"`
}

type ApplicationResponse struct {
	ID              string                          `json:"id"`
	ApplicationCode string                          `json:"application_code"`
	ServiceTypeID   string                          `json:"service_type_id"`
	ServiceTypeName string                          `json:"service_type_name"`
	Status          string                          `json:"status"`
	SubmittedData   json.RawMessage                 `json:"submitted_data"`
	SubmittedAt     time.Time                       `json:"submitted_at"`
	Attachments     []ApplicationAttachmentResponse `json:"attachments,omitempty"`
}

type ApplicationListItem struct {
	ID              string    `json:"id"`
	ApplicationCode string    `json:"application_code"`
	ServiceTypeID   string    `json:"service_type_id"`
	ServiceTypeName string    `json:"service_type_name"`
	Status          string    `json:"status"`
	SubmittedAt     time.Time `json:"submitted_at"`
}
