package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLogger() {
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func getLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func main() {
	initLogger()
	router := gin.Default()

	router.GET("/ping", handlePing)
	router.GET("/serviceInfo", handleInfo)

	router.Run()
}

func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, "everything is up and running")
}

func handleInfo(c *gin.Context) {
	requestId := c.GetHeader("X-Request-ID")
	logger := getLogger()
	if requestId != "" {
		logger = logger.With().Str("request_id", requestId).Logger()
	}

	logger.Info().Msg("handling serviceInfo route")
	c.JSON(http.StatusOK, "this is some information regarding me")
}
