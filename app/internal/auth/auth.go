package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"stocktracker.com/app/internal/db"
	"stocktracker.com/app/internal/model"
)

var (
	GlobalSecureKey string
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Init() {
	var err error
	GlobalSecureKey, err = generateSecureKey(32)
	if err != nil {
		log.Fatalf("Failed to generate secure key: %v", err)
	}
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

	user := model.User{Username: creds.Username, Password: string(hashedPassword)}
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

	var user model.User
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
	tokenString, err := token.SignedString([]byte(GlobalSecureKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not log in"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"login-token": tokenString})
}

// GetUserFromToken extracts the username from the JWT token in the request
func GetUserFromToken(c *gin.Context) (string, error) {
	// Get the token from the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header provided")
	}

	// Bearer token should start with "Bearer "
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return "", errors.New("invalid token format")
	}

	// Parse the token with the claims
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(GlobalSecureKey), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	// Extract the claims
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	return "", errors.New("could not extract user from token")
}

func generateSecureKey(length int) (string, error) {
	// Create a byte slice of the specified length
	key := make([]byte, length)

	// Fill the byte slice with random data
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("error generating key: %v", err)
	}

	// Encode the byte slice to a base64 string for ease of use
	secureKey := base64.StdEncoding.EncodeToString(key)
	return secureKey, nil
}
