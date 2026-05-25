package handlers

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

const AGE_REFRESH_TOKEN = 7 * 24 * 60 * 60
const AGE_LANGUAGE_COOKIE = 365 * 24 * 60 * 60

func cookieSecure() bool {
	return os.Getenv("APP_ENV") == "production"
}

type AuthHandler struct {
	authService AuthService
	logger      ActivityLogger
}

type AuthService interface {
	Register(reqData *dtos.RegisterRequest) (*models.User, error)
	Login(reqData *dtos.LoginRequest) (*models.User, string, string, error)
}

type ActivityLogger interface {
	Log(log *models.ActivityLog) error
}

func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) WithActivityLogger(logger ActivityLogger) *AuthHandler {
	h.logger = logger
	return h
}

func (h *AuthHandler) writeActivityLog(log *models.ActivityLog) {
	if h.logger == nil || log == nil {
		return
	}
	_ = h.logger.Log(log)
}

func (h *AuthHandler) Register(c *echo.Context) error {
	reqData := new(dtos.RegisterRequest)
	if err := c.Bind(reqData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "auth.invalid_request").Wrap(err)
	}

	if err := c.Validate(reqData); err != nil {
		return err
	}

	user, err := h.authService.Register(reqData)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return err
	}

	return c.JSON(http.StatusCreated, utils.Map{
		"user": user,
	})
}

func (h *AuthHandler) Login(c *echo.Context) error {
	reqData := new(dtos.LoginRequest)
	if err := c.Bind(reqData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "auth.invalid_request").Wrap(err)
	}

	if err := c.Validate(reqData); err != nil {
		return err
	}

	user, token, refreshToken, err := h.authService.Login(reqData)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			return echo.NewHTTPError(http.StatusUnauthorized, "auth.user_not_found")
		case errors.Is(err, services.ErrPasswordMismatch):
			return echo.NewHTTPError(http.StatusUnauthorized, "auth.password_mismatch")
		case errors.Is(err, services.ErrUserBlocked):
			return echo.NewHTTPError(http.StatusForbidden, "auth.user_blocked")
		default:
			return err
		}
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
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
		Description: "Citizen login successful",
		Result:      "success",
	})

	return c.JSON(http.StatusOK, utils.Map{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) RefreshTokenHandler(c *echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "auth.missing_refresh_token")
	}

	claims, err := configs.ParseToken(cookie.Value, configs.RefreshTokenType)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "auth.invalid_refresh_token")
	}

	token, err := configs.GenerateAccessToken(&models.User{
		ID:    claims.ID,
		Email: claims.Email,
		Role:  models.UserRole(claims.Role),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}

	return c.JSON(http.StatusOK, utils.Map{
		"token": token,
	})
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	var actorUserID *string
	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		claims, parseErr := configs.ParseToken(cookie.Value, configs.RefreshTokenType)
		if parseErr == nil && claims.ID != "" {
			actorUserID = &claims.ID
		}
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
		Description: "Citizen logout successful",
		Result:      "success",
	})

	return c.JSON(http.StatusOK, utils.Map{
		"message": configs.T(c, "auth.logout_success", nil),
	})
}
