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
	limit := rate.NewLimiter(
		rate.Limit(globals.Global.MaxRequestsPerSecond), globals.Global.BurstThreshold)

	limiter := models.RateLimiter{Limit: limit, Type: "service"}

	return RateLimitMiddleware(limiter)
}

// TODO: Later will enhance this to introspect the headers
// Precedence of rate limiting, Service(highest value) > route > user
func RateLimitMiddleware(limiter models.RateLimiter) gin.HandlerFunc {
	if limiter.Type == "user" {
		rdb := globals.Global.RedisDb
		// For now keeping the window to 5 minutes..
		return PerUserRateLimiter(rdb, 10, 300 * time.Second)
	}

	return func(ctx *gin.Context) {
		logger := applogger.GetCtxLogger(ctx)

		if !limiter.Limit.Allow() {
			// This request has been rate limited
			logger.Debug().Str("limit_type", limiter.Type).Msg("too many requests")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, "too many requests, try again after some time")
			return
		}
		ctx.Next()
	}
}

func PerUserRateLimiter(rdb *redis.Client, limit int64, window time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Also return an error if userId is not found..
		userId := getUserIdFromHeader(ctx)

		// Create a redis key for the current time window
		now := time.Now().Unix()

		// This will make sure that the key is same for current time window..
		windowStart := now - (now % int64(window.Seconds()))
		redisKey := fmt.Sprintf("rate_limit:%s:%d", userId, windowStart)

		var currentCount int64
		pipe := rdb.Pipeline()

		incrCmd := pipe.Incr(context.Background(), redisKey)

		pipe.Expire(context.Background(), redisKey, window+5*time.Second)
		_, err := pipe.Exec(context.Background())
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
