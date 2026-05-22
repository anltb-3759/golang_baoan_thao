package routes

import (
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/middlewares"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
)

type ApiHandler struct {
	AdminAuthHandler        *handlers.AdminAuthHandler
	AdminDashboardHandler   *handlers.AdminDashboardHandler
	AdminUserHandler        *handlers.AdminUserHandler
	AdminDepartmentHandler  *handlers.AdminDepartmentHandler
	AdminCategoryHandler    *handlers.AdminCategoryHandler
	AdminApplicationHandler *handlers.AdminApplicationHandler
	AuthHandler             *handlers.AuthHandler
	CitizenProfileHandler   *handlers.CitizenProfileHandler
	ServiceCatalogHandler   *handlers.ServiceCatalogHandler
	ApplicationHandler      *handlers.ApplicationHandler
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

	// Service types (Super Admin only)
	serviceTypes := admin.Group("/service-types", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	serviceTypes.GET("", handler.ServiceCatalogHandler.ListServiceTypesAdmin)
	serviceTypes.GET("/:id", handler.ServiceCatalogHandler.ShowServiceType)
	serviceTypes.GET("/new", handler.ServiceCatalogHandler.CreateServiceTypeForm)
	serviceTypes.POST("", handler.ServiceCatalogHandler.CreateServiceType)
	serviceTypes.GET("/:id/edit", handler.ServiceCatalogHandler.EditServiceTypeForm)
	serviceTypes.POST("/:id", handler.ServiceCatalogHandler.UpdateServiceType)
	serviceTypes.POST("/:id/delete", handler.ServiceCatalogHandler.DeleteServiceType)

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

	// Department list (Manager + Super Admin)
	deptList := admin.Group("/departments", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	deptList.GET("", handler.AdminDepartmentHandler.ListDepartments)

	// Departments admin actions (Super Admin only)
	depts := admin.Group("/departments", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	depts.GET("/new", handler.AdminDepartmentHandler.ShowCreateForm)
	depts.POST("", handler.AdminDepartmentHandler.CreateDepartment)
	depts.GET("/:id/edit", handler.AdminDepartmentHandler.ShowEditForm)
	depts.POST("/:id/edit", handler.AdminDepartmentHandler.UpdateDepartment)
	depts.POST("/:id/delete", handler.AdminDepartmentHandler.DeleteDepartment)

	// Categories (Super Admin only)
	cats := admin.Group("/categories", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	cats.GET("", handler.AdminCategoryHandler.ListCategories)
	cats.GET("/new", handler.AdminCategoryHandler.ShowCreateForm)
	cats.POST("", handler.AdminCategoryHandler.CreateCategory)
	cats.GET("/:id/edit", handler.AdminCategoryHandler.ShowEditForm)
	cats.POST("/:id/edit", handler.AdminCategoryHandler.UpdateCategory)
	cats.POST("/:id/delete", handler.AdminCategoryHandler.DeleteCategory)
	// Department staff management (Manager + Super Admin)
	deptStaff := admin.Group("/departments/:id/staff", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	deptStaff.GET("", handler.AdminDepartmentHandler.ListDepartmentStaff)
	deptStaff.GET("/assign", handler.AdminDepartmentHandler.ShowAssignStaffForm)
	deptStaff.POST("/assign", handler.AdminDepartmentHandler.AssignStaffToDept)
	deptStaff.POST("/:user_id/remove", handler.AdminDepartmentHandler.RemoveStaffFromDept)

	// Admin applications (manager+)
	apps := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleManager))
	apps.GET("", handler.AdminApplicationHandler.ListApplications)
	apps.GET("/:id", handler.AdminApplicationHandler.ShowApplication)
	apps.GET("/:id/assign", handler.AdminApplicationHandler.ShowAssignForm)
	apps.POST("/:id/assign", handler.AdminApplicationHandler.AssignToStaff)

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
	citizen.GET("/me/applications/:id/status-history", handler.ApplicationHandler.ListMyStatusHistory)
	citizen.POST("/me/applications/:id/supplements", handler.ApplicationHandler.UploadSupplements)

	staff := api.Group("/staff")
	staff.Use(middlewares.JWTMiddleware)
	staff.Use(middlewares.RequireRoles(models.UserRoleStaff))

	manager := api.Group("/managers")
	manager.Use(middlewares.JWTMiddleware)
	manager.Use(middlewares.RequireRoles(models.UserRoleManager))
}
