package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type citizenNotificationSvc interface {
	List(userID string, filter repositories.NotificationFilter, page, limit int) ([]dtos.NotificationResponse, int64, error)
	MarkAsRead(id, userID string) error
	MarkAllAsRead(userID string) error
	CountUnread(userID string) (int64, error)
}

type citizenAppSvc interface {
	SubmitApplication(citizenUserID string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error)
	ListMyApplications(userID string, page, limit int) ([]models.Application, int64, error)
	GetMyApplication(userID, appID string) (*dtos.ApplicationResponse, error)
	ListMyApplicationStatusHistory(userID, appID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error)
	UploadMyApplicationSupplements(userID, appID string, files []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error)
}

type CitizenWebHandler struct {
	authService     AuthService
	logger          ActivityLogger
	notificationSvc citizenNotificationSvc
	appSvc          citizenAppSvc
	catalogSvc      serviceCatalogSvc
}

func NewCitizenWebHandler(authService AuthService) *CitizenWebHandler {
	return &CitizenWebHandler{authService: authService}
}

func (h *CitizenWebHandler) WithActivityLogger(logger ActivityLogger) *CitizenWebHandler {
	h.logger = logger
	return h
}

func (h *CitizenWebHandler) WithNotificationService(svc citizenNotificationSvc) *CitizenWebHandler {
	h.notificationSvc = svc
	return h
}

func (h *CitizenWebHandler) WithCatalogService(svc serviceCatalogSvc) *CitizenWebHandler {
	h.catalogSvc = svc
	return h
}

func (h *CitizenWebHandler) WithApplicationService(svc citizenAppSvc) *CitizenWebHandler {
	h.appSvc = svc
	return h
}

func (h *CitizenWebHandler) unreadCount(userID string) int64 {
	if h.notificationSvc == nil || userID == "" {
		return 0
	}
	n, err := h.notificationSvc.CountUnread(userID)
	if err != nil {
		return 0
	}
	return n
}

func (h *CitizenWebHandler) writeActivityLog(log *models.ActivityLog) {
	if h.logger == nil || log == nil {
		return
	}
	_ = h.logger.Log(log)
}

func citizenCurrentUser(c *echo.Context) *configs.JwtCustomClaims {
	v := c.Get("user")
	if v == nil {
		return nil
	}
	claims, _ := v.(*configs.JwtCustomClaims)
	return claims
}

func (h *CitizenWebHandler) csrfToken(c *echo.Context) string {
	v, _ := c.Get("csrf").(string)
	return v
}

func flashURL(path, kind, msg string) string {
	return path + "?" + url.Values{"flash": {kind}, "msg": {msg}}.Encode()
}

func (h *CitizenWebHandler) ShowLoginPage(c *echo.Context) error {
	data := map[string]interface{}{
		"FullPage": true,
		"Title":    configs.T(c, "ui.form.citizen_login", nil),
		"Flash":    flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "citizen/pages/auth/login.html", data)
}

func (h *CitizenWebHandler) WebLogin(c *echo.Context) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := c.FormValue("password")

	reqData := &dtos.LoginRequest{Email: email, Password: password}

	user, _, refreshToken, err := h.authService.Login(reqData)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			msg = configs.T(c, "auth.user_not_found", nil)
		case errors.Is(err, services.ErrPasswordMismatch):
			msg = configs.T(c, "auth.password_mismatch", nil)
		case errors.Is(err, services.ErrUserBlocked):
			msg = configs.T(c, "auth.user_blocked", nil)
		default:
			msg = configs.T(c, "common.internal_error", nil)
		}
		return c.Render(http.StatusUnauthorized, "citizen/pages/auth/login.html", map[string]interface{}{
			"Error":    msg,
			"Email":    email,
			"FullPage": true,
			"Title":    configs.T(c, "ui.form.citizen_login", nil),
		})
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	if user.Role != models.UserRoleCitizen {
		return c.Render(http.StatusUnauthorized, "citizen/pages/auth/login.html", map[string]interface{}{
			"Error":    configs.T(c, "auth.user_not_found", nil),
			"Email":    email,
			"FullPage": true,
			"Title":    configs.T(c, "ui.form.citizen_login", nil),
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   AGE_REFRESH_TOKEN,
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: &user.ID,
		Action:      "auth.login",
		EntityType:  "user",
		EntityID:    &user.ID,
		Description: "Citizen web login successful",
		Result:      "success",
	})

	return c.Redirect(http.StatusSeeOther, "/citizen")
}

