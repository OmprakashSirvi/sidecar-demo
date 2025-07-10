package routes

import (
	"net/http/httputil"
	"sidecar/config"
	"sidecar/globals"
	"sidecar/handlers"
	"sidecar/middlewares"
	"sidecar/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

// This will run only at the setup
func SetProxyRoutes(router *gin.Engine, proxy *httputil.ReverseProxy, parentLogger zerolog.Logger) {
	routes, err := config.GetRoutesFromConfig(parentLogger)
	if err != nil {
		return
	}

	// This will contain the limiters for individual routes.
	routerLimiters := make(map[string]*rate.Limiter)

	// TODO: If the route is protected, then add a middleware that introspects the ctx header to get the token.
	// This middleware will also be responsible for adding user-ids into the context
	for _, route := range routes {
		logger := parentLogger.With().Str("method", route.Type).Str("path", route.Path).Logger()
		routeHandlers := []gin.HandlerFunc{handlers.ProxyRequestHandler(proxy, route)}
		// Verify that the rate-limit is valid or not, or if there are any misconfiguration or not..
		// We only configure rate limit if it is enabled for that particular route
		if route.EnableRateLimit {
			if route.MaxRequestsPerSecond == nil {
				logger.Debug().Float64("max-rps", globals.Global.MaxRequestsPerSecond).Msg("setting default RPS")
				// Directly pointing it to global variable
				route.MaxRequestsPerSecond = &globals.Global.MaxRequestsPerSecond
			}
			if route.BurstThreshold == nil {
				logger.Debug().Int("burst", globals.Global.BurstThreshold).Msg("setting default burst")
				route.BurstThreshold = &globals.Global.BurstThreshold
			}
			logger.Debug().Msg("added route limiter")
			limiter := rate.NewLimiter(rate.Limit(route.GetMaxRequestsPerSecond()), route.GetBurstThreshold())
			routerLimiters[route.Path] = limiter

			// Add this middleware at the start of the handler functions
			routeHandlers = append([]gin.HandlerFunc{
				middlewares.RateLimitMiddleware(models.RateLimiter{Limit: limiter, Type: "route"})},
				routeHandlers...,
			)
		}

		// TODO: Here need to add more switch cases
		switch route.Type {
		case "GET":
			router.GET(route.Path, routeHandlers...)
		}
	}
}
