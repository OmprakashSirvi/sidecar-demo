// This package will only contain models which are accessed globally
package models

type ProxyRoute struct {
	Type                 string       `mapstructure:"type"`
	Path                 string       `mapstructure:"path"`
	EnableRateLimit      bool         `mapstructure:"enable-rate-limit"`
	MaxRequestsPerSecond *float64     `mapstructure:"max-requests-per-second"`
	EnableUserRateLimit  bool         `mapstructure:"enable-user-rate-limit"`
	BurstThreshold       *int         `mapstructure:"burst-threshold"`
	UserRateLimit        *float64     `mapstructure:"user-rate-limit"`
	UserRateLimitWindow  *int     `mapstructure:"user-rate-limit-window"`
	RequireServiceTicket bool         `mapstructure:"require-service-ticket"`
	RouteTokens          []RouteToken `mapstructure:"tokens"`
}

type RouteToken struct {
	Type string `mapstructure:"user-jwt"`
}

func (route *ProxyRoute) GetMaxRequestsPerSecond() float64 {
	return *route.MaxRequestsPerSecond
}

func (route *ProxyRoute) GetBurstThreshold() int {
	return *route.BurstThreshold
}

// TODO: Search for invalid route configurations
// For now there seems to be no invalid configurations which results in fatal errors
func (route *ProxyRoute) IsValidRoute() (string, bool) {
	// logger := applogger.GetLogger().With().Str("method", route.Type).Str("path", route.Path).Logger()

	return "", true
}
