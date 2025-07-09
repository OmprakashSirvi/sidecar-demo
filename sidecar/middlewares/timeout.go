package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"sidecar/globals"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware creates a Gin middleware for handling request timeouts.
func TimeoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		timeoutSeconds := globals.Global.RequestTimeout
		if timeoutSeconds <= 0 {
			// If timeout is not set or invalid, proceed without a timeout.
			c.Next()
			return
		}

		// Create a context with a timeout.
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(timeoutSeconds)*time.Second)
		defer cancel() // IMPORTANT: Ensures resources are cleaned up.

		// Replace the request's context with our new timed context.
		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal when the handler has finished.
		done := make(chan struct{})

		// Run the next handlers in a goroutine.
		go func() {
			c.Next()
			// Signal that the handler has finished.
			close(done)
		}()

		// Use select to wait for either the handler to finish or the context to time out.
		select {
		case <-done:
			// The handler finished in time. We can just return.
			return
		case <-ctx.Done():
			// The context's deadline was exceeded.
			// Check if the response has already been written.
			if !c.Writer.Written() {
				// We are sending 503 Service Unavailable, which is often more appropriate
				// than 408 Request Timeout, as the server is choosing to drop the request.
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
					"error": fmt.Sprintf("Request timed out after %d seconds", timeoutSeconds),
				})
			}
		}
	}
}
