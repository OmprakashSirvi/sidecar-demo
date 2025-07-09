package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)


func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestId := uuid.New().String()

		reqLogger := log.With().Str("request_id", requestId).Logger()

		c := reqLogger.WithContext(ctx.Request.Context())
		ctx.Request = ctx.Request.WithContext(c)

		start := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method

		reqLogger.Debug().Str("method", method).Str("path", path).Str("ip", ctx.ClientIP()).
			Msg("incoming request")

		ctx.Next()

		latency := time.Since(start)
		statusCode := ctx.Writer.Status()

		reqLogger.Debug().Int("status_code", statusCode).Str("method", method).Str("path", path).Dur("latency", latency).
			Msg("request completed")
	}
}

