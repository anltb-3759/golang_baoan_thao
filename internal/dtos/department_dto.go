package dtos

type DepartmentCreateRequest struct {
	Name         string `form:"name"    validate:"required,max=255"`
	Code         string `form:"code"    validate:"required,max=100"`
	Address      string `form:"address" validate:"omitempty,max=500"`
	LeaderUserID string `form:"leader_user_id"`
}

type DepartmentUpdateRequest struct {
	Name         string `form:"name"    validate:"required,max=255"`
	Code         string `form:"code"    validate:"required,max=100"`
	Address      string `form:"address" validate:"omitempty,max=500"`
	LeaderUserID string `form:"leader_user_id"`
}
