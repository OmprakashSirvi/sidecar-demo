package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Backend service URL
// Need to setup reverse proxy for this and forward request to sidecar here..
// TODO: Make this configurable
var PROXY_BACKEND = "http://backend:8080"


func main() {
	router := gin.Default()

	// Get information regarding sidecar, this will give out the routes it supports,
	//  and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	ginProxy, err := NewReverseProxy(PROXY_BACKEND)
	if err != nil {
		panic("invalid proxy backend configuration")
	}

	router.NoRoute(func(ctx *gin.Context) {
		// TODO: We can use the ctx to filter out any invalid or unregistered routes.
		ginProxy.ServeHTTP(ctx.Writer, ctx.Request)
	})
	router.Run()
}

func handleSidecarInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "basic sidecar information here")
}
