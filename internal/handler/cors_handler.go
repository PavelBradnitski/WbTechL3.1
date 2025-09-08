package handler

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

// HandlerFunc defines the handler function type.
type HandlerFunc func(*ginext.Context)

// corsMiddleware implements CORS headers
func CorsMiddleware() HandlerFunc {
	return func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080") // Replace with your frontend origin
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
