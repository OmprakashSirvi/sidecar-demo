package main

import (
	"fmt"
	"net/http"
	"os"
	"sidecar/config"
	"sidecar/constants"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Backend service URL
// Need to setup reverse proxy for this and forward request to sidecar here..
// TODO: Make this configurable
var PROXY_BACKEND = "http://backend:8080"

type Route struct {
	Type string
	Path string
}

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func getLogger() zerolog.Logger{
	return zerolog.New(os.Stdout)
}

func main() {
	router := gin.Default()
	config.InitConfig()
	initLogging()

	logger := getLogger()

	// Get information regarding sidecar, this will give out the routes it supports,
	//  and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	backendUrl := viper.GetString(config.GetKeyNameForEnv(constants.PROXY_BACKEND))
	logger.Debug().Str("message",fmt.Sprintf( "backend URL : %v", backendUrl))
	ginProxy, err := NewReverseProxy(backendUrl)
	if err != nil {
		panic("invalid proxy backend configuration")
	}

	setProxyRoutes(router, ginProxy, logger)

	// router.NoRoute(func(ctx *gin.Context) {
	// 	// TODO: We can use the ctx to filter out any invalid or unregistered routes.
	// 	ginProxy.ServeHTTP(ctx.Writer, ctx.Request)
	// })
	router.Run()
}

func handleSidecarInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "basic sidecar information here")
}
