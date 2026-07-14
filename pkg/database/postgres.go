package database

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Connect establishes a connection to the PostgreSQL database with connection pooling.
func Connect() {
	dsn := config.AppConfig.DBUrl
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database: " + err.Error())
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Fatal("Failed to get database object: " + err.Error())
	}

	// Connection Pooling
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("Database connection established successfully")
}
