package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

/*
LimitRequestSize set a limit on the request body size.
If the limit is reached, it will return a 413 status code with a JSON error message.

	maxBytes: maximum size in bytes
	Examples for a limit of 10MB:
	- bit shift: LimitRequestSize(10<<20)
	- for var/const value: LimitRequestSize(10 * 1 << 20)
	- explicit value: LimitRequestSize(10485760) or LimitRequestSize(10_485_760) or LimitRequestSize(10 * 1024 * 1024)
*/
func LimitRequestSize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

		contentType := c.ContentType()
		var err error

		switch {
		case strings.HasPrefix(contentType, "multipart/form-data"):
			err = c.Request.ParseMultipartForm(maxBytes)
		case strings.HasPrefix(contentType, "application/json"):
			var tmp map[string]any
			err = c.ShouldBindJSON(&tmp)
		}

		if err != nil && strings.Contains(err.Error(), "http: request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("The request body is too large. Maximum size expected %d Mo.", maxBytes/(1024*1024)),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
