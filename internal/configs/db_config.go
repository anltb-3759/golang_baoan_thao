package configs

import (
	"fmt"
	"os"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("database configuration error: DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		panic(fmt.Sprintf("failed to enable pgcrypto extension: %v", err))
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.CitizenProfile{},
		&models.Department{},
		&models.StaffProfile{},
		&models.ServiceType{},
		&models.Application{},
		&models.ApplicationAttachment{},
		&models.ApplicationStatusLog{},
		&models.ApplicationAssignment{},
		&models.Notification{},
		&models.ActivityLog{},
		&models.ImportExportLog{},
	); err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}

	return db
}
