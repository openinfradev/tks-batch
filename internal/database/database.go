package database

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() (*gorm.DB, error) {
	// Connect to gormDB
	dsn := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s port=%d sslmode=disable TimeZone=Asia/Seoul",
		viper.GetString("dbhost"),
		viper.GetString("dbname"),
		viper.GetString("dbuser"),
		viper.GetString("dbpassword"),
		viper.GetInt("dbport"),
	)

	// Set log level for gorm
	var level logger.LogLevel
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		level = logger.Info
	case "TEST":
		level = logger.Silent
	default:
		level = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