func (h *CitizenWebHandler) ShowRegisterPage(c *echo.Context) error {
	return c.Render(http.StatusOK, "citizen/pages/auth/register.html", map[string]interface{}{
		"FullPage": true,
		"Title":    configs.T(c, "ui.form.citizen_register", nil),
	})
}

func (h *CitizenWebHandler) WebRegister(c *echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	cccd := strings.TrimSpace(c.FormValue("citizen_id_number"))
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	renderError := func(msg string) error {
		return c.Render(http.StatusUnprocessableEntity, "citizen/pages/auth/register.html", map[string]interface{}{
			"FullPage": true,
			"Title":    configs.T(c, "ui.form.citizen_register", nil),
			"Error":    msg,
			"Name":     name,
			"Email":    email,
			"CCCD":     cccd,
		})
	}

	if name == "" || email == "" || cccd == "" || password == "" {
		return renderError(configs.T(c, "auth.invalid_request", nil))
	}
	if password != confirmPassword {
		return renderError(configs.T(c, "auth.password_mismatch", nil))
	}
	if len(cccd) != 12 || !isAllDigits(cccd) {
		return renderError(configs.T(c, "validation.invalid", map[string]string{"field": configs.T(c, "ui.form.cccd", nil)}))
	}
	if len(password) < 6 {
		return renderError(configs.T(c, "validation.min", map[string]string{"field": configs.T(c, "ui.form.password", nil), "param": "6"}))
	}

	reqData := &dtos.RegisterRequest{
		Name:            name,
		Email:           email,
		Password:        password,
		CitizenIDNumber: cccd,
	}

	user, err := h.authService.Register(reqData)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, services.ErrEmailAlreadyExists):
			msg = configs.T(c, "auth.email_exists", nil)
		default:
			msg = configs.T(c, "common.internal_error", nil)
		}
		return renderError(msg)
	}

	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: &user.ID,
		Action:      "auth.register",
		EntityType:  "user",
		EntityID:    &user.ID,
		Description: "Citizen registered via web",
		Result:      "success",
	})

	return c.Redirect(http.StatusSeeOther, "/login?"+url.Values{
		"flash": {"success"},
		"msg":   {configs.T(c, "auth.register_success", nil)},
	}.Encode())
}

func (h *CitizenWebHandler) WebLogout(c *echo.Context) error {
	var actorUserID *string
	if claims := citizenCurrentUser(c); claims != nil && claims.ID != "" {
		actorUserID = &claims.ID
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	h.writeActivityLog(&models.ActivityLog{
		ActorUserID: actorUserID,
		Action:      "auth.logout",
		EntityType:  "user",
		EntityID:    actorUserID,
		Description: "Citizen web logout successful",
		Result:      "success",
	})

	return c.Redirect(http.StatusSeeOther, "/login")
}

func (h *CitizenWebHandler) ShowDashboard(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	var userID string
	if claims != nil {
		userID = claims.ID
	}
	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.citizen.dashboard.title", nil),
		"CurrentPath": "/citizen",
		"CurrentUser": claims,
		"UnreadCount": h.unreadCount(userID),
	}
	return c.Render(http.StatusOK, "citizen/pages/dashboard.html", data)
}

