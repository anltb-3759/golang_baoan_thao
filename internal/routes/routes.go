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
	AdminCitizenHandler     *handlers.AdminCitizenHandler
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

	// Citizens admin (Super Admin only)
	citizens := admin.Group("/citizens", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	citizens.GET("", handler.AdminCitizenHandler.ListCitizens)
	citizens.GET("/template", handler.AdminCitizenHandler.DownloadTemplate)
	citizens.POST("/import", handler.AdminCitizenHandler.ImportCSV)

	// Service types: list (Manager + Super Admin)
	stRead := admin.Group("/service-types", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	stRead.GET("", handler.ServiceCatalogHandler.ListServiceTypesAdmin)
	stRead.GET("/:id", handler.ServiceCatalogHandler.ShowServiceType)

	// Service types: write (Super Admin only)
	stWrite := admin.Group("/service-types", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	stWrite.GET("/export", handler.ServiceCatalogHandler.ExportCSV)
	stWrite.GET("/template", handler.ServiceCatalogHandler.DownloadTemplate)
	stWrite.POST("/import", handler.ServiceCatalogHandler.ImportCSV)
	stWrite.GET("/new", handler.ServiceCatalogHandler.CreateServiceTypeForm)
	stWrite.POST("", handler.ServiceCatalogHandler.CreateServiceType)
	stWrite.GET("/:id/edit", handler.ServiceCatalogHandler.EditServiceTypeForm)
	stWrite.POST("/:id", handler.ServiceCatalogHandler.UpdateServiceType)
	stWrite.POST("/:id/delete", handler.ServiceCatalogHandler.DeleteServiceType)

	// Users: list (Manager + Super Admin)
	usersRead := admin.Group("/users", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	usersRead.GET("", handler.AdminUserHandler.ListUsers)
	usersRead.GET("/:id", handler.AdminUserHandler.ShowUser)

	// Users: write (Super Admin only)
	usersWrite := admin.Group("/users", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	usersWrite.GET("/export/citizens", handler.AdminUserHandler.ExportCitizens)
	usersWrite.GET("/export/staff", handler.AdminUserHandler.ExportStaff)
	usersWrite.GET("/template", handler.AdminUserHandler.DownloadTemplate)
	usersWrite.POST("/import", handler.AdminUserHandler.ImportCSV)
	usersWrite.GET("/new", handler.AdminUserHandler.ShowCreateForm)
	usersWrite.POST("", handler.AdminUserHandler.CreateUser)
	usersWrite.GET("/:id/edit", handler.AdminUserHandler.ShowEditForm)
	usersWrite.POST("/:id/edit", handler.AdminUserHandler.UpdateUser)
	usersWrite.POST("/:id/block", handler.AdminUserHandler.BlockUser)
	usersWrite.POST("/:id/unblock", handler.AdminUserHandler.UnblockUser)
	usersWrite.POST("/:id/delete", handler.AdminUserHandler.DeleteUser)

	// Department list (Manager + Super Admin)
	deptList := admin.Group("/departments", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	deptList.GET("", handler.AdminDepartmentHandler.ListDepartments)
	deptList.GET("/export", handler.AdminDepartmentHandler.ExportCSV)
	deptList.GET("/template", handler.AdminDepartmentHandler.DownloadTemplate)

	// Departments admin actions (Super Admin only)
	depts := admin.Group("/departments", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	depts.POST("/import", handler.AdminDepartmentHandler.ImportCSV)
	depts.GET("/new", handler.AdminDepartmentHandler.ShowCreateForm)
	depts.POST("", handler.AdminDepartmentHandler.CreateDepartment)
	depts.GET("/:id/edit", handler.AdminDepartmentHandler.ShowEditForm)
	depts.POST("/:id/edit", handler.AdminDepartmentHandler.UpdateDepartment)
	depts.POST("/:id/delete", handler.AdminDepartmentHandler.DeleteDepartment)

	// Categories: list (Manager + Super Admin)
	catRead := admin.Group("/categories", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	catRead.GET("", handler.AdminCategoryHandler.ListCategories)

	// Categories: write (Super Admin only)
	catWrite := admin.Group("/categories", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	catWrite.GET("/new", handler.AdminCategoryHandler.ShowCreateForm)
	catWrite.POST("", handler.AdminCategoryHandler.CreateCategory)
	catWrite.GET("/:id/edit", handler.AdminCategoryHandler.ShowEditForm)
	catWrite.POST("/:id/edit", handler.AdminCategoryHandler.UpdateCategory)
	catWrite.POST("/:id/delete", handler.AdminCategoryHandler.DeleteCategory)
	// Department staff management (Manager + Super Admin)
	deptStaff := admin.Group("/departments/:id/staff", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	deptStaff.GET("", handler.AdminDepartmentHandler.ListDepartmentStaff)
	deptStaff.GET("/assign", handler.AdminDepartmentHandler.ShowAssignStaffForm)
	deptStaff.POST("/assign", handler.AdminDepartmentHandler.AssignStaffToDept)
	deptStaff.POST("/:user_id/remove", handler.AdminDepartmentHandler.RemoveStaffFromDept)

	// Admin applications: read (Manager + Super Admin)
	appsRead := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	appsRead.GET("", handler.AdminApplicationHandler.ListApplications)
	appsRead.GET("/export", handler.AdminApplicationHandler.ExportCSV)
	appsRead.GET("/:id", handler.AdminApplicationHandler.ShowApplication)

	// Admin applications: write (Manager only)
	appsWrite := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleManager))
	appsWrite.POST("/:id/process", handler.AdminApplicationHandler.ProcessApplication)
	appsWrite.GET("/:id/assign", handler.AdminApplicationHandler.ShowAssignForm)
	appsWrite.POST("/:id/assign", handler.AdminApplicationHandler.AssignToStaff)

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
