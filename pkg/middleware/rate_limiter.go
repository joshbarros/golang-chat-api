package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimiter() gin.HandlerFunc {
	rateLimit := 1   // Default rate (requests per second)
	burstLimit := 5  // Default burst limit

	limiterMap := make(map[string]*rate.Limiter) // A map to store the limiter per client

	// Create a limiter for a given client (IP or Token)
	createLimiter := func() *rate.Limiter {
		return rate.NewLimiter(rate.Limit(rateLimit), burstLimit)
	}

	return func(c *gin.Context) {
		// Use client IP or JWT as a unique key
		clientKey := c.ClientIP() // Or use some other identifier like JWT

		limiter, exists := limiterMap[clientKey]
		if !exists {
			limiter = createLimiter()
			limiterMap[clientKey] = limiter
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
