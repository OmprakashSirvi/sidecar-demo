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
)

// TODO: Later will enhance this to introspect the headers
// This will run only at the setup
func SetProxyRoutes(router *gin.Engine, proxy *httputil.ReverseProxy, parentLogger *zerolog.Logger) {
	routes, err := config.GetRoutesFromConfig(parentLogger)
	if err != nil {
		return
	}

	for _, route := range routes {
		logger := parentLogger.With().Str("method", route.Method).Str("path", route.Path).Logger()

		// Checking if this route is configured properly
		if msg, ok := route.IsValidRoute(); !ok {
			// This is not a valid route
			logger.Error().Str("reason", msg).Msg("invalid route configuration")
			continue
		}
		setDefaultRateLimitValues(&route, &logger)
		routeHandlers := []gin.HandlerFunc{}

		// Add authorization middleware if the route is protected.
		if len(route.RoutePolicies) != 0 {
			// Should add a middleware here which will validate the JWT tokens beforehand
			routeHandlers = append(routeHandlers, middlewares.ValidateJwtTokens(&route))
			// If authorization is enabled, then this should be the first middleware before ratelimiting
			routeHandlers = append(routeHandlers, middlewares.AuthorizeRequest(&route))
		}

		// TODO: For now RateLimiters configurations could not be set dynamically, need to figure this out

		// Verify that the rate-limit is valid or not, or if there are any misconfiguration or not..
		// We only configure rate limit if it is enabled for that particular route
		if route.EnableRateLimit {
			logger.Debug().Msg("adding route rate limit")

			// Add this middleware to the route handlers
			routeHandlers = append(routeHandlers, middlewares.RateLimitMiddleware(&route, "route"))
		}

		if route.EnableUserRateLimit {
			logger.Debug().Msg("adding per user rate-limit")
			// If per user limit is enabled, then use perUserLimit middleware
			routeHandlers = append(routeHandlers, middlewares.RateLimitMiddleware(&route, "user"))
		}

		// Finally append the proxy request handler
		routeHandlers = append(routeHandlers, handlers.ProxyRequestHandler(proxy, route))

		// TODO: Here need to add more switch cases
		switch route.Method {
		case "GET":
			router.GET(route.Path, routeHandlers...)
		}
	}
}

// Sets the default rate limit values if not provided
func setDefaultRateLimitValues(route *models.ProxyRoute, logger *zerolog.Logger) {
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
	}

	if route.EnableUserRateLimit {
		logger.Debug().Msg("user rate limit is enabled for this route")
		if route.UserRateLimit == nil {
			logger.Debug().Float64("max-rps", globals.Global.MaxRequestsPerSecond).Msg("setting user-rate-limit to default value")
			route.UserRateLimit = &globals.Global.MaxRequestsPerSecond
		}
		if route.UserRateLimitWindow == nil {
			logger.Debug().Int("rate-limit-window", globals.Global.RateLimitWindow).Msg("setting rate-limit window to default value")
			route.UserRateLimitWindow = &globals.Global.RateLimitWindow
		}
	}
}
