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
func SetProxyRoutes(router *gin.Engine, proxy *httputil.ReverseProxy, parentLogger *zerolog.Logger) {
	routes, err := config.GetRoutesFromConfig(parentLogger)
	if err != nil {
		return
	}

	// This will contain the limiters for individual routes.
	routerLimiters := make(map[string]*rate.Limiter)

	for _, route := range routes {
		logger := parentLogger.With().Str("method", route.Type).Str("path", route.Path).Logger()
		routeHandlers := []gin.HandlerFunc{}

		// Add authorization middleware if the route is protected.
		if len(route.RouteTokens) != 0 {
			// If authorization is enabled, then this should be the first middleware before ratelimiting
			routeHandlers = append(routeHandlers, middlewares.AuthorizeRequest())
		}

		// Verify that the rate-limit is valid or not, or if there are any misconfiguration or not..
		// We only configure rate limit if it is enabled for that particular route
		if route.EnableRateLimit {
			setDefaultRateLimitValues(route, &logger)

			logger.Debug().Msg("added route limiter")
			limiter := rate.NewLimiter(rate.Limit(route.GetMaxRequestsPerSecond()), route.GetBurstThreshold())
			routerLimiters[route.Path] = limiter

			// Add this middleware to the route handlers
			routeHandlers = append(routeHandlers, middlewares.RateLimitMiddleware(models.RateLimiter{Limit: limiter, Type: "route"}))
		}

		// If per user limit is enabled, then use perUserLimit middleware
		userRateLimit := models.RateLimiter{Type: "user"}
		routeHandlers = append(routeHandlers, middlewares.RateLimitMiddleware(userRateLimit))

		// Finally append the proxy request handler
		routeHandlers = append(routeHandlers, handlers.ProxyRequestHandler(proxy, route))

		// TODO: Here need to add more switch cases
		switch route.Type {
		case "GET":
			router.GET(route.Path, routeHandlers...)
		}
	}
}

func setDefaultRateLimitValues(route models.ProxyRoute, logger *zerolog.Logger) {
	if route.MaxRequestsPerSecond == nil {
		logger.Debug().Float64("max-rps", globals.Global.MaxRequestsPerSecond).Msg("setting default RPS")
		// Directly pointing it to global variable
		route.MaxRequestsPerSecond = &globals.Global.MaxRequestsPerSecond
	}
	if route.BurstThreshold == nil {
		logger.Debug().Int("burst", globals.Global.BurstThreshold).Msg("setting default burst")
		route.BurstThreshold = &globals.Global.BurstThreshold
	}
}
