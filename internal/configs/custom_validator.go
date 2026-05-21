package configs

import (
	"net/http"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	Validator *validator.Validate
}

type ValidatorError struct {
	Code     int
	Messages []ValidatorMessage
}

type ValidatorMessage struct {
	Field  string
	Key    string
	Params map[string]string
}

func (ve *ValidatorError) Error() string {
	return "validation error"
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make([]ValidatorMessage, 0, len(validationErrors))
			for _, e := range validationErrors {
				switch e.Tag() {
				case "required":
					errorMessages = append(errorMessages, ValidatorMessage{
						Field: e.Field(),
						Key:   "validation.required",
					})
				case "email":
					errorMessages = append(errorMessages, ValidatorMessage{
						Field: e.Field(),
						Key:   "validation.email",
					})
				case "min":
					errorMessages = append(errorMessages, ValidatorMessage{
						Field:  e.Field(),
						Key:    "validation.min",
						Params: map[string]string{"param": e.Param()},
					})
				case "max":
					errorMessages = append(errorMessages, ValidatorMessage{
						Field:  e.Field(),
						Key:    "validation.max",
						Params: map[string]string{"param": e.Param()},
					})
				case "len":
					errorMessages = append(errorMessages, ValidatorMessage{
						Field:  e.Field(),
						Key:    "validation.len",
						Params: map[string]string{"param": e.Param()},
					})
				case "numeric":
					errorMessages = append(errorMessages, ValidatorMessage{
						Field: e.Field(),
						Key:   "validation.numeric",
					})
				case "oneof":
					errorMessages = append(errorMessages, ValidatorMessage{
						Field: e.Field(),
						Key:   "validation.invalid",
					})
				default:
					errorMessages = append(errorMessages, ValidatorMessage{
						Field: e.Field(),
						Key:   "validation.invalid",
					})
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
