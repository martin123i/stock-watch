// src/middleware/authMiddleware.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"stocktracker.com/app/internal/auth"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := auth.GetUserFromToken(c); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
