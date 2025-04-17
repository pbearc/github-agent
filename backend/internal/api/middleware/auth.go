package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS middleware with proper preflight handling
func CORS() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"http://githubagents3-18042025.s3-website-us-east-1.amazonaws.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return cors.New(config)
}

// Auth middleware (placeholder for future implementation)
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add authentication logic here when needed
		c.Next()
	}
}
