// src/portfolio/portfolio.go
package portfolio

import (
	"net/http"
	"path/to/your/auth"
	"path/to/your/db"
	"path/to/your/models"

	"github.com/gin-gonic/gin"
)

func AddFavorite(c *gin.Context) {
	username, _ := auth.GetUserFromToken(c) // Helper function to decode JWT token
	var request struct {
		Symbol string `json:"symbol"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	portfolio := models.Portfolio{UserID: user.ID, Symbol: request.Symbol}
	db.DB.Create(&portfolio)
	c.JSON(http.StatusOK, gin.H{"message": "Stock added to portfolio"})
}

func GetFavorites(c *gin.Context) {
	username, _ := auth.GetUserFromToken(c)
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var portfolio []models.Portfolio
	db.DB.Where("user_id = ?", user.ID).Find(&portfolio)
	c.JSON(http.StatusOK, portfolio)
}
