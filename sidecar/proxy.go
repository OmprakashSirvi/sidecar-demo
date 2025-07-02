package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sidecar/config"
	"sidecar/constants"

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
	logger := getLogger()
	logger.Debug().Msg(fmt.Sprintf("handling route for: %v:%v", route.Type, route.Path))
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
