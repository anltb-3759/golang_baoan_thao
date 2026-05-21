package dtos

type ServiceTypeFormRequest struct {
	Name                    string `form:"name" json:"name" validate:"required"`
	Code                    string `form:"code" json:"code" validate:"required"`
	Category                string `form:"category" json:"category" validate:"required"`
	Description             string `form:"description" json:"description"`
	RequiredDocuments       string `form:"required_documents" json:"required_documents"`
	FormSchema              string `form:"form_schema" json:"form_schema"`
	ProcessingTime          string `form:"processing_time" json:"processing_time"`
	Fee                     string `form:"fee" json:"fee"`
	ResponsibleDepartmentID string `form:"responsible_department_id" json:"responsible_department_id"`
	IsActive                bool   `form:"is_active" json:"is_active"`
}
