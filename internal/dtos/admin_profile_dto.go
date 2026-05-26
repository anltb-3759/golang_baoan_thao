package dtos

type UpdateAdminProfileRequest struct {
	Phone   string `form:"phone" validate:"omitempty"`
	Address string `form:"address" validate:"omitempty"`
}

type ChangeAdminPasswordRequest struct {
	CurrentPassword    string `form:"current_password" validate:"required,min=6,max=72"`
	NewPassword        string `form:"new_password" validate:"required,min=6,max=72"`
	ConfirmNewPassword string `form:"confirm_new_password" validate:"required,min=6,max=72"`
}
