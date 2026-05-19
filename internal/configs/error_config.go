package configs

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	code := http.StatusInternalServerError
	var errorDetails []ErrorDetail

	var ve *ValidatorError
	if errors.As(err, &ve) {
		code = ve.Code
		errorDetails = translateValidationMessages(c, ve.Messages)
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		errorDetails = []ErrorDetail{translateHTTPMessage(c, he.Message)}
	} else {
		errorDetails = []ErrorDetail{
			{
				Code:    "common.internal_error",
				Message: T(c, "common.internal_error", nil),
			},
		}
	}

	errorResponse := map[string]interface{}{
		"errors": errorDetails,
		"code":   code,
	}

	if c.Request().Method == http.MethodHead {
		c.NoContent(code)
	} else {
		c.JSON(code, errorResponse)
	}
}

func translateHTTPMessage(c *echo.Context, message interface{}) ErrorDetail {
	key, ok := message.(string)
	if !ok {
		return ErrorDetail{
			Code:    "error.unknown",
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	return ErrorDetail{
		Code:    key,
		Message: T(c, key, nil),
	}
}

func translateValidationMessages(c *echo.Context, messages []ValidatorMessage) []ErrorDetail {
	translatedMessages := make([]ErrorDetail, 0, len(messages))
	for _, message := range messages {
		translatedMessages = append(translatedMessages, ErrorDetail{
			Field:   message.Field,
			Code:    message.Key,
			Message: T(c, message.Key, message.Params),
		})
	}

	return translatedMessages
}
