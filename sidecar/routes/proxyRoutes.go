package routes

import (
	"net/http/httputil"
	"sidecar/config"
	"sidecar/handlers"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// This will run only at the setup
func SetProxyRoutes(router *gin.Engine, proxy *httputil.ReverseProxy, logger zerolog.Logger) {
	routes, err := config.GetRoutesFromConfig(logger)
	if err != nil {
		return
	}

	for _, route := range routes {
		// TODO: Here need to add more switch cases
		switch route.Type {
		case "GET":
			router.GET(route.Path, handlers.ProxyRequestHandler(proxy, route))
		}
	}
}
