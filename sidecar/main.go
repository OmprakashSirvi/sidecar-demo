package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func getLogger() zerolog.Logger {
	return zerolog.New(os.Stdout)
}

func initSidecar() {
	logger := getLogger()
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
	initLogging()
	initSidecar()

	logger := getLogger()

	// Get information regarding sidecar, this will give out the routes it supports,
	// and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	if globals.Global.ProxyBackend != "" {
		ginProxy, err := NewReverseProxy(globals.Global.ProxyBackend)
		if err != nil {
			panic("invalid proxy backend configuration")
		}

		setProxyRoutes(router, ginProxy, logger)
	}

	// router.NoRoute(func(ctx *gin.Context) {
	// 	// TODO: We can use the ctx to filter out any invalid or unregistered routes.
	// 	ginProxy.ServeHTTP(ctx.Writer, ctx.Request)
	// })
	router.Run()
}

func handleSidecarInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "basic sidecar information here")
}
