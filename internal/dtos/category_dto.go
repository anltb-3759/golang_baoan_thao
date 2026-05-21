package dtos

type CategoryCreateRequest struct {
	Name        string `form:"name"        validate:"required,max=255"`
	Code        string `form:"code"        validate:"required,max=100"`
	Description string `form:"description" validate:"omitempty,max=1000"`
}

type CategoryUpdateRequest struct {
	Name        string `form:"name"        validate:"required,max=255"`
	Code        string `form:"code"        validate:"required,max=100"`
	Description string `form:"description" validate:"omitempty,max=1000"`
}
