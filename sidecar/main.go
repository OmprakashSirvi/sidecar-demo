package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"sidecar/applogger"
	"sidecar/config"
	"sidecar/constants"
	"sidecar/globals"
	"sidecar/middlewares"
	"sidecar/models"
	"sidecar/routes"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var defineFlags sync.Once

func initSidecar() {
	logger := applogger.GetLogger()
	config.InitConfig()

	globals.Global.ProxyBackend = viper.GetString(config.GetKeyName(constants.PROXY_BACKEND))
	logger.Debug().Str("message", fmt.Sprintf("backend URL : %v", globals.Global.ProxyBackend))

	configDir, err := filepath.Abs(globals.Global.ConfigDir)
	if err != nil {
		logger.Fatal().Err(err).Msg("invalid config directory")
	}

	// Set default values for your configuration keys.
	viper.SetDefault(config.GetKeyName(constants.MAX_CONNECTION_LIMIT), 50)
	viper.SetDefault(config.GetKeyName(constants.REQUEST_TIMEOUT), 10)
	viper.SetDefault(config.GetKeyName(constants.MAX_REQUESTS_PER_SECOND), constants.DefaultMaxRequestPerSecond)
	viper.SetDefault(config.GetKeyName(constants.BURST_THRESHOLD), constants.BURST_THRESHOLD)

	// Set the config values into the application's global configuration
	globals.Global.MaxConnectionLimit = viper.GetInt(config.GetKeyName(constants.MAX_CONNECTION_LIMIT))
	globals.Global.RequestTimeout = viper.GetInt(config.GetKeyName(constants.REQUEST_TIMEOUT))
	globals.Global.MaxRequestsPerSecond = viper.GetFloat64(config.GetKeyName(constants.MAX_REQUESTS_PER_SECOND))
	globals.Global.BurstThreshold = viper.GetInt(config.GetKeyName(constants.BURST_THRESHOLD))

	// Load casbin enforcers from authz config
	logger.Debug().Int("length", len(globals.Global.AuthzConfigs)).Msg("number of authz-configs provided")
	for _, authzConfig := range globals.Global.AuthzConfigs {
		switch authzConfig.AuthzType {
		// Load user enforcer
		case constants.USER_ID:
			logger.Debug().Msg("loading user-id authz-config")
			modelPath := filepath.Join(configDir, authzConfig.ModelFile)
			policyPath := filepath.Join(configDir, authzConfig.PolicyFile)
			globals.Global.UserAuthorizer = loadEnforcer(modelPath, policyPath)

		// Load service enforcer
		case constants.SERVICE_ID:
			logger.Debug().Msg("loading service-id authz-config")
			modelPath := filepath.Join(configDir, authzConfig.ModelFile)
			policyPath := filepath.Join(configDir, authzConfig.PolicyFile)
			globals.Global.ServiceAuthorizer = loadEnforcer(modelPath, policyPath)

		// Handle invalid configuration
		default:
			logger.Fatal().Str("type", authzConfig.AuthzType).Msg("unsupported authz-config type provided")
		}
	}

	// Redis connection setup
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		logger.Fatal().Err(err).Msg("could not connect to redis")
	}
	logger.Info().Msg("Successfully connected to redis")
	globals.Global.RedisDb = rdb
}

// loadEnforcer creates a new Casbin SyncedEnforcer from the given model and policy files.
func loadEnforcer(modelPath string, policyPath string) *models.BasicAuthorizer {
	logger := applogger.GetLogger()

	syncedEnforcer, err := casbin.NewSyncedEnforcer(modelPath, policyPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("unable to load enforcer with model")
	}

	syncedEnforcer.EnableLog(true)
	logger.Info().Str("model", modelPath).Msg("Successfully loaded Casbin enforcer")

	return &models.BasicAuthorizer{Enforcer: syncedEnforcer}
}

func main() {
	applogger.InitLogging()
	logger := applogger.GetLogger()
	defineFlags.Do(func() {
		logger.Debug().Msg("loading command line flags")
		flag.StringVar(&globals.Global.ConfigDir, "config-dir", "/conf", "location of config dir, defaults to /conf")

		flag.Parse()
	})

	initSidecar()

	// sidecar router, handles sidecar routes
	sidecarRouter := gin.Default()
	// Request timeouts
	sidecarRouter.Use(middlewares.TimeoutMiddleware())

	// Get information regarding sidecar, this will give out the routes it supports,
	// and some other information, This will be modified in the future.
	sidecarRouter.GET("/info", handleSidecarInfo)
	sidecarRouter.GET("/ticket", handleGetServiceTicket)

	go func() {
		err := sidecarRouter.Run("localhost:8070")
		if err != nil {
			logger.Error().Err(err).Msg("sidecarRouter error")
		}
	}()

	// Reverse proxy router, handles backend routes
	router := gin.Default()
	// For logging purposes
	router.Use(middlewares.LoggingMiddleware())

	// Rate limiter
	router.Use(middlewares.GlobalRateLimiter())

	if globals.Global.ProxyBackend != "" {
		logger.Debug().Str("proxy_backend", globals.Global.ProxyBackend).Msg("enabling reverse proxy for provided backend")
		ginProxy, err := NewReverseProxy(globals.Global.ProxyBackend)
		if err != nil {
			logger.Fatal().Err(err).Msg("invalid proxy backend configuration")
		}

		routes.SetProxyRoutes(router, ginProxy, &logger)
	}

	// Will listen to default port: 8000
	router.Run()
}

// Just trying to run the builder
// TODO: Implement this handler to return information regarding sidecar configurations
func handleSidecarInfo(c *gin.Context) {
	logger := zerolog.Ctx(c.Request.Context())

	logger.Debug().Msg("handling sidecar info")
	c.JSON(http.StatusOK, "basic sidecar information here")
}

// TODO: Implement this handler
func handleGetServiceTicket(c *gin.Context) {

}
