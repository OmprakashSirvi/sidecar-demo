package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/ping", handlePing)

	router.Run()
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, "everything is up and running")
} 
