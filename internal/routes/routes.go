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
	AdminDashboardHandler *handlers.AdminDashboardHandler
	AdminUserHandler      *handlers.AdminUserHandler
	ServiceCatalogHandler *handlers.ServiceCatalogHandler
	CitizenProfileHandler *handlers.CitizenProfileHandler
	ApplicationHandler    *handlers.ApplicationHandler
}

func SetupRoutes(e *echo.Echo, handler *ApiHandler) {
	// Admin auth (public)
	e.GET("/admin/login", handler.AdminAuthHandler.ShowLoginPage)
	e.POST("/admin/login", handler.AdminAuthHandler.WebLogin)
	e.GET("/admin/logout", handler.AdminAuthHandler.WebLogout)
	e.GET("/set-locale", handler.AdminAuthHandler.SetLocale)

	// Admin web routes (protected by cookie auth)
	admin := e.Group("/admin", middlewares.AdminWebMiddleware)
	admin.GET("", handler.AdminDashboardHandler.ShowDashboard)

	// Users (Super Admin only)
	users := admin.Group("/users", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	users.GET("", handler.AdminUserHandler.ListUsers)
	users.GET("/new", handler.AdminUserHandler.ShowCreateForm)
	users.POST("", handler.AdminUserHandler.CreateUser)
	users.GET("/:id", handler.AdminUserHandler.ShowUser)
	users.GET("/:id/edit", handler.AdminUserHandler.ShowEditForm)
	users.POST("/:id/edit", handler.AdminUserHandler.UpdateUser)
	users.POST("/:id/block", handler.AdminUserHandler.BlockUser)
	users.POST("/:id/unblock", handler.AdminUserHandler.UnblockUser)
	users.POST("/:id/delete", handler.AdminUserHandler.DeleteUser)

	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/login", handler.AuthHandler.Login)
	auth.POST("/register", handler.AuthHandler.Register)
	auth.POST("/refresh", handler.AuthHandler.RefreshTokenHandler)
	auth.POST("/logout", handler.AuthHandler.Logout)

	citizen := api.Group("/citizens")
	citizen.Use(middlewares.JWTMiddleware)
	citizen.Use(middlewares.RequireRoles(models.UserRoleCitizen))
	citizen.GET("/me", handler.CitizenProfileHandler.GetMe)
	citizen.PUT("/me", handler.CitizenProfileHandler.UpdateMe)
	citizen.GET("/services", handler.ServiceCatalogHandler.ListServices)
	citizen.GET("/services/:id", handler.ServiceCatalogHandler.GetService)

	// Applications
	citizen.POST("/me/applications", handler.ApplicationHandler.Submit)
	citizen.GET("/me/applications", handler.ApplicationHandler.ListMine)
	citizen.GET("/me/applications/:id", handler.ApplicationHandler.GetMine)

	staff := api.Group("/staff")
	staff.Use(middlewares.JWTMiddleware)
	staff.Use(middlewares.RequireRoles(models.UserRoleStaff))

	manager := api.Group("/managers")
	manager.Use(middlewares.JWTMiddleware)
	manager.Use(middlewares.RequireRoles(models.UserRoleManager))

	superAdmin := api.Group("/super-admins")
	superAdmin.Use(middlewares.JWTMiddleware)
	superAdmin.Use(middlewares.RequireRoles(models.UserRoleSuperAdmin))
}
