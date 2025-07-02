package main

import (
	"fmt"
	"net/http"
	"path/filepath"
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

	viper.SetDefault(constants.MY_ENV, "local")
}

// Get the key name with the current env
func getKeyNameForEnv(key string) string {
	env := viper.GetString(constants.MY_ENV)
	keyName := fmt.Sprintf("%v.%v", env, key)
	if viper.IsSet(keyName) {
		// This means there is an override available for this key in current env configuration
		return keyName
	}
	// No override is available for this key, use the existing one..
	return key
}

func main() {
	router := gin.Default()
	initConfig()

	// Get information regarding sidecar, this will give out the routes it supports,
	//  and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	backendUrl := viper.GetString(getKeyNameForEnv(constants.PROXY_BACKEND))
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
