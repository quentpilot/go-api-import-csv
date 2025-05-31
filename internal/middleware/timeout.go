package middleware

import (
	"context"
	"go-csv-import/internal/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout middleware handles a endpoint handler.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("Run Timeout Middleware")
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace HTTP request
		c.Request = c.Request.WithContext(ctx)

		done := make(chan struct{})
		panicChan := make(chan any)

		// Execute handle within a goroutine
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			c.Next()
			close(done)
		}()

		select {
		case p := <-panicChan:
			panic(p) // Resent panic
		case <-done:
			// Handler correctly finished
		case <-ctx.Done():
			// Timeout reached
			logger.Error("Request timeout reached")
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"message": "request timeout exceeded",
			})
		}
	}
}
