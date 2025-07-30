package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"sidecar/applogger"
	"sidecar/globals"
	"sidecar/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

func GlobalRateLimiter() gin.HandlerFunc {
	return RateLimitMiddleware(nil, "service")
}

// Precedence of rate limiting, service(highest value) > route > user
func RateLimitMiddleware(route *models.ProxyRoute, limitType string) gin.HandlerFunc {
	var limiter *rate.Limiter

	switch limitType {
	case "service":
		limiter = rate.NewLimiter(
			rate.Limit(globals.Global.MaxRequestsPerSecond), globals.Global.BurstThreshold)
	case "route":
		limiter = rate.NewLimiter(rate.Limit(route.GetMaxRequestsPerSecond()), route.GetBurstThreshold())
	case "user":
		limit := *route.UserRateLimit
		window := *route.UserRateLimitWindow

		rdb := globals.Global.RedisDb
		// For now keeping the window to 5 minutes..
		return PerUserRateLimiter(rdb, int64(limit), time.Duration(window*int(time.Second)))
	}

	return func(ctx *gin.Context) {
		logger := applogger.GetCtxLogger(ctx)

		if !limiter.Allow() {
			// This request has been rate limited
			logger.Debug().Str("limit_type", limitType).Msg("too many requests")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, "too many requests, try again after some time")
			return
		}
		ctx.Next()
	}
}

func PerUserRateLimiter(rdb *redis.Client, limit int64, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := applogger.GetCtxLogger(ctx).With().Str("function", "PerUserRateLimiter").Logger()

		// Also return an error if userIdentifier is not found..
		userIdentifier, err := getUserIdFromHeader(ctx)
		if err != nil {
			logger.Error().Err(err)
			userIdentifier = ctx.GetHeader("x-forwarded-for")
		}

		// Create a redis key for the current time window
		now := time.Now().Unix()

		// This will make sure that the key is same for current time window..
		windowStart := now - (now % int64(window.Seconds()))
		// TODO: Will need to encrypt this before storing this as key
		redisKey := fmt.Sprintf("rate_limit:%s:%d", userIdentifier, windowStart)

		var currentCount int64
		pipe := rdb.Pipeline()

		incrCmd := pipe.Incr(context.Background(), redisKey)

		pipe.Expire(context.Background(), redisKey, window+5*time.Second)
		_, err = pipe.Exec(context.Background())
		if err != nil {
			// This error will indicate that the redis is down, log this error and allow this user for now
			ctx.Error(err)
			ctx.Next()
			return
		}

		currentCount, _ = incrCmd.Result()

		if currentCount > limit {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, "too many requests")
			return
		}

		ctx.Next()
	}
}
