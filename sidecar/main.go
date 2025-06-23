package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Get information regarding sidecar, this will give out the routes it supports,
	//  and some other information, This will be modified in the future.
	router.GET("/info", handleSidecarInfo)

	router.Run()
}

func handleSidecarInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "basic sidecar information here")
}
