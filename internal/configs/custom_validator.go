package configs

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	Validator *validator.Validate
}

type ValidatorError struct {
	Code     int
	Messages map[string]string
}

func (ve *ValidatorError) Error() string {
	return "validation error"
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				switch e.Tag() {
				case "required":
					errorMessages[e.Field()] = fmt.Sprintf("%s is required", e.Field())
				case "min":
					errorMessages[e.Field()] = fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param())
				default:
					errorMessages[e.Field()] = fmt.Sprintf("%s is invalid (%s)", e.Field(), e.Tag())
				}
			}
			return &ValidatorError{
				Code:     http.StatusBadRequest,
				Messages: errorMessages,
			}
		}
		return err
	}
	return nil
}
