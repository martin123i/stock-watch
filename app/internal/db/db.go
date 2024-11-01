// src/db/db.go
package db

import (
	"path/to/your/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("stocks.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	// Migrate the schema
	DB.AutoMigrate(&models.User{}, &models.Portfolio{})
}
