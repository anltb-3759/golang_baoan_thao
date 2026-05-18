package configs

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	code := http.StatusInternalServerError
	var message interface{}

	var ve *ValidatorError
	if errors.As(err, &ve) {
		code = ve.Code
		message = ve.Messages
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	} else {
		message = "Internal Server Error"
	}

	errorResponse := map[string]interface{}{
		"errors": []interface{}{message},
		"code":   code,
	}

	if c.Request().Method == http.MethodHead {
		c.NoContent(code)
	} else {
		c.JSON(code, errorResponse)
	}
}
