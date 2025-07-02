package main

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
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

// TODO: Specify the pathname and config file name dynamically
func initConfig() {
	viper.SetConfigName("proxy")
	viper.SetConfigType("yaml")
	// TODO: Should get the config directory dynamically while initializing the service
	// Keep this as default if config directory is not provided
	abs, _ := filepath.Abs("/conf")
	viper.AddConfigPath(abs)
	err := viper.ReadInConfig()
	if err != nil {
		panic("not able to read proxy.yaml")
	}
	viper.AutomaticEnv()
}

func main() {
	router := gin.Default()
	initConfig()

	// Get information regarding sidecar, this will give out the routes it supports,
	//  and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	ginProxy, err := NewReverseProxy(PROXY_BACKEND)
	if err != nil {
		panic("invalid proxy backend configuration")
	}

	setProxyRoutes(router, ginProxy)

	// router.NoRoute(func(ctx *gin.Context) {
	// 	// TODO: We can use the ctx to filter out any invalid or unregistered routes.
	// 	ginProxy.ServeHTTP(ctx.Writer, ctx.Request)
	// })
	router.Run()
}

func handleSidecarInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "basic sidecar information here")
}
