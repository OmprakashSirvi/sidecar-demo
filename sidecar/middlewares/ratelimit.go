package middlewares

import (
	"net/http"
	"sidecar/applogger"
	"sidecar/globals"
	"sidecar/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func GlobalRateLimiter() gin.HandlerFunc {
	limit := rate.NewLimiter(
		rate.Limit(globals.Global.MaxRequestsPerSecond), globals.Global.BurstThreshold)

	limiter := models.RateLimiter{Limit: limit, Type: "service"}

	return RateLimitMiddleware(limiter)
}

// TODO: Later will enhance this to introspect the headers
func RateLimitMiddleware(limiter models.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := applogger.GetCtxLogger(ctx)

		if !limiter.Limit.Allow() {
			logger.Debug().Str("limit_type", limiter.Type).Msg("too many requests")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, "too many requests, try again after some time")
			return
		}
		ctx.Next()
	}
}
