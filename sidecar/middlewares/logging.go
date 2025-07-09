package middlewares

import (
	"sidecar/applogger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LoggingMiddleware generates a request ID, adds it to the request context
// and headers, and logs request details.
func LoggingMiddleware() gin.HandlerFunc {
	logger := applogger.GetLogger()
	return func(ctx *gin.Context) {
		// 1. Check for an existing Request ID from an upstream service
		requestId := ctx.GetHeader("X-Request-ID")
		if requestId == "" {
			// 2. Generate a new one if it doesn't exist
			requestId = uuid.New().String()
		}
		logger.Debug().Str("request_id", requestId).Msg("request id generated")

		// 3. Set the header on the INCOMING request object.
		ctx.Request.Header.Set("X-Request-ID", requestId)

		// Also set the response header so the original client gets the ID back.
		ctx.Header("X-Request-ID", requestId)

		// 4. Create a logger with the request_id field for sidecar's logs
		reqLogger := logger.With().Str("request_id", requestId).Logger()

		// 5. Store the logger in the Gin context for use in other sidecar handlers
		ctx.Set("logger", reqLogger)

		start := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method

		reqLogger.Trace().
			Str("method", method).
			Str("path", path).
			Str("ip", ctx.ClientIP()).
			Msg("incoming request")

		ctx.Next()

		latency := time.Since(start)
		statusCode := ctx.Writer.Status()

		reqLogger.Trace().
			Int("status_code", statusCode).
			Str("method", method).
			Str("path", path).
			Dur("latency", latency).
			Msg("request completed")
	}
}
