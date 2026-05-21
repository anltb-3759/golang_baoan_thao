package dtos

type AdminCreateUserRequest struct {
	Name     string `form:"name"     validate:"required,max=255"`
	Email    string `form:"email"    validate:"required,email"`
	Password string `form:"password" validate:"required,min=6,max=72"`
	Role     string `form:"role"     validate:"required,oneof=citizen staff manager super_admin"`
	Phone    string `form:"phone"    validate:"omitempty,numeric,min=10,max=11"`
	Address  string `form:"address"  validate:"omitempty,max=500"`
}

type AdminUpdateUserRequest struct {
	Name    string `form:"name"    validate:"required,max=255"`
	Role    string `form:"role"    validate:"required,oneof=citizen staff manager super_admin"`
	Phone   string `form:"phone"   validate:"omitempty,numeric,min=10,max=11"`
	Address string `form:"address" validate:"omitempty,max=500"`
}
