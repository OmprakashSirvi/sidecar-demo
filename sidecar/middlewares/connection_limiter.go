package middlewares

import (
	"net/http"
	"sidecar/globals"

	"github.com/gin-gonic/gin"
)

// ConnectionLimiter creates a Gin middleware that limits the number of
// concurrent requests being processed.
func ConnectionLimiter() gin.HandlerFunc {
	maxConnections := globals.Global.MaxConnectionLimit

	// If the provided limit is zero or negative, the middleware is disabled.
	if maxConnections <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// A buffered channel is used as a semaphore. The capacity of the channel
	// is the maximum number of concurrent connections allowed.
	semaphore := make(chan struct{}, maxConnections)

	return func(c *gin.Context) {
		// We use a select statement to attempt to send a value to the semaphore channel.
		// This is a non-blocking operation.
		select {
		case semaphore <- struct{}{}:
			// If the send is successful, it means there was a free slot.
			// We must release the slot when the request is done.
			// 'defer' ensures this runs even if the handler panics.
			defer func() {
				<-semaphore // Release the slot.
			}()

			// Proceed to the next middleware or handler.
			c.Next()

		default:
			// If the send fails (because the channel buffer is full),
			// it means the connection limit has been reached.
			// We abort the request with a 503 Service Unavailable status.
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "Service is busy, please try again later.",
			})
			return
		}
	}
}
