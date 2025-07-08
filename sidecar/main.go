package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sidecar/applogger"
	"sidecar/config"
	"sidecar/constants"
	"sidecar/globals"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type Route struct {
	Type string
	Path string
}

func initSidecar() {
	logger := applogger.GetLogger()
	config.InitConfig()

	globals.Global.ProxyBackend = viper.GetString(config.GetKeyNameForEnv(constants.PROXY_BACKEND))
	logger.Debug().Str("message", fmt.Sprintf("backend URL : %v", globals.Global.ProxyBackend))

	configDir, err := filepath.Abs("/conf")
	if err != nil {
		errStr := fmt.Sprintf("invalid config directory, error: %v", err)
		panic(errStr)
	}

	modelPath := filepath.Join(configDir, "auth_models.conf")
	policyPath := filepath.Join(configDir, "auth_policy.csv")

	syncedEnforcer, err := casbin.NewSyncedEnforcer(modelPath, policyPath)
	if err != nil {
		errStr := fmt.Sprintf("unable to load enforcer: %v", err)
		panic(errStr)
	}
	syncedEnforcer.EnableLog(true)
	globals.Global.UserAuthorizer = &globals.BasicAuthorizer{Enforcer: syncedEnforcer}
}

func main() {
	router := gin.Default()
	applogger.InitLogging()
	initSidecar()

	logger := applogger.GetLogger()

	// For logging purposes
	router.Use(applogger.LoggingMiddleware())

	// Get information regarding sidecar, this will give out the routes it supports,
	// and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	if globals.Global.ProxyBackend != "" {
		logger.Debug().Str("proxy_backend", globals.Global.ProxyBackend).Msg("enabling reverse proxy for provided backend")
		ginProxy, err := NewReverseProxy(globals.Global.ProxyBackend)
		if err != nil {
			panic("invalid proxy backend configuration")
		}

		setProxyRoutes(router, ginProxy, logger)
	}

	router.Run()
}
 
func handleSidecarInfo(c *gin.Context) {
	logger := zerolog.Ctx(c.Request.Context())

	c.JSON(http.StatusOK, "basic sidecar information here")
	logger.Debug().Msg("sidecar info handled")
}
