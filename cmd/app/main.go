package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/docs"
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/middlewares"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/routes"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	templates "github.com/awesome-academy/golang_baoan_thao/internal/templates"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := configs.LoadI18nMessages("locales"); err != nil {
		log.Fatalf("failed to load i18n messages: %v", err)
	}

	db := configs.InitDB()

	e := echo.New()

	if r, err := templates.NewRenderer("templates"); err == nil {
		e.Renderer = r
	} else {
		log.Fatalf("failed to initialize templates: %v", err)
	}

	e.Validator = &configs.CustomValidator{Validator: validator.New()}

	configs.CustomLogger(e)
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(services.MaxTotalAttachmentBytes))
	e.Use(middlewares.LocaleMiddleware)

	e.HTTPErrorHandler = configs.CustomHTTPErrorHandler

	uploadDir := utils.EnvOr("UPLOAD_DIR", "./uploads")
	e.Static("/uploads", uploadDir)

	userRepo := repositories.NewUserRepo(db)
	citizenProfileRepo := repositories.NewCitizenProfileRepository(db)
	applicationRepo := repositories.NewApplicationRepository(db)
	serviceCatalogRepo := repositories.NewServiceTypeRepository(db)

	// Auth
	authService := services.NewAuthService(db, userRepo, citizenProfileRepo)
	authHandler := handlers.NewAuthHandler(authService)
	adminAuthHandler := handlers.NewAdminAuthHandler(authService)

	serviceCatalogSvc := services.NewServiceCatalogService(serviceCatalogRepo)
	adminUserSvc := services.NewAdminUserService(userRepo)
	serviceCatalogHandler := handlers.NewServiceCatalogHandler(serviceCatalogSvc, adminUserSvc)

	citizenProfileSvc := services.NewCitizenProfileService(userRepo, citizenProfileRepo, applicationRepo)
	citizenProfileHandler := handlers.NewCitizenProfileHandler(citizenProfileSvc)

	storage := utils.NewLocalDiskStorage(uploadDir, "/uploads")

	smtpCfg := services.LoadSMTPConfigFromEnv()
	mailer := services.NewSMTPMailer(smtpCfg)

	applicationSvc := services.NewApplicationService(applicationRepo, serviceCatalogRepo, userRepo, storage, mailer)
	applicationHandler := handlers.NewApplicationHandler(applicationSvc)

	adminUserHandler := handlers.NewAdminUserHandler(adminUserSvc)
	adminDashboardHandler := handlers.NewAdminDashboardHandler()

	departmentRepo := repositories.NewDepartmentRepo(db)
	staffProfileRepo := repositories.NewStaffProfileRepo(db)
	departmentSvc := services.NewDepartmentService(departmentRepo, staffProfileRepo)
	staffProfileSvc := services.NewStaffProfileService(staffProfileRepo, userRepo)
	adminDepartmentHandler := handlers.NewAdminDepartmentHandler(departmentSvc, adminUserSvc, staffProfileSvc)

	// application assignment service + admin handler
	applicationAssignmentRepo := repositories.NewApplicationAssignmentRepo(db)
	applicationAssignmentSvc := services.NewApplicationAssignmentService(applicationRepo, applicationAssignmentRepo, userRepo)
	adminApplicationSvc := services.NewAdminApplicationService(applicationRepo, applicationAssignmentSvc, storage)
	adminApplicationHandler := handlers.NewAdminApplicationHandler(adminApplicationSvc, adminUserSvc, staffProfileSvc)

	categoryRepo := repositories.NewCategoryRepo(db)
	categorySvc := services.NewCategoryService(categoryRepo)
	adminCategoryHandler := handlers.NewAdminCategoryHandler(categorySvc)

	importExportSvc := services.NewImportExportService(db, departmentRepo, userRepo, citizenProfileRepo, serviceCatalogRepo, staffProfileRepo)
	adminCitizenHandler := handlers.NewAdminCitizenHandler(citizenProfileRepo, importExportSvc)
	adminUserHandler = adminUserHandler.WithImportExport(importExportSvc).WithDeptAndStaffRepos(departmentRepo, staffProfileRepo)
	adminDepartmentHandler = adminDepartmentHandler.WithImportExport(importExportSvc)
	serviceCatalogHandler = serviceCatalogHandler.WithImportExport(importExportSvc)

	docs.SetupSwaggerRoutes(e)
	routes.SetupRoutes(e, &routes.ApiHandler{
		AuthHandler:             authHandler,
		AdminAuthHandler:        adminAuthHandler,
		AdminDashboardHandler:   adminDashboardHandler,
		AdminUserHandler:        adminUserHandler,
		AdminDepartmentHandler:  adminDepartmentHandler,
		AdminApplicationHandler: adminApplicationHandler,
		AdminCitizenHandler:     adminCitizenHandler,
		ServiceCatalogHandler:   serviceCatalogHandler,
		CitizenProfileHandler:   citizenProfileHandler,
		ApplicationHandler:      applicationHandler,
		AdminCategoryHandler:    adminCategoryHandler,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sc := echo.StartConfig{
		Address:         ":8080",
		GracefulTimeout: 5 * time.Second,
	}

	if err := sc.Start(ctx, e); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
