// This package will only contain models which are accessed globally
package models

type ProxyRoute struct {
	Type                 string       `mapstructure:"type"`
	Path                 string       `mapstructure:"path"`
	EnableRateLimit      bool         `mapstructure:"enable-rate-limit"`
	MaxRequestsPerSecond *float64     `mapstructure:"max-requests-per-second"`
	BurstThreshold       *int         `mapstructure:"burst-threshold"`
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
