package middlewares

import (
	"net/http"
	"sidecar/applogger"
	"sidecar/globals"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func AuthorizeRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := applogger.GetCtxLogger(ctx)
		authorizeRequest(ctx, &logger)
		ctx.Next()
	}
}

func authorizeRequest(c *gin.Context, parentLogger *zerolog.Logger) bool {
	userId := getUserIdFromHeader(c)

	path := c.Request.URL.Path
	method := c.Request.Method
	logger := parentLogger.With().Str("userId", userId).Str("method", method).Str("path", path).Logger()

	// This checks if the user has relevant permissions to access this endpoint
	ok, err := globals.Global.UserAuthorizer.Enforcer.Enforce(userId, path, method)
	if err != nil {
		msg := "something went wrong while authorizing user"
		logger.Error().Err(err).Msg(msg)
		c.AbortWithStatusJSON(http.StatusBadRequest, msg)
		return false
	}
	if !ok {
		logger.Error().Msg("user not allowed to access this route")
		c.AbortWithStatusJSON(http.StatusForbidden, "user not authorized")
		return false
	}

	logger.Debug().Msg("allowed access")
	return true
}

// TODO: This will be enhanced to verify the jwt, introspect the token to get userID
func getUserIdFromHeader(c *gin.Context) string {
	return c.GetHeader("x-user-id")
}
