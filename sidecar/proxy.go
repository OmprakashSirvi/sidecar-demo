package main

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sidecar/config"
	"sidecar/constants"
	"sidecar/globals"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewReverseProxy(upstream string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	// ginProxy := &httputil.ReverseProxy {
	// 	ModifyResponse: UpstreamResponseModifier,
	// }
	proxy := httputil.NewSingleHostReverseProxy(url)
	// We can add metrics of some other modifications on the response from the backend service
	proxy.ModifyResponse = UpstreamResponseModifier

	return proxy, nil
}

func UpstreamResponseModifier(r *http.Response) error {
	// Can modify the response, and see to the response and give out an error if something does not seems right
	// Can be used in logging and tracking
	return nil
}

func setProxyRoutes(router *gin.Engine, proxy *httputil.ReverseProxy, logger zerolog.Logger) {
	routes, err := getRoutesFromConfig(logger)
	if err != nil {
		return
	}

	for _, route := range routes {
		// TODO: Here need to add more switch cases
		switch route.Type {
		case "GET":
			router.GET(route.Path, proxyRequestHandler(proxy, route))
		}
	}
}

// Errors are already logged
func getRoutesFromConfig(logger zerolog.Logger) ([]Route, error){
	if ok := viper.IsSet(constants.PROXY_ROUTES); !ok {
		errMsg := "proxy-routes is not set, hence not configuring any routes"
		logger.Debug().Msg(errMsg)
		return nil, errors.New(errMsg)
	}

	var routes []Route
	err := viper.UnmarshalKey(config.GetKeyName(constants.PROXY_ROUTES), &routes)
	if err != nil {
		logger.Fatal().Err(err).Msg("invalid proxy-routes configuration")
		return nil, err
	}

	return routes, nil
}

func proxyRequestHandler(proxy *httputil.ReverseProxy, route Route) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := zerolog.Ctx(ctx.Request.Context())

		if ok := authorizeRequest(ctx, logger); !ok {
			return
		}

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
