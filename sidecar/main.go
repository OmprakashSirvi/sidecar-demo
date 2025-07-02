package main

import (
	"fmt"
	"net/http"
	"sidecar/config"
	"sidecar/constants"

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

func main() {
	router := gin.Default()
	config.InitConfig()

	// Get information regarding sidecar, this will give out the routes it supports,
	//  and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	backendUrl := viper.GetString(config.GetKeyNameForEnv(constants.PROXY_BACKEND))
	fmt.Printf("backend URL : %v\n", backendUrl)
	ginProxy, err := NewReverseProxy(backendUrl)
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
