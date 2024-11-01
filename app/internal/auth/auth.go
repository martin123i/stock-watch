// src/auth/auth.go
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"path/to/your/db"
	"path/to/your/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(c *gin.Context) {
	var creds Credentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		return
	}

	user := models.User{Username: creds.Username, Password: string(hashedPassword)}
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var creds Credentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := db.DB.Where("username = ?", creds.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   creds.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secureKey, err := generateSecureKey()
	tokenString, err := token.SignedString(secureKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not log in"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"login-token": tokenString})
}

func generateSecureKey() (string, error) {
	key := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
