package db

import (
	"go-csv-import/internal/config"
	"go-csv-import/internal/model"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var connected bool

func Connect(c config.DbConfig) error {
	connected = false
	var err error
	DB, err = gorm.Open(mysql.Open(c.Dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		connected = true
	}

	return err
}

func AutoMigrate() {
	if connected {
		DB.AutoMigrate(&model.Contact{})
	}
}