func (h *CitizenWebHandler) ListNotifications(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil || h.notificationSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	page, limit := parsePagination(c)

	filterRead := strings.TrimSpace(c.QueryParam("is_read"))
	filterType := strings.TrimSpace(c.QueryParam("type"))
	filter := repositories.NotificationFilter{Type: filterType}
	switch filterRead {
	case "true":
		v := true
		filter.IsRead = &v
	case "false":
		v := false
		filter.IsRead = &v
	}

	items, total, err := h.notificationSvc.List(claims.ID, filter, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	data := map[string]interface{}{
		"Title":         configs.T(c, "ui.nav.citizen.notifications", nil),
		"CurrentPath":   "/citizen/notifications",
		"CurrentUser":   claims,
		"UnreadCount":   h.unreadCount(claims.ID),
		"Notifications": items,
		"Pagination":    utils.NewPagination(page, limit, total),
		"FilterIsRead":  filterRead,
		"FilterType":    filterType,
		"Flash":         flashFromQuery(c),
		"CSRFToken":     h.csrfToken(c),
	}
	return c.Render(http.StatusOK, "citizen/pages/notifications/list.html", data)
}

func (h *CitizenWebHandler) MarkNotificationRead(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil || h.notificationSvc == nil {
		return c.Redirect(http.StatusSeeOther, "/citizen/notifications")
	}
	id := c.Param("id")
	if err := h.notificationSvc.MarkAsRead(id, claims.ID); err != nil {
		log.Printf("mark notification %s as read failed for user %s: %v", id, claims.ID, err)
	}
	return c.Redirect(http.StatusSeeOther, "/citizen/notifications")
}

func (h *CitizenWebHandler) MarkAllNotificationsRead(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil || h.notificationSvc == nil {
		return c.Redirect(http.StatusSeeOther, "/citizen/notifications")
	}
	if err := h.notificationSvc.MarkAllAsRead(claims.ID); err != nil {
		return c.Redirect(http.StatusSeeOther, "/citizen/notifications?"+url.Values{
			"flash": {"error"},
			"msg":   {configs.T(c, "common.internal_error", nil)},
		}.Encode())
	}
	return c.Redirect(http.StatusSeeOther, "/citizen/notifications?"+url.Values{
		"flash": {"success"},
		"msg":   {configs.T(c, "notification.all_marked_read", nil)},
	}.Encode())
}

// ShowServiceCatalog handles GET /citizen/services
func (h *CitizenWebHandler) ShowServiceCatalog(c *echo.Context) error {
	if h.catalogSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	claims := citizenCurrentUser(c)
	var userID string
	if claims != nil {
		userID = claims.ID
	}

	page, limit := parsePagination(c)
	search := strings.TrimSpace(c.QueryParam("search"))
	category := strings.TrimSpace(c.QueryParam("category"))

	result, err := h.catalogSvc.List(c.Request().Context(), repositories.ListFilter{
		Search:          search,
		Category:        category,
		Page:            page,
		Limit:           limit,
		IncludeInactive: false,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	cats, _ := h.catalogSvc.ListCategories(c.Request().Context())

	return c.Render(http.StatusOK, "citizen/pages/services/list.html", map[string]interface{}{
		"Title":          configs.T(c, "ui.citizen.services.title", nil),
		"CurrentPath":    "/citizen/services",
		"CurrentUser":    claims,
		"UnreadCount":    h.unreadCount(userID),
		"ServiceTypes":   result.Items,
		"Pagination":     utils.NewPagination(page, limit, result.Total),
		"Search":         search,
		"FilterCategory": category,
		"Categories":     cats,
	})
}

// ShowServiceDetail handles GET /citizen/services/:id
func (h *CitizenWebHandler) ShowServiceDetail(c *echo.Context) error {
	if h.catalogSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	claims := citizenCurrentUser(c)
	var userID string
	if claims != nil {
		userID = claims.ID
	}

	st, err := h.catalogSvc.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "service_type.not_found")
	}

	return c.Render(http.StatusOK, "citizen/pages/services/detail.html", map[string]interface{}{
		"Title":       st.Name,
		"CurrentPath": "/citizen/services",
		"CurrentUser": claims,
		"UnreadCount": h.unreadCount(userID),
		"ServiceType": st,
	})
}

// ShowApplicationsList handles GET /citizen/applications
func (h *CitizenWebHandler) ShowApplicationsList(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "common.unauthorized")
	}
	if h.appSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	page, limit := parsePagination(c)
	apps, total, err := h.appSvc.ListMyApplications(claims.ID, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.Render(http.StatusOK, "citizen/pages/applications/list.html", map[string]interface{}{
		"Title":        configs.T(c, "ui.citizen.applications.title", nil),
		"CurrentPath":  "/citizen/applications",
		"CurrentUser":  claims,
		"UnreadCount":  h.unreadCount(claims.ID),
		"Applications": apps,
		"Pagination":   utils.NewPagination(page, limit, total),
		"Flash":        flashFromQuery(c),
	})
}

