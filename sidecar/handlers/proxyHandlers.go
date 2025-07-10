package handlers

import (
	"net/http"
	"net/http/httputil"
	"sidecar/globals"
	"sidecar/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// This will be the handler for all the requests
func ProxyRequestHandler(proxy *httputil.ReverseProxy, route models.ProxyRoute) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := zerolog.Ctx(ctx.Request.Context())

		if ok := authorizeRequest(ctx, logger); !ok {
			return
		}

		// We can implement rate limit, which can be set in the route object here..

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}


func authorizeRequest(c *gin.Context, parentLogger *zerolog.Logger) bool {
	userId := getUserIdFromHeader(c)

	path := c.Request.URL.Path
	method := c.Request.Method
	logger := parentLogger.With().Str("userId", userId).Str("method", method).Str("path", path).Logger()
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

