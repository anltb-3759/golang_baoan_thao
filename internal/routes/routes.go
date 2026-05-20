package routes

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/middlewares"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

type ApiHandler struct {
	AuthHandler           *handlers.AuthHandler
	AdminAuthHandler      *handlers.AdminAuthHandler
	ServiceCatalogHandler *handlers.ServiceCatalogHandler
}

func SetupRoutes(e *echo.Echo, handler *ApiHandler) {
	// Admin web routes (HTML templates)
	e.GET("/admin/login", handler.AdminAuthHandler.ShowLoginPage)
	e.POST("/admin/login", handler.AdminAuthHandler.WebLogin)
	e.GET("/admin/logout", handler.AdminAuthHandler.WebLogout)
	e.GET("/set-locale", handler.AdminAuthHandler.SetLocale)

	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/login", handler.AuthHandler.Login)
	auth.POST("/register", handler.AuthHandler.Register)
	auth.POST("/refresh", handler.AuthHandler.RefreshTokenHandler)
	auth.POST("/logout", handler.AuthHandler.Logout)

	citizen := api.Group("/citizens")
	citizen.Use(middlewares.JWTMiddleware)
	citizen.Use(middlewares.RequireRoles(models.UserRoleCitizen))

	staff := api.Group("/staff")
	staff.Use(middlewares.JWTMiddleware)
	staff.Use(middlewares.RequireRoles(models.UserRoleStaff))

	manager := api.Group("/managers")
	manager.Use(middlewares.JWTMiddleware)
	manager.Use(middlewares.RequireRoles(models.UserRoleManager))

	superAdmin := api.Group("/super-admins")
	superAdmin.Use(middlewares.JWTMiddleware)
	superAdmin.Use(middlewares.RequireRoles(models.UserRoleSuperAdmin))

	services := api.Group("/services")
	services.GET("", handler.ServiceCatalogHandler.ListServices)
	services.GET("/:id", handler.ServiceCatalogHandler.GetService)
}
