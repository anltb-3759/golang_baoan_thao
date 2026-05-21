package handlers

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

type CategoryService interface {
	ListCategories(filter repositories.CategoryFilter, page, limit int) ([]models.Category, int64, error)
	GetCategory(id string) (*models.Category, error)
	CreateCategory(req *dtos.CategoryCreateRequest, createdBy string) (*models.Category, error)
	UpdateCategory(id string, req *dtos.CategoryUpdateRequest, updatedBy string) (*models.Category, error)
	DeleteCategory(id string, deletedBy string) error
}

type AdminCategoryHandler struct {
	svc CategoryService
}

func NewAdminCategoryHandler(svc CategoryService) *AdminCategoryHandler {
	return &AdminCategoryHandler{svc: svc}
}

func catFlashURL(flash, msg string) string {
	return "/admin/categories?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func (h *AdminCategoryHandler) ListCategories(c *echo.Context) error {
	search := c.QueryParam("search")
	page, limit := parsePagination(c)

	filter := repositories.CategoryFilter{Search: search}
	cats, total, err := h.svc.ListCategories(filter, page, limit)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.categories.title", nil),
		"CurrentPath": "/admin/categories",
		"CurrentUser": adminCurrentUser(c),
		"Categories":  cats,
		"Pagination":  newPagination(page, limit, total),
		"Search":      search,
		"Flash":       flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/categories/list.html", data)
}

func (h *AdminCategoryHandler) ShowCreateForm(c *echo.Context) error {
	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.categories.form.create_title", nil),
		"CurrentPath": "/admin/categories",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      false,
	}
	return c.Render(http.StatusOK, "admin/pages/categories/form.html", data)
}

func (h *AdminCategoryHandler) CreateCategory(c *echo.Context) error {
	req := new(dtos.CategoryCreateRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, extractFieldErrors(c, err), "")
	}

	_, err := h.svc.CreateCategory(req, actorID(c))
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrCategoryCodeExists) {
			msg = configs.T(c, "category.code_exists", nil)
		}
		return h.renderFormErrors(c, false, nil, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, catFlashURL("success", configs.T(c, "ui.msg.category_created", nil)))
}

func (h *AdminCategoryHandler) ShowEditForm(c *echo.Context) error {
	id := c.Param("id")
	cat, err := h.svc.GetCategory(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, catFlashURL("error", configs.T(c, "category.not_found", nil)))
	}

	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.categories.form.edit_title", nil),
		"CurrentPath": "/admin/categories",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      true,
		"Category":    cat,
	}
	return c.Render(http.StatusOK, "admin/pages/categories/form.html", data)
}

func (h *AdminCategoryHandler) UpdateCategory(c *echo.Context) error {
	id := c.Param("id")

	cat, err := h.svc.GetCategory(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, catFlashURL("error", configs.T(c, "category.not_found", nil)))
	}

	req := new(dtos.CategoryUpdateRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, true, cat, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, true, cat, req, extractFieldErrors(c, err), "")
	}

	if _, err := h.svc.UpdateCategory(id, req, actorID(c)); err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrCategoryCodeExists) {
			msg = configs.T(c, "category.code_exists", nil)
		}
		return h.renderFormErrors(c, true, cat, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, catFlashURL("success", configs.T(c, "ui.msg.category_updated", nil)))
}

func (h *AdminCategoryHandler) DeleteCategory(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteCategory(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, catFlashURL("error", configs.T(c, "ui.msg.category_delete_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, catFlashURL("success", configs.T(c, "ui.msg.category_deleted", nil)))
}

func (h *AdminCategoryHandler) renderFormErrors(c *echo.Context, isEdit bool, cat *models.Category, req interface{}, fieldErrors map[string]string, globalError string) error {
	titleKey := "ui.categories.form.create_title"
	if isEdit {
		titleKey = "ui.categories.form.edit_title"
	}
	data := map[string]interface{}{
		"Title":       configs.T(c, titleKey, nil),
		"CurrentPath": "/admin/categories",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      isEdit,
		"Category":    cat,
		"Req":         req,
		"FieldErrors": fieldErrors,
		"Error":       globalError,
	}
	return c.Render(http.StatusUnprocessableEntity, "admin/pages/categories/form.html", data)
}