type parsedFormSchema struct {
	Fields   []string
	Required map[string]bool
}

func parseFormSchema(raw []byte) parsedFormSchema {
	if len(raw) == 0 {
		return parsedFormSchema{Required: map[string]bool{}}
	}
	var s struct {
		Fields         []string `json:"fields"`
		RequiredFields []string `json:"required_fields"`
		Required       []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return parsedFormSchema{Required: map[string]bool{}}
	}
	req := map[string]bool{}
	for _, f := range s.RequiredFields {
		req[f] = true
	}
	for _, f := range s.Required {
		req[f] = true
	}
	return parsedFormSchema{Fields: s.Fields, Required: req}
}

// ShowApplyForm handles GET /citizen/applications/new
func (h *CitizenWebHandler) ShowApplyForm(c *echo.Context) error {
	if h.catalogSvc == nil || h.appSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	claims := citizenCurrentUser(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "common.unauthorized")
	}

	stID := strings.TrimSpace(c.QueryParam("service_type_id"))
	if stID == "" {
		return c.Redirect(http.StatusSeeOther, "/citizen/services")
	}

	st, err := h.catalogSvc.GetByID(c.Request().Context(), stID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "service_type.not_found")
	}

	schema := parseFormSchema(st.FormSchema)

	return c.Render(http.StatusOK, "citizen/pages/applications/new.html", map[string]interface{}{
		"Title":       configs.T(c, "ui.citizen.applications.new.title", nil),
		"CurrentPath": "/citizen/applications",
		"CurrentUser": claims,
		"UnreadCount": h.unreadCount(claims.ID),
		"ServiceType": st,
		"Schema":      schema,
		"Values":      map[string]string{},
		"CSRFToken":   h.csrfToken(c),
	})
}

