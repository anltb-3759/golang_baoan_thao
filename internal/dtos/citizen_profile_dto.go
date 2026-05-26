package dtos

import (
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
)

type ChangeMyPasswordRequest struct {
	CurrentPassword    string `json:"current_password" validate:"required,min=6,max=72"`
	NewPassword        string `json:"new_password" validate:"required,min=6,max=72"`
	ConfirmNewPassword string `json:"confirm_new_password" validate:"required,min=6,max=72"`
}


type UpdateCitizenProfileRequest struct {
	Name                     *string    `json:"name" validate:"omitempty,min=1"`
	Phone                    *string    `json:"phone"`
	Address                  *string    `json:"address"`
	Gender                   *string    `json:"gender"`
	PermanentAddress         *string    `json:"permanent_address"`
	DateOfBirth              *time.Time `json:"date_of_birth"`
	EmailNotificationEnabled *bool      `json:"email_notification_enabled"`
}

type CitizenProfileResponse struct {
	UserID                   string     `json:"user_id"`
	Name                     string     `json:"name"`
	Email                    string     `json:"email"`
	Phone                    string     `json:"phone"`
	Address                  string     `json:"address"`
	CitizenIDNumber          string     `json:"citizen_id_number"`
	DateOfBirth              *time.Time `json:"date_of_birth"`
	Gender                   string     `json:"gender"`
	PermanentAddress         string     `json:"permanent_address"`
	EmailNotificationEnabled bool       `json:"email_notification_enabled"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

func NewCitizenProfileResponse(user *models.User, profile *models.CitizenProfile) *CitizenProfileResponse {
	return &CitizenProfileResponse{
		UserID:                   user.ID,
		Name:                     user.Name,
		Email:                    user.Email,
		Phone:                    user.Phone,
		Address:                  user.Address,
		CitizenIDNumber:          profile.CitizenIDNumber,
		DateOfBirth:              profile.DateOfBirth,
		Gender:                   profile.Gender,
		PermanentAddress:         profile.PermanentAddress,
		EmailNotificationEnabled: profile.EmailNotificationEnabled,
		CreatedAt:                profile.CreatedAt,
		UpdatedAt:                profile.UpdatedAt,
	}
}
