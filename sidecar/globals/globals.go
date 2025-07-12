// Stores the global variables of the service
// Can implement thread safe mechanisms to access and modify these variables
package globals

import (
	"sidecar/models"

	"github.com/redis/go-redis/v9"
)

type globalVars struct {
	// Backend service URL
	// Need to setup reverse proxy for this and forward request to sidecar here..
	ProxyBackend         string
	ConfigDir            string
	AuthzConfigs         []models.AuthzConfig
	MaxConnectionLimit   int
	RequestTimeout       int
	MaxRequestsPerSecond float64
	BurstThreshold       int
	RedisDb              *redis.Client

	UserAuthorizer    *models.BasicAuthorizer
	ServiceAuthorizer *models.BasicAuthorizer
}

// Stores the global variables
var Global = globalVars{}
