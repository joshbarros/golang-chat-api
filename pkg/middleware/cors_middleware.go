package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS middleware to allow cross-origin requests
func SetupCORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Origin, Accept")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

        // If it's a preflight request, we just send 204 and return early.
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        // Pass the request down the middleware chain
        c.Next()
    }
}