// SubmitApplication handles POST /citizen/applications
func (h *CitizenWebHandler) SubmitApplication(c *echo.Context) error {
	if h.catalogSvc == nil || h.appSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	claims := citizenCurrentUser(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "common.unauthorized")
	}

	stID := strings.TrimSpace(c.FormValue("service_type_id"))
	if stID == "" {
		return c.Redirect(http.StatusSeeOther, "/citizen/services")
	}
	st, err := h.catalogSvc.GetByID(c.Request().Context(), stID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "service_type.not_found")
	}
	schema := parseFormSchema(st.FormSchema)

	values := map[string]string{}
	fieldData := map[string]interface{}{}
	for _, f := range schema.Fields {
		v := c.FormValue(f)
		values[f] = v
		fieldData[f] = v
	}

	submittedData, _ := json.Marshal(fieldData)

	var files []*multipart.FileHeader
	if form, ferr := c.MultipartForm(); ferr == nil && form != nil {
		files = form.File["attachments[]"]
	}

	req := &dtos.SubmitApplicationRequest{
		ServiceTypeID: stID,
		SubmittedData: submittedData,
	}

	result, err := h.appSvc.SubmitApplication(claims.ID, req, files)
	if err != nil {
		var errMsg string
		switch {
		case errors.Is(err, services.ErrMissingRequiredField):
			errMsg = configs.T(c, "application.missing_required_field", nil)
		case errors.Is(err, services.ErrAttachmentRequired):
			errMsg = configs.T(c, "application.attachment_required", nil)
		case errors.Is(err, services.ErrTooManyAttachments):
			errMsg = configs.T(c, "application.attachment_too_many", nil)
		case errors.Is(err, services.ErrAttachmentTooLarge):
			errMsg = configs.T(c, "application.attachment_too_large", nil)
		case errors.Is(err, services.ErrAttachmentInvalidType):
			errMsg = configs.T(c, "application.attachment_invalid_type", nil)
		default:
			errMsg = configs.T(c, "common.internal_error", nil)
		}
		return c.Render(http.StatusUnprocessableEntity, "citizen/pages/applications/new.html", map[string]interface{}{
			"Title":       configs.T(c, "ui.citizen.applications.new.title", nil),
			"CurrentPath": "/citizen/applications",
			"CurrentUser": claims,
			"UnreadCount": h.unreadCount(claims.ID),
			"ServiceType": st,
			"Schema":      schema,
			"Values":      values,
			"Error":       errMsg,
			"CSRFToken":   h.csrfToken(c),
		})
	}

	return c.Redirect(http.StatusSeeOther, flashURL("/citizen/applications/"+result.ID, "success", configs.T(c, "ui.citizen.applications.flash.submitted", nil)))
}

// ShowApplicationDetail handles GET /citizen/applications/:id
func (h *CitizenWebHandler) ShowApplicationDetail(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "common.unauthorized")
	}
	if h.appSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	appID := c.Param("id")
	app, err := h.appSvc.GetMyApplication(claims.ID, appID)
	if err != nil {
		if errors.Is(err, services.ErrApplicationNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "application.not_found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	logs, _, err := h.appSvc.ListMyApplicationStatusHistory(claims.ID, appID, 1, 50, nil)
	if err != nil {
		logs = nil
	}

	return c.Render(http.StatusOK, "citizen/pages/applications/detail.html", map[string]interface{}{
		"Title":         app.ApplicationCode,
		"CurrentPath":   "/citizen/applications",
		"CurrentUser":   claims,
		"UnreadCount":   h.unreadCount(claims.ID),
		"Application":   app,
		"StatusHistory": logs,
		"Flash":         flashFromQuery(c),
		"CSRFToken":     h.csrfToken(c),
	})
}

// UploadApplicationSupplements handles POST /citizen/applications/:id/supplements
func (h *CitizenWebHandler) UploadApplicationSupplements(c *echo.Context) error {
	claims := citizenCurrentUser(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "common.unauthorized")
	}
	if h.appSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	appID := c.Param("id")
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return c.Redirect(http.StatusSeeOther, flashURL("/citizen/applications/"+appID, "error", configs.T(c, "application.invalid_request", nil)))
	}

	files := form.File["attachments[]"]
	_, err = h.appSvc.UploadMyApplicationSupplements(claims.ID, appID, files)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, services.ErrApplicationNotFound):
			return echo.NewHTTPError(http.StatusNotFound, "application.not_found")
		case errors.Is(err, services.ErrSupplementNotAllowed):
			msg = configs.T(c, "application.supplement_not_allowed", nil)
		case errors.Is(err, services.ErrTooManyAttachments):
			msg = configs.T(c, "application.attachment_too_many", nil)
		case errors.Is(err, services.ErrAttachmentTooLarge):
			msg = configs.T(c, "application.attachment_too_large", nil)
		default:
			msg = configs.T(c, "common.internal_error", nil)
		}
		return c.Redirect(http.StatusSeeOther, flashURL("/citizen/applications/"+appID, "error", msg))
	}

	return c.Redirect(http.StatusSeeOther, flashURL("/citizen/applications/"+appID, "success", configs.T(c, "ui.citizen.applications.flash.supplement_uploaded", nil)))
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
