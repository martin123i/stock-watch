// src/db/db.go
package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"stocktracker.com/app/internal/model"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("stocks.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	// Migrate the schema
	DB.AutoMigrate(&model.User{}, &model.Portfolio{})
}
