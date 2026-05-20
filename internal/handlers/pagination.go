package handlers

import (
	"strconv"

	"github.com/labstack/echo/v5"
)

func parsePagination(c *echo.Context) (page, limit int) {
	page = parseIntParam(c, "page", 1)
	limit = parseIntParam(c, "limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return
}

func parseIntParam(c *echo.Context, key string, fallback int) int {
	v, err := strconv.Atoi(c.QueryParam(key))
	if err != nil {
		return fallback
	}
	return v
}
