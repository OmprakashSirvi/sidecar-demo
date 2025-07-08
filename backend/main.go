package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/ping", handlePing)
	router.GET("/serviceInfo", handleInfo)

	router.Run()
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, "everything is up and running")
}

func handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "this is some information regarding me")
}
