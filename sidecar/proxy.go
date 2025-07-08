package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sidecar/applogger"
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

// panic if something goes wrong
func setProxyRoutes(router *gin.Engine, proxy *httputil.ReverseProxy, logger zerolog.Logger) {
	// This will be a good location to get the proxy routes
	if ok := viper.IsSet(constants.PROXY_ROUTES); !ok {
		logger.Debug().Msg("proxy-routes is not set, hence not configuring any routes")
		return
	}

	var routes []Route
	err := viper.UnmarshalKey(config.GetKeyNameForEnv(constants.PROXY_ROUTES), &routes)
	if err != nil {
		panic("invalid proxy-routes configuration")
	}

	for _, route := range routes {
		// TODO: Here need to add more switch cases
		switch route.Type {
		case "GET":
			router.GET(route.Path, proxyRequestHandler(proxy, route))
		}
	}
}

func proxyRequestHandler(proxy *httputil.ReverseProxy, route Route) gin.HandlerFunc {
	logger := applogger.GetLogger()
	logger.Debug().Msg(fmt.Sprintf("handling route for: %v:%v", route.Type, route.Path))
	return func(ctx *gin.Context) {
		userId := ctx.GetHeader("x-user-id")
		path := ctx.Request.URL.Path
		method := ctx.Request.Method
		ok, err := globals.Global.UserAuthorizer.Enforcer.Enforce(userId, path, method)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, "something went wrong while authorizing user")
			return
		}
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, "user not authorized")
			return
		}
		msg := fmt.Sprintf("the user: %v, is allowed access to %v:%v", userId, method, path)
		logger.Debug().Msg(msg)

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
